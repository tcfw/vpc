package sbs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockCacheGetSetGet(t *testing.T) {
	data := []byte{1, 2, 3, 4}
	bc := newBlockCache()

	d := bc.get(1234567890)
	assert.Nil(t, d)

	bc.set(1234567890, data)

	d = bc.get(1234567890)
	assert.Equal(t, data, d.data)
}
