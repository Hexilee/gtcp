package gtcp

type Actor interface {
	TCPBox
	OnConnect() error
	OnMessage(data []byte) error
	OnError(err error) error
	OnClose()
}

type ActorType struct {
	*TCPConn
}

func (a *ActorType) ReInstallTCPConn(conn *TCPConn) {
	if a.IsScanning(){
		a.CloseOnce()
	}
	a.TCPConn = conn
}


func (a *ActorType) InstallTCPConn(conn *TCPConn) {
	a.TCPConn = conn
}
