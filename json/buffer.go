package json

import (
	"bytes"
	"sync"
)

type buffer struct {
	bytes.Buffer
}

var bufPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		buf := new(buffer)
		// buf.Grow(5 * 1024 * 1024)
		return buf
	},
}

func GetBuffer() *buffer {
	b := bufPool.Get().(*buffer)
	b.Reset()
	return b
}

func (b *buffer) Return() {
	bufPool.Put(b)
}
