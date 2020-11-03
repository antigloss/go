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

/*
	Package sync provides extra synchronization facilities such as semaphore in addition to the standard sync package.
*/
package sync

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// SemaphoreResource holds resources acquired from a Semaphore object.
type SemaphoreResource struct {
	sema unsafe.Pointer
}

// Release releases resources acquired from a Semaphore object.
func (sr *SemaphoreResource) Release() {
	var new unsafe.Pointer
	old := atomic.SwapPointer(&sr.sema, new)
	if old != nil {
		(*Semaphore)(old).release()
	}
}

// Semaphore is a mimic of the POSIX semaphore based on sync.Cond. It could be used to limit the number of concurrent running goroutines.
// Basic example:
//	// Creates a ready-to-use semaphore
//	sema := concurrent.NewSemaphore(InitValue)
//	// Decrements the semaphore, blocks if value of the semaphore is less than 1
//	semaResource := sema.Acquire()
//	// Increments the semaphore. If the semaphore's value consequently becomes greater than zero,
//	// then another goroutine blocked in sema.Acquire() will be woken up and acquire the resources.
//	semaResource.Release()
type Semaphore struct {
	cond       *sync.Cond
	value      int
	waitingNum int
}

// NewSemaphore creates a ready-to-use Semaphore.
//   value: Initial value of the Semaphore.
func NewSemaphore(value int) *Semaphore {
	return &Semaphore{cond: sync.NewCond(new(sync.Mutex)), value: value}
}

// Acquire decrements the semaphore, blocks if value of the semaphore is less than 1.
func (s *Semaphore) Acquire() *SemaphoreResource {
	s.cond.L.Lock()
	for {
		if s.value > 0 {
			s.value--
			break
		} else {
			s.waitingNum++
			s.cond.Wait()
			s.waitingNum--
		}
	}
	s.cond.L.Unlock()

	return &SemaphoreResource{sema: unsafe.Pointer(s)}
}

// TryAcquire tries to decrement the semaphore. It returns nil if the decrement cannot be done immediately.
func (s *Semaphore) TryAcquire() (sr *SemaphoreResource) {
	s.cond.L.Lock()
	if s.value > 0 {
		s.value--
		sr = &SemaphoreResource{sema: unsafe.Pointer(s)}
	}
	s.cond.L.Unlock()
	return
}

/*
// TimedAcquire waits at most a given time to decrement the semaphore. It returns nil if the decrement cannot be done after the given timeout.
func (s *Semaphore) TimedAcquire(duration duration.Duration) (sr *SemaphoreResource) {
	sr = s.TryAcquire()
	if sr != nil {
		return
	}

	// TODO
	_ = duration

	return nil
}
*/

// release increments the semaphore. If the semaphore’s value consequently becomes greater than zero,
// then another goroutine blocked in sema.Acquire() will be woken up and acquire the resources.
func (s *Semaphore) release() {
	s.cond.L.Lock()
	s.value++
	if s.waitingNum > 0 {
		s.cond.Signal()
	}
	s.cond.L.Unlock()
}
