/*
 *
 * logger - A package for writing logs
 * Copyright (C) 2020 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

package logger

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LogLevel int // LogLevel is used to exclude logs with lower level.

const (
	LogLevelTrace LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelPanic // Call panic() after log is written.
	LogLevelFatal // Call os.Exit(-1) after log is written.
	LogLevelCount // Number of different log levels.
)
const (
	kLogLevelTrace = iota
	kLogLevelInfo
	kLogLevelWarn
	kLogLevelError
	kLogLevelPanic
	kLogLevelFatal
	kLogLevelCount // Number of different log levels.
)

type LogDest uint32 // LogDest controls where the logs are written.

const (
	LogDestFile    LogDest = 1 << iota // Write logs to files.
	LogDestConsole                     // Write logs to console.
	LogDestNone    = 0                 // Don't write logs.
)
const (
	LogDestBoth = LogDestFile | LogDestConsole // Write logs both to files and console.
)
const (
	kLogDestFile = 1 << iota
	kLogDestConsole
	kLogDestNone = 0
)

type ControlFlag int // ControlFlag controls how the logs are written. Use `|`(Or operator) to mix multiple flags.

const (
	ControlFlagLogThrough  ControlFlag = 1 << iota // Controls if logs with higher level are written to lower level log files.
	ControlFlagLogFuncName                         // Controls if function name is prepended to the logs.
	ControlFlagLogLineNum                          // Controls if filename and line number are prepended to the logs.
	ControlFlagLogDate                             // Controls if a date string formatted as '20201201' is prepended to the logs.
	ControlFlagNone        = 0
)

// Config contains options for creating a new Logger object.
type Config struct {
	// Directory to hold the log files. If left empty, current working directory is used.
	// Should you need to create multiple Logger objects, better to associate them with different directories.
	LogDir string
	// Name of a log file is formatted as `LogFilenamePrefix.LogLevel.DateTime.log`.
	// 3 placeholders are pre-defined: %P, %H and %U. When used in the prefix,
	// %P will be replaced with the program's name, %H will be replaced with hostname,
	// and %U will be replaced with username.
	// If LogFilenamePrefix is left empty, it'll be defaulted to `%P.%H.%U`.
	// If you create multiple Logger objects with the same directory, you must associate them with different prefixes.
	LogFilenamePrefix string
	// Latest log files of each level are associated with symbolic links. Name of a symlink is formatted as `LogSymlinkPrefix.LogLevel`.
	// If LogSymlinkPrefix is left empty, it'll be defaulted to `%P.%U`.
	// If you create multiple Logger objects with the same directory, you must associate them with different prefixes.
	LogSymlinkPrefix string
	// Limit the maximum size in MB for a single log file. 0 means unlimited.
	LogFileMaxSize uint32
	// Limit the maximum number of log files under `LogDir`. `LogFileNumToDel` log files will be deleted if reached. <=0 means unlimited.
	LogFileMaxNum int
	// Number of log files to be deleted when `LogFileMaxNum` reached. <=0 means don't delete.
	LogFileNumToDel int
	// Don't write logs below `LogLevel`.
	LogLevel LogLevel
	// Where the logs are written.
	LogDest LogDest
	// How the logs are written.
	Flag ControlFlag
}

// Init is used to create the global Logger object with cfg. It must be called once and only once
// before any other function backed by the global Logger object can be used.
// It returns nil if all goes well, otherwise it returns the corresponding error.
func Init(cfg *Config) (err error) {
	defLoggerLock.Lock()
	defer defLoggerLock.Unlock()

	if defLogger == nil {
		defLogger, err = New(cfg)
	}
	return
}

// SetLogLevel is used to tell the global Logger object created by Init not to write logs below logLevel.
func SetLogLevel(logLevel LogLevel) {
	defLogger.SetLogLevel(logLevel)
}

// Trace uses the global Logger object created by Init to write a log with trace level.
func Trace(args ...interface{}) {
	defLogger.log(kLogLevelTrace, args)
}

// Tracef uses the global Logger object created by Init to write a log with trace level.
func Tracef(format string, args ...interface{}) {
	defLogger.logf(kLogLevelTrace, format, args)
}

// Info uses the global Logger object created by Init to write a log with info level.
func Info(args ...interface{}) {
	defLogger.log(kLogLevelInfo, args)
}

// Infof uses the global Logger object created by Init to write a log with info level.
func Infof(format string, args ...interface{}) {
	defLogger.logf(kLogLevelInfo, format, args)
}

// Warn uses the global Logger object created by Init to write a log with warning level.
func Warn(args ...interface{}) {
	defLogger.log(kLogLevelWarn, args)
}

// Warnf uses the global Logger object created by Init to write a log with warning level.
func Warnf(format string, args ...interface{}) {
	defLogger.logf(kLogLevelWarn, format, args)
}

// Error uses the global Logger object created by Init to write a log with error level.
func Error(args ...interface{}) {
	defLogger.log(kLogLevelError, args)
}

// Errorf uses the global Logger object created by Init to write a log with error level.
func Errorf(format string, args ...interface{}) {
	defLogger.logf(kLogLevelError, format, args)
}

// Panic uses the global Logger object created by Init to write a log with panic level followed by a call to panic("Panicf").
func Panic(args ...interface{}) {
	defLogger.log(kLogLevelPanic, args)
	panic("Panic")
}

// Panicf uses the global Logger object created by Init to write a log with panic level followed by a call to panic("Panicf").
func Panicf(format string, args ...interface{}) {
	defLogger.logf(kLogLevelPanic, format, args)
	panic("Panicf")
}

// Fatal uses the global Logger object created by Init to write a log with fatal level followed by a call to os.Exit(-1).
func Fatal(args ...interface{}) {
	defLogger.log(kLogLevelFatal, args)
	os.Exit(-1)
}

// Fatalf uses the global Logger object created by Init to write a log with fatal level followed by a call to os.Exit(-1).
func Fatalf(format string, args ...interface{}) {
	defLogger.logf(kLogLevelFatal, format, args)
	os.Exit(-1)
}

// Logger can be used to write logs with different severity levels to files, console, or both.
// Logs with different severity levels are written to different files. It is goroutine-safe and supports the following features:
//
//  1. Auto rotation: It'll create a new logfile whenever day changes or size of the current logfile exceeds the configured size limit.
//  2. Auto purging: It'll delete some oldest logfiles whenever the number of logfiles exceeds the configured limit.
//  3. Log-through: Logs with higher severity level will be written to all the logfiles with lower severity level.
//  4. Log levels: 6 different levels are supported. Logs with different levels are written to different logfiles. By setting the Logger object to a higher log level, lower level logs will be filtered out.
//  5. Logs are not buffered, they are written to logfiles immediately with os.(*File).Write().
//  6. It'll create symlinks that link to the most current logfiles.
type Logger struct {
	// Variables not allowed to be changed at runtime go here
	logDir         string
	logPathPrefix  string
	logFileMaxSize int64
	logFileMaxNum  int
	logFilesToDel  int
	flag           ControlFlag

	// Variables allowed to be changed at runtime go here
	logLevel int32
	logDest  uint32

	// Variables used by the log-purging goroutine go here
	logFileCurNum    int // number of log files under `logDir` currently
	logFilenameRegex *regexp.Regexp
	logFilePurgeCh   chan bool

	// Logger implementation
	bufPool bufferPool
	loggers [kLogLevelCount]logger
}

// New can be used to create as many Logger objects as desired, while the global Logger object created by Init should be enough for most cases.
// Should you need to create multiple Logger objects, better to associate them with different directories, at least with different filename prefixes(including symlink prefixes),
// otherwise they will not work properly.
func New(cfg *Config) (logger *Logger, err error) {
	logDir := cfg.LogDir
	if len(logDir) > 0 {
		err = os.MkdirAll(logDir, 0755)
		if err != nil {
			return
		}
		if logDir[len(logDir)-1] != os.PathSeparator {
			logDir += string(os.PathSeparator)
		}
	} else {
		logDir, err = os.Getwd()
		if err != nil {
			return
		}
		logDir += string(os.PathSeparator)
	}

	logger = &Logger{
		logDir:        logDir,
		logFileMaxNum: cfg.LogFileMaxNum,
		logFileCurNum: cfg.LogFileMaxNum, // Force to check if purging needed at startup
		logFilesToDel: cfg.LogFileNumToDel,
		logLevel:      int32(cfg.LogLevel),
		logDest:       uint32(cfg.LogDest),
		flag:          cfg.Flag,
	}

	if cfg.LogFileMaxSize > 0 {
		logger.logFileMaxSize = int64(cfg.LogFileMaxSize) * 1024 * 1024
	} else {
		logger.logFileMaxSize = kMaxInt64 - (1024 * 1024 * 1024 * 1024)
	}

	err = logger.initLoggerImpl(cfg.LogFilenamePrefix, cfg.LogSymlinkPrefix)
	if err != nil {
		logger = nil
	}
	return
}

// Close should be call once and only once to destroy the Logger object.
func (l *Logger) Close() error {
	atomic.StoreUint32(&l.logDest, kLogDestNone)
	for i := kLogLevelTrace; i != kLogLevelCount; i++ {
		l.loggers[i].close()
	}
	l.logFilePurgeCh <- false

	return nil
}

// SetLogLevel tells the Logger object not to write logs below `logLevel`.
func (l *Logger) SetLogLevel(logLevel LogLevel) {
	atomic.StoreInt32(&l.logLevel, int32(logLevel))
}

// Trace writes a log with trace level.
func (l *Logger) Trace(args ...interface{}) {
	l.log(kLogLevelTrace, args)
}

// Tracef writes a log with trace level.
func (l *Logger) Tracef(format string, args ...interface{}) {
	l.logf(kLogLevelTrace, format, args)
}

// Info writes a log with info level.
func (l *Logger) Info(args ...interface{}) {
	l.log(kLogLevelInfo, args)
}

// Infof writes a log with info level.
func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf(kLogLevelInfo, format, args)
}

// Warn writes a log with warning level.
func (l *Logger) Warn(args ...interface{}) {
	l.log(kLogLevelWarn, args)
}

// Warnf writes a log with warning level.
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf(kLogLevelWarn, format, args)
}

// Error writes a log with error level.
func (l *Logger) Error(args ...interface{}) {
	l.log(kLogLevelError, args)
}

// Errorf writes a log with error level.
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf(kLogLevelError, format, args)
}

// Panic writes a log with panic level followed by a call to panic("Panic").
func (l *Logger) Panic(args ...interface{}) {
	l.log(kLogLevelPanic, args)
	panic("Panic")
}

// Panicf writes a log with panic level followed by a call to panic("Panicf").
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logf(kLogLevelPanic, format, args)
	panic("Panicf")
}

// Fatal writes a log with fatal level followed by a call to os.Exit(-1).
func (l *Logger) Fatal(args ...interface{}) {
	l.log(kLogLevelFatal, args)
	os.Exit(-1)
}

// Fatalf writes a log with fatal level followed by a call to os.Exit(-1).
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf(kLogLevelFatal, format, args)
	os.Exit(-1)
}

func (l *Logger) initLoggerImpl(filenamePrefix, symlinkPrefix string) (err error) {
	if len(filenamePrefix) == 0 {
		filenamePrefix = "%P.%H.%U" // Default value
	}
	filenamePrefix = strings.Replace(filenamePrefix, "%P", kProgramName, -1)
	filenamePrefix = strings.Replace(filenamePrefix, "%H", kHostname, -1)
	filenamePrefix = strings.Replace(filenamePrefix, "%U", kUsername, -1)
	l.logPathPrefix = l.logDir + filenamePrefix + "."

	if len(symlinkPrefix) == 0 {
		symlinkPrefix = "%P.%U" // Default value
	}
	symlinkPrefix = strings.Replace(symlinkPrefix, "%P", kProgramName, -1)
	symlinkPrefix = strings.Replace(symlinkPrefix, "%H", kHostname, -1)
	symlinkPrefix = strings.Replace(symlinkPrefix, "%U", kUsername, -1)
	symlinkPrefix += "."

	for i := int32(kLogLevelTrace); i != kLogLevelCount; i++ {
		l.loggers[i].level = i
		l.loggers[i].parent = l
		l.loggers[i].symlinkFullPath = l.logDir + symlinkPrefix + kLogLevelNames[i]
	}

	if l.logFileMaxNum > 0 && l.logFilesToDel > 0 {
		var sb strings.Builder
		sb.WriteByte('^')
		sb.WriteString(regexp.QuoteMeta(filenamePrefix))
		sb.WriteString(`\.(`)
		lastLevelNameIdx := len(kLogLevelNames) - 1
		for i := 0; i < lastLevelNameIdx; i++ {
			sb.WriteString(kLogLevelNames[i])
			sb.WriteByte('|')
		}
		sb.WriteString(kLogLevelNames[lastLevelNameIdx])
		sb.WriteString(`)\.\d{20}\.log$`)

		l.logFilenameRegex, err = regexp.Compile(sb.String())
		if err == nil {
			l.logFilePurgeCh = make(chan bool, 4096)
			go l.purgeLogFiles() // Purge old log files in another goroutine
		}
	}

	return
}

func (l *Logger) purgeLogFiles() {
	l.tryPurgeOldLogFiles()

	for r := range l.logFilePurgeCh {
		if !r {
			return
		}

		l.logFileCurNum++
		l.tryPurgeOldLogFiles()
	}
}

func (l *Logger) tryPurgeOldLogFiles() {
	if l.logFileCurNum < l.logFileMaxNum {
		return
	}

	files, err := l.getLogFilenames()
	if err != nil {
		l.Errorf("Failed to purge old log files: %s", err)
		return
	}
	l.logFileCurNum = len(files)

	if l.logFileCurNum >= l.logFileMaxNum {
		sort.Sort(byCreatedTime(files))
		nFiles := l.logFileCurNum - l.logFileMaxNum + l.logFilesToDel
		if nFiles > l.logFileCurNum {
			nFiles = l.logFileCurNum
		}
		for i := 0; i < nFiles; i++ {
			err := os.RemoveAll(l.logDir + files[i])
			if err == nil {
				l.logFileCurNum--
			} else {
				l.Errorf("RemoveAll failed: %v", err)
			}
		}
	}
}

func (l *Logger) getLogFilenames() ([]string, error) {
	var filenames []string
	f, err := os.Open(l.logDir)
	if err == nil {
		filenames, err = f.Readdirnames(0)
		f.Close()
		if err == nil {
			nFiles := len(filenames)
			for i := 0; i < nFiles; {
				if l.logFilenameRegex.MatchString(filenames[i]) {
					i++
				} else {
					nFiles--
					filenames[i] = filenames[nFiles]
					filenames = filenames[:nFiles]
				}
			}
		}
	}
	return filenames, err
}

func (l *Logger) log(logLevel int32, args []interface{}) {
	lowestLogLevel := atomic.LoadInt32(&l.logLevel)
	logDest := atomic.LoadUint32(&l.logDest)
	if lowestLogLevel > logLevel || logDest == kLogDestNone {
		return
	}

	buf := l.bufPool.getBuffer()

	t := time.Now()
	l.genLogPrefix(buf, logLevel, 3, t)
	fmt.Fprintln(buf, args...)
	output := buf.Bytes()
	if logDest&kLogDestFile != kLogDestNone {
		if l.flag&ControlFlagLogThrough != ControlFlagNone {
			for i := logLevel; i >= lowestLogLevel; i-- {
				l.loggers[i].log(t, output)
			}
		} else {
			l.loggers[logLevel].log(t, output)
		}
	}
	if logDest&kLogDestConsole != kLogDestNone {
		os.Stdout.Write(output)
	}

	l.bufPool.putBuffer(buf)
}

func (l *Logger) logf(logLevel int32, format string, args []interface{}) {
	lowestLogLevel := atomic.LoadInt32(&l.logLevel)
	logDest := atomic.LoadUint32(&l.logDest)
	if lowestLogLevel > logLevel || logDest == kLogDestNone {
		return
	}

	buf := l.bufPool.getBuffer()

	t := time.Now()
	l.genLogPrefix(buf, logLevel, 3, t)
	fmt.Fprintf(buf, format, args...)
	buf.WriteByte('\n')
	output := buf.Bytes()
	if logDest&kLogDestFile != kLogDestNone {
		if l.flag&ControlFlagLogThrough != ControlFlagNone {
			for i := logLevel; i >= lowestLogLevel; i-- {
				l.loggers[i].log(t, output)
			}
		} else {
			l.loggers[logLevel].log(t, output)
		}
	}
	if logDest&kLogDestConsole != kLogDestNone {
		os.Stdout.Write(output)
	}

	l.bufPool.putBuffer(buf)
}

func (l *Logger) genLogPrefix(buf *buffer, logLevel int32, skip int, t time.Time) {
	h, m, s := t.Clock()

	// time
	buf.tmp[0] = kLogLevelChar[logLevel]
	surplus := 0
	if l.flag&ControlFlagLogDate != ControlFlagNone {
		year, mon, day := t.Date()
		buf.nDigits(4, 1, year, '0')
		buf.nDigits(2, 5, int(mon), '0')
		buf.nDigits(2, 7, day, '0')
		buf.tmp[9] = ' '
		surplus = 9
	}
	buf.twoDigits(1+surplus, h)
	buf.tmp[3+surplus] = ':'
	buf.twoDigits(4+surplus, m)
	buf.tmp[6+surplus] = ':'
	buf.twoDigits(7+surplus, s)
	buf.Write(buf.tmp[:9+surplus])

	var pc uintptr
	var ok bool
	if l.flag&ControlFlagLogLineNum != ControlFlagNone {
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
	if l.flag&ControlFlagLogFuncName != ControlFlagNone {
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

type logger struct {
	file   *os.File
	day    int
	size   int64
	closed bool
	lock   sync.Mutex // Protects variables above

	// Variables that won't be changed at runtime go here
	level           int32
	symlinkFullPath string
	parent          *Logger
}

func (l *logger) close() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.file.Close()
	l.file = nil
	l.closed = true
}

func (l *logger) log(t time.Time, data []byte) {
	y, m, d := t.Date()

	l.lock.Lock()
	defer l.lock.Unlock()

	if !l.closed {
		if l.size >= l.parent.logFileMaxSize || l.day != d || l.file == nil {
			hour, min, sec := t.Clock()
			filename := fmt.Sprintf("%s%s.%d%02d%02d%02d%02d%02d%06d.log", l.parent.logPathPrefix, kLogLevelNames[l.level],
				y, m, d, hour, min, sec, t.Nanosecond()/1000)
			newFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				l.errLog(t, data, err)
				return
			}

			l.file.Close()
			l.file = newFile
			l.day = d
			l.size = 0

			err = os.RemoveAll(l.symlinkFullPath)
			if err != nil {
				l.errLog(t, nil, err)
			}
			err = os.Symlink(path.Base(filename), l.symlinkFullPath)
			if err != nil {
				l.errLog(t, nil, err)
			}

			if l.parent.logFilePurgeCh != nil {
				l.parent.logFilePurgeCh <- true
			}
		}

		n, _ := l.file.Write(data)
		l.size += int64(n)
	}
}

// errLog should only be called within (*logger).log()
func (l *logger) errLog(t time.Time, originLog []byte, err error) {
	buf := l.parent.bufPool.getBuffer()

	l.parent.genLogPrefix(buf, l.level, 2, t)
	buf.WriteString(err.Error())
	buf.WriteByte('\n')
	if l.file != nil {
		n, _ := l.file.Write(buf.Bytes())
		l.size += int64(n)
		if len(originLog) > 0 {
			n, _ = l.file.Write(originLog)
			l.size += int64(n)
		}
	} else {
		os.Stderr.Write(buf.Bytes())
		if len(originLog) > 0 {
			os.Stderr.Write(originLog)
		}
	}

	l.parent.bufPool.putBuffer(buf)
}

// sort files by created time embedded in the filename
type byCreatedTime []string

func (a byCreatedTime) Len() int {
	return len(a)
}

func (a byCreatedTime) Less(i, j int) bool {
	s1, s2 := a[i], a[j]
	l1, l2 := len(s1), len(s2)
	return s1[l1-24:l1-4] < s2[l2-24:l2-4]
}

func (a byCreatedTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// init is called after all the variable declarations in the package have evaluated their initializers,
// and those are evaluated only after all the imported packages have been initialized.
// Besides initializations that cannot be expressed as declarations, a common use of init functions is to verify
// or repair correctness of the program state before real execution begins.
func init() {
	tmpStrArr := strings.Split(path.Base(os.Args[0]), "\\") // for compatible with `go run` under Windows
	kProgramName = tmpStrArr[len(tmpStrArr)-1]

	var err error
	kHostname, err = os.Hostname()
	if err != nil {
		kHostname = "Unknown"
	}

	curUser, err := user.Current()
	if err == nil {
		tmpStrArr = strings.Split(curUser.Username, "\\") // for compatible with Windows
		kUsername = tmpStrArr[len(tmpStrArr)-1]
	} else {
		kUsername = "Unknown"
	}
}

const (
	kMaxInt64     = int64(^uint64(0) >> 1)
	kLogLevelChar = "TIWEPF"
)

var (
	kLogLevelNames = [kLogLevelCount]string{"TRACE", "INFO", "WARN", "ERROR", "PANIC", "FATAL"}

	kProgramName string
	kHostname    string
	kUsername    string

	defLoggerLock sync.Mutex // protects defLogger
	defLogger     *Logger
)
