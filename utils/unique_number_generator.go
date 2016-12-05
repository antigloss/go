// Author: https://github.com/antigloss

/*
Package utils packs with many useful handy code.
*/
package utils

import (
	"sync/atomic"
)

// MonoIncSeqNumGenerator32 is a goroutine-safe Monotonically Increasing Sequence Number Generator which generates 32bit unsigned ints.
type MonoIncSeqNumGenerator32 uint32

// NewMonoIncSeqNumGenerator32 is an easy way to get a new, ready-to-use MonoIncSeqNumGenerator32 object with an initial value.
//
//   initVal: Initial value for this MonoIncSeqNumGenerator32
func NewMonoIncSeqNumGenerator32(initVal uint32) *MonoIncSeqNumGenerator32 {
	m32 := MonoIncSeqNumGenerator32(initVal)
	return &m32
}

// GetSeqNum returns a sequence number that is bigger than the previously sequence number by 1.
func (m32 *MonoIncSeqNumGenerator32) GetSeqNum() uint32 {
	seq := atomic.AddUint32((*uint32)(m32), 1)
	for ; seq == 0; {
		seq = atomic.AddUint32((*uint32)(m32), 1)
	}
	return seq
}

// MonoIncSeqNumGenerator64 is a goroutine-safe Monotonically Increasing Sequence Number Generator which generates 64bit unsigned ints.
type MonoIncSeqNumGenerator64 uint64

// NewMonoIncSeqNumGenerator64 is an easy way to get a new, ready-to-use MonoIncSeqNumGenerator64 object with an initial value.
//
//   initVal: Initial value for this MonoIncSeqNumGenerator64
func NewMonoIncSeqNumGenerator64(initVal uint32) *MonoIncSeqNumGenerator64 {
	m64 := MonoIncSeqNumGenerator64(initVal)
	return &m64
}

// GetSeqNum returns a sequence number that is bigger than the previously sequence number by 1.
func (m64 *MonoIncSeqNumGenerator64) GetSeqNum() uint64 {
	seq := atomic.AddUint64((*uint64)(m64), 1)
	for ; seq == 0; {
		seq = atomic.AddUint64((*uint64)(m64), 1)
	}
	return seq
}