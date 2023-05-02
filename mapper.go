package natsrpc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	ErrReceiverNotFound = errors.New("receiver not found")
	ErrMethodNotFound   = errors.New("method not found")
)

var (
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

type receiverMethod struct {
	receiverValue reflect.Value
	method        reflect.Method
	paramsType    reflect.Type
	resultType    reflect.Type
}

type Mapper struct {
	receivers map[string]map[string]receiverMethod
	sync.Mutex
}

func NewMapper() *Mapper {
	return &Mapper{}
}

func (m *Mapper) Register(name string, receiver any) {
	m.Lock()
	if m.receivers == nil {
		m.receivers = make(map[string]map[string]receiverMethod)
	}

	if m.receivers[name] == nil {
		m.receivers[name] = make(map[string]receiverMethod)
	}
	m.Unlock()

	receiverType := reflect.TypeOf(receiver)

	for i := 0; i < receiverType.NumMethod(); i++ {
		method := receiverType.Method(i)
		methodType := method.Type

		if !method.IsExported() {
			continue
		}

		if methodType.NumIn() != 2 {
			continue
		}

		paramsType := methodType.In(1)
		if !m.isExportedOrBuiltin(paramsType) {
			continue
		}

		if methodType.NumOut() != 2 {
			continue
		}

		resultReturnType := methodType.Out(0)
		if !m.isExportedOrBuiltin(resultReturnType) {
			continue
		}

		errReturnType := methodType.Out(1)
		if errReturnType != typeOfError {
			continue
		}

		m.Lock()
		m.receivers[name][method.Name] = receiverMethod{
			receiverValue: reflect.ValueOf(receiver),
			method:        method,
			paramsType:    paramsType,
			resultType:    resultReturnType,
		}
		m.Unlock()
	}
}

func (m *Mapper) IsDefined(receiver, method string) bool {
	_, err := m.get(receiver, method)
	return err == nil
}

func (m *Mapper) Call(receiver, method string, params reflect.Value) (reflect.Value, error) {
	mm, err := m.get(receiver, method)
	if err != nil {
		return reflect.Value{}, err
	}

	if !params.IsValid() {
		err := fmt.Errorf("%s.%s received invalid parameters", receiver, method)
		panic(err)
	}

	paramsType := params.Type()
	if !paramsType.AssignableTo(mm.paramsType) {
		err := fmt.Errorf("%s.%s expects parameter to be of type %s but type %s was passed instead",
			receiver, method, mm.paramsType, paramsType)
		panic(err)
	}

	returnValue := mm.method.Func.Call([]reflect.Value{
		mm.receiverValue,
		params,
	})
	if v := returnValue[1].Interface(); v != nil {
		return reflect.Value{}, v.(error)
	}

	return returnValue[0], nil
}

func (m *Mapper) Params(receiver, method string) (reflect.Value, error) {
	mm, err := m.get(receiver, method)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.New(mm.paramsType), nil
}

func (m *Mapper) Result(receiver, method string) (reflect.Value, error) {
	mm, err := m.get(receiver, method)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.New(mm.resultType), nil
}

func (m *Mapper) get(receiver, method string) (receiverMethod, error) {
	m.Lock()
	defer m.Unlock()
	_, ok := m.receivers[receiver]
	if !ok {
		return receiverMethod{}, ErrReceiverNotFound
	}

	mm, ok := m.receivers[receiver][method]
	if !ok {
		return receiverMethod{}, ErrMethodNotFound
	}

	return mm, nil
}

// isExported returns true of a string is an exported (upper case) name.
func (m *Mapper) isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

// isExportedOrBuiltin returns true if a type is exported or a builtin.
func (m *Mapper) isExportedOrBuiltin(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return m.isExported(t.Name()) || t.PkgPath() == ""
}
