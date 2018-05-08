package ftp_pool

import (
	"container/list"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
)

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

type FTPPool struct {
	cond       *sync.Cond
	freeList   list.List
	curConnNum int // 当前FTP连接数
	waitingNum int // 当前正在等待FTP连接的goroutine数量
	// readonly variables
	maxCachedNum int // 最大缓存FTP连接数
	connLimit    int // 最大FTP连接数
	addr         string
	user         string
	passwd       string
}

func (pool *FTPPool) Get() (conn *ftp.ServerConn, err error) {
	pool.cond.L.Lock()
	for {
		elem := pool.freeList.Front()
		if elem != nil { // 有可用的连接，直接使用
			conn = elem.Value.(*ftpConnNode).conn
			pool.freeList.Remove(elem)
			break
		} else if pool.curConnNum < pool.connLimit { // 还能建立多余的连接
			pool.curConnNum++ // 先把当前连接数加1，如何后面连接建立不成功，再减1
			break
		} else { // 无法建立更多连接，等待有空闲连接后重试
			pool.waitingNum++
			pool.cond.Wait()
			pool.waitingNum--
		}
	}
	pool.cond.L.Unlock()

	if conn != nil {
		return
	}

	for i := 0; i < 2; i++ { // 如果连不上FTP，则尝试重连一次
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
		pool.curConnNum-- // 新建连接失败，把当前连接数减1
		if pool.waitingNum > 0 {
			pool.cond.Signal()
		}
		pool.cond.L.Unlock()
	}

	return
}

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
		conn.Quit() // Quit可能是比较耗时的操作，所以放到锁外面执行
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

// 定期和FTP服务器心跳，避免被关闭
func (pool *FTPPool) keepalive() {
	for {
		time.Sleep(5 * time.Second) // 每隔5秒扫描一次
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
