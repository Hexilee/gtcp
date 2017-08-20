package gtcp

import (
	"sync"
	"sync/atomic"
)

var (
	connMu         sync.RWMutex
	connP          ConnPool
	isConnPoolOpen uint32
)

// TCPConn pool
type ConnPool chan *TCPConn

// Get conn pool instance
func GetConnPool() ConnPool {
	connMu.RLock()
	defer connMu.RUnlock()
	return connP
}

// Receive size to Open TCPConn pool
func OpenConnPool(size uint) {
	if !IsConnPoolOpen() {
		connMu.Lock()
		connP = make(chan *TCPConn, size)
		connMu.Unlock()
		atomic.StoreUint32(&isConnPoolOpen, 1)
	}
}

// Return true if TCPConn is open else false
func IsConnPoolOpen() bool {
	return atomic.LoadUint32(&isConnPoolOpen) != 0
}

// Reopen TCPConn
func ReopenConnPool(size uint) {
	DropConnPool()
	OpenConnPool(size)
}

// Drop TCPConn
func DropConnPool() {
	connMu.Lock()
	connP = nil
	connMu.Unlock()
	atomic.StoreUint32(&isConnPoolOpen, 0)
}
