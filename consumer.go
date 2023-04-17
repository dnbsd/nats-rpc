package gonatsrpc

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

type consumer struct {
	nc            *nats.Conn
	subscriptions map[string]*nats.Subscription
}

func newConsumer(nc *nats.Conn) *consumer {
	return &consumer{
		nc:            nc,
		subscriptions: make(map[string]*nats.Subscription),
	}
}

func (s *consumer) isSubscribed(subject string) bool {
	_, exists := s.subscriptions[subject]
	return exists
}

func (s *consumer) Subscribe(subject string, outputCh chan *nats.Msg) error {
	if s.isSubscribed(subject) {
		return fmt.Errorf("already subscribed to subject '%s'", subject)
	}

	subscription, err := s.nc.ChanSubscribe(subject, outputCh)
	if err != nil {
		return err
	}

	s.subscriptions[subject] = subscription
	return nil
}

func (s *consumer) Unsubscribe(subject string) error {
	subscription, exists := s.subscriptions[subject]
	if !exists {
		return nil
	}

	err := subscription.Unsubscribe()
	if err != nil {
		return err
	}

	delete(s.subscriptions, subject)
	return nil
}

func (s *consumer) Close() error {
	for subject := range s.subscriptions {
		err := s.Unsubscribe(subject)
		if err != nil {
			return err
		}
	}
	return nil
}
