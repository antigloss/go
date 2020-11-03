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
	"container/list"
	"sync"
	"sync/atomic"
	"time"
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

// Semaphore is a mimic of the POSIX semaphore based on channel and sync.Mutex. It could be used to limit the number of concurrent running goroutines.
// Basic example:
//	// Creates a ready-to-use semaphore
//	sema := concurrent.NewSemaphore(InitValue)
//	// Decrements the semaphore, blocks if value of the semaphore is less than 1
//	semaResource := sema.Acquire()
//	// Increments the semaphore. If the semaphore's value consequently becomes greater than zero,
//	// then another goroutine blocked in sema.Acquire() will be woken up and acquire the resources.
//	semaResource.Release()
type Semaphore struct {
	lock    sync.Mutex
	value   int
	waiters list.List
}

// NewSemaphore creates a ready-to-use Semaphore.
//   value: Initial value of the Semaphore.
func NewSemaphore(value int) *Semaphore {
	return &Semaphore{value: value}
}

// Acquire decrements the semaphore, blocks if value of the semaphore is less than 1.
func (s *Semaphore) Acquire() *SemaphoreResource {
	s.lock.Lock()
	if s.value > 0 {
		s.value--
		s.lock.Unlock()
		return &SemaphoreResource{sema: unsafe.Pointer(s)}
	}

	ready := make(chan bool)
	s.waiters.PushBack(ready)
	s.lock.Unlock()

	<-ready
	return &SemaphoreResource{sema: unsafe.Pointer(s)}
}

// TryAcquire tries to decrement the semaphore. It returns nil if the decrement cannot be done immediately.
func (s *Semaphore) TryAcquire() (sr *SemaphoreResource) {
	s.lock.Lock()
	if s.value > 0 {
		s.value--
		sr = &SemaphoreResource{sema: unsafe.Pointer(s)}
	}
	s.lock.Unlock()
	return
}

// TimedAcquire waits at most a given time to decrement the semaphore. It returns nil if the decrement cannot be done after the given timeout.
func (s *Semaphore) TimedAcquire(duration time.Duration) (sr *SemaphoreResource) {
	s.lock.Lock()
	if s.value > 0 {
		s.value--
		s.lock.Unlock()
		sr = &SemaphoreResource{sema: unsafe.Pointer(s)}
		return
	}

	ready := make(chan bool)
	elem := s.waiters.PushBack(ready)
	s.lock.Unlock()

	timer := time.NewTimer(duration)
	select {
	case <-timer.C:
		s.lock.Lock()
		select {
		case <-ready:
			sr = &SemaphoreResource{sema: unsafe.Pointer(s)}
		default:
			s.waiters.Remove(elem)
		}
		s.lock.Unlock()
	case <-ready:
		timer.Stop()
		sr = &SemaphoreResource{sema: unsafe.Pointer(s)}
	}

	return
}

// release increments the semaphore. If the semaphore’s value consequently becomes greater than zero,
// then another goroutine blocked in sema.Acquire() will be woken up and acquire the resources.
func (s *Semaphore) release() {
	s.lock.Lock()
	waiter := s.waiters.Front()
	if waiter == nil {
		s.value++
	} else {
		s.waiters.Remove(waiter)
		close(waiter.Value.(chan bool))
	}
	s.lock.Unlock()
}
