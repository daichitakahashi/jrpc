package jrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

const (
	typeNumber  byte = 1
	typeString  byte = 2
	typeUnknown byte = 4
)

var (
	// UnknownID represents explicit but null ID. Ex) {..."id": null}
	UnknownID = ID{
		n:     json.Number("null"),
		nType: typeUnknown,
	}

	// NoID represents omitted ID
	NoID = ID{}
)

// ID represents "id" member of JSON-RPC object.
// An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
// If it is not included it is assumed to be a notification.
// The value SHOULD normally not be Null and Numbers SHOULD NOT contain fractional parts
//
// The Server MUST reply with the same value in the Response object if included.
// This member is used to correlate the context between the two objects.
type ID struct {
	n     json.Number
	nType byte
}

// NewID is
// NoID is zero value
func NewID(v interface{}) ID {
	if v == nil {
		return UnknownID
	}
	switch val := v.(type) {
	case int:
		return ID{
			n:     json.Number(strconv.FormatInt(int64(val), 10)),
			nType: typeNumber,
		}
	case int8:
		return ID{
			n:     json.Number(strconv.FormatInt(int64(val), 10)),
			nType: typeNumber,
		}
	case int16:
		return ID{
			n:     json.Number(strconv.FormatInt(int64(val), 10)),
			nType: typeNumber,
		}
	case int32:
		return ID{
			n:     json.Number(strconv.FormatInt(int64(val), 10)),
			nType: typeNumber,
		}
	case int64:
		return ID{
			n:     json.Number(strconv.FormatInt(val, 10)),
			nType: typeNumber,
		}
	case uint:
		return ID{
			n:     json.Number(strconv.FormatUint(uint64(val), 10)),
			nType: typeNumber,
		}
	case uint8:
		return ID{
			n:     json.Number(strconv.FormatUint(uint64(val), 10)),
			nType: typeNumber,
		}
	case uint16:
		return ID{
			n:     json.Number(strconv.FormatUint(uint64(val), 10)),
			nType: typeNumber,
		}
	case uint32:
		return ID{
			n:     json.Number(strconv.FormatUint(uint64(val), 10)),
			nType: typeNumber,
		}
	case uint64:
		return ID{
			n:     json.Number(strconv.FormatUint(val, 10)),
			nType: typeNumber,
		}
	case float32:
		return ID{
			n:     json.Number(strconv.FormatFloat(float64(val), 'f', -1, 32)),
			nType: typeNumber,
		}
	case float64:
		return ID{
			n:     json.Number(strconv.FormatFloat(val, 'f', -1, 64)),
			nType: typeNumber,
		}
	case string:
		return ID{
			n:     json.Number(val),
			nType: typeString,
		}
	}
	panic(errors.New("jrpc: jrpc.ID: invalid type of \"id\" member"))
}

// String is
func (id ID) String() string {
	return string(id.n)
}

// Int is
func (id ID) Int() (int, error) {
	i64, err := strconv.ParseInt(string(id.n), 10, strconv.IntSize)
	return int(i64), err
}

// Int64 is
func (id ID) Int64() (int64, error) {
	return strconv.ParseInt(string(id.n), 10, 64)
}

// Uint is
func (id ID) Uint() (uint, error) {
	ui64, err := strconv.ParseUint(string(id.n), 10, strconv.IntSize)
	return uint(ui64), err
}

// Uint64 is
func (id ID) Uint64() (uint64, error) {
	return strconv.ParseUint(string(id.n), 10, 64)
}

// Float32 is
func (id ID) Float32() (float32, error) {
	f64, err := strconv.ParseFloat(string(id.n), 32)
	return float32(f64), err
}

// Float64 is
func (id ID) Float64() (float64, error) {
	return strconv.ParseFloat(string(id.n), 64)
}

// Equals is
func (id ID) Equals(v interface{}) bool {
	switch id.nType {
	case typeNumber:
		var equal bool
		var err error

		switch val := v.(type) {
		case int:
			var i int
			i, err = id.Int()
			equal = val == i
		case int8:
			var i64 int64
			i64, err = id.Int64()
			equal = int64(val) == i64
		case int16:
			var i64 int64
			i64, err = id.Int64()
			equal = int64(val) == i64
		case int32:
			var i64 int64
			i64, err = id.Int64()
			equal = int64(val) == i64
		case int64:
			var i64 int64
			i64, err = id.Int64()
			equal = val == i64
		case uint:
			var ui uint
			ui, err = id.Uint()
			equal = val == ui
		case uint8:
			var ui64 uint64
			ui64, err = id.Uint64()
			equal = uint64(val) == ui64
		case uint16:
			var ui64 uint64
			ui64, err = id.Uint64()
			equal = uint64(val) == ui64
		case uint32:
			var ui64 uint64
			ui64, err = id.Uint64()
			equal = uint64(val) == ui64
		case uint64:
			var ui64 uint64
			ui64, err = id.Uint64()
			equal = uint64(val) == ui64
		case float32:
			var f32 float32
			f32, err = id.Float32()
			equal = val == f32
		case float64:
			var f64 float64
			f64, err = id.Float64()
			equal = val == f64
		}
		if err != nil {
			return false
		}
		return equal
	case typeString:
		if val, ok := v.(string); ok {
			return string(id.n) == val
		}
		return false
	case typeUnknown:
		return v == nil
	default:
		return false
	}
}

// Type returns number, string or unknown
func (id ID) Type() string {
	switch id.nType {
	case typeNumber:
		return "number"
	case typeString:
		return "string"
	case typeUnknown, 0:
		return "unknown"
	default:
		panic("jrpc: ID: invalid stringify of invalid value ID")
	}
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

func (id ID) encodeTo(buf *Buffer) error {
	if id == UnknownID || id == NoID {
		buf.AppendString("null")
		return nil
	}

	switch id.nType {
	case typeNumber:
		buf.AppendString(string(id.n))
	case typeString:
		b, _ := json.Marshal(string(id.n))
		buf.Write(b)
		//buf.AppendQuote(string(id.n))
	default:
		return errors.New("jrpc: invalid type of \"id\" member")
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler
func (id *ID) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		id.n = "null"
		id.nType = typeUnknown
		return nil
	}
	err := json.Unmarshal(b, &id.n)
	if err != nil {
		return err
	}

	if b[0] == '"' {
		id.nType = typeString
	} else {
		id.nType = typeNumber
	}
	return nil
}
