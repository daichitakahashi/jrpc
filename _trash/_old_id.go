package jrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// ID represents "id" member of JSON-RPC object.
// An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
// If it is not included it is assumed to be a notification.
// The value SHOULD normally not be Null and Numbers SHOULD NOT contain fractional parts
//
// The Server MUST reply with the same value in the Response object if included.
// This member is used to correlate the context between the two objects.
type ID struct {
	val interface{}
}

// NewID is
func NewID(v interface{}) ID {
	id := ID{}
	if v == nil {
		// id.val = nil
		return id
	}
	switch val := v.(type) {
	case int:
		id.val = int64(val)
		return id
	case int8:
		id.val = int64(val)
		return id
	case int16:
		id.val = int64(val)
		return id
	case int32:
		id.val = int64(val)
		return id
	case int64:
		id.val = val
		return id
	case uint:
		id.val = int64(val)
		return id
	case uint8:
		id.val = int64(val)
		return id
	case uint16:
		id.val = int64(val)
		return id
	case uint32:
		id.val = int64(val)
		return id
	case uint64:
		id.val = int64(val)
		return id
	case float32:
		id.val = float64(val)
		return id
	case float64:
		id.val = val
		return id
	case string:
		id.val = val
		return id
	}
	panic(errors.New("jrpc: jrpc.ID: invalid type of \"id\" member"))
}

func (id ID) String() string {
	if str, ok := id.val.(string); ok {
		return str
	}
	return fmt.Sprint(id.val)
}

// Int64 is
func (id ID) Int64() int64 {
	return id.val.(int64)
}

// Float64 is
func (id ID) Float64() float64 {
	return id.val.(float64)
}

// EqualsValue is
func (id ID) EqualsValue(v interface{}) bool {
	return id.val == v
}

// MarshalJSON implements json.Marshaler
func (id ID) MarshalJSON() (b []byte, err error) {
	buf := bufferpool.Get()
	err = id.encodeTo(buf)
	if err != nil {
		buf.Free()
		return
	}
	b = make([]byte, buf.Len())
	copy(b, buf.Bytes())
	buf.Free()
	return
}

func (id *ID) encodeTo(buf *Buffer) error {
	if id == nil {
		buf.AppendString("null")
		return nil
	} else if *id == UnknownID || *id == NoID {
		buf.AppendString("null")
		return nil
	}
	switch val := id.val.(type) {
	case int64:
		buf.AppendInt(val)
	case float64:
		buf.AppendFloat(val, 64)
	case string:
		buf.AppendQuote(val)
	default:
		return errors.New("jrpc: jrpc.ID: invalid content of \"id\" member")
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler
func (id *ID) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		id.val = unknown{}
		return nil
	}
	var num json.Number
	err := json.Unmarshal(b, &num)
	if err != nil {
		return err
	}

	if b[0] == '"' {
		id.val = num.String()
		return nil
	}
	if n, err := num.Int64(); err == nil {
		id.val = n
		return nil
	}
	if n, err := num.Float64(); err == nil {
		id.val = n
		return nil
	}
	return errors.New("jrpc: jrpc.ID: invalid content of \"id\" member")
}

type unknown struct{}

// UnknownID represents explicit but null ID. Ex) {..."id": null}
var UnknownID = ID{val: unknown{}}

// NoID represents omitted ID
var NoID = ID{val: nil}
