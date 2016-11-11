package logger

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

func BenchmarkLogger(b *testing.B) {
	//Init("./log", 10, 2, 2, false)

	benchmark(b, 1, 1, 100000)
	benchmark(b, 1, 100000, 1)
	benchmark(b, 2, 100000, 1)
	benchmark(b, 4, 100000, 1)
	benchmark(b, 8, 100000, 1)
}

func benchmark(b *testing.B, nProcs, nGoroutines, nWrites int) {
	runtime.GOMAXPROCS(nProcs)
	bn := fmt.Sprintf("%d Procs %d Goroutines each makes %d writes", nProcs, nGoroutines, nWrites)

	b.Run(bn, func(b *testing.B) {
		var wg sync.WaitGroup
		for i := 0; i < nGoroutines; i++ {
			wg.Add(1)
			go func() {
				for j := 0; j < nWrites; j++ {
					Info("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
				}
				wg.Add(-1)
			}()
		}
		wg.Wait()
	})
}
