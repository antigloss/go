/*
 *
 * sync - Synchronization facilities.
 * Copyright (C) 2019 Antigloss Huang (https://github.com/antigloss) All rights reserved.
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

package sync_test

import (
	"github.com/antigloss/go/sync"
	"time"
)

// This example shows the basic usage of Semaphore.
func ExampleNewSemaphore() {
	// Create a ready-to-use semaphore
	sema := sync.NewSemaphore(10)
	// Block to acquire resources from the semaphore
	semaResource := sema.Acquire()
	// Release resources acquired from a semaphore
	semaResource.Release()
	// Try to acquire resources from the semaphore, returns nil if resources cannot be acquired immediately
	semaResource = sema.TryAcquire()
	if semaResource != nil {
		// Release resources acquired from a semaphore
		semaResource.Release()
	}
	// Wait at most 2 seconds to acquire resources from the semaphore, returns nil if resources cannot be acquired after timeout
	semaResource = sema.TimedAcquire(2 * time.Second)
	if semaResource != nil {
		// Release resources acquired from a semaphore
		semaResource.Release()
	}
}
