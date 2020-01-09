// Package ftp_pool implements a ftp pool base on "github.com/jlaffaye/ftp"
package ftp_pool

import (
	"container/list"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
)

// FTPPool is an ftp pool.
type FTPPool struct {
	cond       *sync.Cond
	freeList   list.List
	curConnNum int // Current ftp connection number
	waitingNum int // Number of goroutines waiting for ftp connection currently
	// readonly variables
	maxCachedNum int    // Max pooled ftp connections
	connLimit    int    // Max ftp connections
	addr         string // ftp address
	user         string // ftp username
	passwd       string // ftp password
}

// NewFTPPool is the only way to get a new, ready-to-use FTPPool object.
//
//   addr: ftp address
//   user: ftp username
//   passwd: ftp password
//   maxCachedConn: Max pooled ftp connections
//   connLimit: Max ftp connections
//
// Example:
//
//	 ftpPool := NewFTPPool(Addr, User, Passwd, 10, 100)
//   ftpConn, _ := ftpPool.Get() // Gets an ftp connection from the pool, or creates a new one if the pool is empty
//   ftpPool.Put(ftpConn, false) // Puts an ftp connection back to the pool
func NewFTPPool(addr, user, passwd string, maxCachedConn, connLimit int) *FTPPool {
	pool := &FTPPool{
		cond:         sync.NewCond(new(sync.Mutex)),
		maxCachedNum: maxCachedConn,
		connLimit:    connLimit,
		addr:         addr,
		user:         user,
		passwd:       passwd,
	}
	pool.freeList.Init()
	go pool.keepalive()

	return pool
}

// Get gets an ftp connection from the pool. If no free connection is available and MaxConnLimit not reached,
// a new connection will be created. If MaxConnLimit is reached, Get blocks waiting to get/create a connection.
func (pool *FTPPool) Get() (conn *ftp.ServerConn, err error) {
	pool.cond.L.Lock()
	for {
		elem := pool.freeList.Front()
		if elem != nil { // Get a connection from the pool
			conn = elem.Value.(*ftpConnNode).conn
			pool.freeList.Remove(elem)
			break
		} else if pool.curConnNum < pool.connLimit { // Can still create more connection
			pool.curConnNum++ // Increase it anyway and decrease it later
			break
		} else { // waiting for permission to get/create a connection
			pool.waitingNum++
			pool.cond.Wait()
			pool.waitingNum--
		}
	}
	pool.cond.L.Unlock()

	if conn != nil {
		return
	}

	for i := 0; i < 2; i++ { // Try again one more time if failed
		conn, err = ftp.DialTimeout(pool.addr, 5*time.Second)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		err = conn.Login(pool.user, pool.passwd)
		if err == nil {
			break
		}

		conn.Quit()
		conn = nil
	}
	if conn == nil {
		pool.cond.L.Lock()
		pool.curConnNum--
		if pool.waitingNum > 0 {
			pool.cond.Signal()
		}
		pool.cond.L.Unlock()
	}

	return
}

// Put returns an ftp connection to the pool. If MaxCachedConn had been reached, the connection will be discarded.
//
//   conn: ftp connection to be returned
//   forceFree: the connection will be discarded anyway if true is passed
func (pool *FTPPool) Put(conn *ftp.ServerConn, forceFree bool) {
	pool.cond.L.Lock()
	if !forceFree && pool.freeList.Len() < pool.maxCachedNum {
		pool.freeList.PushBack(&ftpConnNode{conn, time.Now()})
	} else {
		forceFree = true
		pool.curConnNum--
	}
	if pool.waitingNum > 0 {
		pool.cond.Signal()
	}
	pool.cond.L.Unlock()

	if forceFree {
		conn.Quit()
	}
}

func (pool *FTPPool) Addr() string {
	return pool.addr
}

func (pool *FTPPool) User() string {
	return pool.user
}

func (pool *FTPPool) Password() string {
	return pool.passwd
}

func (pool *FTPPool) MaxCachedConnNum() int {
	return pool.maxCachedNum
}

type ftpConnNode struct {
	conn        *ftp.ServerConn
	lastActTime time.Time
}

// Keepalive with the ftp server
func (pool *FTPPool) keepalive() {
	for {
		time.Sleep(5 * time.Second)
		tNow := time.Now()
		pool.cond.L.Lock()
		for nextElem := pool.freeList.Front(); nextElem != nil; {
			node := nextElem.Value.(*ftpConnNode)
			if tNow.Sub(node.lastActTime).Seconds() < 10 {
				break
			}

			curElem := nextElem
			nextElem = nextElem.Next()
			pool.freeList.Remove(curElem)
			go func(conn *ftp.ServerConn, pool *FTPPool) {
				err := conn.NoOp()
				pool.Put(conn, err != nil)
			}(node.conn, pool)
		}
		pool.cond.L.Unlock()
	}
}
