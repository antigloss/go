/*
 *
 * sync - Synchronization facilities.
 * Copyright (C) 2018 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

package utils

import (
	"runtime"
	"sort"
	"sync"
	"testing"
)

const (
	kGoRoutineNum   = 100
	kGetSeqNumTimes = 100000
	kTotalSeqNum    = kGoRoutineNum * kGetSeqNumTimes
)

func TestMonoIncSeqNumGenerator64(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	seqGenerator := NewMonoIncSeqNumGenerator64(0)
	var seqBuf [kGoRoutineNum][kGetSeqNumTimes]int
	var wg sync.WaitGroup
	wg.Add(kGoRoutineNum)
	for i := 0; i != kGoRoutineNum; i++ {
		go func(n int) {
			for i := 0; i != kGetSeqNumTimes; i++ {
				seqBuf[n][i] = int(seqGenerator.GetSeqNum())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	// Verification
	resultBuf := seqBuf[0][0:]
	for i := 1; i != kGoRoutineNum; i++ {
		resultBuf = append(resultBuf, seqBuf[i][0:]...)
	}
	sort.Ints(resultBuf)
	for i := 0; i != kTotalSeqNum; i++ {
		if i+1 != resultBuf[i] {
			t.Error("Invalid result:", i+1, resultBuf[i])
			break
		}
	}
}
