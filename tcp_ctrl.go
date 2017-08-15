package gtcp

import (
	"errors"
)

type TCPCtrlInterface interface {
	TCPBox
	InstallActor(actor Actor)
}

func NewTCPCtrl(actor Actor) *TCPCtrl {
	return &TCPCtrl{actor}
}

type TCPCtrl struct {
	Actor
}

func (t *TCPCtrl) Close() error {
	select {
	case <-t.Done():
		return errors.New("TCPCtrl has already closed")
	default:
	}
	t.OnClose()
	err := t.Actor.Close()
	return err
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
