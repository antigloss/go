// Author: https://github.com/antigloss

/*
Package logger is a logging facility which provides functions Trace, Info, Warn, Error, Panic and Abort to
write logs with different severity levels. Logs with different severity levels are written to different logfiles.

Sorry for my poor English, I've tried my best.

Features:

	1. Auto rotation: It'll create a new logfile whenever day changes or size of the current logfile exceeds the configured size limit.
	2. Auto purging: It'll delete some oldest logfiles whenever the number of logfiles exceeds the configured limit.
	3. Log-through: Logs with higher severity level will be written to all the logfiles with lower severity level.
	4. Logs are not buffered, they are written to logfiles immediately with os.(*File).Write().
	5. Symlinks `PROG_NAME`.`SEVERITY_LEVEL` will always link to the most current logfiles.
	6. Goroutine-safe.

Basic example:

	// logger.Init must be called first to setup logger
	logger.Init("./log", // specify the directory to save the logfiles
			400, // maximum logfiles allowed under the specified log directory
			20, // number of logfiles to delete when number of logfiles exceeds the configured limit
			100, // maximum size of a logfile in MB
			false) // whether logs with Trace level are written down
	logger.Info("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
	logger.Warn("Failed to parse protocol! uid=%d plid=%d cmd=%s", 1234, 678942, "getplayer")

Performance:

	package main

	import (
		"fmt"
		"github.com/antigloss/logger"
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
*/
package logger

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"sync"
	"time"
)

// consts
const (
	kMaxInt64          = int64(^uint64(0) >> 1)
	kLogCreatedTimeLen = 21
	kLogFilenameMinLen = 36
)

// log level
const (
	kLogLevelTrace = iota
	kLogLevelInfo
	kLogLevelWarn
	kLogLevelError
	kLogLevelPanic
	kLogLevelAbort

	kLogLevelMax
)

// log flags
const (
	kFlagLogTrace = 1 << iota
	kFlagLogThrough
	kFlagLogFuncName
	kFlagLogFilenameLineNum
)

const gLogLevelChar = "TIWEPA"

// Init must be called first, otherwise this logger will not function properly!
// It returns nil if all goes well, otherwise it returns the corresponding error.
//   maxfiles: Must be greater than 0 and less than or equal to 100000.
//   nfilesToDel: Number of files deleted when number of log files reaches `maxfiles`.
//                Must be greater than 0 and less than or equal to `maxfiles`.
//   maxsize: Maximum size of a log file in MB, 0 means unlimited.
//   logTrace: If set to false, `logger.Trace("xxxx")` will be mute.
func Init(logpath string, maxfiles, nfilesToDel int, maxsize uint32, logTrace bool) error {
	err := os.MkdirAll(logpath, 0755)
	if err != nil {
		return err
	}

	if maxfiles <= 0 || maxfiles > 100000 {
		return fmt.Errorf("maxfiles must be greater than 0 and less than or equal to 100000: %d", maxfiles)
	}

	if nfilesToDel <= 0 || nfilesToDel > maxfiles {
		return fmt.Errorf("nfilesToDel must be greater than 0 and less than or equal to maxfiles! toDel=%d maxfiles=%d",
			nfilesToDel, maxfiles)
	}

	// get names form the directory `logpath`
	files, err := getDirnames(logpath)
	if err != nil {
		return err
	}

	gConf.setLogPath(logpath)
	gConf.setFlags(kFlagLogTrace, logTrace)
	gConf.maxfiles = maxfiles
	gConf.curfiles = calcLogfileNum(files)
	gConf.nfilesToDel = nfilesToDel
	gConf.setMaxSize(maxsize)
	return nil
}

// SetLogThrough sets whether to write log to all the logfiles with less severe log level.
// By default, logthrough is turn on. You can turn it off for better performance.
func SetLogThrough(on bool) {
	gConf.setFlags(kFlagLogThrough, on)
}

// SetLogFunctionName sets whether to log down the function name where the log takes place.
// By default, function name is not logged down for better performance.
func SetLogFunctionName(on bool) {
	gConf.setFlags(kFlagLogFuncName, on)
}

// SetLogFilenameLineNum sets whether to log down the filename and line number where the log takes place.
// By default, filename and line number are logged down. You can turn it off for better performance.
func SetLogFilenameLineNum(on bool) {
	gConf.setFlags(kFlagLogFilenameLineNum, on)
}

