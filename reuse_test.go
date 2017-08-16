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
	ReInitPoolWithCtx(100, context.Background())
	doAllTest(t)
	OpenPool()
	doAllTest(t)
}