// Author: https://github.com/antigloss

/*
Package concurrent provides some concurrent control utilities.
*/
package concurrent

import (
	"sync"
)

// Semaphore is a mimic of the POSIX semaphore based on sync.Cond. It could be used to limit the number of concurrent running goroutines.
// Basic example:
//	// Creates a ready-to-use semaphare
//	sema := concurrent.NewSemaphore(InitValue)
//	// Decrements the semaphore, blocks if the semaphore is less than 1
//	sema.Acquire()
//	// Increments the semaphore. If the semaphore’s value consequently becomes greater than zero,
//  // then another goroutine blocked in sema.Acquire() will be woken up and proceed to lock the semaphore.
//	sema.Release()
type Semaphore struct {
	cond       *sync.Cond
	nsems      int
	waitingNum int
}

// NewSemaphore creates a ready-to-use Semaphore.
//   value: Initial value of the Semaphore.
func NewSemaphore(value int) *Semaphore {
	return &Semaphore{cond: sync.NewCond(new(sync.Mutex)), nsems: value}
}

// Acquire decrements the semaphore, blocks if the semaphore is less than 1
func (this *Semaphore) Acquire() {
	this.cond.L.Lock()
	for {
		if this.nsems > 0 {
			this.nsems--
			break
		} else {
			this.waitingNum++
			this.cond.Wait()
			this.waitingNum--
		}
	}
	this.cond.L.Unlock()
}

// TryAcquire tries to decrement the semaphore. It returns true if the decrement is done, false otherwise.
func (this *Semaphore) TryAcquire() (ret bool) {
	this.cond.L.Lock()
	if this.nsems > 0 {
		this.nsems--
		ret = true
	}
	this.cond.L.Unlock()
	return
}

// Release increments the semaphore. If the semaphore’s value consequently becomes greater than zero,
// then another goroutine blocked in sema.Acquire() will be woken up and proceed to lock the semaphore.
func (this *Semaphore) Release() {
	this.cond.L.Lock()
	this.nsems++
	if this.waitingNum > 0 {
		this.cond.Signal()
	}
	this.cond.L.Unlock()
}
