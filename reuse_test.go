package gtcp

import (
	"testing"
	"context"
)

func doAllTest(t *testing.T) {
	TestTCPCtrlInterface_Server(t)
	TestTCPCtrlInterface_Client(t)
	TestTCPConnInterface(t)
	TestTCPConn(t)
	TestTCPConn_InstallCtx(t)
}

func TestConnPool(t *testing.T) {
	InitPool(30)
	doAllTest(t)
	ReInitPool(100)
	doAllTest(t)
	ReInitPoolWithCtx(10000, context.Background())
	doAllTest(t)
	OpenPool()
	doAllTest(t)
	ReopenConnPool()
	doAllTest(t)
	ReopenConnPoolWithCtx(context.Background())
	doAllTest(t)
	CloseConnPool()
	doAllTest(t)
}

//func TestHighPerformance(t *testing.T) {
//	for i:=1; i<100; i++ {
//		go TestConnPool(t)
//	}
//}
