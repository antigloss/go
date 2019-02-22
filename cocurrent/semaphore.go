package cocurrent

import (
	"sync"
)

func NewSemaphore(nsems int) *Semaphore {
	return &Semaphore{cond: sync.NewCond(new(sync.Mutex)), nsems: nsems}
}

type Semaphore struct {
	cond       *sync.Cond
	nsems      int
	waitingNum int
}

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

func (this *Semaphore) Release() {
	this.cond.L.Lock()
	this.nsems++
	if this.waitingNum > 0 {
		this.cond.Signal()
	}
	this.cond.L.Unlock()
}
