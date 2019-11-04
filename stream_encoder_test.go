package jrpc

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoder_Encode(t *testing.T) {
	r, w := io.Pipe()
	enc := NewEncoder(nil)

	t.Run("unary", func(t *testing.T) {
		enc.Reset(w)
		resp := &Response{
			Version: "2.0",
			ID:      NewID(99.99),
		}
		resp.EncodeAndSetResult(true)

		go func() {
			err := enc.Encode([]*Response{resp}, false)
			require.NoError(t, err)
		}()

		b := make([]byte, 1000)
		n, err := r.Read(b)
		if err != nil {
			panic(err)
		}
		require.Equal(t, `{"jsonrpc":"2.0","result":true,"id":99.99}`+"\n", string(b[:n]))
	})

	t.Run("batch(but len(resp)==1)", func(t *testing.T) {
		enc.Reset(w)
		resp := &Response{
			Version: "2.0",
			ID:      NewID(99.99),
		}
		resp.EncodeAndSetResult(true)

		go func() {
			err := enc.Encode([]*Response{resp}, true)
			require.NoError(t, err)
		}()

		b := make([]byte, 1000)
		n, err := r.Read(b)
		if err != nil {
			panic(err)
		}
		require.Equal(t, `[{"jsonrpc":"2.0","result":true,"id":99.99}]`+"\n", string(b[:n]))
	})

	t.Run("batch", func(t *testing.T) {
		enc.Reset(w)

		resp1 := &Response{
			Version: "2.0",
			Error:   ErrInvalidParams(),
			ID:      NewID("111"),
		}
		resp2 := &Response{
			Version: "2.0",
			ID:      NewID("222"),
		}
		resp2.EncodeAndSetResult("result sentence")
		resp3 := &Response{
			Version: "2.0",
			ID:      NewID("333"),
		}
		resp3.EncodeAndSetResult(3333333333)

		resps := []*Response{resp1, resp2, resp3}

		go func() {
			err := enc.Encode(resps, true)
			require.NoError(t, err)
		}()

		b := make([]byte, 1000)
		n, err := r.Read(b)
		if err != nil {
			panic(err)
		}
		require.Equal(t, `[{"jsonrpc":"2.0","error":{"code":-32602,"message":"Invalid params"},"id":"111"},`+
			`{"jsonrpc":"2.0","result":"result sentence","id":"222"},`+
			`{"jsonrpc":"2.0","result":3333333333,"id":"333"}]`+"\n",
			string(b[:n]))
	})

	t.Run("empty", func(t *testing.T) {
		enc.Reset(w)
		err := enc.Encode([]*Response{}, true)
		require.NoError(t, err)
	})

	t.Run("encode error", func(t *testing.T) {
		enc.Reset(w)
		resp := &Response{
			Version: "2.0",
			ID:      ID{nType: 99},
		}

		err := enc.Encode([]*Response{resp}, false)
		require.Error(t, err)
	})

	enc = NewEncoderSize(w, 20)

	t.Run("interrupt", func(t *testing.T) {
		resp1 := &Response{
			Version: "2.0",
			Error:   ErrInvalidParams(),
			ID:      NewID("111"),
		}
		resp2 := &Response{
			Version: "2.0",
			ID:      NewID("222"),
		}
		resp2.EncodeAndSetResult("result sentence")
		resp3 := &Response{
			Version: "2.0",
			ID:      NewID("333"),
		}
		resp3.EncodeAndSetResult(3333333333)

		resps := []*Response{resp1, resp2, resp3}

		r.Close()

		err := enc.Encode(resps, true)
		require.Error(t, err)
	})
}

func TestEncoder_EncodeContext(t *testing.T) {
	enc := NewEncoder(bufferpool.Get())

	resp := &Response{
		Version: "2.0",
		ID:      NewID(99.99),
	}
	resp.EncodeAndSetResult(true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := enc.EncodeContext(ctx, []*Response{resp}, false)
	require.Equal(t, ctx.Err(), err)

}
