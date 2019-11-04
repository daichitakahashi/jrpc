package jrpc

import (
	"context"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

/*
TODO:

*/

// Client is
type Client struct {
	Transport ClientTransport
	dec       *json.Decoder
	options   clientOptions
}

// NewClient is
func NewClient(t ClientTransport, opts ...ClientOption) *Client {
	client := &Client{
		Transport: t,
		options:   defaultClientOption,
	}
	for _, opt := range opts {
		opt.apply(&client.options)
	}
	return client
}

// ClientTransport is
type ClientTransport interface {
	SendRequest(ctx context.Context, r io.Reader) error
	ReceivedResponse(ctx context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error)
	Close() error
}

// Call is
// レスポンスを待つかどうか、ClientTransportの設定次第
// タイムアウトは、http.Clientのものも使用できるし、contextのDeadlineでも
func (c *Client) Call(ctx context.Context, req *Request) (*Response, error) {
	buf, err := req.encodeLine()
	if err != nil {
		return nil, err
	}
	var resp Response
	err = c.call(ctx, buf, &resp, nil)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// CallBatch is
// レスポンスを待つかどうか、ClientTransportの設定次第
func (c *Client) CallBatch(ctx context.Context, reqs BatchRequest) (BatchResponse, error) {
	buf, err := reqs.encodeLine()
	if err != nil {
		return nil, err
	}

	var resp Response
	resps := make(BatchResponse, 0, defaultBatchCapacity)
	err = c.call(ctx, buf, &resp, resps)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		resps = append(resps[:0], &resp)
	}
	return resps, nil
}

// Do is shorthand, use IDFactory
// レスポンスを待つかどうか、ClientTransportの設定次第
func (c *Client) Do(ctx context.Context, method string, params, result interface{}) error {
	req, err := NewRequest(method, params, c.options.idFactory.CreateID())
	if err != nil {
		return err
	}
	buf, err := req.encodeLine()
	if err != nil {
		return err
	}

	var resp Response
	err = c.call(ctx, buf, &resp, nil)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	} else if result == nil || resp.Result == nil {
		return nil
	}
	// if Result is 'null', try to decode nil object.
	return resp.DecodeResult(result)
}

/*
var subtractResult int
var subtractErr *jrpc.Error
var notifyErr *jrpc.Error

client.Batch(ctx, func(register jrpc.RegisterFunc){
	register("subtract", []int{1, 2}, &subtractResult, &subtractErr)
	register("notify", []int{1, 2, 3, 4, 5}, nil, &notifyErr)
	register("xxxx", nil, nil, nil)
}, nil)
*/

// Batch is
// result is nil -> notify
// jrpcErr is nil -> mere ignorant
func (c *Client) Batch(ctx context.Context, registerer func(RegisterFunc), unknownErrs *[]*Error) error {
	rr := requestRegisterer{
		batchReq:  make(BatchRequest, 0, defaultBatchCapacity),
		targets:   make(map[ID]responseTarget),
		idFactory: c.options.idFactory.BatchIDFactory(),
	}
	registerer(rr.Register)
	if len(rr.batchReq) == 0 {
		return errors.New("no request")
	}

	batchResp, err := c.CallBatch(ctx, rr.batchReq)
	if err != nil {
		return err
	}

	for _, resp := range batchResp {
		if resp.ID == NoID { // ignore
			continue
		} else if resp.ID == UnknownID { // unexpected
			if resp.Error != nil && unknownErrs != nil { // notification error?
				*unknownErrs = append(*unknownErrs, resp.Error)
			} else { // truly unknown
				continue
			}
		}
		target, ok := rr.targets[resp.ID]
		if !ok {
			// if you connect to JSON-RPC server that has possibility of modify ID of request,
			// use Client.CallBatch.
			continue
		}
		if target.result != nil {
			err = json.Unmarshal(*resp.Result, target.result)
			if err != nil {
				return err
			}
		}
		if target.errPtr != nil {
			*target.errPtr = resp.Error
		}
	}
	return nil
}

type (
	// RegisterFunc is
	RegisterFunc func(method string, params, result interface{}, jrpcErr **Error) error

	requestRegisterer struct {
		batchReq  BatchRequest
		targets   map[ID]responseTarget
		idFactory IDFactory
	}

	responseTarget struct {
		result interface{}
		errPtr **Error
	}
)

func (rr *requestRegisterer) Register(method string, params, result interface{}, jrpcErr **Error) error {
	var id ID
	if result != nil {
		id = rr.idFactory.CreateID()
		rr.targets[id] = responseTarget{
			result: result,
			errPtr: jrpcErr,
		}
	}
	req, err := NewRequest(method, params, id)
	if err != nil {
		return err
	}
	rr.batchReq.Add(req)
	return nil
}

// Notify is
// 戻り値について同様
// レスポンス可能性を完全に無視する（何か返ってきていても無視する）
func (c *Client) Notify(ctx context.Context, method string, params interface{}) error {
	req, err := NewRequest(method, params, NoID)
	if err != nil {
		return err
	}
	buf, err := req.encodeLine()
	if err != nil {
		return err
	}
	return c.call(ctx, buf, nil, nil)
}

// NotifyWithError is
// レスポンス可能性を積極的に拾いに行くが、最終的にはClientTransportの設定次第
func (c *Client) NotifyWithError(ctx context.Context, method string, params interface{}) error {
	req, err := NewRequest(method, params, NoID)
	if err != nil {
		return err
	}
	buf, err := req.encodeLine()
	if err != nil {
		return err
	}
	var resp Response
	err = c.call(ctx, buf, &resp, nil)
	if err != nil {
		return err
	} else if resp.Error != nil {
		return resp.Error
	}
	// invalid server implementation
	return nil
}

func (c *Client) call(ctx context.Context, r io.Reader, resp *Response, batchResp BatchResponse) error {
	err := c.Transport.SendRequest(ctx, r)
	if err != nil {
		return nil
	}

	responseReader, updated, shouldClose, err := c.Transport.ReceivedResponse(ctx)
	if err != nil {
		return err
	} else if responseReader == nil {
		return nil
	}
	defer func() {
		if shouldClose {
			responseReader.Close()
		}
	}()

	if resp == nil && batchResp == nil {
		return nil
	} else if updated || (c.dec == nil) {
		c.dec = json.NewDecoder(responseReader)
	}

	switch {
	case batchResp != nil:
		var raw json.RawMessage
		err = c.dec.Decode(raw)
		if err != nil {
			return err
		}
		switch raw[0] {
		case '[':
			return json.Unmarshal(raw, batchResp)
		case '{':
			return json.Unmarshal(raw, resp)
		default:
			return errors.New("invalid character found")
		}
	default:
		return c.dec.Decode(resp)
	}
}

// Close closing connection
func (c *Client) Close() error {
	return c.Transport.Close()
}