// Trace logs down a log with trace level.
// If parameter logTrace of logger.Init() is set to be false, no trace logs will be logged down.
func Trace(format string, args ...interface{}) {
	if gConf.logTrace() {
		log(kLogLevelTrace, format, args)
	}
}

// Info logs down a log with info level.
func Info(format string, args ...interface{}) {
	log(kLogLevelInfo, format, args)
}

// Warn logs down a log with warning level.
func Warn(format string, args ...interface{}) {
	log(kLogLevelWarn, format, args)
}

// Error logs down a log with error level.
func Error(format string, args ...interface{}) {
	log(kLogLevelError, format, args)
}

// Panic logs down a log with panic level and then panic("panic log") is called.
func Panic(format string, args ...interface{}) {
	log(kLogLevelPanic, format, args)
	panic("panic log")
}

// Abort logs down a log with abort level and then os.Exit(-1) is called.
func Abort(format string, args ...interface{}) {
	log(kLogLevelAbort, format, args)
	os.Exit(-1)
}

// logger configuration
type config struct {
	logPath     string
	pathPrefix  string
	logflags    uint32
	maxfiles    int   // limit the number of log files under `logPath`
	curfiles    int   // number of files under `logPath` currently
	nfilesToDel int   // number of files deleted when reaching the limit of the number of log files
	maxsize     int64 // limit size of a log file
	purgeLock   sync.Mutex
}

func (conf *config) setFlags(flag uint32, on bool) {
	if on {
		conf.logflags = conf.logflags | flag
	} else {
		conf.logflags = conf.logflags & ^flag
	}
}

func (conf *config) logTrace() bool {
	return (conf.logflags & kFlagLogTrace) != 0
}

func (conf *config) logThrough() bool {
	return (conf.logflags & kFlagLogThrough) != 0
}

func (conf *config) logFuncName() bool {
	return (conf.logflags & kFlagLogFuncName) != 0
}

func (conf *config) logFilenameLineNum() bool {
	return (conf.logflags & kFlagLogFilenameLineNum) != 0
}

func (conf *config) setMaxSize(maxsize uint32) {
	if maxsize > 0 {
		conf.maxsize = int64(maxsize) * 1024 * 1024
	} else {
		conf.maxsize = kMaxInt64 - (1024 * 1024 * 1024 * 1024 * 1024)
	}
}

// logger
type logger struct {
	file  *os.File
	level int
	day   int
	size  int64
	lock  sync.Mutex
}

func (l *logger) log(t time.Time, data []byte) {
	y, m, d := t.Date()

	l.lock.Lock()
	defer l.lock.Unlock()
	if l.size >= gConf.maxsize || l.day != d || l.file == nil {
		gConf.purgeLock.Lock()
		hasLocked := true
		defer func() {
			if hasLocked {
				gConf.purgeLock.Unlock()
			}
		}()
		// reaches limit of number of log files
		if gConf.curfiles >= gConf.maxfiles {
			files, err := getDirnames(gConf.logPath)
			if err != nil {
				l.errlog(t, data, err)
				return
			}

			if len(files) >= gConf.maxfiles {
				sort.Sort(byCreatedTime(files))
				nfiles := gConf.curfiles - gConf.maxfiles + gConf.nfilesToDel
				if nfiles > len(files) {
					nfiles = len(files)
					gConf.curfiles = nfiles
				}
				for i := 0; i != nfiles; i++ {
					err := os.RemoveAll(gConf.logPath + files[i])
					if err == nil {
						gConf.curfiles--
					} else {
						l.errlog(t, nil, err)
					}
				}
			} else {
				gConf.curfiles = calcLogfileNum(files)
			}
		}

		filename := fmt.Sprintf("%s%s.%d%02d%02d-%012d", gConf.pathPrefix, gLogLevelNames[l.level],
			y, m, d, (t.UnixNano()/1000)%1000000000000)
		newfile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			l.errlog(t, data, err)
			return
		}
		gConf.curfiles++
		gConf.purgeLock.Unlock()
		hasLocked = false

		l.file.Close()
		l.file = newfile
		l.day = d
		l.size = 0

		err = os.RemoveAll(gFullSymlinks[l.level])
		if err != nil {
			l.errlog(t, nil, err)
		}
		err = os.Symlink(path.Base(filename), gFullSymlinks[l.level])
		if err != nil {
			l.errlog(t, nil, err)
		}
	}

	n, _ := l.file.Write(data)
	l.size += int64(n)
}

