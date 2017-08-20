package gtcp

import (
	"context"
)

// All TCPCtrl Type should implement TCPCtrlInterface
type TCPCtrlInterface interface {
	TCPBox
	InstallActor(actor Actor)
}

// Get a new TCPCtrl
func NewTCPCtrl(actor Actor) *TCPCtrl {
	return &TCPCtrl{Actor: actor}
}

// if Pool is open, and there are some TCPCtrl in it, get a TCPCtrl pointer from it
// else return a new TCPCtrl
func GetTCPCtrl(actor Actor) *TCPCtrl {
	tcpCtrl, ok := GetCtrlFromPool(actor)
	if ok {
		return tcpCtrl
	}
	return NewTCPCtrl(actor)
}

// TCPCtrl struct
type TCPCtrl struct {
	Actor
}

// Executed only once in a life cycle
func (t *TCPCtrl) CloseOnce() {
	defer SendCtrlToPool(t)
	err := t.OnClose()
	for err != nil {
		err = t.OnError(err)
	}
}

// Replace embedded actor
func (t *TCPCtrl) InstallActor(actor Actor) {
	t.Close()
	SendActorToPool(t.Actor)
	t.Actor = actor
}

// Go Scan
func (t *TCPCtrl) Start() {
	t.InstallCtx(context.Background())
	go t.Scan()
}

// Go Scan with father ctx
func (t *TCPCtrl) StartWithCtx(ctx context.Context) {
	t.InstallCtx(ctx)
	go t.Scan()
}

// Get data from TCPConn and execute OnConnect, OnMessage, OnError and OnClose event function.
func (t *TCPCtrl) Scan() {
	defer t.CloseOnce()
	err := t.OnConnect()
	for err != nil {
		err = t.OnError(err)
	}

	go t.Actor.Scan()

	dataChan := t.GetDataChan()
	errChan := t.GetErrChan()

Circle:
	for {
		select {
		case <-t.Done():
			break Circle
		case data := <-dataChan:
			err := t.OnMessage(data)
			for err != nil {
				err = t.OnError(err)
			}
		case err := <-errChan:
			for err != nil {
				err = t.OnError(err)
			}
		}
	}
}
