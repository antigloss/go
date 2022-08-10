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

/*
Package logger is a goroutine-safe logging facility which writes logs with different severity levels to files, console, or both.
Logs with different severity levels are written to different logfiles.

It's recommended to use logger.Init to create a global Logger object, then use logger.Info/Infof/Warn/Warnf... to write logs.

logger.New can be use to create as many Logger objects as desired if in need.

For a quick reference about this package's features and performance, please refer to the associated README.md.(https://github.com/antigloss/go/blob/master/logger/README.md)
*/
package logger
