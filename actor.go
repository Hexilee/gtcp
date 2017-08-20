package gtcp

// All Event-callback objects defined by user should implement Actor interface
type Actor interface {
	TCPBox
	OnConnect() error
	OnMessage(data []byte) error
	// if OnError always return err, maybe fall into trap of dead loop
	OnError(err error) error
	OnClose() error
}

// All Event-callback struct defined by user is supported to embed ActorType
type ActorType struct {
	*TCPConn
}

// To embed pointer of TCPConn in actor
func (a *ActorType) InstallTCPConn(conn *TCPConn) {
	a.TCPConn = conn
}
