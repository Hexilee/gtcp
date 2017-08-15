package gtcp

import (
	"sync"
	"testing"
	//"context"
)

func TestTCPConn_InstallCtx(t *testing.T) {
	var (
		wg1 sync.WaitGroup
		wg2 sync.WaitGroup
	)
	wg1.Add(1)

	TCPChan := make(chan TCPCtrlInterface)

	listener, err := NewTCPListener(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		actorTestType := &ActorTestType{
			Wg1:  &wg1,
			Wg2:  &wg2,
			T:    t,
			Data: make([]string, 0),
		}
		tcpConn, err := listener.AcceptTCPCtrl(actorTestType)
		if err != nil {
			t.Errorf("tcp listener err: %s", err.Error())
		}
		TCPChan <- tcpConn
	}()

	client, err := DialTCP(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	assertEqual(t, testChanData[0], client.ReadString(), "conn onConnect write err (server)")

	var server TCPConnInterface
	select {
	case server = <-TCPChan:
		break
	}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
		wg2.Add(1)
	}
	wg2.Wait()

	//ctx, cancel := context.WithCancel(context.Background())
	//server.InstallCtx(ctx)
	//cancel()
	server.Cancel()
	//server.Close()
	wg1.Wait()
}
