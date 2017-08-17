package gtcp

import (
	"testing"
)

func doAllTest(t *testing.T) {
	TestTCPCtrlInterface_Server(t)
	TestTCPCtrlInterface_Client(t)
	TestTCPConnInterface(t)
	TestTCPConn(t)
	TestTCPConn_InstallCtx(t)
}

func TestConnPool(t *testing.T) {
	OpenConnPool(30)
	doAllTest(t)
	ReopenConnPool(100)
	doAllTest(t)
}

//func TestHighPerformance(t *testing.T) {
//	for i:=1; i<100; i++ {
//		TestConnPool(t)
//	}
//}
