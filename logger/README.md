# Overview

Package logger is a goroutine-safe logging facility which  writes logs with different severity levels to files, console, or both. Logs with different severity levels are written to different logfiles.

# Features

1. Auto rotation: It'll create a new logfile whenever day changes or size of the current logfile exceeds the configured size limit.
2. Auto purging: It'll delete some oldest logfiles whenever the number of logfiles exceeds the configured limit.
3. Log-through: Logs with higher severity level will be written to all the logfiles with lower severity level.
4. Log levels: 6 different levels are supported. Logs with different levels are written to different logfiles. By setting the Logger object to a higher log level, lower level logs will be filtered out.
5. Logs are not buffered, they are written to logfiles immediately with os.(*File).Write().
6. It'll create symlinks that link to the most current logfiles.

# Basic example

## Use the global Logger object
	// logger.Init must be called first to create the global Logger object
	err := logger.Init(&logger.Config{
    	LogDir:          "./logs",
    	LogFileMaxSize:  200,
    	LogFileMaxNum:   500,
    	LogFileNumToDel: 50,
    	LogLevel:        logger.LogLevelInfo,
    	LogDest:         logger.LogDestFile,
    	Flag:            logger.ControlFlagLogLineNum,
    })
    if err != nil {
        panic(err)
    }
	logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
	logger.Warnf("Failed to parse protocol! uid=%d plid=%d cmd=%s", 1234, 678942, "getplayer")

## Create a new Logger object

While the global Logger object should be enough for most cases, you can also create as many Logger objects as desired.

    newLogger, err := logger.New(&logger.Config{
    	LogDir:          "./logs",
    	LogFileMaxSize:  200,
    	LogFileMaxNum:   500,
    	LogFileNumToDel: 50,
    	LogLevel:        logger.LogLevelInfo,
    	LogDest:         logger.LogDestFile,
    	Flag:            logger.ControlFlagLogLineNum,
    })
    if err != nil {
        panic(err)
    }
	newLogger.Info("abc", 123, 444.77)
	newLogger.Error("abc", 123, 444.77)
	newLogger.Close()

# Documentation

Please refer to [godoc](https://godoc.org/github.com/antigloss/go/logger) or [pkg.go.dev](https://pkg.go.dev/github.com/antigloss/go/logger).

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
        logger.Init(&logger.Config{
            LogDir:          "./logs",
            LogFileMaxSize:  200,
            LogFileMaxNum:   500,
            LogFileNumToDel: 50,
            LogLevel:        logger.LogLevelInfo,
            LogDest:         logger.LogDestFile,
            Flag:            logger.ControlFlagLogLineNum,
        })
    
        fmt.Print("Single goroutine (1000000 writes), GOMAXPROCS(1): ")
        tSaved := time.Now()
        for i := 0; i != 1000000; i++ {
            logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
        }
        fmt.Println(time.Now().Sub(tSaved))
    
        fmt.Print("100000 goroutines (each makes 10 writes), GOMAXPROCS(1): ")
        test()
    
        fmt.Print("100000 goroutines (each makes 10 writes), GOMAXPROCS(2): ")
        runtime.GOMAXPROCS(2)
        test()
    
        fmt.Print("100000 goroutines (each makes 10 writes), GOMAXPROCS(4): ")
        runtime.GOMAXPROCS(4)
        test()
    
        fmt.Print("100000 goroutines (each makes 10 writes), GOMAXPROCS(8): ")
        runtime.GOMAXPROCS(8)
        test()
    
        fmt.Print("100000 goroutines (each makes 10 writes), GOMAXPROCS(16): ")
        runtime.GOMAXPROCS(16)
        test()
    }
    
    func test() {
        tSaved := time.Now()
        for i := 0; i != 100000; i++ {
            wg.Add(1)
            go func() {
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                logger.Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
                wg.Add(-1)
            }()
        }
        wg.Wait()
        fmt.Println(time.Now().Sub(tSaved))
    }

Running this test program under Win10, i7-9700 @ 3.00GHz(8 cores, 8 threads), WDS500G2B0A-00SM50(SSD):

    Single goroutine (1000000 writes), GOMAXPROCS(1): 3.2784983s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(1): 4.0297497s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(2): 3.7067683s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(4): 3.8241992s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(8): 4.2463802s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(16): 4.1943385s

About 233 ~ 303 thousand writes per second.

Under macOS Catalina 10.15.6，i9-9880H @ 2.3GHz(8 cores, 16 threads), APPLE SSD AP2048N:

    Single goroutine (1000000 writes), GOMAXPROCS(1): 4.311780023s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(1): 6.627393432s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(2): 5.553245254s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(4): 6.579052789s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(8): 5.725944572s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(16): 5.952777505s

About 149 ~ 231 thousand writes per second.

Under a cloud server, Ubuntu 18.04, Xeon Gold 6149 @ 3.10GHz(16 cores), unknown brand HDD:

    Single goroutine (1000000 writes), GOMAXPROCS(1): 3.183363298s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(1): 3.992387381s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(2): 3.96368545s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(4): 3.936505055s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(8): 3.843349416s
    100000 goroutines (each makes 10 writes), GOMAXPROCS(16): 3.864881766s

About 250 ~ 312 thousand writes per second.
