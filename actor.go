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

func (a *ActorType) ReInstallTCPConn(conn *TCPConn) (err error) {
	if a.IsScanning() && !a.IsClosed() {
		err = a.Close()
	}
	a.TCPConn = conn
	return err
}

func (a *ActorType) InstallTCPConn(conn *TCPConn) {
	a.TCPConn = conn
}
