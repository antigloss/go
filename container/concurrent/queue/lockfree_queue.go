/*
 *
 * queue - Goroutine-safe Queue implementations
 * Copyright (C) 2016 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package queue offers goroutine-safe Queue implementations such as LockfreeQueue(Lock free queue).
package queue

import (
	"sync/atomic"
	"unsafe"
)

// LockfreeQueue is a goroutine-safe Queue implementation.
// The overall performance of LockfreeQueue is much better than List+Mutex(standard package).
type LockfreeQueue struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	dummy lfqNode
}

// NewLockfreeQueue is the only way to get a new, ready-to-use LockfreeQueue.
//
// Example:
//
//	lfq := queue.NewLockfreeQueue()
//	lfq.Push(100)
//	v := lfq.Pop()
func NewLockfreeQueue() *LockfreeQueue {
	var lfq LockfreeQueue
	lfq.head = unsafe.Pointer(&lfq.dummy)
	lfq.tail = lfq.head
	return &lfq
}

// Pop returns (and removes) an element from the front of the queue, or nil if the queue is empty.
// It performs about 100% better than list.List.Front() and list.List.Remove() with sync.Mutex.
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

// Push inserts an element to the back of the queue.
// It performs exactly the same as list.List.PushBack() with sync.Mutex.
func (lfq *LockfreeQueue) Push(val interface{}) {
	node := unsafe.Pointer(&lfqNode{val: val})
	for {
		t := atomic.LoadPointer(&lfq.tail)
		rt := (*lfqNode)(t)
		if atomic.CompareAndSwapPointer(&rt.next, nil, node) {
			// It'll be a dead loop if atomic.StorePointer() is used.
			// Don't know why.
			// atomic.StorePointer(&lfq.tail, node)
			atomic.CompareAndSwapPointer(&lfq.tail, t, node)
			return
		} else {
			continue
		}
	}
}

type lfqNode struct {
	val  interface{}
	next unsafe.Pointer
}
