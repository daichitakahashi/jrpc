package jrpc

import (
	"encoding/json"
	"errors"
	"fmt"
)

/*
TODO:
- Err... シリーズ、Dataに、Error()の結果入れる？

*/

// ErrNilData is
var ErrNilData = errors.New("jrpc: Error.Data is nil")

// ErrorCode is
type ErrorCode int

const (
	// ErrorCodeParse Invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	ErrorCodeParse ErrorCode = -32700
	// ErrorCodeInvalidRequest The JSON sent is not a valid Request object.
	ErrorCodeInvalidRequest ErrorCode = -32600
	// ErrorCodeMethodNotFound The method does not exist / is not available.
	ErrorCodeMethodNotFound ErrorCode = -32601
	// ErrorCodeInvalidParams Invalid method parameter(s).
	ErrorCodeInvalidParams ErrorCode = -32602
	// ErrorCodeInternal Internal JSON-RPC error.
	ErrorCodeInternal ErrorCode = -32603
	/*
		-32000 to -32099	Server error	Reserved for implementation-defined server-errors.
		The remainder of the space is available for application defined errors.
	*/
)

// Error represents JSON-RPC error object.
// When a rpc call encounters an error, the Response Object MUST contain the error member with a value that is a Object with the following members:
//
// code
// A Number that indicates the error type that occurred.
// This MUST be an integer.
//
// message
// A String providing a short description of the error.
// The message SHOULD be limited to a concise single sentence.
//
// data
// A Primitive or Structured value that contains additional information about the error.
// This may be omitted.
// The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
//
// Either the result member or error member MUST be included, but both members MUST NOT be included.
type Error struct {
	Code    ErrorCode        `json:"code"`
	Message string           `json:"message"`
	Data    *json.RawMessage `json:"data"`
	err     error
}

// NewError is
func NewError(code ErrorCode, message string, data interface{}) (*Error, error) {
	e := &Error{
		Code:    code,
		Message: message,
	}
	if data != nil {
		err := e.EncodeAndSetData(data)
		if err != nil {
			return nil, err
		}
	}
	return e, nil
}

// FromError is
func FromError(code ErrorCode, err error) *Error {
	e := &Error{
		Code:    code,
		Message: err.Error(),
		err:     err,
	}
	e.EncodeAndSetData(err)
	return e
}

func (e *Error) Error() string {
	if e.Data == nil {
		return fmt.Sprintf("jsonrpc: code: %d, message: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("jsonrpc: code: %d, message: %s, data: %s", e.Code, e.Message, *e.Data)
}

// DecodeData is
func (e *Error) DecodeData(v interface{}) error {
	if e.Data == nil {
		return ErrNilData
	}
	return json.Unmarshal(*e.Data, v)
}

// EncodeAndSetData is
func (e *Error) EncodeAndSetData(v interface{}) error {
	data, err := encodeValue(v)
	if err != nil {
		return err
	}
	e.Data = &data
	return nil
}

// Cause is
func (e *Error) Cause() error {
	return e.err
}

// MarshalJSON implements json.Marshaler
func (e *Error) MarshalJSON() (b []byte, err error) {
	buf, _ := e.encode()
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}

func (e *Error) encode() (*Buffer, error) {
	buf := bufferpool.Get()
	_ = e.encodeTo(buf)
	return buf, nil
}

func (e *Error) encodeTo(buf *Buffer) error {
	buf.AppendString(`{"code":`)
	buf.AppendInt(int64(e.Code))
	buf.AppendString(`,"message":`)
	b, _ := json.Marshal(e.Message)
	buf.Write(b)
	if e.Data != nil {
		buf.AppendString(`,"data":`)
		buf.Write(*e.Data)
	}
	buf.AppendByte('}')
	return nil
}

var _ encoder = (*Error)(nil)

// ErrParse returns parse error.
func ErrParse(err error) *Error {
	return &Error{
		Code:    ErrorCodeParse,
		Message: "Parse error",
		err:     err,
	}
}

// ErrInvalidRequest returns invalid request error.
func ErrInvalidRequest(err error) *Error {
	return &Error{
		Code:    ErrorCodeInvalidRequest,
		Message: "Invalid Request",
		err:     err,
	}
}

// ErrMethodNotFound returns method not found error.
func ErrMethodNotFound() *Error {
	return &Error{
		Code:    ErrorCodeMethodNotFound,
		Message: "Method not found",
	}
}

// ErrInvalidParams returns invalid params error.
func ErrInvalidParams() *Error {
	return &Error{
		Code:    ErrorCodeInvalidParams,
		Message: "Invalid params",
	}
}

// ErrInternal returns internal error.
func ErrInternal(err error) *Error {
	return &Error{
		Code:    ErrorCodeInternal,
		Message: "Internal error",
		err:     err,
	}
}

// RecoveredError is
type RecoveredError struct {
	Request   *Request
	Recovered interface{}
}

func (re *RecoveredError) Error() string {
	/*req, err := re.Request.encode()
	if err != nil {
		return fmt.Sprintf("recovered: %+v, request: %+v", re.Recovered, *re.Request)
	}*/

	req := struct {
		Version string
		Method  string
		Params  string
		ID      ID
	}{
		re.Request.Version,
		re.Request.Method,
		"",
		re.Request.ID,
	}
	if re.Request.Params != nil {
		req.Params = string(*re.Request.Params)
	}
	return fmt.Sprintf("recovered: %+v, request: %+v", re.Recovered, req)
}
