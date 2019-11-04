package adaptor

import (
	"context"
	"encoding/json"

	"github.com/daichitakahashi/jrpc"
	"github.com/daichitakahashi/jrpc/positional"
)

/*
repository.Register(adaptor.Adapt("multiply", &Multiply{}))

type(
	Multiply struct{}

	MultiplyParams struct{
		A int
		B int
	}

	MultiplyResult int
)

func (m *Multipy) ServeJRPC(ctx context.Context, paramsPtr, resultPtr interface{}) *jrpc.Error {
	params := paramsPtr.(*MultiplyParams)

	r := resultPtr.(*MultiplyResult)
	*r = p.A * p.B

	return nil
}

func (m *Multiply) NewParamsPtr() interface{} {
	return &MultiplyParams{}
}

func (m *Multiply) NewResultPtr() interface{} {
	return new(MultipyResult)
}
*/

// Handler is
type Handler interface {
	ServeJRPC(ctx context.Context, paramsPtr, resultPtr interface{}) *jrpc.Error
	NewParamsPtr() interface{}
	NewResultPtr() interface{}
}

// Adapt is
func Adapt(method string, handler Handler) (m string, h jrpc.Handler, p, r interface{}) {
	m = method
	af := &adaptorFunc{Handler: handler}
	p = handler.NewParamsPtr()
	r = handler.NewResultPtr()
	_, af.positional = p.(positional.Unmarshaler)
	h = af
	return
}

type adaptorFunc struct {
	Handler
	positional bool
}

func (af *adaptorFunc) ServeJSONRPC(ctx context.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
	paramsPtr := af.NewParamsPtr()
	if params != nil && paramsPtr != nil {
		var err error
		if af.positional {
			err = positional.Unmarshal(*params, paramsPtr)
		} else {
			err = json.Unmarshal(*params, paramsPtr)
		}
		if err != nil {
			return nil, jrpc.ErrInvalidParams()
		}
	}
	resultPtr := af.NewResultPtr()

	jrpcErr := af.ServeJRPC(ctx, paramsPtr, resultPtr)
	if jrpcErr != nil {
		return nil, jrpcErr
	}
	return resultPtr, nil
}
