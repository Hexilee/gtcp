package gtcp

import (
	"testing"
	"bytes"
	"encoding/binary"
	"sync"
	"fmt"
	"io"
)

var testChanData = [...]string{
		"hello world!",
		"12312u4h18yg4912g49",
		"你好世界！",
		"こんにちは世界",
		`Contributing

To hack on this project:

Install as usual (go get -u github.com/firstrow/tcp_server)
Create your feature branch (git checkout -b my-new-feature)
Ensure everything works and the tests pass (go test)
Commit your changes (git commit -am 'Add some feature')
Contribute upstream:

Fork it on GitHub
Add your remote (git remote add fork git@github.com:firstrow/tcp_server.git)
Push to the branch (git push fork my-new-feature)
Create a new Pull Request on GitHub
Notice: Always use the original import path by installing with go get.`,
	}

const (
	Addr = "127.0.0.1:8005"
)

func assertEqual(t *testing.T, expect string, got string, msg string) {
	if expect != got {
		t.Errorf(`%s: expect %s(%d),\n but got %s(%d)`, msg, expect, len(expect), got, len(got))
	}
}

func AddHeader(data []byte) string {
	length := uint32(len(data))
	head := make([]byte, 4)
	binary.LittleEndian.PutUint32(head, length)
	buf := bytes.NewBuffer(head)
	buf.Write(data)
	return buf.String()
}

func TestTCPConnInterface(t *testing.T) {
	TCPChan := make(chan TCPConnInterface)

	listener, err := NewTCPListenser(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		tcpConn, err := listener.Accept()
		if err != nil {
			t.Errorf("tcp listener err: %s", err.Error())
		}
		TCPChan <- tcpConn
	}()

	client, err := DialTCP(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}

	var server TCPConnInterface
	select {
	case server = <-TCPChan:
		break
	}
	testChanData := []string{
		"hello world!",
		"12312u4h18yg4912g49",
		"你好世界！",
		"こんにちは世界",
		`Contributing

To hack on this project:

Install as usual (go get -u github.com/firstrow/tcp_server)
Create your feature branch (git checkout -b my-new-feature)
Ensure everything works and the tests pass (go test)
Commit your changes (git commit -am 'Add some feature')
Contribute upstream:

Fork it on GitHub
Add your remote (git remote add fork git@github.com:firstrow/tcp_server.git)
Push to the branch (git push fork my-new-feature)
Create a new Pull Request on GitHub
Notice: Always use the original import path by installing with go get.`,
	}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
		resultBytes := <-server.GetDataChan()
		assertEqual(t, string(resultBytes), AddHeader([]byte(testStr)), "TCP add header data err (client -> server)")
	}

	for _, testStr := range testChanData {
		_, _ = server.Write([]byte(testStr))
		resultBytes := <-client.GetDataChan()
		assertEqual(t, string(resultBytes), AddHeader([]byte(testStr)), "TCP add header data err (server -> client)")
	}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
		resultBytes := server.ReadData()
		assertEqual(t, string(resultBytes), testStr, "TCP data err (client -> server)")
	}

	for _, testStr := range testChanData {
		_, _ = server.Write([]byte(testStr))
		resultBytes := client.ReadData()
		assertEqual(t, string(resultBytes), testStr, "TCP data err (server -> client)")
	}
}

func TestTCPConn(t *testing.T) {
	TCPChan := make(chan *TCPConn)

	listener, err := NewTCPListenser(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			t.Errorf("tcp listener err: %s", err.Error())
		}
		TCPChan <- tcpConn
	}()

	client, err := DialTCP(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}

	var server TCPConnInterface
	select {
	case server = <-TCPChan:
		break
	}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
		resultBytes := <-server.GetDataChan()
		assertEqual(t, string(resultBytes), AddHeader([]byte(testStr)), "TCP add header data err (client -> server)")
	}

	for _, testStr := range testChanData {
		_, _ = server.Write([]byte(testStr))
		resultBytes := <-client.GetDataChan()
		assertEqual(t, string(resultBytes), AddHeader([]byte(testStr)), "TCP add header data err (server -> client)")
	}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
		resultBytes := server.ReadData()
		assertEqual(t, string(resultBytes), testStr, "TCP data err (client -> server)")
	}

	for _, testStr := range testChanData {
		_, _ = server.Write([]byte(testStr))
		resultBytes := client.ReadData()
		assertEqual(t, string(resultBytes), testStr, "TCP data err (server -> client)")
	}
	server.Close()
}


type ServerType struct {
	ActorType
	Wg *sync.WaitGroup
	T *testing.T
	Data []string
}

func (s *ServerType) OnConnect() error{
	_, err := s.Write([]byte(testChanData[0]))
	return err
}

func (s *ServerType) OnMessage(data []byte) error {
	s.Data = append(s.Data, string(data))
	s.Wg.Done()
	return nil
}

func (s *ServerType) OnClose() error {
	for n, data := range testChanData {
		assertEqual(s.T, AddHeader([]byte(data)), s.Data[n], "Test TCP Type Err")
	}
	s.Wg.Done()
	return nil
}

func (s *ServerType) OnError(err error) error {
	fmt.Println(err)
	if err != io.EOF {
		s.Close()
	}
	return nil
}


func TestTCPCtrlInterface_Server(t *testing.T) {
	var wg sync.WaitGroup
	//wg.Add(1)

	TCPChan := make(chan TCPCtrlInterface)

	listener, err := NewTCPListenser(Addr)
	if err != nil {
		t.Errorf("tcp listener err: %s", err.Error())
	}
	defer listener.Close()

	go func() {
		serverType := &ServerType{
			Wg:&wg,
			T: t,
			Data:make([]string, 0),
		}
		tcpConn, err := listener.AcceptTCPCtrl()
		tcpConn.InstallActor(serverType)
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

	//var server TCPConnInterface
	//select {
	//case server = <-TCPChan:
	//	break
	//}

	for _, testStr := range testChanData {
		_, _ = client.Write([]byte(testStr))
		wg.Add(1)
	}
	wg.Wait()
	//server.Close()
}