// Author: https://github.com/antigloss

// Package mux (short for connection multiplexer) is a multiplexing package for Golang.
//
// In some rare cases, we can only open a few connections to a remote server,
// but we want to program like we can open unlimited connections.
// Should you encounter this rare cases, then this package is exactly what you need.
package mux

import (
	"fmt"
	"io"
	"net"
	"time"

	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	kSimpleMuxMinHeaderSz = 9
	kSimpleMuxMaxHeaderSz = 1024
)

// SimpleMuxHeader
type SimpleMuxHeader interface {
	SessionID() uint64
	BodyLen() int64
}

// Packet holds the protocol data receive from the remote server.
type Packet struct {
	Header SimpleMuxHeader // protocol header
	Body   []byte          // protocol body
}

// NewSimpleMux is the only way to get a new, ready-to-use SimpleMux.
//
//   conn: Connection to the remote server. Once a connection has been assigned to a SimpleMux,
//         you should never use it elsewhere, otherwise it might cause the SimpleMux to malfunction.
//   hdrSz: Size (in bytes) of protocol header for communicating with the remote server.
//   hdrParser: Function to parser the header. Returns (hdr, nil) on success, or (nil, err) on error.
//   defHandler: Handler function for handling packets without an associated session. Could be nil.
func NewSimpleMux(conn net.Conn, hdrSz int,
	hdrParser func(hdr []byte) (SimpleMuxHeader, error),
	defHandler func(*Packet)) (*SimpleMux, error) {
	if hdrSz < kSimpleMuxMinHeaderSz || hdrSz > kSimpleMuxMaxHeaderSz {
		return nil, fmt.Errorf("`hdrSz` should be [%d, %d].", kSimpleMuxMinHeaderSz, kSimpleMuxMaxHeaderSz)
	}
	if hdrParser == nil {
		return nil, fmt.Errorf("`hdrParser` must not be nil!")
	}

	mux := &SimpleMux{
		conn:      conn,
		hdrSz:     hdrSz,
		hdrParser: hdrParser,
		allSess:   make(map[uint64]*Session),
	}
	if defHandler != nil {
		mux.defHandler = defHandler
		mux.defPacketQ = newPacketQueue()
		mux.defNotiChnl = make(chan bool, 1)
		mux.defQuitChnl = make(chan bool, 1)
		go mux.procNonSessionPackets()
	}
	go mux.loop()

	return mux, nil
}

// SimpleMux is a connection multiplexer. It is very useful when under the following constraints:
//
//   1. Can only open a few connections (probably only 1 connection) to a remote server,
//       but want to program like there can be unlimited connections.
//   2. The remote server has its own protocol format which could not be changed.
//   3. Fortunately, we can set 8 bytes of information to the protocol header which
//       will remain the same in the server's response.
//
// All methods of SimpleMux are goroutine-safe.
//
// Seek to simple_mux_test.go for detailed usage.
type SimpleMux struct {
	closed      bool // Determine if this `SimpleMux` has been closed
	conn        net.Conn
	hdrSz       int
	hdrParser   func(hdr []byte) (SimpleMuxHeader, error)
	nextSessID  uint32
	sessLock    sync.RWMutex
	allSess     map[uint64]*Session
	defHandler  func(*Packet) // defHandler will be invoke if session not found
	defPacketQ  *packetQueue  // Non-session-packets will be pushed into it for defHandler
	defNotiChnl chan bool     // Notify defHandler that there is incoming non-session-packet
	defQuitChnl chan bool     // Notify defHandler to quit
}

// NewSession is used to create a new session.
// You can create as many sessions as you want.
// All sessions are base on the single connecions of the SimpleMux,
// but they act like they are separate connections.
//
//   Note: Methods of Session are not goroutine-safe.
//         One session is intended to be used within one goroutine.
func (mux *SimpleMux) NewSession() (sess *Session, err error) {
	id := mux.getNextSessID()
	sess = &Session{
		id:         id,
		mux:        mux,
		packets:    newPacketQueue(),
		packetNoti: make(chan bool, 1),
		err:        make(chan error, 1),
	}
	mux.sessLock.Lock()
	if !mux.closed {
		mux.allSess[id] = sess
	} else {
		sess = nil
		err = kSimpleMuxClosed
	}
	mux.sessLock.Unlock()
	return
}

// Close is used to close the SimpleMux (including its underlying connection)
// and all sessions.
//
//   Note: After finish using a SimpleMux, Close must be called to release resources.
func (mux *SimpleMux) Close() {
	mux.close(kSimpleMuxClosed)
}

func (mux *SimpleMux) loop() {
	var muxHdr SimpleMuxHeader
	var err error
	hdr := make([]byte, mux.hdrSz)
	for {
		_, err = io.ReadFull(mux.conn, hdr)
		if err != nil {
			break
		}

		muxHdr, err = mux.hdrParser(hdr)
		if err != nil {
			break
		}

		packet := &Packet{Header: muxHdr}
		bodyLen := muxHdr.BodyLen()
		if bodyLen > 0 {
			packet.Body = make([]byte, bodyLen)
			_, err = io.ReadFull(mux.conn, packet.Body)
			if err != nil {
				break
			}
		}

		mux.sessLock.RLock()
		if mux.closed {
			break
		}
		sess := mux.allSess[muxHdr.SessionID()]
		mux.sessLock.RUnlock()
		if sess != nil {
			sess.packets.push(packet)
			asyncNotify(sess.packetNoti)
		} else {
			if mux.defHandler != nil {
				mux.defPacketQ.push(packet)
				asyncNotify(mux.defNotiChnl)
			}
		}
	}

	mux.close(err)
}

