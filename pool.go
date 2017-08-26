package gtcp

import (
	"net"
	"sync"
	"sync/atomic"
)

var (
	pool       = new(Pool)
	poolMu     sync.RWMutex
	isPoolOpen uint32
)

// Pool Container
// Contain TCPCtrl pool, Actor pool and TCPConn pool
type Pool struct {
	ctrls  chan *TCPCtrl
	actors chan Actor
	conns  chan *TCPConn
}

// Get a TCPCtrl pointer from pool
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

// Get an Actor pointer from pool
func GetActorFromPool() (actor Actor, ok bool) {
	if IsPoolOpen() {
		select {
		case actor = <-pool.actors:
			return actor, true
		default:

		}
	}
	return
}

// Get a TCPConn pointer from pool
func GetConnFromPool(conn *net.TCPConn) (tcpConn *TCPConn, ok bool) {
	if IsPoolOpen() || IsConnPoolOpen() {
		select {
		case tcpConn = <-pool.conns:
			tcpConn.ReInstallNetConn(conn)
			ok = true
		case tcpConn = <-connP:
			tcpConn.ReInstallNetConn(conn)
			ok = true
		default:
		}
	}
	return
}

// Send a TCPCtrl pointer to pool
func SendCtrlToPool(ctrl *TCPCtrl) {
	if IsPoolOpen() {
		select {
		case pool.ctrls <- ctrl:
		default:
		}
	}
}

// Send an Actor pointer to pool
func SendActorToPool(actor Actor) {
	if IsPoolOpen() {
		select {
		case pool.actors <- actor:
		default:
		}
	}
}

// Send a TCPConn pointer to pool
func SendConnToPool(conn *TCPConn) {
	if IsPoolOpen() || IsConnPoolOpen() {
		select {
		case <-conn.Context.Done():
			select {
			case pool.conns <- conn:
			case connP <- conn:
			default:
			}
		default:
		}
	}
}

// Get TCPCtrl pool instance
func (p *Pool) GetCtrls() <-chan *TCPCtrl {
	return p.ctrls
}

// Get Actor pool instance
func (p *Pool) GetActors() <-chan Actor {
	return p.actors
}

// Get TCPConn pool instance
func (p *Pool) GetConns() <-chan *TCPConn {
	return p.conns
}

// Get pool container
func GetPool() *Pool {
	return pool
}

// receive size and open pool
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

// return true if container is open else false
func IsPoolOpen() bool {
	poolMu.RLock()
	defer poolMu.RUnlock()
	return atomic.LoadUint32(&isPoolOpen) != 0
}

// reopen container
func ReopenPool(size uint) {
	DropPool()
	OpenPool(size)
}

// drop container
func DropPool() {
	poolMu.Lock()
	defer poolMu.Unlock()
	pool = new(Pool)
	atomic.StoreUint32(&isPoolOpen, 0)
}
