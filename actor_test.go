package gtcp

import (
	"sync"
	"testing"
	"fmt"
	"io"
)

type ActorTestType struct {
	ActorType
	Wg1  *sync.WaitGroup
	Wg2  *sync.WaitGroup
	T    *testing.T
	Data []string
}

func (s *ActorTestType) OnConnect() error {
	_, err := s.Write([]byte(testChanData[0]))
	return err
}

func (s *ActorTestType) OnMessage(data []byte) error {
	s.Data = append(s.Data, string(data))
	s.Wg2.Done()
	return nil
}

func (s *ActorTestType) OnClose() {
	for n, data := range testChanData {
		assertEqual(s.T, AddHeader([]byte(data)), s.Data[n], "Test TCP Type Err")
	}
	s.Wg1.Done()
}

func (s *ActorTestType) OnError(err error) error {
	fmt.Println(err)
	if err != io.EOF {
		s.CloseOnce()
	}
	return nil
}

func TestTCPCtrlInterface_Server(t *testing.T) {
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
		tcpConn.Start()
		TCPChan <- tcpConn
	}()

	client, err := DialTCP(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	client.Start()
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
	server.Close()
	wg1.Wait()
}

func TestTCPCtrlInterface_Client(t *testing.T) {
	var (
		wg1 sync.WaitGroup
		wg2 sync.WaitGroup
	)
	wg1.Add(1)

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
		Wg1:  &wg1,
		Wg2:  &wg2,
		T:    t,
		Data: make([]string, 0),
	}
	client, err := DialTCPCtrl(Addr, actorTestType)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	client.Start()

	var server TCPConnInterface
	select {
	case server = <-TCPChan:
		break
	}

	assertEqual(t, testChanData[0], server.ReadString(), "conn onConnect write err (client)")

	for _, testStr := range testChanData {
		_, _ = server.Write([]byte(testStr))
		wg2.Add(1)
	}
	wg2.Wait()
	client.Close()
	wg1.Wait()
}
