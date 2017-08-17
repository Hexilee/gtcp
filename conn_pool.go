package gtcp

import (
	"sync"
	"net"
)

var (
	connP          ConnPool
	connPMu        sync.RWMutex
	isConnPoolOpen bool
)

type ConnPool chan *TCPConn

func GetConnFromConnPool(conn *net.TCPConn) (tcpConn *TCPConn, ok bool) {
	if IsConnPoolOpen() {
		select {
		case tcpConn = <-connP:
			tcpConn.InstallNetConn(conn)
			return tcpConn, true
		default:
		}
	}
	return
}

func SendConnToConnPool(conn *TCPConn) {
	if IsConnPoolOpen() {
		conn.Clear()
		select {
		case connP <- conn:
		default:
		}
	}
}

func GetConnPool() ConnPool {
	return connP
}

func OpenConnPool(size uint) {
	if !IsConnPoolOpen() {
		connPMu.Lock()
		connP = make(chan *TCPConn, size)
		isConnPoolOpen = true
		connPMu.Unlock()
	}
}

func IsConnPoolOpen() bool {
	connPMu.RLock()
	defer connPMu.RUnlock()
	return isConnPoolOpen
}

func ReopenConnPool(size uint) {
	DropConnPool()
	OpenConnPool(size)
}

func DropConnPool() {
	connPMu.Lock()
	connP = make(ConnPool)
	isConnPoolOpen = false
	connPMu.Unlock()
}
