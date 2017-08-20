package gtcp

import (
	"sync"
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
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		select {
		case <-time.After(1 * time.Second):
			t.Error("Pool is not open")
		case <-pool.GetConns():
		case <-pool.GetActors():
		case <-pool.GetCtrls():
		case <-GetConnPool():
		}
		wg.Done()
	}()
	OpenPool(30)
	doAllTest(t)
	ReopenPool(100)
	doAllTest(t)
	OpenConnPool(1000)
	doAllTest(t)
	ReopenConnPool(2000)
	doAllTest(t)
	DropPool()
	doAllTest(t)
	DropConnPool()
	doAllTest(t)
	wg.Wait()
}

//func TestMulti(t *testing.T) {
//	for i:=1; i<= 50; i++ {
//		TestConnPool(t)
//	}
//}
