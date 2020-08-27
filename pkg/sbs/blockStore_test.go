package sbs

import (
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"

	volumesAPI "github.com/tcfw/vpc/pkg/api/v1/volumes"
	"github.com/tcfw/vpc/pkg/utils"
)

func TestBlockStoreCmdEncoding(t *testing.T) {
	bsc := &StoreCommand{
		Op:     OpWrite,
		Offset: 8611823616,
		Length: 10,
		Data:   []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
	}

	p, _ := bsc.Encode()

	bsc2 := &StoreCommand{}
	if err := bsc2.Decode(p); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, bsc.Op, bsc2.Op)
	assert.Equal(t, bsc.Offset, bsc2.Offset)
	assert.Equal(t, bsc.Length, bsc2.Length)
	assert.Equal(t, bsc.Data, bsc2.Data)
}

func TestFileOffset(t *testing.T) {
	bs := &BlockStore{}

	offset := bs.blockOffset(250 * 1 << 20)
	assert.Equal(t, uint64(2), offset)

	offset = bs.blockOffset(150 * 1 << 20)
	assert.Equal(t, uint64(1), offset)

	offset = bs.blockOffset(50 * 1 << 20)
	assert.Equal(t, uint64(0), offset)
}

func TestWriteApplyRead(t *testing.T) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 10, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	//Data to write
	rData := []byte{86, 85, 35, 34, 35, 36}
	data := make([]byte, BlockSize)
	copy(data, rData)
	var offset uint64 = 2097152000 //100MB - 3 byte offset

	bcmd := &StoreCommand{
		Op:     OpWrite,
		Offset: offset,
		Length: BlockSize,
		Data:   data,
	}

	bcmdBytes, _ := bcmd.Encode()

	//Apply as if from raft
	resp := bs.Apply(&raft.Log{Type: raft.LogCommand, Data: bcmdBytes})
	if err, ok := resp.(error); ok {
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	//Read back
	buf := make([]byte, len(data))
	n, err := bs.ReadAt(buf, offset)
	if assert.NoError(t, err) {
		assert.Equal(t, len(data), n)
		assert.Equal(t, data[:n], buf[:n])
	}
}

func TestSeqReadWrite(t *testing.T) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 2, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	dataSize := 16 * 1024
	data := []byte{}
	for i := 0; i < dataSize; i++ {
		data = append(data, byte(i%256))
	}

	assert.Equal(t, dataSize, len(data))

	errN := 0
	errs := []error{}

	t.Log("starting write")

	for i := uint64(0); i < bs.size; i += uint64(len(data)) {
		bLen := uint32(len(data))
		if i+uint64(bLen) > bs.size {
			bLen = uint32(bs.size - i)
		}
		if bLen < uint32(len(data)) {
			t.Logf("truncated last block to: %d", bLen)
		}
		if err := bs.write(&StoreCommand{
			Op:     OpWrite,
			Offset: i,
			Length: bLen,
			Data:   data[:bLen],
		}); err != nil {
			errN++
			errs = append(errs, err)
		}
	}

	if !assert.Zero(t, errN) {
		t.Fatal(errs)
	}

	t.Log("starting read ")

	buf := make([]byte, len(data))
	for i := uint64(0); i < bs.size; i += uint64(len(data)) {
		bLen := uint32(len(data))
		if i+uint64(bLen) > bs.size {
			bLen = uint32(bs.size - i)
		}
		n, err := bs.ReadAt(buf, i)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(data), n)
		assert.Equal(t, data[:bLen], buf[:bLen], "at offset %d", i)
	}
}

func TestRead(t *testing.T) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 10, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	buf := make([]byte, 4096)
	n, err := bs.ReadAt(buf, 4430233600)
	assert.NoError(t, err)
	assert.Equal(t, 4096, n)
}

