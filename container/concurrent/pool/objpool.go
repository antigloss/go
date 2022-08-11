/*
 *
 * pool - Goroutine-safe object pools.
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

// Package pool provides some goroutine-safe object pools such as ObjectPool.
package pool

import (
	"sync"
)

// CreateFunc is used by ObjectPool to create a new object when it's empty.
type CreateFunc[T any] func() *T

// ClearFunc is used by ObjectPool to reset a used object to it's initial state for reuse.
type ClearFunc[T any] func(*T)

// NewObjectPool is the only way to get a new, ready-to-use ObjectPool for objects of a specified type.
//
// If you use `var op pool.ObjectPool`, or `new(pool.ObjectPool)`, or the like to obtain an ObjectPool, it'll
// crash when you call Get().
//
//	maxObjectNum: Maximum number of objects that will be pooled in ObjectPool.
//	createObj: Called to create a new object when ObjectPool is empty. Cannot be nil.
//	clearObj: Called to reset a used object to it's initial state for reuse. Could be nil if it need not be reset.
//	          ObjectPool will perform about 12% faster if `clearObj` is nil.
//
// Example:
//
//	// create an ObjectPool for bytes.Buffer
//	op := pool.NewObjectPool[bytes.Buffer](10000,
//	                         func() *bytes.Buffer { return new(bytes.Buffer) },
//	                         func(obj *bytes.Buffer) { obj.Reset() })
//	obj := op.Get() // get a ready-to-use bytes.Buffer
//	// do something with `buf`
//	op.Put(obj) // return obj to ObjectPool.
func NewObjectPool[T any](maxObjectNum int, createObj CreateFunc[T], clearObj ClearFunc[T]) *ObjectPool[T] {
	return &ObjectPool[T]{maxObjNum: maxObjectNum, createFunc: createObj, clearFunc: clearObj}
}

// ObjectPool is a goroutine-safe generic pool for objects of any type.
type ObjectPool[T any] struct {
	lock       sync.Mutex
	freeList   *object[T]
	freeObjNum int
	maxObjNum  int
	createFunc CreateFunc[T]
	clearFunc  ClearFunc[T]
}

// Get returns a ready-to-use object.
func (op *ObjectPool[T]) Get() *T {
	op.lock.Lock()
	o := op.freeList
	if o != nil {
		op.freeList = o.next
		op.freeObjNum--
	}
	op.lock.Unlock()

	var obj *T
	if o != nil {
		obj = o.obj
		if op.clearFunc != nil {
			op.clearFunc(obj)
		}
		o.obj = nil
		o.next = nil
	} else {
		obj = op.createFunc()
	}
	return obj
}

// Put returns an object to ObjectPool.
func (op *ObjectPool[T]) Put(obj *T) {
	op.lock.Lock()
	if op.freeObjNum < op.maxObjNum {
		op.freeList = &object[T]{obj, op.freeList}
		op.freeObjNum++
	}
	op.lock.Unlock()
}

// object holds an object of arbitrary type for reuse.
type object[T any] struct {
	obj  *T
	next *object[T]
}
