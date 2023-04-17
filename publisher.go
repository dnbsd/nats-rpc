package gonatsrpc

import (
	"context"

	"github.com/nats-io/nats.go"
)

type publisher struct {
	nc *nats.Conn
}

func newPublisher(nc *nats.Conn) *publisher {
	return &publisher{
		nc: nc,
	}
}

func (p *publisher) Start(ctx context.Context, inputCh chan *nats.Msg) error {
	for {
		select {
		case msg := <-inputCh:
			err := p.nc.PublishMsg(msg)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return nil
		}
	}
}