func TestWriteZeros(t *testing.T) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 1, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	//Data to write
	data := []byte{86, 85, 35, 34, 35, 36}
	offset := uint64(250 * 1 << 20) //250MB offset

	bcmdWrite := &StoreCommand{
		Op:     OpWrite,
		Offset: offset,
		Length: BlockSize,
		Data:   data,
	}
	bcmdWriteBytes, _ := bcmdWrite.Encode()

	bcmdZero := &StoreCommand{
		Op:     OpZeros,
		Offset: offset,
		Length: BlockSize,
	}
	bcmdZeroBytes, _ := bcmdZero.Encode()

	//Apply as if from raft
	//Write initial data
	resp := bs.Apply(&raft.Log{Type: raft.LogCommand, Data: bcmdWriteBytes})
	if err, ok := resp.(error); ok {
		assert.Error(t, err)
	}

	//Zero out
	resp = bs.Apply(&raft.Log{Type: raft.LogCommand, Data: bcmdZeroBytes})
	if err, ok := resp.(error); ok {
		assert.NoError(t, err)
	}

	//Read back
	buf := make([]byte, bcmdZero.Length)
	n, err := bs.ReadAt(buf, offset)
	if assert.NoError(t, err) {
		assert.Equal(t, int(bcmdZero.Length), n)
		assert.Equal(t, make([]byte, bcmdZero.Length), buf)
	}
}

func TestOutOfBounds(t *testing.T) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 1, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	//Data to write
	data := []byte{86, 85, 35, 34, 35, 36}
	offset := uint64(1 << 30) //1GB offset

	bcmd := &StoreCommand{
		Op:     OpWrite,
		Offset: offset,
		Length: BlockSize,
		Data:   data,
	}
	bcmdBytes, _ := bcmd.Encode()

	//Apply as if from raft
	resp := bs.Apply(&raft.Log{Type: raft.LogCommand, Data: bcmdBytes})
	if err, ok := resp.(error); ok {
		assert.Error(t, err)
	}
}

func TestBlockStoreStats(t *testing.T) {
	srv := NewServerWithPort(rand.Intn(1750) + 34000)
	defer cleanUpServerData([]*Server{srv})

	stats, err := srv.BlockStoreStats()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotZero(t, stats.Available)
	assert.Zero(t, stats.Used)
	assert.Zero(t, stats.Allocated)

	vol := &volumesAPI.Volume{Id: "vol-test", Size: 1}

	if err := srv.AddVolume(vol, []string{}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	stats, err = srv.BlockStoreStats()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(1*1<<30), stats.Allocated)

	vStats, err := srv.BlockVolumeStats(srv.Volume(vol.Id))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, uint64(1*1<<30), vStats.Allocated)
	assert.Zero(t, vStats.Used)
	assert.NotZero(t, vStats.Available)
}

func BenchmarkWriteApply(b *testing.B) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 1, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	//Data to write
	data := []byte{86, 85, 35, 34, 35, 36}
	offset := uint64(250 * 1 << 20) //250MB offset

	bcmd := &StoreCommand{
		Op:     OpWrite,
		Offset: offset,
		Length: BlockSize,
		Data:   data,
	}
	bcmdBytes, _ := bcmd.Encode()

	b.Run("direct-write", func(b *testing.B) {
		b.SetBytes(int64(bcmd.Length))
		for i := 0; i < b.N; i++ {
			bs.write(bcmd)
		}
	})

	b.Run("apply-write", func(b *testing.B) {
		b.SetBytes(int64(bcmd.Length))
		for i := 0; i < b.N; i++ {
			bs.Apply(&raft.Log{Type: raft.LogCommand, Data: bcmdBytes})
		}
	})

	b.Run("parrallel-apply", func(b *testing.B) {
		b.SetBytes(int64(bcmd.Length))
		b.RunParallel(func(b *testing.PB) {
			for b.Next() {
				bs.Apply(&raft.Log{Type: raft.LogCommand, Data: bcmdBytes})
			}
		})
	})
}

func BenchmarkRead(b *testing.B) {
	vol := &volumesAPI.Volume{
		Id:   "vol-test",
		Size: 1, //GB
	}

	dir := os.TempDir()
	defer os.RemoveAll(dir + "/" + vol.Id)

	bs := NewBlockStore("", vol, dir, utils.DefaultLogger())

	//Data to write
	data := []byte{86, 85, 35, 34, 35, 36}
	offset := uint64(250 * 1 << 20) //250MB offset

	bcmd := &StoreCommand{
		Op:     OpWrite,
		Offset: offset,
		Length: BlockSize,
		Data:   data,
	}
	bs.write(bcmd)

	b.Run("read", func(b *testing.B) {
		b.SetBytes(BlockSize)
		p := make([]byte, BlockSize)
		for i := 0; i < b.N; i++ {
			n, err := bs.ReadAt(p, bcmd.Offset)
			if n == 0 {
				b.Fatal("nothing read")
			}
			if err != nil {
				b.Fatal(err)
			}
		}

	})
}
