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

package pool

import (
	"sync"
)

// NewGoRoutinePool is the only way to get a new, ready-to-use GoRoutinePool.
//
// If you use `var goPool pool.GoRoutinePool`, or `new(pool.GoRoutinePool)`,
// or the like to obtain a GoRoutinePool, it'll still work,
// but it won't pool even a single goroutine.
//
//	maxGoRoutineNum: Maximum number of goroutines that will be pooled in GoRoutinePool
//
// Example:
//
//	goPool := pool.NewGoRoutinePool(100)
//	goPool.Run(func(){ fmt.Println("Hello, GoRoutinePool!") }) // runs a function using a pooled goroutine
func NewGoRoutinePool(maxGoRoutineNum int) *GoRoutinePool {
	return &GoRoutinePool{maxNum: maxGoRoutineNum}
}

// GoRoutinePool is a goroutine-safe pool for goroutines.
//
// After benchmarking, I found that use raw `go` keyword performs much better than this GoRoutinePool.
// So it makes no sense to use this GoRoutinePool.
type GoRoutinePool struct {
	lock     sync.Mutex
	freeList *goroutine
	freeNum  int
	maxNum   int
}

// Run executes a function using a pooled goroutine.
func (goPool *GoRoutinePool) Run(f func()) {
	goPool.lock.Lock()
	gr := goPool.freeList
	if gr != nil {
		goPool.freeList = gr.next
		goPool.freeNum--
	}
	goPool.lock.Unlock()

	if gr == nil {
		gr = &goroutine{
			ch:     make(chan func(), 1),
			goPool: goPool,
		}
		go gr.worker()
	}
	gr.ch <- f
}

// put returns a goroutine to the GoRoutinePool.
func (goPool *GoRoutinePool) put(gr *goroutine) {
	goPool.lock.Lock()
	if goPool.freeNum < goPool.maxNum {
		gr.next = goPool.freeList
		goPool.freeList = gr
		goPool.freeNum++
	} else {
		gr.ch <- nil
	}
	goPool.lock.Unlock()
}

// goroutine holds a channel for communicating with the goroutine worker
type goroutine struct {
	ch     chan func()
	goPool *GoRoutinePool
	next   *goroutine
}

// goroutine worker
func (gr *goroutine) worker() {
	for {
		f := <-gr.ch
		if f != nil {
			f()
			gr.goPool.put(gr)
		} else {
			break
		}
	}
}
