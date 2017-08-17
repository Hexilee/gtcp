package gtcp

import (
	"sync"
	"context"
)

type TCPCtrlInterface interface {
	TCPBox
	InstallActor(actor Actor)
}

func NewTCPCtrl(actor Actor) *TCPCtrl {
	return &TCPCtrl{Actor: actor}
}

func GetTCPCtrl (actor Actor) (*TCPCtrl){
	tcpCtrl, ok := GetCtrlFromPool(actor)
	if ok {
		return tcpCtrl
	}
	return &TCPCtrl{Actor: actor}
}

type TCPCtrl struct {
	Actor
	OnceOnClose sync.Once
}

func (t *TCPCtrl) Clear() {
	t.OnceOnClose = sync.Once{}
}

func (t *TCPCtrl) Close() error {
	defer SendCtrlToPool(t)
	t.OnceOnClose.Do(t.OnClose)
	t.Actor.CloseOnce()
	return nil
}

func (t *TCPCtrl) InstallActor(actor Actor){
	t.Close()
	ctrlP.RecycleActor(t.Actor)
	t.Actor = actor
}

func (t *TCPCtrl) Start () {
	go t.Scan()
}

func (t *TCPCtrl) StartWithCtx(ctx context.Context) {
	t.InstallCtx(ctx)
	go t.Scan()
}

func (t *TCPCtrl) Scan() {
	defer t.Close()
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
