package jrpcmock

/*
import (
	"encoding/json"
	"time"

	"github.com/daichitakahashi/httpjrpc/jrpc"
	"github.com/daichitakahashi/httpjrpc/jrpc/jrpcmock/notify"
	"github.com/daichitakahashi/httpjrpc/jrpc/jrpcmock/subtract"
	"github.com/daichitakahashi/httpjrpc/jrpc/jrpcmock/sum"
	"github.com/pkg/errors"
)

// Repository returns mock
func Repository() *jrpc.RootRepository {
	r := jrpc.NewRootRepository(
		jrpc.NewCore())

	r.Assign("", subtract.Package())
	r.Assign("", sum.Package())
	r.Assign("notify", notify.Package())

	return r.With(waiter).(*jrpc.RootRepository)
}

func waiter(next jrpc.InterceptorHandler) jrpc.InterceptorHandler {
	return jrpc.InterceptorFunc(func(c jrpc.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
		req := c.Value(jrpc.RequestKey{}).(*jrpc.Request)

		var id interface{}
		err := jrpc.Unmarshal(req.RawID, &id)
		if err != nil {
			panic(errors.Wrap(err, "invalid request ID"))
		}

		if id != nil {
			d, ok := id.(time.Duration)
			if !ok {
				panic(errors.Wrap(err, "mock's request ID must be integer"))
			}
			time.Sleep(time.Millisecond * 10 * d)
		}

		return next.Intercept(c, params)

		/*
			var n time.Duration
			err := jrpc.Unmarshal(req.ID, &n)
			if err != nil {
				panic(errors.Wrap(err, "mock's request ID must be integer"))
			}

			time.Sleep(time.Millisecond * 10 * n)
			return next.Intercept(c, params)
		*
	})
}
*/
