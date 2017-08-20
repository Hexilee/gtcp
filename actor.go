package gtcp

type Actor interface {
	TCPBox
	OnConnect() error
	OnMessage(data []byte) error
	OnError(err error) error
	OnClose() error
}

type ActorType struct {
	*TCPConn
}

func (a *ActorType) InstallTCPConn(conn *TCPConn) {
	a.TCPConn = conn
}
