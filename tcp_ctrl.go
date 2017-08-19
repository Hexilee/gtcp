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
	return &TCPCtrl{Actor: actor, OnceOnClose: new(sync.Once), mu: new(sync.RWMutex)}
}

func GetTCPCtrl(actor Actor) (*TCPCtrl) {
	tcpCtrl, ok := GetCtrlFromPool(actor)
	if ok {
		return tcpCtrl
	}
	return NewTCPCtrl(actor)
}

type TCPCtrl struct {
	Actor
	OnceOnClose *sync.Once
	mu          *sync.RWMutex
}

func (t *TCPCtrl) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.OnceOnClose = new(sync.Once)
}

func (t *TCPCtrl) Close() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	defer SendCtrlToPool(t)
	t.OnceOnClose.Do(t.OnClose)
	t.Actor.CloseOnce()
	return nil
}

func (t *TCPCtrl) InstallActor(actor Actor) {
	t.Close()
	t.mu.Lock()
	defer t.mu.Unlock()
	SendActorToPool(t.Actor)
	t.Actor = actor
}

func (t *TCPCtrl) Start() {
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
