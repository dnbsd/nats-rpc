package natsrpc

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	nc        *nats.Conn
	consumer  *consumer
	publisher *publisher
	services  map[string]Service
}

func NewServer(nc *nats.Conn) *Server {
	return &Server{
		nc:        nc,
		consumer:  newConsumer(nc),
		publisher: newPublisher(nc),
		services:  make(map[string]Service),
	}
}

func (s *Server) StartWithContext(ctx context.Context) error {
	defer s.consumer.Close()

	if len(s.services) == 0 {
		return errors.New("cannot start the server: no services registered")
	}

	group, groupCtx := errgroup.WithContext(ctx)
	for subject, service := range s.services {
		// TODO: set capacity from the service!
		chCapacity := 1024
		reqCh := make(chan *nats.Msg, chCapacity)
		respCh := make(chan *nats.Msg, chCapacity)

		err := s.consumer.Subscribe(subject, reqCh)
		if err != nil {
			return err
		}

		group.Go(func() error {
			return service.Start(groupCtx, reqCh, respCh)
		})

		group.Go(func() error {
			return s.publisher.Start(groupCtx, respCh)
		})
	}

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Register(subject string, service Service) {
	s.services[subject] = service
}
