package gtcp

import (
	"sync"
)

var (
	connMu         sync.RWMutex
	connP          ConnPool
	isConnPoolOpen bool
)

type ConnPool chan *TCPConn

func GetConnPool() ConnPool {
	return connP
}

func OpenConnPool(size uint) {
	if !IsConnPoolOpen() {
		connMu.Lock()
		connP = make(chan *TCPConn, size)
		isConnPoolOpen = true
		connMu.Unlock()
	}
}

func IsConnPoolOpen() bool {
	connMu.RLock()
	defer connMu.RUnlock()
	return isConnPoolOpen
}

func ReopenConnPool(size uint) {
	DropConnPool()
	OpenConnPool(size)
}

func DropConnPool() {
	connMu.Lock()
	connP = nil
	isConnPoolOpen = false
	connMu.Unlock()
}
