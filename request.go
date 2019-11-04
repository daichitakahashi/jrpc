package jrpc

import (
	"encoding/json"
	"errors"
	"io"
)

/*
TODO:

*/

// ErrNilParams is
var ErrNilParams = errors.New("jrpc: Request.Params is nil")

// Request represents JSON-RPC Request object.
// A rpc call is represented by sending a Request object to a Server.
// The Request object has the following members:
//
// jsonrpc
// A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
//
// method
// A String containing the name of the method to be invoked.
// Method names that begin with the word rpc followed by a period character (U+002E or ASCII 46) are reserved for
// rpc-internal methods and extensions and MUST NOT be used for anything else.
//
// params
// A Structured value that holds the parameter values to be used during the invocation of the method.
// This member MAY be omitted.
//
// id
// An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
// If it is not included it is assumed to be a notification.
// The value SHOULD normally not be Null and Numbers SHOULD NOT contain fractional parts
//
// The Server MUST reply with the same value in the Response object if included. This member is used to correlate the context between the two objects.
type Request struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	Params  *json.RawMessage `json:"params"`
	ID      ID               `json:"id"`
	err     error
}

// NewRequest is
func NewRequest(method string, params interface{}, id ID) (*Request, error) {
	req := &Request{
		Version: "2.0",
		Method:  method,
		ID:      id,
	}
	if params == nil {
		return req, nil
	}
	err := req.EncodeAndSetParams(params)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// DecodeParams is
func (req *Request) DecodeParams(v interface{}) error {
	if req.Params == nil {
		return ErrNilParams
	}
	return json.Unmarshal(*req.Params, v)
}

// EncodeAndSetParams is
func (req *Request) EncodeAndSetParams(v interface{}) error {
	data, err := encodeValue(v)
	if err != nil {
		return err
	}
	req.Params = &data
	return nil
}

func (req *Request) toResponse() *Response {
	if req != nil {
		return &Response{
			Version: req.Version,
			ID:      req.ID,
		}
	}
	return &Response{
		Version: "2.0",
	}
}

// Reader is.
func (req *Request) Reader() (io.Reader, error) {
	buf, err := req.encodeLine()
	if err != nil {
		return nil, err
	}
	// buf.Free() will be called after io.EOF found
	return buf, nil
}

// MarshalJSON implements json.Marshaler
func (req *Request) MarshalJSON() (b []byte, err error) {
	buf, err := req.encode()
	if err != nil {
		return
	}
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}

func (req *Request) encode() (*Buffer, error) {
	buf := bufferpool.Get()

	err := req.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	return buf, nil
}

func (req *Request) encodeTo(buf *Buffer) (err error) {
	buf.AppendString(`{"jsonrpc":`)
	b, _ := json.Marshal(req.Version)
	buf.Write(b)
	//buf.AppendQuote(req.Version)
	buf.AppendString(`,"method":`)
	b, _ = json.Marshal(req.Method)
	buf.Write(b)
	//buf.AppendQuote(req.Method)
	if req.Params != nil {
		buf.AppendString(`,"params":`)
		buf.Write(*req.Params)
	}
	if req.ID != NoID {
		buf.AppendString(`,"id":`)
		err = req.ID.encodeTo(buf)
		if err != nil {
			return err
		}
	}
	buf.AppendByte('}')

	return nil
}

func (req *Request) encodeLine() (*Buffer, error) {
	buf := bufferpool.Get()

	err := req.encodeTo(buf)
	if err != nil {
		buf.Free()
		return nil, err
	}
	buf.AppendByte('\n')
	return buf, nil
}

var _ encoderLine = (*Request)(nil)
