// Author: https://github.com/antigloss

// Package pool offers goroutine-safe object pools such as BufferPool
//
// Sorry for my poor English, I've tried my best.
package pool

import (
	"bytes"
	"sync"
)

// NewBufferPool is the only way to get a new, ready-to-use BufferPool from which bytes.Buffer could be get.
//
// If you use `var bp pool.BufferPool`, or `new(pool.BufferPool)`, or the like to obtain a BufferPool, it'll
// still work, but it won't pool even a single bytes.Buffer.
//
//   maxBufferNum: Maximum number of bytes.Buffer that will be pooled in BufferPool
//   initBufferSize: Initial size in bytes for a newly created bytes.Buffer
//
// Example:
//
//   bp := NewBufferPool(1000, 512)
//   buf := bp.Get() // get a ready-to-use bytes.Buffer
//   // do something with `buf`
//   bp.Put(buf) // return buf to BufferPool
func NewBufferPool(maxBufferNum, initBufferSize int) *BufferPool {
	return &BufferPool{maxBufNum: maxBufferNum, initBufSz: initBufferSize}
}

// BufferPool is a goroutine-safe pool for bytes.Buffer.
type BufferPool struct {
	lock       sync.Mutex
	freeList   *buffer
	freeBufNum int
	maxBufNum  int
	initBufSz  int
}

// Get returns a ready-to-use bytes.Buffer.
func (bp *BufferPool) Get() *bytes.Buffer {
	bp.lock.Lock()
	b := bp.freeList
	if b != nil {
		bp.freeList = b.next
		bp.freeBufNum--
	}
	bp.lock.Unlock()

	var buf *bytes.Buffer
	if b != nil {
		buf = b.buf
		buf.Reset()
		b.buf = nil
		b.next = nil
	} else {
		buf = new(bytes.Buffer)
		buf.Grow(bp.initBufSz)
	}
	return buf
}

// Put returns a bytes.Buffer to the BufferPool.
func (bp *BufferPool) Put(buf *bytes.Buffer) {
	bp.lock.Lock()
	if bp.freeBufNum < bp.maxBufNum {
		bp.freeList = &buffer{buf, bp.freeList}
		bp.freeBufNum++
	}
	bp.lock.Unlock()
}

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	buf  *bytes.Buffer
	next *buffer
}
