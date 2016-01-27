package queue

import (
	"sync/atomic"
	"unsafe"
)

func NewLockfreeQueue() *LockfreeQueue {
	var lfq LockfreeQueue
	lfq.head = unsafe.Pointer(&lfq.dummy)
	lfq.tail = unsafe.Pointer(&lfq.dummy)
	return &lfq
}

type LockfreeQueue struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	dummy lfqNode
}

func (lfq *LockfreeQueue) Pop() interface{} {
	for {
		h := atomic.LoadPointer(&lfq.head)
		rh := (*lfqNode)(h)
		n := (*lfqNode)(atomic.LoadPointer(&rh.next))
		if n != nil {
			if atomic.CompareAndSwapPointer(&lfq.head, h, rh.next) {
				return n.val
			} else {
				continue
			}
		} else {
			return nil
		}
	}
}

func (lfq *LockfreeQueue) Push(val interface{}) {
	node := unsafe.Pointer(&lfqNode{val: val})
	for {
		t := atomic.LoadPointer(&lfq.tail)
		rt := (*lfqNode)(t)
		if atomic.CompareAndSwapPointer(&rt.next, nil, node) {
			atomic.StorePointer(&lfq.tail, node)
			break
		} else {
			continue
		}
	}
}

type lfqNode struct {
	val  interface{}
	next unsafe.Pointer
}
