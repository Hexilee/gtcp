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
	InitConnPool(30)
	doAllTest(t)
	ReInitConnPool(100)
	doAllTest(t)
	ReInitConnPoolWithCtx(10000, context.Background())
	doAllTest(t)
	OpenConnPool()
	doAllTest(t)
	ReopenConnPool()
	doAllTest(t)
	ReopenConnPoolWithCtx(context.Background())
	doAllTest(t)
	CloseConnPool()
	doAllTest(t)
}
