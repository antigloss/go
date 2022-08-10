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

package queue

import (
	"runtime"
	"sort"
	"sync"
	"testing"
)

const (
	kGoRoutineNum = 10
	kPushingNum   = 500000
	kBufSz        = kGoRoutineNum * kPushingNum
)

var out *testing.T
var wg sync.WaitGroup
var lfq = NewLockfreeQueue()
var popBuf [kGoRoutineNum][]int

func TestLockfreeQueue(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	out = t
	// init popBuf
	for i := 0; i != kGoRoutineNum; i++ {
		popBuf[i] = make([]int, 0, kBufSz)
	}

	// Push() simultaneously
	wg.Add(kGoRoutineNum)
	for i := 0; i != kGoRoutineNum; i++ {
		go push()
	}
	wg.Wait()
	// Pop() simultaneously
	wg.Add(kGoRoutineNum)
	for i := 0; i != kGoRoutineNum; i++ {
		go pop_only()
	}
	wg.Wait()

	// Push() and Pop() simultaneously
	wg.Add(kGoRoutineNum * 2)
	for i := 0; i != kGoRoutineNum; i++ {
		go push()
		go pop_while_pushing(i)
	}
	wg.Wait()
	// Verification
	resultBuf := popBuf[0]
	for i := 1; i != kGoRoutineNum; i++ {
		resultBuf = append(resultBuf, popBuf[i]...)
	}
	// in case there are some elements left in the queue
	for v := lfq.Pop(); v != nil; v = lfq.Pop() {
		resultBuf = append(resultBuf, v.(int))
	}
	sort.Ints(resultBuf)
	for i := 0; i != kPushingNum; i++ {
		for j := 0; j != kGoRoutineNum; j++ {
			if resultBuf[(i*kGoRoutineNum)+j] != i {
				t.Error("Invalid result:", i, j, resultBuf[(i*kGoRoutineNum)+j])
			}
		}
	}
}

func push() {
	for i := 0; i != kPushingNum; i++ {
		lfq.Push(i)
	}
	wg.Done()
}

func pop_only() {
	for i := 0; i != kPushingNum; i++ {
		v := lfq.Pop()
		if v == nil {
			out.Error("Should never be nil!")
		}
	}
	wg.Done()
}

func pop_while_pushing(n int) {
	for i := 0; i != kPushingNum*2; i++ {
		v := lfq.Pop()
		if v != nil {
			popBuf[n] = append(popBuf[n], v.(int))
		}
	}
	wg.Done()
}
