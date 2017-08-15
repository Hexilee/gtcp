package gtcp

import "net"

func getConn(addr string) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func DialTCP(addr string) (*TCPConn, error) {

	conn, err := getConn(addr)
	if err != nil {
		conn.Close()
		return nil, err
	}
	tcpConn := NewTCPConn(conn)
	go tcpConn.Scan()
	return tcpConn, nil
}

func DialTCPCtrl(addr string, TCPCtrl TCPCtrlInterface) (TCPCtrlInterface, error) {
	conn, err := getConn(addr)
	if err != nil {
		conn.Close()
		return nil, err
	}
	tcpConn := NewTCPConn(conn)
	TCPCtrl.InstallTCPConn(tcpConn)
	go TCPCtrl.Scan()
	return TCPCtrl, err
}
