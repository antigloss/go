package mux

import (
	"io"
	"net"
	"time"

	"sync"
	"sync/atomic"
)

func NewSimpleMux(conn net.Conn, hdrSz int, hdrParser func(hdr []byte) (sessID uint64, bodyLen int)) *SimpleMux {
	mux := &SimpleMux{
		conn:      conn,
		hdrSz:     hdrSz,
		hdrParser: hdrParser,
		allSess:   make(map[uint64]*Session),
	}
	go mux.loop() // TODO how to stop the loop?

	return mux
}

type SimpleMux struct {
	conn       net.Conn
	hdrSz      int
	hdrParser  func(hdr []byte) (sessID uint64, bodyLen int)
	nextSessID uint32
	sessLock   sync.RWMutex
	allSess    map[uint64]*Session
}

func (mux *SimpleMux) NewSession() *Session { // TODO determine if conn is available
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
	return sess
}

func (mux *SimpleMux) loop() {
	for {
		packet := &Packet{Header: make([]byte, mux.hdrSz)}
		_, err := io.ReadFull(mux.conn, packet.Header)
		if err != nil {
			// TODO
			break
		}

		id, bodyLen := mux.hdrParser(packet.Header)
		if bodyLen < 0 {
			// TODO
			break
		}

		if bodyLen > 0 {
			packet.Body = make([]byte, bodyLen)
			_, err := io.ReadFull(mux.conn, packet.Body)
			if err != nil {
				// TODO
				break
			}
		}

		mux.sessLock.RLock()
		sess := mux.allSess[id]
		sess.buf <- packet
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
	buf  chan *Packet
	err  chan error
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
	return nil
}

type Packet struct {
	Header []byte
	Body   []byte
}
