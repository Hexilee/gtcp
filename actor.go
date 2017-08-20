package gtcp

// Actor is interface that all Event-callback objects defined by user should implement it.
type Actor interface {
	TCPBox
	OnConnect() error
	OnMessage(data []byte) error
	// if OnError always return err, maybe fall into trap of dead loop
	OnError(err error) error
	OnClose() error
}

// ActorType is a struct that all Event-callback struct defined by user is supported to embed it.
type ActorType struct {
	*TCPConn
}

// InstallTCPConn is a method to embed pointer of TCPConn in actor
func (a *ActorType) InstallTCPConn(conn *TCPConn) {
	a.TCPConn = conn
}
