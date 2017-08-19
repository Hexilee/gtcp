package gtcp

import (
	"sync"
	"net"
	"sync/atomic"
)

var (
	pool       = new(Pool)
	poolMu     sync.RWMutex
	isPoolOpen uint32
)

type Pool struct {
	ctrls  chan *TCPCtrl
	actors chan Actor
	conns  chan *TCPConn
}

func GetCtrlFromPool(actor Actor) (tcpCtrl *TCPCtrl, ok bool) {
	if IsPoolOpen() {
		select {
		case tcpCtrl = <-pool.ctrls:
			tcpCtrl.InstallActor(actor)
			return tcpCtrl, true
		default:
		}
	}
	return
}

func GetActorFromPool() (actor Actor, ok bool) {
	if IsPoolOpen() {
		select {
		case actor = <-pool.actors:
			return actor, true
		}
	}
	return
}

func GetConnFromPool(conn *net.TCPConn) (tcpConn *TCPConn, ok bool) {
	if IsPoolOpen() || IsConnPoolOpen() {
		select {
		case tcpConn = <-pool.conns:
			tcpConn.InstallNetConn(conn)
			ok = true
		case tcpConn = <-connP:
			tcpConn.InstallNetConn(conn)
			ok = true
		default:
		}
	}
	return
}

func SendCtrlToPool(ctrl *TCPCtrl) {
	if IsPoolOpen() {
		ctrl.Clear()
		select {
		case pool.ctrls <- ctrl:
		default:
		}
	}
}

func SendActorToPool(actor Actor) {
	if IsPoolOpen() {
		select {
		case pool.actors <- actor:
		default:
		}
	}
}

func SendConnToPool(conn *TCPConn) {
	if IsPoolOpen() || IsConnPoolOpen() {
		conn.Clear()
		select {
		case pool.conns <- conn:
		case connP <- conn:
		default:
		}
	}
}

func (p *Pool) GetCtrls() <-chan *TCPCtrl {
	return p.ctrls
}

func (p *Pool) GetActors() <-chan Actor {
	return p.actors
}

func (p *Pool) GetConns() <-chan *TCPConn {
	return p.conns
}

func GetPool() *Pool {
	return pool
}

func OpenPool(size uint) {
	if !IsPoolOpen() {
		poolMu.Lock()
		pool.ctrls = make(chan *TCPCtrl, size)
		pool.actors = make(chan Actor, size)
		pool.conns = make(chan *TCPConn, size)
		poolMu.Unlock()
		atomic.StoreUint32(&isPoolOpen, 1)
	}
}

func IsPoolOpen() bool {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return atomic.LoadUint32(&isPoolOpen) != 0
}

func ReopenPool(size uint) {
	DropPool()
	OpenPool(size)
}

func DropPool() {
	poolMu.Lock()
	defer poolMu.Unlock()
	pool = new(Pool)
	atomic.StoreUint32(&isPoolOpen, 0)
}
