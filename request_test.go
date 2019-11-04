package jrpc

import (
	"bytes"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRequest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		params := []int{1, 2, 3, 4, 5, 6}
		req, err := NewRequest("success.method", params, NewID(99))

		require.NoError(t, err)
		require.Equal(t, "2.0", req.Version)
		require.Equal(t, "success.method", req.Method)
		encodedParams := []byte(*req.Params)
		require.True(t,
			bytes.Equal(
				[]byte(`[1,2,3,4,5,6]`),
				encodedParams,
			),
		)
		require.Equal(t, "99", req.ID.String())

		req, err = NewRequest("", nil, NoID)
		require.NoError(t, err)
		require.NotNil(t, req)
	})

	t.Run("error", func(t *testing.T) {
		params := errMarshaler{}
		req, err := NewRequest("failed.method", params, UnknownID)

		require.Error(t, err)
		require.Nil(t, req)
	})
}

func TestRequest_DecodeParams(t *testing.T) {
	t.Run("Params exists", func(t *testing.T) {
		req, _ := NewRequest("", 99, NoID)
		var i int
		err := req.DecodeParams(&i)
		require.NoError(t, err)
		require.Equal(t, 99, i)
	})

	t.Run("Params not exitsts", func(t *testing.T) {
		req, _ := NewRequest("", nil, NoID)
		var i int
		err := req.DecodeParams(&i)
		require.Equal(t, ErrNilParams, err)
		require.Zero(t, i)
	})
}

func TestRequest_Reader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req, _ := NewRequest("reader.test", 99, NewID("000-0000"))
		r, err := req.Reader()
		require.NoError(t, err)

		b, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		require.Equal(t, `{"jsonrpc":"2.0","method":"reader.test","params":99,"id":"000-0000"}`+"\n", string(b))
	})

	t.Run("error", func(t *testing.T) {
		id := ID{
			nType: 16,
		}
		req, _ := NewRequest("", nil, id)
		r, err := req.Reader()
		require.Error(t, err)
		require.Nil(t, r)
	})
}

func TestRequest_MarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req, _ := NewRequest("reader.test", nil, NoID)
		b, err := req.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, `{"jsonrpc":"2.0","method":"reader.test"}`, string(b))
	})

	t.Run("error", func(t *testing.T) {
		id := ID{
			nType: 16,
		}
		req, _ := NewRequest("", nil, id)
		b, err := req.MarshalJSON()
		require.Error(t, err)
		require.Nil(t, b)
	})
}

type errMarshaler struct{}

func (em errMarshaler) MarshalJSON() ([]byte, error) { return nil, errors.New("dummy error") }
