package gtcp

import (
	"testing"
	"fmt"
)

type ActorTestType struct {
	ActorType
	T    *testing.T
	Data []string
}

func (s *ActorTestType) OnConnect() error {
	_, err := s.Write([]byte(testChanData[0]))
	return err
}

func (s *ActorTestType) OnMessage(data []byte) error {
	s.Data = append(s.Data, string(data))
	return nil
}

func (s *ActorTestType) OnClose() error {
	for n, data := range s.Data {
		assertEqual(s.T, AddHeader([]byte(testChanData[n])), data, "Test TCP Type Err")
	}
	return nil
}

func (s *ActorTestType) OnError(err error) error {
	fmt.Println(err)
	s.Close()
	return nil
}

func TestTCPCtrlInterface_Server(t *testing.T) {
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
		tcpConn.Start()
		TCPChan <- tcpConn
	}()

	client, err := DialTCP(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	client.Start()
	assertEqual(t, testChanData[0], client.ReadString(), "conn onConnect write err (server)")

	var server *TCPCtrl
	select {
	case server = <-TCPChan:
		break
	}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
	}
	server.Close()
}

func TestTCPCtrlInterface_Client(t *testing.T) {
	var (
	)

	TCPChan := make(chan *TCPConn)

	listener, err := NewTCPListener(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			t.Errorf("tcp listener err: %s", err.Error())
		}
		tcpConn.Start()
		TCPChan <- tcpConn
	}()

	actorTestType := &ActorTestType{
		T:    t,
		Data: make([]string, 0),
	}
	client, err := DialTCPCtrl(Addr, actorTestType)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	client.Start()
	select {
	case server := <-TCPChan:
		assertEqual(t, testChanData[0], server.ReadString(), "conn onConnect write err (client)")
		for _, testStr := range testChanData {
			_, _ = server.Write([]byte(testStr))
		}
	}
	client.Close()
}
