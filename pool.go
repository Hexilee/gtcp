package gtcp

import (
	"sync"
	"context"
	"net"
)

type Pool struct {
	pool    chan *TCPConn
	recycle chan *TCPConn
	mu      sync.RWMutex
	isInit  bool
	isOpen  bool
	Ctx     context.Context
	Cancel  func()
}

var p = new(Pool)

func (p *Pool) InstallCtx(ctx context.Context) {
	p.Ctx, p.Cancel = context.WithCancel(ctx)
}

func (p *Pool) reuse(conn *TCPConn) {
	select {
	case p.pool <- conn:
	default:
	}
}

func (p *Pool) ClearPool() {
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
	if IsPoolInit() {
		select {
		case tcpConn = <-p.pool:
			tcpConn.InstallNetConn(conn)
			return tcpConn, ok
		default:
		}
	}
	return
}

func SendConnToPool(conn *TCPConn) {
	if IsPoolInit() {
		select {
		case p.recycle <- conn:
		default:
		}
	}
}

func (p *Pool) GetPool() <-chan *TCPConn {
	return p.pool
}

func (p *Pool) GetRecycle() chan<- *TCPConn {
	return p.recycle
}

func GetPool() *Pool {
	return p
}

func InitPool(size uint) {
	if IsPoolInit() {
		return
	}
	p.InstallCtx(context.Background())
	p.pool = make(chan *TCPConn, size)
	p.recycle = make(chan *TCPConn, size)
	p.mu.Lock()
	p.isInit = true
	p.mu.Unlock()
}

func InitPoolWithCtx(size uint, ctx context.Context) {
	InitPool(size)
	p.InstallCtx(ctx)
}

func IsPoolInit() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isInit
}

func IsPoolOpen() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isOpen
}

func OpenPool() {
	if IsPoolOpen() {
		return
	}
	p.mu.Lock()
	p.isOpen = true
	p.mu.Unlock()
	go p.ClearPool()
}

func ClosePool() {
	p.mu.Lock()
	p.isOpen = false
	p.mu.Unlock()
	p.Cancel()
}

func ReopenPool() {
	ClosePool()
	p.InstallCtx(context.Background())
	go p.ClearPool()
}

func ReopenPoolWithCtx(ctx context.Context) {
	ClosePool()
	p.InstallCtx(ctx)
	go p.ClearPool()
}

func DropPool() {
	p.Cancel()
	p = new(Pool)
}

func ReInitPool(size uint) {
	DropPool()
	InitPool(size)
}

func ReInitPoolWithCtx(size uint, ctx context.Context) {
	DropPool()
	InitPoolWithCtx(size, ctx)
}
