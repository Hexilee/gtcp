package gtcp

import (
	"net"
)

func NewTCPListener(addr string) (*TCPListener, error) {
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
	tcpConn := GetTCPConn(conn)
	return tcpConn, err
}

func (t *TCPListener) Accept() (TCPConnInterface, error) {
	conn, err := t.TCPListener.AcceptTCP()
	tcpConn := GetTCPConn(conn)
	return tcpConn, err
}

func (t *TCPListener) AcceptTCPCtrl(actor Actor) (*TCPCtrl, error) {
	conn, err := t.TCPListener.AcceptTCP()

	if err != nil {
		return nil, err
	}
	tcpConn := GetTCPConn(conn)
	TCPCtrl := GetTCPCtrl(actor)
	TCPCtrl.InstallTCPConn(tcpConn)
	return TCPCtrl, err
}
