package jrpc

import (
	"context"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

/*
TODO:

*/

// Execute is
func (c *Core) Execute(ctx context.Context, requests []*Request, batch bool) ([]*Response, error) {
	if len(requests) == 0 {
		resp := (*Request)(nil).toResponse()
		resp.Error = ErrInvalidRequest(nil)
		resp.ID = UnknownID
		return []*Response{
			resp,
		}, nil
	}

	resps := make([]*Response, 0, len(requests))

	if len(requests) == 1 {
		resp, err := c.DoMethod(ctx, requests[0], batch)
		if err != nil {
			return nil, err
		} else if resp.isSend() {
			resps = append(resps, resp)
		}
		return resps, nil
	} else if c.options.disableConcurrentCall {
		for _, req := range requests {
			resp, err := c.DoMethod(ctx, req, batch)
			if err != nil {
				return nil, err
			} else if resp.isSend() {
				resps = append(resps, resp)
			}
		}
		return resps, nil
	}

	eg, ctxBatch := errgroup.WithContext(ctx)
	m := sync.Mutex{}

	for _, req := range requests {
		req := req
		eg.Go(func() error {
			resp, err := c.DoMethod(ctxBatch, req, batch)
			if err != nil {
				return err
			} else if resp.isSend() {
				m.Lock()
				resps = append(resps, resp)
				m.Unlock()
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	return resps, nil
}

// DoMethod is
func (c *Core) DoMethod(ctx context.Context, req *Request, batch bool) (resp *Response, err error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if req.Version != "2.0" {
		resp := req.toResponse()
		resp.Error = ErrInvalidRequest(nil) // 独自のエラー入れる？ unsupported version of JSON-RPC
		return resp, nil
	} else if strings.HasPrefix(req.Method, "rpc.") {
		return callRPCInternal(req), nil
	}

	resp = req.toResponse()
	v, ok := c.methods.Load(req.Method)
	if !ok {
		resp.Error = ErrMethodNotFound()
		return resp, nil
	}
	md := v.(*Metadata)

	reqInfo := RequestInfo{
		MethodFullName: req.Method,
		IsBatch:        batch,
		ID:             req.ID,
	}

	defer func() {
		rvr := recover()
		if rvr != nil {
			if c.options.panicHandler != nil {
				c.options.panicHandler(req, rvr)
			}
			resp = callRPCInternal(&Request{
				Method: rpcInternalError,
				err: &RecoveredError{
					Request:   req,
					Recovered: rvr,
				},
			})
		}
	}()

	var result interface{}
	result, resp.Error = md.InterceptorChain(ctx, req.Params, &reqInfo, md.Handler)
	if resp.Error == nil {
		err := resp.EncodeAndSetResult(result)
		if err != nil {
			// この段階でエンコードエラーが出る
			// ということは、レスポンスのエンコード時には、ストリームエラーしか発生しえない
			resp.Error = ErrInternal(err)
		}
	}
	return resp, nil
}

const (
	rpcInternalError  = "rpc.internalError"
	rpcParseError     = "rpc.parseError"
	rpcInvalidRequest = "rpc.invalidRequest"
)

func callRPCInternal(req *Request) *Response {
	resp := req.toResponse()

	switch req.Method {
	case rpcInternalError:
		resp.Error = ErrInternal(req.err)
	case rpcParseError:
		resp.Error = ErrParse(req.err)
	case rpcInvalidRequest:
		resp.Error = ErrInvalidRequest(req.err)
	default:
		resp.Error = ErrMethodNotFound()
		// errors.New("Method names that begin with the word 'rpc.' are reserved for rpc-internal methods")
	}
	if resp.ID == NoID {
		resp.ID = UnknownID
	}
	return resp
}
