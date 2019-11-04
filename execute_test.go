package jrpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	repository := newMock()

	t.Run("0 request", func(t *testing.T) {
		resps, err := repository.Execute(context.Background(), nil, false)
		require.Nil(t, err)
		require.Equal(t, 1, len(resps))
		require.Equal(t, ErrorCodeInvalidRequest, resps[0].Error.Code)
	})

	t.Run("1 request", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			reqs := []*Request{
				&Request{
					Version: "2.0",
					Method:  "sum",
					ID:      NewID(0),
				},
			}
			reqs[0].EncodeAndSetParams([]int{1, 2, 3})
			resps, err := repository.Execute(context.Background(), reqs, false)
			require.Nil(t, err)
			require.Equal(t, 1, len(resps))
			require.Nil(t, resps[0].Error)
			var result int
			resps[0].DecodeResult(&result)
			require.Equal(t, 6, result)
		})

		t.Run("error(ctx.Cancelled)", func(t *testing.T) {
			reqs := []*Request{
				&Request{},
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			resps, err := repository.Execute(ctx, reqs, false)
			require.Nil(t, resps)
			require.Equal(t, ctx.Err(), err)
		})

		t.Run("notify", func(t *testing.T) {
			reqs := []*Request{
				&Request{
					Version: "2.0",
					Method:  "sum",
				},
			}
			reqs[0].EncodeAndSetParams([]int{1, 2, 3})
			resps, err := repository.Execute(context.Background(), reqs, false)
			require.Nil(t, err)
			require.NotNil(t, resps)
			require.Equal(t, 0, len(resps))
		})
	})

	batchReqs := []*Request{
		&Request{
			Version: "2.0",
			Method:  "wait.sayHello",
			ID:      NewID(70),
		},
		&Request{
			Version: "2.0",
			Method:  "wait.sayHello",
			ID:      NewID(20),
		},
		&Request{
			Version: "2.0",
			Method:  "wait.sayHello",
			ID:      NewID(50),
		},
		&Request{
			Version: "2.0",
			Method:  "wait.sayHello",
			ID:      NewID(40),
		},
		&Request{
			Version: "2.0",
			Method:  "wait.sayHello",
			ID:      NewID(80),
		},
		&Request{
			Version: "2.0",
			Method:  "err.panicMaker",
		},
	}

	t.Run("batch(concurrent)", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			resps, err := repository.Execute(context.Background(), batchReqs, true)
			require.Nil(t, err)
			require.Equal(t, 6, len(resps))
			for i, resp := range resps {
				switch i {
				case 0:
					require.Equal(t, UnknownID, resp.ID)
					require.Equal(t, ErrorCodeInternal, resp.Error.Code)
				case 1:
					require.Equal(t, NewID(20), resp.ID)
				case 2:
					require.Equal(t, NewID(40), resp.ID)
				case 3:
					require.Equal(t, NewID(50), resp.ID)
				case 4:
					require.Equal(t, NewID(70), resp.ID)
				case 5:
					require.Equal(t, NewID(80), resp.ID)
				}
			}
		})

		t.Run("ctx.Cancelled(already)", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			resps, err := repository.Execute(ctx, batchReqs, true)
			<-ctx.Done()
			require.Equal(t, err, ctx.Err())
			require.Nil(t, resps)
		})

		t.Run("ctx.Cancelled(too lazy)", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*40)
			defer cancel()
			resps, err := repository.Execute(ctx, batchReqs, true)
			<-ctx.Done()
			require.NotEqual(t, err, ctx.Err())
			require.NotNil(t, resps)
		})
	})

	t.Run("batch(disable concurrent)", func(t *testing.T) {
		repository.options.disableConcurrentCall = true
		t.Run("success", func(t *testing.T) {
			resps, err := repository.Execute(context.Background(), batchReqs, true)
			require.Nil(t, err)
			require.Equal(t, 6, len(resps))
			for i, resp := range resps {
				switch i {
				case 0:
					require.Equal(t, NewID(70), resp.ID)
				case 1:
					require.Equal(t, NewID(20), resp.ID)
				case 2:
					require.Equal(t, NewID(50), resp.ID)
				case 3:
					require.Equal(t, NewID(40), resp.ID)
				case 4:
					require.Equal(t, NewID(80), resp.ID)
				case 5:
					require.Equal(t, UnknownID, resp.ID)
					require.Equal(t, ErrorCodeInternal, resp.Error.Code)
				}
			}
		})

		t.Run("ctx.Cancelled", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*40)
			defer cancel()
			resps, err := repository.Execute(ctx, batchReqs, true)
			<-ctx.Done()
			require.Equal(t, err, ctx.Err())
			require.Nil(t, resps)
		})
	})
}

