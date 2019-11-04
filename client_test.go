package jrpc

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCall(t *testing.T) {
	req, _ := NewRequest("add", [2]int{1, 2}, NewID(0))

	transport := &transport{}
	transport.send = func(ctx context.Context, r io.Reader) error {
		buf := bufferpool.GetForBufferSlice()
		defer buf.Free()
		b := buf.Bytes()

		n, err := r.Read(b)
		assert.Nil(t, err)
		read := string(b[:n])
		assert.Equal(t, `{"jsonrpc":"2.0","method":"add","params":[1,2],"id":0}`+"\n", read)
		return nil
	}

	transport.recv = func(ctx context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error) {
		buf := bufferpool.Get()
		defer buf.Free()

		resp := Response{
			Version: "2.0",
			ID:      NewID(0),
		}
		err = resp.EncodeAndSetResult(3)
		assert.Nil(t, err)
		err = resp.encodeTo(buf)
		buf.AppendByte('\n')
		assert.Nil(t, err)
		assert.Equal(t, `{"jsonrpc":"2.0","result":3,"id":0}`+"\n", buf.String())

		return ioutil.NopCloser(buf), true, false, nil
	}

	client := NewClient(transport)

	resp, err := client.Call(context.Background(), req)
	assert.Nil(t, err)

	assert.Equal(t, "2.0", resp.Version)
	assert.Equal(t, "3", string(*resp.Result))
	assert.Equal(t, NewID(0), resp.ID)

	var result int
	err = resp.DecodeResult(&result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result)
}

type transport struct {
	send func(ctx context.Context, r io.Reader) error
	recv func(ctx context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error)
}

func (t *transport) SendRequest(ctx context.Context, r io.Reader) error { // responseExpected bool
	return t.send(ctx, r)
}

func (t *transport) ReceivedResponse(ctx context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error) {
	return t.recv(ctx)
}

func (t *transport) Close() error {
	return nil
}

func BenchmarkClient(b *testing.B) {
	transport := &transport{}
	transport.send = func(ctx context.Context, r io.Reader) error {
		buf := bufferpool.GetForBufferSlice()
		defer buf.Free()
		b := buf.Bytes()

		r.Read(b)
		//read := string(b[:n])
		return nil
	}

	transport.recv = func(ctx context.Context) (recv io.ReadCloser, updated, shouldClose bool, err error) {
		buf := bufferpool.Get()
		defer buf.Free()

		resp := Response{
			Version: "2.0",
			ID:      NewID(0),
		}
		resp.EncodeAndSetResult(3)
		resp.encodeTo(buf)
		buf.AppendByte('\n')

		return ioutil.NopCloser(buf), true, false, nil
	}

	client := NewClient(transport)

	b.Run("Call", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := NewRequest("add", [2]int{1, 2}, NewID(0))
			resp, _ := client.Call(context.Background(), req)

			var result int
			resp.DecodeResult(&result)
		}
	})

	b.Run("Do", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result int
			if client.Do(context.Background(), "add", [2]int{1, 2}, &result) != nil {
				panic("")
			}
		}
	})

	b.Run("Notify", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if client.Notify(context.Background(), "add", [2]int{1, 2}) != nil {
				panic("")
			}
		}
	})

	b.Run("NotifyWError", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if client.NotifyWithError(context.Background(), "add", [2]int{1, 2}) != nil {
				panic("")
			}
		}
	})
}

/*
func TestClientCall(t *testing.T) {
	req, _ := NewRequest("add", [2]int{1, 2}, NewID(0))

	transport := transportFunc(func(ctx context.Context, ger GetterEncodedRequest, responseExpected bool) error {
		buf := bufferpool.Get()
		defer buf.Free()

		var err error
		srr, _ := ger.GetEncodedRequest(ctx, buf)
		/*if err != nil {
			return err
		}*
		assert.Equal(t, `{"jsonrpc":"2.0","method":"add","params":[1,2],"id":0}`+"\n", buf.String())

		buf.Reset()

		resp := Response{
			Version: "2.0",
			ID:      NewID(0),
		}
		err = resp.EncodeAndSetResult(3)
		assert.Nil(t, err)
		err = resp.encodeTo(buf)
		buf.AppendByte('\n')
		assert.Nil(t, err)
		assert.Equal(t, `{"jsonrpc":"2.0","result":3,"id":0}`+"\n", buf.String())

		return srr.SetReceivedResponse(ctx, buf)
	})

	client := NewClient(transport)

	resp, err := client.Call(context.Background(), req)
	assert.Nil(t, err)

	assert.Equal(t, "2.0", resp.Version)
	assert.Equal(t, "3", string(*resp.Result))
	assert.Equal(t, NewID(0), resp.ID)

	var result int
	err = resp.DecodeResult(&result)
	assert.Nil(t, err)
	assert.Equal(t, 3, result)
	//assert.Equal(t, 3, resp.Result)
}

type transportFunc func(context.Context, GetterEncodedRequest, bool) error

func (tf transportFunc) Transport(ctx context.Context, ger GetterEncodedRequest, responseExpected bool) error {
	return tf(ctx, ger, responseExpected)
}

var _ ClientTransport = (transportFunc)(nil)
*/
