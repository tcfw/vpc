package sbs

import (
	"sync"
)

const (
	maxEntries = 12800
)

type blockCache struct {
	blocks   []*blockCacheEntry
	mu       sync.RWMutex
	bufAlloc *bufAlloc
}

type blockCacheEntry struct {
	offset uint64
	data   []byte
}

func newBlockCache() *blockCache {
	bc := &blockCache{
		blocks:   make([]*blockCacheEntry, maxEntries),
		bufAlloc: newBufAlloc(BlockSize * 2),
	}
	return bc
}

func (bc *blockCache) get(offset uint64) *blockCacheEntry {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	k := offset % uint64(len(bc.blocks))
	entry := bc.blocks[k]
	if entry == nil || entry.offset != offset {
		return nil
	}

	return entry
}

func (bc *blockCache) set(offset uint64, data []byte) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	buf := make([]byte, len(data))
	copy(buf, data)

	k := offset % uint64(len(bc.blocks))
	bc.blocks[k] = &blockCacheEntry{
		offset: offset,
		data:   buf,
	}
}

func (bc *blockCache) invalidate(offset uint64) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	k := offset % uint64(len(bc.blocks))
	bc.blocks[k] = nil
}
