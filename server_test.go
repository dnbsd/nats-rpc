package gonatsrpc

import (
	"context"
	"fmt"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type echoService struct{}

func (s *echoService) Start(ctx context.Context, reqCh <-chan *nats.Msg, respCh chan<- *nats.Msg) error {
	sendResponse := func(msg *nats.Msg) {
		select {
		case respCh <- msg:
		case <-ctx.Done():
			return
		}
	}

	for {
		select {
		case msg := <-reqCh:
			msg = &nats.Msg{
				Subject: msg.Reply,
				Data:    msg.Data,
			}
			sendResponse(msg)

		case <-ctx.Done():
			return nil
		}
	}
}

func newNatsServerAndConnection(t *testing.T) *nats.Conn {
	opts := natsserver.DefaultTestOptions
	opts.NoLog = false
	opts.Port = 14444
	s := natsserver.RunServer(&opts)
	uri := fmt.Sprintf("nats://%s:%d", opts.Host, opts.Port)
	nc, err := nats.Connect(uri)
	assert.NoError(t, err)
	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})
	return nc
}

func TestServer_Service(t *testing.T) {
	const rpcSubject = "test_service.rpc"
	nc := newNatsServerAndConnection(t)
	s := NewServer(nc)
	s.Register(rpcSubject, &echoService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := s.StartWithContext(ctx)
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	reqMsg := &nats.Msg{
		Subject: rpcSubject,
		Reply:   nats.NewInbox(),
		Data:    []byte("hello world!"),
	}
	respMsg, err := nc.RequestMsg(reqMsg, 5*time.Second)
	assert.NoError(t, err)
	assert.Equal(t, reqMsg.Data, respMsg.Data)
}
