package httpjrpc

import (
	"context"
	"io"
	"net/http"

	"github.com/daichitakahashi/jrpc"
	"github.com/pkg/errors"
)

/*
TODO:
- bufferpool
*/

// Transport is
type Transport struct {
	url    string
	client http.Client
	resp   *http.Response
}

// SendRequest is
func (ht *Transport) SendRequest(ctx context.Context, r io.Reader) error {

	req, err := http.NewRequest("POST", ht.url, r)
	/*for _, ri := range ht.options.requestInterceptors {
		ri(req)
	}*/
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	ht.resp, err = ht.client.Do(req)
	if err != nil {
		return err
	}
	return nil

}

// ReceivedResponse is
func (ht *Transport) ReceivedResponse(_ context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error) {
	if ht.resp.StatusCode != http.StatusOK {
		return nil, false, false, errors.Errorf("HTTPTransport: status code %d", ht.resp.StatusCode)
	}
	if ht.resp.ContentLength == 0 {
		return nil, false, false, nil
	}
	return ht.resp.Body, true, true, nil
}

// Close is no-op
func (ht *Transport) Close() error {
	return nil
}

var _ jrpc.ClientTransport = (*Transport)(nil)

/*

// HTTPTransport is
type HTTPTransport struct {
	client  *http.Client
	url     string
	options *httpClientOptions
}

// NewHTTPTransport is
func NewHTTPTransport(url string, opts ...HTTPClientOption) *HTTPTransport {
	transport := defaultHTTPTransport(url)
	for _, opt := range opts {
		opt.applyHTTP(transport.options)
	}
	return transport
}

// NewHTTPClient is
func NewHTTPClient(url string, opts ...jrpc.ClientOption) *jrpc.Client {
	transport := defaultHTTPTransport(url)
	for _, opt := range opts {
		if hco, ok := opt.(*HTTPClientOption); ok {
			hco.applyHTTP(transport.options) //
		}
	}
	return jrpc.NewClient(transport, opts...)
}

func defaultHTTPTransport(url string) *HTTPTransport {
	transport := &HTTPTransport{
		client: http.DefaultClient,
		url:    url,
	}
	transport.options = defaultHTTPClientOptions(&transport.client)
	return transport
}

// Transport is
func (ht *HTTPTransport) Transport(ctx context.Context, t jrpc.Transmitter) error {
	buf := jrpc.Pool{}.Get() // pseudo
	defer buf.Free()

	r, err := t.Transmit(ctx, buf)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", ht.url, buf)
	for _, ri := range ht.options.requestInterceptors {
		ri(req)
	}
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(buf.Len())
	req = req.WithContext(ctx)

	resp, err := ht.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New("HTTPTransport: status code ") // =============================================
	}

	err = r.Receive(ctx, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

var _ jrpc.ClientTransport = (*HTTPTransport)(nil)

*/