// (l *logger).errlog() should only be used within (l *logger).log()
func (l *logger) errlog(t time.Time, originLog []byte, err error) {
	buf := gBufPool.getBuffer()

	genLogPrefix(buf, l.level, 2, t)
	buf.WriteString(err.Error())
	buf.WriteByte('\n')
	if l.file != nil {
		l.file.Write(buf.Bytes())
		if len(originLog) > 0 {
			l.file.Write(originLog)
		}
	} else {
		fmt.Fprint(os.Stderr, buf.String())
		if len(originLog) > 0 {
			fmt.Fprint(os.Stderr, string(originLog))
		}
	}

	gBufPool.putBuffer(buf)
}

// sort files by created time embedded in the filename
type byCreatedTime []string

func (a byCreatedTime) Len() int {
	return len(a)
}

func (a byCreatedTime) Less(i, j int) bool {
	s1, s2 := a[i], a[j]
	if len(s1) < kLogFilenameMinLen {
		if isSymlink[s1] {
			return false
		}
		return true
	} else if len(s2) < kLogFilenameMinLen {
		if isSymlink[s2] {
			return true
		}
		return false
	} else {
		return s1[len(s1)-kLogCreatedTimeLen:] < s2[len(s2)-kLogCreatedTimeLen:]
	}
}

func (a byCreatedTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// init is called after all the variable declarations in the package have evaluated their initializers,
// and those are evaluated only after all the imported packages have been initialized.
// Besides initializations that cannot be expressed as declarations, a common use of init functions is to verify
// or repair correctness of the program state before real execution begins.
func init() {
	for i := 0; i != kLogLevelMax; i++ {
		gLoggers[i].level = i
		gSymlinks[i] = gProgname + "." + gLogLevelNames[i]
		isSymlink[gSymlinks[i]] = true
	}

	gConf.setLogPath("./log")
}

// helpers
func getDirnames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err == nil {
		defer f.Close()
		return f.Readdirnames(0)
	}
	return nil, err
}

func calcLogfileNum(files []string) int {
	curfiles := 0
	for _, filename := range files {
		if isSymlink[filename] == false {
			curfiles++
		}
	}
	return curfiles
}

func genLogPrefix(buf *buffer, logLevel, skip int, t time.Time) {
	h, m, s := t.Clock()

	// time
	buf.tmp[0] = gLogLevelChar[logLevel]
	buf.twoDigits(1, h)
	buf.tmp[3] = ':'
	buf.twoDigits(4, m)
	buf.tmp[6] = ':'
	buf.twoDigits(7, s)
	buf.Write(buf.tmp[:9])

	var pc uintptr
	var ok bool
	if gConf.logFilenameLineNum() {
		var file string
		var line int
		pc, file, line, ok = runtime.Caller(skip)
		if ok {
			buf.WriteByte(' ')
			buf.WriteString(path.Base(file))
			buf.tmp[0] = ':'
			n := buf.someDigits(1, line)
			buf.Write(buf.tmp[:n+1])
		}
	}
	if gConf.logFuncName() {
		if !ok {
			pc, _, _, ok = runtime.Caller(skip)
		}
		if ok {
			buf.WriteByte(' ')
			buf.WriteString(runtime.FuncForPC(pc).Name())
		}
	}

	buf.WriteString("] ")
}

func log(logLevel int, format string, args []interface{}) {
	buf := gBufPool.getBuffer()

	t := time.Now()
	genLogPrefix(buf, logLevel, 3, t)
	fmt.Fprintf(buf, format, args...)
	buf.WriteByte('\n')
	output := buf.Bytes()
	if gConf.logThrough() {
		for i := logLevel; i != kLogLevelTrace; i-- {
			gLoggers[i].log(t, output)
		}
		if gConf.logTrace() {
			gLoggers[kLogLevelTrace].log(t, output)
		}
	} else {
		gLoggers[logLevel].log(t, output)
	}

	gBufPool.putBuffer(buf)
}

var gProgname = path.Base(os.Args[0])

var gLogLevelNames = [kLogLevelMax]string{
	"TRACE", "INFO", "WARN", "ERROR", "PANIC", "ABORT",
}

var gConf = config{
	logflags:    kFlagLogFilenameLineNum | kFlagLogThrough,
	maxfiles:    400,
	nfilesToDel: 10,
	maxsize:     100 * 1024 * 1024,
}

var gSymlinks [kLogLevelMax]string
var isSymlink = map[string]bool{}
var gFullSymlinks [kLogLevelMax]string
var gBufPool bufferPool
var gLoggers [kLogLevelMax]logger
