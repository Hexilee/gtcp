package gtcp

import (
	"context"
	"testing"
)

func TestTCPConn_InstallCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	TCPChan := make(chan *TCPCtrl)

	listener, err := NewTCPListener(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		actorTestType := &ActorTestType{
			T:    t,
			Data: make([]string, 0),
		}
		tcpConn, err := listener.AcceptTCPCtrl(actorTestType)
		if err != nil {
			t.Errorf("tcp listener err: %s", err.Error())
		}
		tcpConn.StartWithCtx(ctx)
		TCPChan <- tcpConn
	}()

	client, err := DialTCP(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	client.Start()
	assertEqual(t, testChanData[0], client.ReadString(), "conn onConnect write err (server)")

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
	}
	cancel()
}
