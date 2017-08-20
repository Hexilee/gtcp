package gtcp

import (
	"testing"
	"time"
)

func doAllTest(t *testing.T) {
	TestTCPCtrlInterface_Server(t)
	TestTCPCtrlInterface_Client(t)
	TestTCPConnInterface(t)
	TestTCPConn(t)
	TestTCPConn_InstallCtx(t)
}

func TestConnPool(t *testing.T) {
	go func() {
		select {
		case <- time.After(1 * time.Second):
			t.Error("There is no tcpconn in pool.conns")
		case <- pool.GetConns():
		}
		select {
		case <- time.After(1 * time.Second):
			t.Error("There is no actor in pool.actors")
		case <- pool.GetActors():
		}
		select {
		case <- time.After(1 * time.Second):
			t.Error("There is no tcpconn in pool.ctrls")
		case <- pool.GetCtrls():
		}
		select {
		case <- time.After(1 * time.Second):
			t.Error("There is no tcpconn in ConnPool")
		case <- GetConnPool():
		}
	}()
	OpenPool(30)
	doAllTest(t)
	ReopenPool(100)
	doAllTest(t)
	OpenConnPool(1000)
	doAllTest(t)
	ReopenConnPool(2000)
	for i := 0; i<=50;i++ {
		doAllTest(t)
	}
}