func (mux *SimpleMux) procNonSessionPackets() {
	var closed bool
	var packet *Packet
	for {
		packet = mux.defPacketQ.pop()
		if packet != nil {
			mux.defHandler(packet)
		} else {
			select {
			case <-mux.defNotiChnl:
			case closed = <-mux.defQuitChnl:
			}
			if closed {
				break
			}
		}
	}
}

func (mux *SimpleMux) close(err error) {
	mux.sessLock.Lock()
	if !mux.closed {
		// Notify all sessions that error occurs
		for _, sess := range mux.allSess {
			asyncNotifyError(sess.err, err)
		}
		if mux.defHandler != nil {
			mux.defQuitChnl <- true
		}
		mux.allSess = nil
		mux.closed = true
		mux.conn.Close()
	}
	mux.sessLock.Unlock()
}

func (mux *SimpleMux) closeSession(sessID uint64) {
	mux.sessLock.Lock()
	if !mux.closed {
		delete(mux.allSess, sessID)
	}
	mux.sessLock.Unlock()
}

func (mux *SimpleMux) getNextSessID() uint64 {
	baseID := atomic.AddUint32(&(mux.nextSessID), 1)
	for baseID == 0 {
		baseID = atomic.AddUint32(&(mux.nextSessID), 1)
	}
	return ((uint64(time.Now().Unix()) << 32) | uint64(baseID))
}

func asyncNotify(ch chan bool) {
	select {
	case ch <- true:
	default:
	}
}

func asyncNotifyError(ch chan error, err error) {
	select {
	case ch <- err:
	default:
	}
}

var kSimpleMuxClosed = fmt.Errorf("This SimpleMux object has already been closed.")

//------------------------------------------------------------------
// Session
//------------------------------------------------------------------

// Session is created from a SimpleMux. You can create as many sessions as you want.
// All sessions are base on the single connecions of the SimpleMux,
// but they act like they are separate connections.
//
// Session supports bi-directional communication and server-side push.
//
//   Note: Methods of Session are not goroutine-safe.
//         One session is intended to be used within one goroutine.
type Session struct {
	id         uint64
	mux        *SimpleMux
	packets    *packetQueue
	rdTimeout  time.Duration
	packetNoti chan bool
	err        chan error
}

// ID returns the ID of this session.
func (sess *Session) ID() uint64 {
	return sess.id
}

// Send is used to write to the session.
// For some good reasons, Send dosen't support timeout.
func (sess *Session) Send(b []byte) (int, error) {
	if sess.mux != nil {
		return sess.mux.conn.Write(b)
	}
	return 0, kSessionClosed
}

// Recv reads data from the session.
// Returns net.Error at timeout, use err.(net.Error).Timeout()
// to determine if timeout occurs.
func (sess *Session) Recv() (packet *Packet, err error) {
	for {
		packet = sess.packets.pop()
		if packet != nil {
			return
		}

		var flag bool
		var timeout <-chan time.Time
		if sess.rdTimeout > 0 {
			timeout = time.After(sess.rdTimeout)
		}
		select {
		case flag = <-sess.packetNoti:
		case err = <-sess.err:
		case <-timeout:
		}

		if flag {
			continue
		}

		if err == nil {
			err = kSessionRdTimeout
		}
		return
	}
}

// SetRecvTimeout sets timeout to the session.
// After calling this method, all subsequent calls to Recv() will
// timeout after the specified `timeout`.
//
// Should you want to cancel the timeout setting, just call SetRecvTimeout(0)
//
//   Example:
//       sess.SetRecvTimeout(5 * time.Millisecond)
func (sess *Session) SetRecvTimeout(timeout time.Duration) {
	sess.rdTimeout = timeout
}

// Close is used to close the session.
// After finish using a Session, Close() must be called to release resources.
func (sess *Session) Close() {
	if sess.mux != nil {
		sess.mux.closeSession(sess.ID())
		sess.mux = nil
	}
}

type timeoutError string

func (e timeoutError) Error() string {
	return string(e)
}

func (e timeoutError) Timeout() bool {
	return true
}

func (e timeoutError) Temporary() bool {
	return true
}

var kSessionClosed = fmt.Errorf("This session has already been closed.")
var kSessionRdTimeout = timeoutError("This session has already been closed.")

//--------------------------------------------------------
// packetQueue
//--------------------------------------------------------

func newPacketQueue() *packetQueue {
	var pq packetQueue
	pq.head = unsafe.Pointer(&pq.dummy)
	pq.tail = pq.head
	return &pq
}

// packetQueue is a goroutine-safe Queue implementation.
// The overall performance of packetQueue is much better than List+Mutex(standard package).
type packetQueue struct {
	head  unsafe.Pointer
	tail  unsafe.Pointer
	dummy packetNode
}

func (pq *packetQueue) pop() *Packet {
	for {
		h := atomic.LoadPointer(&pq.head)
		rh := (*packetNode)(h)
		n := (*packetNode)(atomic.LoadPointer(&rh.next))
		if n != nil {
			if atomic.CompareAndSwapPointer(&pq.head, h, rh.next) {
				return n.val
			} else {
				continue
			}
		} else {
			return nil
		}
	}
}

func (pq *packetQueue) push(val *Packet) {
	node := unsafe.Pointer(&packetNode{val: val})
	for {
		t := atomic.LoadPointer(&pq.tail)
		rt := (*packetNode)(t)
		if atomic.CompareAndSwapPointer(&rt.next, nil, node) {
			// It'll be a dead loop if atomic.StorePointer() is used.
			// Don't know why.
			// atomic.StorePointer(&lfq.tail, node)
			atomic.CompareAndSwapPointer(&pq.tail, t, node)
			return
		} else {
			continue
		}
	}
}

type packetNode struct {
	val  *Packet
	next unsafe.Pointer
}
