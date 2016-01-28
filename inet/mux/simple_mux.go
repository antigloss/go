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

func NewSimpleMux(conn net.Conn, hdrSz int, hdrParser func(hdr []byte) SimpleMuxHeader) (*SimpleMux, error) {
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
	go mux.loop() // TODO how to stop the loop?

	return mux, nil
}

//-----------------------

type SimpleMux struct {
	conn       net.Conn
	hdrSz      int
	hdrParser  func(hdr []byte) SimpleMuxHeader
	nextSessID uint32
	sessLock   sync.RWMutex
	allSess    map[uint64]*Session
}

func (mux *SimpleMux) NewSession() (*Session, error) { // TODO determine if conn is available
	id := mux.getNextSessID()
	sess := &Session{
		id:   id,
		conn: mux.conn,
		buf:  make(chan *Packet, 10),
		err:  make(chan error, 1),
	}
	mux.sessLock.Lock()
	mux.allSess[id] = sess
	mux.sessLock.Unlock()
	return sess, nil
}

func (mux *SimpleMux) Close() error {
	// TODO
	return nil
}

func (mux *SimpleMux) loop() {
	hdr := make([]byte, mux.hdrSz)
	for {
		_, err := io.ReadFull(mux.conn, hdr)
		if err != nil {
			// TODO
			break
		}

		muxHdr := mux.hdrParser(hdr)
		bodyLen := muxHdr.BodyLen()
		if bodyLen < 0 {
			// TODO
			break
		}

		packet := &Packet{Header: muxHdr}
		if bodyLen > 0 {
			packet.Body = make([]byte, bodyLen)
			_, err := io.ReadFull(mux.conn, packet.Body)
			if err != nil {
				// TODO
				break
			}
		}

		mux.sessLock.RLock()
		sess := mux.allSess[muxHdr.SessionID()]
		if sess != nil {
			sess.buf <- packet
		} else {
			// TODO default
		}
		mux.sessLock.RUnlock()
	}
}

func (mux *SimpleMux) getNextSessID() uint64 {
	baseID := atomic.AddUint32(&(mux.nextSessID), 1)
	for baseID == 0 {
		baseID = atomic.AddUint32(&(mux.nextSessID), 1)
	}
	return ((uint64(time.Now().Unix()) << 32) | uint64(baseID))
}

type Session struct {
	id   uint64
	conn net.Conn
	// TODO Packet List
	buf chan *Packet
	err chan error
	// channel
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
	select {
	case packet = <-sess.buf:
	case err = <-sess.err:
	}
	return
}

func (sess *Session) Close() error {
	// TODO remove from SimpleMux
	sess.conn = nil
	close(sess.buf) // TODO what happens when some other goroutines are now writing to this channel?
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
