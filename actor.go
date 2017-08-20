package gtcp

type Actor interface {
	TCPBox
	OnConnect() error
	OnMessage(data []byte) error
	OnError(err error) error //
	OnClose() error
}

type ActorType struct {
	*TCPConn
}

//func (a *ActorType) ReInstallTCPConn(conn *TCPConn) {
//	//if a.IsScanning() {
//	//	a.CloseOnce()
//	//}
//	//a.Close()
//	if a.Context != nil {
//		if a.IsDone() {
//			SendConnToPool(a.TCPConn)
//		}
//	}
//	a.TCPConn = conn
//}

func (a *ActorType) InstallTCPConn(conn *TCPConn) {
	a.TCPConn = conn
}
