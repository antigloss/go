package mux

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"runtime"
	"sync"
	"testing"
	"time"
)

var wg sync.WaitGroup
var out *testing.T

const (
	kGoroutineNum = 10
	kSendTimes    = 10000
	kLoopTimes    = kGoroutineNum * kSendTimes
)

func TestSimpleMux(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	out = t

	ln, _ := net.Listen("tcp", ":6666")
	go func() {
		var pkg bytes.Buffer
		binary.Write(&pkg, binary.BigEndian, Header{})

		buf := make([]byte, 16)
		conn, _ := ln.Accept()
		for i := 0; ; i++ {
			io.ReadFull(conn, buf)
			conn.Write(buf)
			if i%kGoroutineNum == 0 {
				conn.Write(pkg.Bytes())
			}
		}
	}()

	conn, _ := net.Dial("tcp", "localhost:6666")
	simpleMux, _ := NewSimpleMux(conn, 12, hdrParser, defHandler)
	wg.Add(kGoroutineNum)
	for i := 0; i != kGoroutineNum; i++ {
		go test(simpleMux, i)
	}
	wg.Wait()

	if gHdlrCallTimes != kSendTimes {
		out.Errorf("Handler called times should be %d! ct=%d", kSendTimes, gHdlrCallTimes)
	}

	sess, _ := simpleMux.NewSession()
	sess.SetRecvTimeout(1 * time.Second)
	sess.Send(make([]byte, 16))
	tSav := time.Now()
	_, err := sess.Recv()
	if err == nil {
		out.Error("Should be timeout!")
	}
	if _, ok := err.(timeoutError); !ok {
		out.Error("Should be timeoutError!")
	}
	sess.Recv() // Recv and timeout again
	tdiff := time.Now().Sub(tSav).Seconds()
	if int(tdiff) != 2 {
		out.Errorf("Timeout should be 1 seconds! %v", tdiff)
	}

	sess.Close()
	simpleMux.Close()
}

func test(simpleMux *SimpleMux, n int) {
	sess, _ := simpleMux.NewSession()
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, Header{Len: 4, ID: sess.ID()})
	for i := 0; i != kSendTimes; i++ {
		buf.Truncate(12)
		binary.Write(&buf, binary.BigEndian, int32(i))
		sess.Send(buf.Bytes())

		packet, _ := sess.Recv()
		hdr := packet.Header.(*Header)
		if hdr.SessionID() != sess.ID() {
			out.Errorf("Session ID mismatch! %d %d", hdr.SessionID(), sess.ID())
		}

		var nn int32
		r := bytes.NewReader(packet.Body)
		binary.Read(r, binary.BigEndian, &nn)
		if int(nn) != i {
			out.Errorf("Out of order! %d %d", i, nn)
		}
	}
	sess.Close()
	wg.Done()
}

type Header struct {
	Len int32
	ID  uint64
}

func (h *Header) BodyLen() int64 {
	return int64(h.Len)
}

func (h *Header) SessionID() uint64 {
	return h.ID
}

func hdrParser(hdr []byte) (SimpleMuxHeader, error) {
	r := bytes.NewReader(hdr)

	var h Header
	binary.Read(r, binary.BigEndian, &h)

	return &h, nil
}

var gHdlrCallTimes int

func defHandler(packet *Packet) {
	gHdlrCallTimes++
}
