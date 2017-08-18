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

type ConnPool chan *TCPConn

func GetConnPool() ConnPool {
	connMu.RLock()
	defer connMu.RUnlock()
	return connP
}

func OpenConnPool(size uint) {
	if !IsConnPoolOpen() {
		connMu.Lock()
		connP = make(chan *TCPConn, size)
		connMu.Unlock()
		atomic.StoreUint32(&isConnPoolOpen, 1)
	}
}

func IsConnPoolOpen() bool {
	return atomic.LoadUint32(&isConnPoolOpen) != 0
}

func ReopenConnPool(size uint) {
	DropConnPool()
	OpenConnPool(size)
}

func DropConnPool() {
	connMu.Lock()
	connP = nil
	connMu.Unlock()
	atomic.StoreUint32(&isConnPoolOpen, 0)
}
