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
	kTotalSeqNum	= kGoRoutineNum * kGetSeqNumTimes
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
		if i + 1 != resultBuf[i] {
			t.Error("Invalid result:", i + 1, resultBuf[i])
			break
		}
	}
}
