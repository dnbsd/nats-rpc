package natsrpc

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type echoReceiver struct{}

type EchoParams struct {
	Message string
}

type EchoResult struct {
	Message string
}

func (s *echoReceiver) Echo(params *EchoParams) (*EchoResult, error) {
	println("echo:", params.Message)
	return &EchoResult{
		Message: params.Message,
	}, nil
}

func (s *echoReceiver) Echo2(params EchoParams) (EchoResult, error) {
	println("echo:", params.Message)
	return EchoResult{
		Message: params.Message,
	}, nil
}

func (s *echoReceiver) Echos(params string) (string, error) {
	println("echo:", params)
	return params, nil
}

func TestMapper_CallEcho(t *testing.T) {
	const message = "hello world!"
	mapper := NewMapper()
	mapper.Register("Echo", &echoReceiver{})
	params, err := mapper.Params("Echo", "Echo")
	assert.NoError(t, err)
	params.Elem().Set(reflect.ValueOf(&EchoParams{Message: message}))
	result, err := mapper.Call("Echo", "Echo", params.Elem())
	assert.NoError(t, err)
	resultMessage := result.Interface().(*EchoResult)
	assert.Equal(t, message, resultMessage.Message)
}

func TestMapper_CallEcho2(t *testing.T) {
	const message = "hello world!"
	mapper := NewMapper()
	mapper.Register("Echo", &echoReceiver{})
	params, err := mapper.Params("Echo", "Echo2")
	assert.NoError(t, err)
	params.Elem().Set(reflect.ValueOf(EchoParams{Message: message}))
	result, err := mapper.Call("Echo", "Echo2", params.Elem())
	assert.NoError(t, err)
	resultMessage := result.Interface().(EchoResult)
	assert.Equal(t, message, resultMessage.Message)
}

func TestMapper_CallEchos(t *testing.T) {
	mapper := NewMapper()
	mapper.Register("Echo", &echoReceiver{})
	paramsStr := "hello world!"
	params, err := mapper.Params("Echo", "Echos")
	assert.NoError(t, err)
	params.Elem().Set(reflect.ValueOf(paramsStr))
	result, err := mapper.Call("Echo", "Echos", params.Elem())
	assert.NoError(t, err)
	resultMessage := result.Interface().(string)
	assert.Equal(t, paramsStr, resultMessage)
}

func TestMapper_CallUndefinedMethod(t *testing.T) {
	mapper := NewMapper()
	mapper.Register("Echo", &echoReceiver{})
	params := reflect.ValueOf(nil)
	_, err := mapper.Call("Echo", "Undefined", params)
	assert.ErrorIs(t, err, ErrMethodNotFound)
}

func TestMapper_CallUnregisteredReceiver(t *testing.T) {
	mapper := NewMapper()
	params := reflect.ValueOf(nil)
	_, err := mapper.Call("Undefined", "Undefined", params)
	assert.ErrorIs(t, err, ErrReceiverNotFound)
}
