package gtcp

import (
	"sync"
)

type TCPCtrlInterface interface {
	TCPBox
	InstallActor(actor Actor)
}

func NewTCPCtrl(actor Actor) *TCPCtrl {
	return &TCPCtrl{Actor: actor}
}

type TCPCtrl struct {
	Actor
	OnceOnClose sync.Once
}

func (t *TCPCtrl) Close() error {
	t.OnceOnClose.Do(t.OnClose)
	t.Actor.CloseOnce()
	return nil
}

func (t *TCPCtrl) InstallActor(actor Actor) {
	t.Actor = actor
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
