package natsrpc

import (
	"context"
	"errors"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	nc        *nats.Conn
	pipelines []pipeline
}

func NewServer(nc *nats.Conn) *Server {
	return &Server{
		nc: nc,
	}
}

// StartWithContext starts a new server and blocks until context is closed or an error occurs.
func (s *Server) StartWithContext(ctx context.Context) error {
	if len(s.pipelines) == 0 {
		return errors.New("cannot start the server: no services registered")
	}

	defer func() {
		for _, p := range s.pipelines {
			_ = p.consumer.Close()
		}
	}()

	group, groupCtx := errgroup.WithContext(ctx)
	for _, p := range s.pipelines {
		// TODO: set capacity from the service!
		chCapacity := 1024
		reqCh := make(chan *nats.Msg, chCapacity)
		respCh := make(chan *nats.Msg, chCapacity)

		err := p.consumer.Subscribe(p.subject, p.group, reqCh)
		if err != nil {
			return err
		}

		group.Go(func() error {
			return p.service.Start(groupCtx, reqCh, respCh)
		})

		group.Go(func() error {
			return p.publisher.Start(groupCtx, respCh)
		})
	}

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}

// Register registers a service with a NATS subject. Many services can be registered to the same subject, all services
// will receive their copy of the message. If group is specified, only of services subscribed to the same subject with
// the same group nam will receive a message.
func (s *Server) Register(subject, group string, service Service) {
	p := pipeline{
		subject:   subject,
		group:     group,
		consumer:  newConsumer(s.nc),
		publisher: newPublisher(s.nc),
		service:   service,
	}
	s.pipelines = append(s.pipelines, p)
}
