package sbs

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/kpango/fastime"
	"github.com/sirupsen/logrus"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
)

//StoreCommandOp block command operation
type StoreCommandOp uint16

const (
	//OpWrite write block
	OpWrite StoreCommandOp = iota
	//OpZeros replace block with zeros
	OpZeros
	//OpTrim _may_ be discarded
	OpTrim
	//OpFlush asked to persist all data
	OpFlush

	//BlockSize block size per raft log
	BlockSize = 16 * 1 << 10 //16KB
	//BlocksPerFile maximum number of blocks that should be in a file
	BlocksPerFile        = 100 * 1 << 20 / BlockSize //100MB files
	blockMaxSize  uint64 = BlockSize * BlocksPerFile

	//VolDescSizeMultiplier is the conversion factor for the volume description to bytes
	VolDescSizeMultiplier = 1 << 30 //volume definition size is in GB
)

//OutOfBounds when the given offset+data len is greater than allocated volume size
type OutOfBounds uint64

func (oob OutOfBounds) Error() string {
	return fmt.Sprintf("offset %d out of bounds", oob)
}

//StoreCommand block store mutation command
type StoreCommand struct {
	Op     StoreCommandOp `json:"op"`
	Offset uint64         `json:"offset"`
	Length uint32         `json:"length"`
	Data   []byte         `json:"data,omitempty"`
}

//Encode encodes the block command to binary formats transmission
func (bsc *StoreCommand) Encode() ([]byte, error) {
	p := bytes.NewBuffer(nil)

	binary.Write(p, binary.LittleEndian, bsc.Op)
	binary.Write(p, binary.LittleEndian, bsc.Offset)
	binary.Write(p, binary.LittleEndian, bsc.Length)
	length := len(bsc.Data)
	for length > 0 {
		n, err := p.Write(bsc.Data)
		if err != nil {
			return nil, err
		}
		length -= n
	}

	return p.Bytes(), nil
}

//Decode decodes the block command from transmission
func (bsc *StoreCommand) Decode(p []byte) error {
	bsc.Op = StoreCommandOp(binary.LittleEndian.Uint16(p[0:2]))
	bsc.Offset = binary.LittleEndian.Uint64(p[2:10])
	bsc.Length = binary.LittleEndian.Uint32(p[10:14])
	bsc.Data = p[14:]

	return nil
}

//BlockStore block store backing e.g. local disks
type BlockStore struct {
	log     *logrus.Logger
	backing map[uint64]*block
	BaseDir string
	def     *volumesAPI.Volume
	mu      sync.RWMutex
	size    uint64
}

type block struct {
	f           *os.File
	mu          sync.Mutex
	lastTouched time.Time
}

func (b *block) Close() error {
	return b.f.Close()
}

//NewBlockStore provides a new store for volumes
func NewBlockStore(peerID string, d *volumesAPI.Volume, baseDir string, l *logrus.Logger) *BlockStore {
	dir := fmt.Sprintf("%s/%s/%s", baseDir, peerID, d.GetId())

	os.MkdirAll(fmt.Sprintf("%s/blocks", dir), os.FileMode(0740)|os.ModeDir)

	return &BlockStore{
		def: d, log: l,
		backing: map[uint64]*block{},
		BaseDir: dir,
		size:    uint64(d.Size * VolDescSizeMultiplier),
	}
}

//SizeOnDisk goes through each block and queries the underlying filesystem for it's current size
func (b *BlockStore) SizeOnDisk() (uint64, error) {
	var fSize uint64

	err := filepath.Walk(fmt.Sprintf("%s/blocks", b.BaseDir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fSize += uint64(info.Size())
		}
		return err
	})
	return fSize, err
}

//Apply log is invoked once a log entry is committed and commits
//the change to the block store
func (b *BlockStore) Apply(l *raft.Log) interface{} {
	switch l.Type {
	case raft.LogCommand:
		cmd := &StoreCommand{}
		if err := cmd.Decode(l.Data); err != nil {
			return err
		}
		return b.apply(cmd)
	case raft.LogNoop, raft.LogAddPeerDeprecated:
		return nil
	default:
		return fmt.Errorf("unsupported log type")
	}
}

//Snapshot is used to support log compaction
func (b *BlockStore) Snapshot() (raft.FSMSnapshot, error) {
	return &blockStoreSnapshotter{store: b}, nil
}

//Restore is used to restore the block store data from a snapshot
func (b *BlockStore) Restore(r io.ReadCloser) error {
	b.log.Info("Starting raft snapshot restore")
	buf := make([]byte, BlockSize)
	var seek uint64

	//unwrap snapshot via zlib
	zReader, err := zlib.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to open snapshot: %s", err)
	}

	defer zReader.Close()
	defer r.Close()

	for {
		n, err := zReader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := b.write(&StoreCommand{
			Op:     OpWrite,
			Offset: seek,
			Length: uint32(n),
			Data:   buf[:n],
		}); err != nil {
			return err
		}
		seek += uint64(n)
	}

	return nil
}

func (b *BlockStore) clearAllLocalStorage() error {
	b.Close()

	b.mu.Lock()
	defer b.mu.Unlock()

	return os.RemoveAll(b.BaseDir)
}

