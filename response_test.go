package jrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResponse_DecodeResult(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		result := json.RawMessage(`-99`)
		resp := &Response{
			Result: &result,
		}
		var i int
		err := resp.DecodeResult(&i)
		require.NoError(t, err)
		require.Equal(t, -99, i)
	})

	t.Run("error", func(t *testing.T) {
		resp := &Response{}
		var i int
		err := resp.DecodeResult(&i)
		require.Equal(t, ErrNilResult, err)
		require.Zero(t, i)
	})
}

func TestResponse_MarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Run("no result", func(t *testing.T) {
			resp := &Response{
				Version: "2.0",
				Result:  nil,
				ID:      NewID(0),
			}
			b, err := resp.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, `{"jsonrpc":"2.0","result":null,"id":0}`, string(b))
		})

		result := json.RawMessage(`"text"`)

		t.Run("with result", func(t *testing.T) {

			resp := &Response{
				Version: "2.0",
				Result:  &result,
				ID:      NewID(123),
			}
			b, err := resp.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, `{"jsonrpc":"2.0","result":"text","id":123}`, string(b))
		})

		t.Run("error with result(ignored)", func(t *testing.T) {
			resp := &Response{
				Version: "2.0",
				Result:  &result,
				Error: &Error{
					Code:    -1,
					Message: "dummy",
				},
				ID: NewID(987),
			}
			b, err := resp.MarshalJSON()
			require.NoError(t, err)
			require.Equal(t, `{"jsonrpc":"2.0","error":{"code":-1,"message":"dummy"},"id":987}`, string(b))
		})
	})

	t.Run("error", func(t *testing.T) {
		id := ID{
			nType: 16,
		}
		resp := &Response{
			Version: "2.0",
			ID:      id,
		}
		b, err := resp.MarshalJSON()
		require.Error(t, err)
		require.Nil(t, b)
	})

	t.Run("encodeLine", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			resp := &Response{
				Version: "2.0",
				Result:  nil,
				ID:      NewID(0),
			}
			buf, err := resp.encodeLine()
			require.NoError(t, err)
			require.Equal(t, `{"jsonrpc":"2.0","result":null,"id":0}`+"\n", buf.String())
		})

		t.Run("error", func(t *testing.T) {
			id := ID{
				nType: 16,
			}
			resp := &Response{
				Version: "2.0",
				ID:      id,
			}
			buf, err := resp.encodeLine()
			require.Error(t, err)
			require.Nil(t, buf)
		})
	})
}
