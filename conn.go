package gtcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

const (
	headerLen = 4
)

type TCPConnInterface interface {
	net.Conn
	ReadData() []byte
	ReadString() string
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
	InstallCtx(ctx context.Context)
	GetDataChan() <-chan []byte
	GetInfoChan() <-chan string
	GetErrChan() <-chan error
	Scan()
	Start()
	StartWithCtx(ctx context.Context)
	Done() <-chan struct{}
	IsDone() bool
	ReInstallNetConn(conn *net.TCPConn)
	CloseOnce()
	//Clear()
}

type TCPBox interface {
	TCPConnInterface
	InstallTCPConn(conn *TCPConn)
	//ReInstallTCPConn(conn *TCPConn)
}

func NewTCPConn(conn *net.TCPConn) *TCPConn {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &TCPConn{
		data:    make(chan []byte),
		info:    make(chan string),
		error:   make(chan error),
		TCPConn: conn,
		mu:      new(sync.RWMutex),
		//OnceClose: new(sync.Once),
		Context: ctx,
		cancel:  cancelFunc,
	}
}

func GetTCPConn(conn *net.TCPConn) (tcpConn *TCPConn) {
	tcpConn, ok := GetConnFromPool(conn)
	if !ok {
		tcpConn = NewTCPConn(conn)
	}
	return
}

type TCPConn struct {
	data  chan []byte
	info  chan string
	error chan error
	mu    *sync.RWMutex
	//OnceClose  *sync.Once
	//isScanning uint32
	*net.TCPConn
	Context context.Context
	cancel  context.CancelFunc
}

func (t *TCPConn) Start() {
	go t.Scan()
}

func (t *TCPConn) StartWithCtx(ctx context.Context) {
	t.InstallCtx(ctx)
	go t.Scan()
}

//func (t *TCPConn) Clear() {
//	//t.mu.Lock()
//	//defer t.mu.Unlock()
//	//t.OnceClose = new(sync.Once)
//	//atomic.StoreUint32(&t.isScanning, 0)
//	t.InstallCtx(context.Background())
//}

func (t *TCPConn) CloseOnce() {
	err := t.TCPConn.Close()
	if err != nil {
		t.error <- err
	}
	//SendConnToPool(t)
}

func (t *TCPConn) ReInstallNetConn(conn *net.TCPConn) {
	//select {
	//case <-t.Done():
	//default:
	//	t.Clear()
	//}
	if !t.IsDone() {
		panic(errors.New("Unclosed TCPConn Cannot reinstall conn!"))
	}
	t.TCPConn = conn
}

func (t *TCPConn) Done() <- chan struct{} {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Context.Done()
}

func (t *TCPConn) IsDone() bool {
	done := t.Done()
	select {
	case <- done:
		return true
	default:
		return false
	}
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

func (t *TCPConn) Close() error {
	t.cancel()
	return nil
}

func (t *TCPConn) InstallCtx(ctx context.Context) {
	if t.cancel != nil {
		t.cancel()
	}
	t.mu.Lock()
	t.Context, t.cancel = context.WithCancel(ctx)
	t.mu.Unlock()
}

func (t *TCPConn) split(data []byte, atEOF bool) (adv int, token []byte, err error) {
	length := len(data)
	if length < headerLen {
		return 0, nil, nil
	}
	if length > 1048576 { //1024*1024=1048576
		t.Close()
		return 0, nil, errors.New(fmt.Sprintf("Read Error. Addr: %s; Err: too large data!", t.RemoteAddr().String()))
	}
	var lhead uint32
	buf := bytes.NewReader(data)
	binary.Read(buf, binary.LittleEndian, &lhead)

	tail := length - headerLen
	if lhead > 1048576 {
		t.Close()
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
	defer t.CloseOnce()
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
		return data[headerLen:]
	}
}

func (t *TCPConn) ReadString() string {
	select {
	case data := <-t.data:
		return string(data[headerLen:])
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