func TestDoMethod(t *testing.T) {
	repository := newMock()

	t.Run("ctx cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		resp, err := repository.DoMethod(ctx, nil, false)
		require.Nil(t, resp)
		require.Equal(t, ctx.Err(), err)
	})

	t.Run("not 2.0", func(t *testing.T) {
		req := &Request{
			Version: "1.1",
		}
		resp, _ := repository.DoMethod(context.Background(), req, false)
		require.Nil(t, resp.Result)
		require.Equal(t, ErrorCodeInvalidRequest, resp.Error.Code)
	})

	t.Run("rpc internal", func(t *testing.T) {
		req := &Request{
			Version: "2.0",
			Method:  rpcInternalError,
		}
		resp, _ := repository.DoMethod(context.Background(), req, false)
		require.Nil(t, resp.Result)
		require.Equal(t, ErrorCodeInternal, resp.Error.Code)
	})

	t.Run("method not found", func(t *testing.T) {
		req := &Request{
			Version: "2.0",
			Method:  "pow",
		}
		resp, _ := repository.DoMethod(context.Background(), req, false)
		require.Nil(t, resp.Result)
		require.Equal(t, ErrorCodeMethodNotFound, resp.Error.Code)
	})

	t.Run("panic occur", func(t *testing.T) {
		req := &Request{
			Version: "2.0",
			Method:  "err.panicMaker",
		}
		req.EncodeAndSetParams("panicsentence")
		var called bool
		repository.options.panicHandler = func(_req *Request, _rvr interface{}) {
			require.Equal(t, *req, *_req)
			require.Equal(t, "panicsentence", _rvr)
			called = true
		}
		resp, _ := repository.DoMethod(context.Background(), req, false)
		require.True(t, called)
		require.Nil(t, resp.Result)
		rvr, ok := resp.Error.Cause().(*RecoveredError)
		require.True(t, ok)
		require.Equal(t, *req, *rvr.Request)
		require.Equal(t, "panicsentence", rvr.Recovered)
	})

	t.Run("encode error", func(t *testing.T) {
		req := &Request{
			Version: "2.0",
			Method:  "err.encodeError",
		}
		resp, _ := repository.DoMethod(context.Background(), req, false)
		require.Nil(t, resp.Result)
		require.Equal(t, ErrorCodeInternal, resp.Error.Code)
	})

	t.Run("success", func(t *testing.T) {
		req := &Request{
			Version: "2.0",
			Method:  "subtract",
		}
		req.EncodeAndSetParams([]int{23, 42})
		resp, _ := repository.DoMethod(context.Background(), req, false)
		var result int
		resp.DecodeResult(&result)
		require.Equal(t, -19, result)
	})

	t.Run("error", func(t *testing.T) {
		req := &Request{
			Version: "2.0",
			Method:  "err.maker",
		}
		resp, _ := repository.DoMethod(context.Background(), req, false)
		require.Nil(t, resp.Result)
		require.NotNil(t, resp.Error)
	})
}

func TestCallRPCInternal(t *testing.T) {

	type TestCase struct {
		desc          string
		requestMethod string
		errorCode     ErrorCode
	}
	testcases := []TestCase{
		{
			desc:          "internal error",
			requestMethod: rpcInternalError,
			errorCode:     ErrorCodeInternal,
		}, {
			desc:          "parse error",
			requestMethod: rpcParseError,
			errorCode:     ErrorCodeParse,
		}, {
			desc:          "invalid request",
			requestMethod: rpcInvalidRequest,
			errorCode:     ErrorCodeInvalidRequest,
		}, {
			desc:          "invalid rpc internal",
			requestMethod: "rpc.unknown",
			errorCode:     ErrorCodeMethodNotFound,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			resp := callRPCInternal(&Request{
				Version: "2.0",
				Method:  testcase.requestMethod,
			})
			require.Equal(t, "2.0", resp.Version)
			require.Equal(t, testcase.errorCode, resp.Error.Code)
		})
	}
}
