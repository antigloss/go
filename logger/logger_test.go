package logger

import (
	"testing"
)

func init() {
	Init(&Config{
		LogDir:          "./logs",
		LogFileMaxSize:  200,
		LogFileMaxNum:   500,
		LogFileNumToDel: 50,
		LogLevel:        LogLevelInfo,
		LogDest:         LogDestFile,
		Flag:            ControlFlagLogLineNum,
	})
}

func BenchmarkLogger(b *testing.B) {
	b.Run("benchmarkInfo", func(b *testing.B) {
		for i := 0; i != b.N; i++ {
			Info("Failed to find player! uid", 1234, "plid", 678942, "cmd=getplayer xxx", 102020101)
		}
	})
	b.Run("benchmarkInfof", func(b *testing.B) {
		for i := 0; i != b.N; i++ {
			Infof("Failed to find player! uid=%d plid=%d cmd=%s xxx=%d", 1234, 678942, "getplayer", 102020101)
		}
	})
}
