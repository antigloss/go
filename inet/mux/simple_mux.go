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

type SimpleMuxHeader interface {
	SessionID() uint64
	BodyLen() int64
}

type Packet struct {
	Header SimpleMuxHeader
	Body   []byte
}

func NewSimpleMux(conn net.Conn, hdrSz int, hdrParser func(hdr []byte) (SimpleMuxHeader, error)) (*SimpleMux, error) {
	if hdrSz < kSimpleMuxMinHeaderSz || hdrSz > kSimpleMuxMaxHeaderSz {
		return nil, fmt.Errorf("`hdrSz` should be [%d, %d].", kSimpleMuxMinHeaderSz, kSimpleMuxMaxHeaderSz)
	}
	if hdrParser == nil {
		return nil, fmt.Errorf("`hdrParser` must not be nil!")
	}

	mux := &SimpleMux{
		conn:        conn,
		hdrSz:       hdrSz,
		hdrParser:   hdrParser,
		allSess:     make(map[uint64]*Session),
		defQuitChnl: make(chan bool, 1),
	}
	go mux.loop()

	return mux, nil
}

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

// SetDefaultHandler
func (mux *SimpleMux) SetDefaultHandler(handler func(*Packet)) {
	if handler != nil && mux.defHandler == nil {
		mux.defPacketQ = newPacketQueue()
		mux.defNotiChnl = make(chan bool, 1)
		go mux.procNonSessionPackets()
	}
	mux.defHandler = handler
}

func (mux *SimpleMux) NewSession() (sess *Session, err error) {
	id := mux.getNextSessID()
	sess = &Session{
		id:         id,
		conn:       mux.conn,
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
			if mux.defHandler != nil {
				mux.defHandler(packet)
			}
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
		mux.defQuitChnl <- true
		mux.allSess = nil
		mux.closed = true
		mux.conn.Close()
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

type Session struct {
	id         uint64
	conn       net.Conn
	packets    *packetQueue
	packetNoti chan bool
	err        chan error
}

func (sess *Session) ID() uint64 {
	return sess.id
}

// TODO
// Send writes data to the session.
func (sess *Session) Send(b []byte) (n int, err error) {
	return sess.conn.Write(b)
}

// TODO
// Recv reads data from the session.
func (sess *Session) Recv() (packet *Packet, err error) {
	var flag bool
	for {
		packet = sess.packets.pop()
		if packet != nil {
			return
		}

		select {
		case flag = <-sess.packetNoti:
		case err = <-sess.err:
		}
		// TODO Check periodically if packet available or session already closed
		// in case Recv() is called simultaneously from multiple goroutines

		if flag {
			continue
		}

		return
	}
}

func (sess *Session) Close() error {
	// TODO remove from SimpleMux
	sess.conn = nil
	close(sess.packetNoti) // TODO what happens when some other goroutines are now writing to this channel?
	close(sess.err)
	return nil // TODO return error
}

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
