# Overview

Package logger is a logging facility which provides functions Trace, Info, Warn, Error, Panic and Abort to
write logs with different severity levels. Logs with different severity levels are written to different logfiles.

# Features

	1. Auto rotation: It'll create a new logfile whenever day changes or size of the current logfile exceeds the configured size limit.
	2. Auto purging: It'll delete some oldest logfiles whenever the number of logfiles exceeds the configured limit.
	3. Log-through: Logs with higher severity level will be written to all the logfiles with lower severity level.
	4. Logs are not buffered, they are written to logfiles immediately with os.(*File).Write().
	5. Symlinks `PROG_NAME`.`USER_NAME`.`SEVERITY_LEVEL` will always link to the most current logfiles.
	6. Goroutine-safe.

# Basic example

	// logger.Init must be called first to setup logger
	logger.Init("./log", // specify the directory to save the logfiles
				400, // maximum logfiles allowed under the specified log directory
				20, // number of logfiles to delete when number of logfiles exceeds the configured limit
				100, // maximum size of a logfile in MB
				false) // whether logs with Trace level are written down
	logger.Info("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
	logger.Warn("Failed to parse protocol! uid=%d plid=%d cmd=%s", 1234, 678942, "getplayer")

# Performance

	package main

	import (
		"fmt"
		"github.com/antigloss/go/logger"
		"runtime"
		"sync"
		"time"
	)

	var wg sync.WaitGroup

	func main() {
		logger.Init("./log", 10, 2, 2, false)

		fmt.Print("Single goroutine (200000 writes), GOMAXPROCS(1): ")
		tSaved := time.Now()
		for i := 0; i != 200000; i++ {
			logger.Info("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
		}
		fmt.Println(time.Now().Sub(tSaved))

		fmt.Print("200000 goroutines (each makes 1 write), GOMAXPROCS(1): ")
		test()

		fmt.Print("200000 goroutines (each makes 1 write), GOMAXPROCS(2): ")
		runtime.GOMAXPROCS(2)
		test()

		fmt.Print("200000 goroutines (each makes 1 write), GOMAXPROCS(4): ")
		runtime.GOMAXPROCS(4)
		test()

		fmt.Print("200000 goroutines (each makes 1 write), GOMAXPROCS(8): ")
		runtime.GOMAXPROCS(8)
		test()
	}

	func test() {
		tSaved := time.Now()
		for i := 0; i != 200000; i++ {
			wg.Add(1)
			go func() {
				logger.Info("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
				wg.Add(-1)
			}()
		}
		wg.Wait()
		fmt.Println(time.Now().Sub(tSaved))
	}

Running this testing program under my development VM (i5-4590 3.3G 2 cores, Samsung SSD 840 EVO):

	Single goroutine (200000 writes), GOMAXPROCS(1): 675.824756ms
	200000 goroutines (each makes 1 write), GOMAXPROCS(1): 1.306264354s
	200000 goroutines (each makes 1 write), GOMAXPROCS(2): 755.983595ms
	200000 goroutines (each makes 1 write), GOMAXPROCS(4): 903.31128ms
	200000 goroutines (each makes 1 write), GOMAXPROCS(8): 1.080061112s

Running this testing program under a cloud server (Unknown brand CPU 2.6G 8 cores, Unknown brand HDD):

	Single goroutine (200000 writes), GOMAXPROCS(1): 1.298951897s
	200000 goroutines (each makes 1 write), GOMAXPROCS(1): 2.403048438s
	200000 goroutines (each makes 1 write), GOMAXPROCS(2): 1.577390142s
	200000 goroutines (each makes 1 write), GOMAXPROCS(4): 2.079531449s
	200000 goroutines (each makes 1 write), GOMAXPROCS(8): 2.452058765s
