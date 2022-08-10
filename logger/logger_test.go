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
