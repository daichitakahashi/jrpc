package jrpc

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		e, err := NewError(-1, "test NewError", 99)
		require.NoError(t, err)
		require.Equal(t, ErrorCode(-1), e.Code)
		require.Equal(t, "test NewError", e.Message)
		require.Equal(t, "99", string(*e.Data))
	})

	t.Run("error", func(t *testing.T) {
		e, err := NewError(-1, "test NewError", errMarshaler{})
		require.Error(t, err)
		require.Nil(t, e)
	})
}

func TestFromError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		err := sampleError{
			Value: "piyopiyo",
		}
		e := FromError(-99, err)
		require.Equal(t, ErrorCode(-99), e.Code)
		require.Equal(t, "sampleError: piyopiyo", e.Message)
		require.Equal(t, `{"Value":"piyopiyo"}`, string(*e.Data))

		require.Equal(t, err, e.Cause())
	})

	t.Run("error", func(t *testing.T) {
		e := FromError(-99, sampleErrorMarshaler{})
		require.Equal(t, ErrorCode(-99), e.Code)
		require.Equal(t, "sampleErrorMarshaler", e.Message)
		require.Nil(t, e.Data)
	})
}

func TestError_Error(t *testing.T) {
	t.Run("no Data", func(t *testing.T) {
		e, _ := NewError(100, "one hundred", nil)
		require.Equal(t, "jsonrpc: code: 100, message: one hundred", e.Error())
	})

	t.Run("with Data", func(t *testing.T) {
		type sample struct {
			First  int
			Second int
		}
		e, _ := NewError(100, "one hundred", sample{
			First:  100,
			Second: 50,
		})
		require.Equal(t, `jsonrpc: code: 100, message: one hundred, data: {"First":100,"Second":50}`, e.Error())
	})
}

func TestError_DecodeData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		raw := json.RawMessage(`999`)
		e := &Error{
			Data: &raw,
		}
		var i int
		err := e.DecodeData(&i)
		require.NoError(t, err)
		require.Equal(t, 999, i)
	})

	t.Run("error", func(t *testing.T) {
		e := &Error{}
		var i int
		err := e.DecodeData(&i)
		require.Equal(t, ErrNilData, err)
	})
}

func TestError_MarshalJSON(t *testing.T) {
	t.Run("no Data", func(t *testing.T) {
		e, _ := NewError(100, "one hundred", nil)
		b, _ := e.MarshalJSON()
		require.Equal(t, `{"code":100,"message":"one hundred"}`, string(b))
	})

	t.Run("with Data", func(t *testing.T) {
		type sample struct {
			First  int
			Second int
		}
		e, _ := NewError(100, "one hundred", sample{
			First:  100,
			Second: 50,
		})
		b, _ := e.MarshalJSON()
		require.Equal(t, `{"code":100,"message":"one hundred","data":{"First":100,"Second":50}}`, string(b))
	})
}

func TestRecoveredError(t *testing.T) {
	type recoveredObject struct {
		First  int
		Second int
		third  int
	}
	params := json.RawMessage(`null_params`)
	rvr := RecoveredError{
		Request: &Request{
			Version: "2.0",
			Method:  "recover.maker",
			Params:  &params,
			ID:      NewID("999-888"),
		},
		Recovered: recoveredObject{100, 99, 0},
	}
	require.Equal(t, `recovered: {First:100 Second:99 third:0}, request: {Version:2.0 Method:recover.maker Params:null_params ID:999-888}`, rvr.Error())
}

type sampleError struct {
	Value string
}

func (se sampleError) Error() string {
	return "sampleError: " + se.Value
}

type sampleErrorMarshaler struct{}

func (sem sampleErrorMarshaler) Error() string {
	return "sampleErrorMarshaler"
}

func (sem sampleErrorMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal error")
}

/*
type EncodedPart struct {
	json.RawMessage
}

func (part *EncodedPart) Unmarshal(v interface{}) error {
	if part == nil {
		return errors.New("ErrEmptyPart")
	} else if part.RawMessage == nil {
		return ErrNilParams
	}
	return json.Unmarshal(part.RawMessage, v)
}

func (part *EncodedPart) MarshalAndSet(v interface{}) error {
	if part == nil {
		return errors.New("ErrEmptyPart")
	}
	b, err := encodeValue(v)
	if err != nil {
		return err
	}
	part.RawMessage = b
	return nil
}

/*
req.Params.Unmarshal(&i)
req.Params.MarshalAndSet(str)
*

func TestA(t *testing.T) {
	type example struct {
		Params *EncodedPart     `json:"Params,omitempty"`
		Result *json.RawMessage `json:"Result,omitempty"`
	}
	src := []byte(`{"Params":null,"Result":null}`)
	e := example{}
	err := json.Unmarshal(src, &e)
	require.NoError(t, err)
	require.Nil(t, e.Params)
	require.Nil(t, e.Result)
	/*
		src = []byte(`{"Params":null,"Result":null}`)
		e = example{}
		err = json.Unmarshal(src, &e)
		require.NoError(t, err)
		require.Equal(t, `"null"`, string(e.Params.RawMessage))
		require.Nil(t, e.Result)
	*
	b, err := json.Marshal(e)
	require.NoError(t, err)
	require.Equal(t, `"null"`, string(b))
}
*/
