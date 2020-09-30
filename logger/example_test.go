/*
	Copyright 2020 Antigloss

	This library is free software; you can redistribute it and/or
	modify it under the terms of the GNU Lesser General Public
	License as published by the Free Software Foundation; either
	version 3 of the License, or (at your option) any later version.

	This library is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
	Lesser General Public License for more details.

	You should have received a copy of the GNU Lesser General Public
	License along with this library; if not, write to the Free Software
	Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA
*/

package logger_test

import (
	"github.com/antigloss/go/logger"
)

// This example shows how to write logs with the global Logger object.
func ExampleInit_globalLoggerObject() {
	// Create the global Logger object
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
	logger.Info(123, 456.789, "abc", 432)
	logger.Infof("This is an %s", "example.")
}

// This example shows how to create multiple Logger objects and write logs with them.
func ExampleNew_multiLoggerObject() {
	lg1, err := logger.New(&logger.Config{
		LogDir:          "./logs1", // Better to associate different Logger object with different directory.
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
	defer lg1.Close() // Don't forget to close the Logger object when you've done with it.

	lg2, err := logger.New(&logger.Config{
		LogDir:          "./logs2", // Better to associate different Logger object with different directory.
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
	defer lg2.Close() // Don't forget to close the Logger object when you've done with it.

	lg1.Error(333, 444.55, "This", "is", "an", "example.")
	lg2.Warnf("This is %s %s %s", "yet", "another", "example.")
}
