package jrpc

import (
	"encoding/json"
	"errors"
)

/*
TODO:

*/

// ErrNilResult is
var ErrNilResult = errors.New("jrpc: Response.Result is nil")

// Response represents JSON-RPC Response object.
// When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications.
// The Response is expressed as a single JSON Object, with the following members:
//
// jsonrpc
// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
//
// result
// This member is REQUIRED on success.
// This member MUST NOT exist if there was an error invoking the method.
// The value of this member is determined by the method invoked on the Server.
//
// error
// This member is REQUIRED on error.
// This member MUST NOT exist if there was no error triggered during invocation.
// The value for this member MUST be an Object as defined in section 5.1.
//
// id
// This member is REQUIRED.
// It MUST be the same as the value of the id member in the Request Object.
// If there was an error in detecting the id in the Request object (e.g. Parse error/Invalid Request), it MUST be Null.
//
// Either the result member or error member MUST be included, but both members MUST NOT be included.
type Response struct {
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   *Error           `json:"error"`
	ID      ID               `json:"id"`
}

// DecodeResult is
func (resp *Response) DecodeResult(v interface{}) error {
	if resp.Result == nil {
		return ErrNilResult
		// return errors.New("Response.Result is nil")
	}
	return json.Unmarshal(*resp.Result, v)
}

// EncodeAndSetResult is
func (resp *Response) EncodeAndSetResult(v interface{}) error {
	data, err := encodeValue(v)
	if err != nil {
		return err
	}
	resp.Result = &data
	return nil
}

// MarshalJSON implements json.Marshaler
func (resp *Response) MarshalJSON() (b []byte, err error) {
	buf, err := resp.encode()
	if err != nil {
		return
	}
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}

func (resp *Response) isSend() bool {
	return resp.ID != NoID || resp.Error != nil
}

func (resp *Response) encode() (*Buffer, error) {
	buf := bufferpool.Get()

	err := resp.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	return buf, nil
}

func (resp *Response) encodeTo(buf *Buffer) (err error) {
	buf.AppendString(`{"jsonrpc":`)
	b, _ := json.Marshal(resp.Version)
	buf.Write(b)
	//buf.AppendQuote(resp.Version)
	if resp.Error == nil {
		buf.AppendString(`,"result":`)
		if resp.Result != nil {
			buf.Write(*resp.Result)
		} else {
			buf.AppendString("null")
		}
	} else {
		buf.AppendString(`,"error":`)
		err = resp.Error.encodeTo(buf)
		if err != nil {
			return
		}
	}
	buf.AppendString(`,"id":`)
	err = resp.ID.encodeTo(buf)
	if err != nil {
		return
	}
	buf.AppendByte('}')
	return
}

func (resp *Response) encodeLine() (*Buffer, error) {
	buf := bufferpool.Get()

	err := resp.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	buf.AppendByte('\n')
	return buf, nil
}

var _ encoderLine = (*Response)(nil)
