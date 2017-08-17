package gtcp

import (
	"sync"
	"context"
	"net"
)

type CtrlPool struct {
	pool    chan *TCPCtrl
	recycle chan *TCPCtrl
	mu      sync.RWMutex
	isInit  bool
	isOpen  bool
	Ctx     context.Context
	Cancel  func()
}

var ctrlP = new(CtrlPool)

func (p *CtrlPool) InstallCtx(ctx context.Context) {
	p.Ctx, p.Cancel = context.WithCancel(ctx)
}

func (p *CtrlPool) reuse(ctrl *TCPCtrl) {
	select {
	case p.pool <- ctrl:
	default:
	}
}

func (p *CtrlPool) ClearPool() {
Circle:
	for {
		select {
		case ctrl := <-p.recycle:
			ctrl.Clear()
			p.reuse(ctrl)
		case <-p.Ctx.Done():
			break Circle
		}
	}
}

func GetCtrlFromPool(conn *net.TCPConn) (tcpCtrl *TCPCtrl, ok bool) {
	if IsCtrlPoolInit() {
		select {
		case tcpCtrl = <-ctrlP.pool:
			tcpCtrl.InstallNetConn(conn)
			return tcpCtrl, ok
		default:
		}
	}
	return
}

func SendCtrlToPool(ctrl *TCPCtrl) {
	if IsCtrlPoolInit() {
		select {
		case ctrlP.recycle <- ctrl:
		default:
		}
	}
}

func (p *CtrlPool) GetPool() <-chan *TCPCtrl {
	return p.pool
}

func (p *CtrlPool) GetRecycle() chan<- *TCPCtrl {
	return p.recycle
}

func GetCtrlPool() *CtrlPool {
	return ctrlP
}

func InitCtrlPool(size uint) {
	if IsCtrlPoolInit() {
		return
	}
	ctrlP.InstallCtx(context.Background())
	ctrlP.pool = make(chan *TCPCtrl, size)
	ctrlP.recycle = make(chan *TCPCtrl, size)
	ctrlP.mu.Lock()
	ctrlP.isInit = true
	ctrlP.mu.Unlock()
}

func InitCtrlPoolWithCtx(size uint, ctx context.Context) {
	InitCtrlPool(size)
	ctrlP.InstallCtx(ctx)
}

func IsCtrlPoolInit() bool {
	ctrlP.mu.RLock()
	defer ctrlP.mu.RUnlock()
	return ctrlP.isInit
}

func IsCtrlPoolOpen() bool {
	ctrlP.mu.RLock()
	defer ctrlP.mu.RUnlock()
	return ctrlP.isOpen
}

func OpenCtrlPool() {
	if IsCtrlPoolOpen() {
		return
	}
	ctrlP.mu.Lock()
	ctrlP.isOpen = true
	ctrlP.mu.Unlock()
	go ctrlP.ClearPool()
}

func CloseCtrlPool() {
	ctrlP.mu.Lock()
	ctrlP.isOpen = false
	ctrlP.mu.Unlock()
	ctrlP.Cancel()
}

func ReopenCtrlPool() {
	CloseCtrlPool()
	ctrlP.InstallCtx(context.Background())
	go ctrlP.ClearPool()
}

func ReopenCtrlPoolWithCtx(ctx context.Context) {
	CloseCtrlPool()
	ctrlP.InstallCtx(ctx)
	go ctrlP.ClearPool()
}

func DropCtrlPool() {
	ctrlP.Cancel()
	ctrlP = new(CtrlPool)
}

func ReInitCtrlPool(size uint) {
	DropCtrlPool()
	InitCtrlPool(size)
}

func ReInitCtrlPoolWithCtx(size uint, ctx context.Context) {
	DropCtrlPool()
	InitCtrlPoolWithCtx(size, ctx)
}