//apply switches out the various commands
func (b *BlockStore) apply(c *StoreCommand) error {
	switch c.Op {
	case OpZeros, OpWrite:
		return b.write(c)
	case OpFlush:
		return b.flush()
	default:
		return fmt.Errorf("unknown op")
	}
}

//write applies the write operations to the offsets in the block set
func (b *BlockStore) write(c *StoreCommand) error {
	//vol bounds check
	if c.Offset+uint64(c.Length) > b.size {
		return OutOfBounds(c.Offset + uint64(c.Length))
	}

	//Allow for lengths larger than the given bytes
	if c.Op == OpZeros {
		d := make([]byte, c.Length)
		copy(d, c.Data)
		c.Data = d
	}

	//support writes that span multiple blocks
	length := uint64(c.Length)
	offset := c.Offset
	var bStart uint64
	var nextBlockStart uint64

	for length > 0 {
		bn := b.blockOffset(offset)
		bl, err := b.getBlock(bn)
		if err != nil {
			return err
		}

		bLen := length
		nextBlockStart = (bn + 1) * blockMaxSize
		if bLen+offset > nextBlockStart {
			bLen = nextBlockStart - offset
		}

		fOffset := offset % blockMaxSize

		bl.mu.Lock()
		n, err := bl.f.WriteAt(c.Data[bStart:bStart+bLen], int64(fOffset))
		if err != nil {
			bl.mu.Unlock()
			return err
		}
		bl.lastTouched = fastime.Now()
		bl.mu.Unlock()

		n64 := uint64(n)

		length -= n64
		offset += n64
		bStart += n64
	}

	return nil
}

//ReadAt allows for the block store to be read from block sets
func (b *BlockStore) ReadAt(p []byte, offset uint64) (int, error) {
	//bounds check
	if offset > b.size {
		return 0, OutOfBounds(offset)
	}

	if offset < 0 {
		return 0, fmt.Errorf("invalid offset")
	}

	length := uint64(len(p))
	var nn uint64
	var nextBlockStart uint64

	for length > 0 {
		bn := b.blockOffset(offset)
		bl, err := b.getBlock(bn)
		if err != nil {
			return 0, fmt.Errorf("failed to get block: %s", err)
		}

		bLen := length
		nextBlockStart = (bn + 1) * blockMaxSize
		if bLen+offset > nextBlockStart {
			bLen = nextBlockStart - offset
		}
		fOffset := offset % blockMaxSize

		bl.mu.Lock()
		n, err := bl.f.ReadAt(p[nn:nn+bLen], int64(fOffset))
		if err != nil {
			bl.mu.Unlock()
			return 0, fmt.Errorf("failed to read at local offset %d: %s", fOffset, err)
		}
		bl.lastTouched = fastime.Now()
		bl.mu.Unlock()

		n64 := uint64(n)

		nn += n64
		offset += n64
		length -= n64
	}

	return int(nn), nil
}

//flush forces os to sync blocks to persistent storage
func (b *BlockStore) flush() error {
	for _, bl := range b.backing {
		bl.mu.Lock()
		if err := bl.f.Sync(); err != nil {
			bl.mu.Unlock()
			return err
		}
		bl.mu.Unlock()
	}

	return nil
}

//getBlock storage of individual block sets
func (b *BlockStore) getBlock(n uint64) (*block, error) {
	b.mu.RLock()
	if block, ok := b.backing[n]; ok {
		b.mu.RUnlock()
		return block, nil
	}
	b.mu.RUnlock()

	fLoc := fmt.Sprintf("%s/blocks/%d.raw", b.BaseDir, n)

	f, err := os.OpenFile(fLoc, os.O_CREATE|os.O_RDWR, os.FileMode(0640))
	if err != nil {
		return nil, err
	}
	if err := f.Truncate(int64(blockMaxSize)); err != nil {
		return nil, err
	}

	bl := &block{f: f}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.backing[n] = bl

	return bl, nil
}

//Close closes all open blocks
func (b *BlockStore) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, block := range b.backing {
		block.Close()
		delete(b.backing, i)
	}
}

//blockOffset which file number the offset should be in
func (b *BlockStore) blockOffset(offset uint64) uint64 {
	return offset / blockMaxSize
}

//blockStoreSnapshotter helps create snapshots of the block store
type blockStoreSnapshotter struct {
	store *BlockStore
	seek  uint64
}

// Persist should dump all necessary state to the WriteCloser 'sink',
// and call sink.Close() when finished or call sink.Cancel() on error.
func (bss *blockStoreSnapshotter) Persist(sink raft.SnapshotSink) error {
	buf := make([]byte, BlockSize)

	//Wrap snapshots in zlip compression
	zWrite := zlib.NewWriter(sink)

	for bss.seek < bss.store.size {
		n, err := bss.store.ReadAt(buf, bss.seek)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		nWrite, err := zWrite.Write(buf[:n])
		if err != nil {
			return err
		}
		bss.seek += uint64(nWrite)
	}

	zWrite.Close()
	sink.Close()

	return nil
}

// Release is invoked when we are finished with the snapshot.
func (bss *blockStoreSnapshotter) Release() {
	bss.store.Close()
	return
}
