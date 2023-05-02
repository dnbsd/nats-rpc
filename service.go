package natsrpc

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Service interface {
	Start(ctx context.Context, reqCh <-chan *nats.Msg, respCh chan<- *nats.Msg) error
}
