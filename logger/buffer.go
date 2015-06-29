// Author: https://github.com/antigloss

package logger

import (
	"bytes"
	"sync"
)

// most of the following code is extracted from https://github.com/golang/glog

const digits = "0123456789"

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	bytes.Buffer
	tmp  [32]byte // temporary byte array for creating headers.
	next *buffer
}

// twoDigits formats a zero-prefixed two-digit integer at buf.tmp[i].
func (buf *buffer) twoDigits(i, d int) {
	buf.tmp[i+1] = digits[d%10]
	d /= 10
	buf.tmp[i] = digits[d%10]
}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func (buf *buffer) nDigits(n, i, d int, pad byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		buf.tmp[i+j] = digits[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		buf.tmp[i+j] = pad
	}
}

// someDigits formats a zero-prefixed variable-width integer at buf.tmp[i].
func (buf *buffer) someDigits(i, d int) int {
	// Print into the top, then copy down. We know there's space for at least
	// a 10-digit number.
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy(buf.tmp[i:], buf.tmp[j:])
}

// bufferPool
type bufferPool struct {
	lock       sync.Mutex
	freeList   *buffer
	freeBufNum int
}

// getBuffer returns a new, ready-to-use buffer.
func (bp *bufferPool) getBuffer() *buffer {
	bp.lock.Lock()
	b := bp.freeList
	if b != nil {
		bp.freeList = b.next
		bp.freeBufNum--
	}
	bp.lock.Unlock()

	if b == nil {
		b = new(buffer)
		b.Grow(512)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (bp *bufferPool) putBuffer(b *buffer) {
	bp.lock.Lock()
	if bp.freeBufNum < 1000 {
		b.next = bp.freeList
		bp.freeList = b
		bp.freeBufNum++
	}
	bp.lock.Unlock()
}
