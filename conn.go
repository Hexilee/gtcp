package gtcp

import (
	"net"
	"os"
	"io"
	"time"
	"syscall"
	"encoding/binary"
	"context"
	"errors"
	"bytes"
	"fmt"
	"bufio"
)

const (
	headerLen = 4
)

type TCPConnInterface interface {
	net.Conn
	ReadData() []byte
	CloseRead() error
	CloseWrite() error
	File() (f *os.File, err error)
	ReadFrom(r io.Reader) (int64, error)
	SetKeepAlive(keepalive bool) error
	SetKeepAlivePeriod(d time.Duration) error
	SetLinger(sec int) error
	SetNoDelay(noDelay bool) error
	SetReadBuffer(bytes int) error
	SetWriteBuffer(bytes int) error
	SyscallConn() (syscall.RawConn, error)
	InstallCtx(ctx context.Context)
	GetDataChan() <-chan []byte
	GetInfoChan() <-chan string
	GetErrChan() <-chan error
	Cancel()
	Scan()
}

type TCPTypeInterface interface {
	TCPConnInterface
	OnConnect() error
	OnMessage(data []byte) error
	OnError(err error) error
	OnClose() error
	InstallTCPConn(conn *TCPConn)
}

func NewTCPConn(conn *net.TCPConn) *TCPConn {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &TCPConn{
		data:    make(chan []byte),
		info:    make(chan string),
		error:   make(chan error),
		TCPConn: conn,
		Context: ctx,
		cancel:  cancelFunc,
	}
}

type TCPConn struct {
	data    chan []byte
	info    chan string
	error   chan error
	*net.TCPConn
	Context context.Context
	cancel  context.CancelFunc
}

func (t *TCPConn) Write(data []byte) (int, error) {
	length := uint32(len(data))
	head := make([]byte, 4)
	binary.LittleEndian.PutUint32(head, length)
	buf := bytes.NewBuffer(head)
	buf.Write(data)
	n, err := t.TCPConn.Write(buf.Bytes())
	if err != nil {
		return n, err
	}
	return n - headerLen, nil
}

func (t *TCPConn) Cancel() {
	t.cancel()
}

func (t *TCPConn) InstallCtx(ctx context.Context) {
	newCtx, cancelFunc := context.WithCancel(ctx)
	t.Context = newCtx
	t.cancel = cancelFunc
}

func (t *TCPConn) split(data []byte, atEOF bool) (adv int, token []byte, err error) {
	length := len(data)
	if length < headerLen {
		return 0, nil, nil
	}
	if length > 1048576 { //1024*1024=1048576
		t.Cancel()
		return 0, nil, errors.New(fmt.Sprintf("Read Error. Addr: %s; Err: too large data!", t.RemoteAddr().String()))
	}
	var lhead uint32
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &lhead)

	tail := length - headerLen
	if lhead > 1048576 {
		t.Cancel()
		return 0, nil, errors.New(fmt.Sprintf("Read Error. Addr: %s; Err: too large data!", t.RemoteAddr().String()))
	}
	if uint32(tail) < lhead {
		return 0, nil, nil
	}
	adv = headerLen + int(lhead)
	token = data[:adv]
	return adv, token, nil
}

func (t *TCPConn) Scan() {
	scanner := bufio.NewScanner(t)
	scanner.Split(t.split)

Circle:
	for scanner.Scan() {
		select {
		case <-t.Context.Done():
			t.info <- fmt.Sprintf("Conn Done. Addr: %s", t.RemoteAddr().String())
			break Circle
		default:
		}

		data := scanner.Bytes()
		msg := make([]byte, len(data))
		copy(msg, data)
		t.data <- msg
	}
	if err := scanner.Err(); err != nil {
		t.error <- err
	}
}

func (t *TCPConn) ReadData() []byte {
	select {
	case data := <-t.data:
		return data[headerLen: ]
	}
}

func (t *TCPConn) ReadString() string {
	select {
	case data := <-t.data:
		return string(data[headerLen: ])
	}
}

func (t *TCPConn) GetDataChan() <-chan []byte {
	return t.data
}

func (t *TCPConn) GetInfoChan() <-chan string {
	return t.info
}

func (t *TCPConn) GetErrChan() <-chan error {
	return t.error
}

type TCPType struct {
	*TCPConn
}

func (t *TCPType) Close() error {
	t.OnClose()
	t.Cancel()
	err := t.TCPConn.Close()
	return err
}

func (t *TCPType) InstallTCPConn(conn *TCPConn) {
	t.TCPConn = conn
}

func (t *TCPType) OnConnect() error {
	return nil
}
func (t *TCPType) OnMessage(data []byte) error {
	return nil
}
func (t *TCPType) OnError(err error) error {
	return nil
}
func (t *TCPType) OnClose() error {
	return nil
}

func (t *TCPType) Scan() {
	defer t.Close()
	err := t.OnConnect()
	for err != nil {
		err = t.OnError(err)
	}

	scanner := bufio.NewScanner(t)
	scanner.Split(t.split)

Circle:
	for scanner.Scan() {
		select {
		case <-t.Context.Done():
			t.info <- fmt.Sprintf("Conn Done. Addr: %s", t.RemoteAddr().String())
			break Circle
		default:
		}

		data := scanner.Bytes()
		msg := make([]byte, len(data))
		copy(msg, data)
		err := t.OnMessage(msg)
		for err != nil {
			err = t.OnError(err)
			break Circle
		}
	}
	err = scanner.Err()
	for err != nil {
		err = t.OnError(err)
	}
}
