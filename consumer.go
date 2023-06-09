package natsrpc

import (
	"errors"
	"github.com/nats-io/nats.go"
)

type consumer struct {
	nc           *nats.Conn
	subscription *nats.Subscription
}

func newConsumer(nc *nats.Conn) *consumer {
	return &consumer{
		nc: nc,
	}
}

func (s *consumer) Subscribe(subject, group string, outputCh chan *nats.Msg) error {
	if s.subscription != nil {
		return errors.New("already subscribed")
	}

	subscription, err := s.nc.ChanQueueSubscribe(subject, group, outputCh)
	if err != nil {
		return err
	}

	s.subscription = subscription
	return nil
}

func (s *consumer) Close() error {
	if s.subscription == nil {
		return nil
	}

	err := s.subscription.Drain()
	if err != nil {
		return err
	}

	return nil
}
