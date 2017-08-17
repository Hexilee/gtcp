package gtcp

import (
	"sync"
	"context"
	"net"
)

type ConnPool struct {
	pool    chan *TCPConn
	recycle chan *TCPConn
	mu      sync.RWMutex
	isInit  bool
	isOpen  bool
	Ctx     context.Context
	Cancel  func()
}

var connP = new(ConnPool)

func (p *ConnPool) InstallCtx(ctx context.Context) {
	p.Ctx, p.Cancel = context.WithCancel(ctx)
}

func (p *ConnPool) reuse(conn *TCPConn) {
	select {
	case p.pool <- conn:
	default:
	}
}

func (p *ConnPool) ClearPool() {
Circle:
	for {
		select {
		case conn := <-p.recycle:
			conn.Clear()
			p.reuse(conn)
		case <-p.Ctx.Done():
			break Circle
		}
	}
}

func GetConnFromPool(conn *net.TCPConn) (tcpConn *TCPConn, ok bool) {
	if IsConnPoolInit() {
		select {
		case tcpConn = <-connP.pool:
			tcpConn.InstallNetConn(conn)
			return tcpConn, ok
		default:
		}
	}
	return
}

func SendConnToPool(conn *TCPConn) {
	if IsConnPoolInit() {
		select {
		case connP.recycle <- conn:
		default:
		}
	}
}

func (p *ConnPool) GetPool() <-chan *TCPConn {
	return p.pool
}

func (p *ConnPool) GetRecycle() chan<- *TCPConn {
	return p.recycle
}

func GetConnPool() *ConnPool {
	return connP
}

func InitConnPool(size uint) {
	if IsConnPoolInit() {
		return
	}
	connP.InstallCtx(context.Background())
	connP.pool = make(chan *TCPConn, size)
	connP.recycle = make(chan *TCPConn, size)
	connP.mu.Lock()
	connP.isInit = true
	connP.mu.Unlock()
}

func InitConnPoolWithCtx(size uint, ctx context.Context) {
	InitConnPool(size)
	connP.InstallCtx(ctx)
}

func IsConnPoolInit() bool {
	connP.mu.RLock()
	defer connP.mu.RUnlock()
	return connP.isInit
}

func IsConnPoolOpen() bool {
	connP.mu.RLock()
	defer connP.mu.RUnlock()
	return connP.isOpen
}

func OpenConnPool() {
	if IsConnPoolOpen() {
		return
	}
	connP.mu.Lock()
	connP.isOpen = true
	connP.mu.Unlock()
	go connP.ClearPool()
}

func CloseConnPool() {
	connP.mu.Lock()
	connP.isOpen = false
	connP.mu.Unlock()
	connP.Cancel()
}

func ReopenConnPool() {
	CloseConnPool()
	connP.InstallCtx(context.Background())
	go connP.ClearPool()
}

func ReopenConnPoolWithCtx(ctx context.Context) {
	CloseConnPool()
	connP.InstallCtx(ctx)
	go connP.ClearPool()
}

func DropConnPool() {
	connP.Cancel()
	connP = new(ConnPool)
}

func ReInitConnPool(size uint) {
	DropConnPool()
	InitConnPool(size)
}

func ReInitConnPoolWithCtx(size uint, ctx context.Context) {
	DropConnPool()
	InitConnPoolWithCtx(size, ctx)
}
