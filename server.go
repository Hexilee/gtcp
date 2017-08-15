package gtcp

import (
	"net"
)

func NewTCPListenser(addr string) (*TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}

	return &TCPListener{listener}, nil
}

type TCPListener struct {
	*net.TCPListener
}

func (t *TCPListener) AcceptTCP() (*TCPConn, error) {
	conn, err := t.TCPListener.AcceptTCP()
	tcpConn := NewTCPConn(conn)
	go tcpConn.Scan()
	return tcpConn, err
}

func (t *TCPListener) Accept() (TCPConnInterface, error) {
	conn, err := t.TCPListener.AcceptTCP()
	tcpConn := NewTCPConn(conn)
	go tcpConn.Scan()
	return tcpConn, err
}

func (t *TCPListener) AcceptTCPCtrl(actor Actor) (TCPCtrlInterface, error) {
	conn, err := t.TCPListener.AcceptTCP()

	if err != nil {
		return nil, err
	}
	tcpConn := NewTCPConn(conn)
	TCPCtrl := NewTCPCtrl(actor)
	TCPCtrl.InstallTCPConn(tcpConn)
	go TCPCtrl.Scan()
	return TCPCtrl, err
}
