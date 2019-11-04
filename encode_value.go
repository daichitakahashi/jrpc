package jrpc

import "encoding/json"

func encodeValue(v interface{}) (json.RawMessage, error) {
	buf := bufferpool.Get()
	defer buf.Free()
	if v == nil {
		buf.AppendString("null")
	} else {
		switch target := v.(type) {
		case int:
			buf.AppendInt(int64(target))
		case int8:
			buf.AppendInt(int64(target))
		case int16:
			buf.AppendInt(int64(target))
		case int32:
			buf.AppendInt(int64(target))
		case int64:
			buf.AppendInt(target)
		case uint:
			buf.AppendUint(uint64(target))
		case uint8: // byte
			buf.AppendUint(uint64(target))
		case uint16:
			buf.AppendUint(uint64(target))
		case uint32:
			buf.AppendUint(uint64(target))
		case uint64:
			buf.AppendUint(target)
		case float32:
			buf.AppendFloat(float64(target), 32)
		case float64:
			buf.AppendFloat(target, 64)
		case bool:
			buf.AppendBool(target)
		/*case string: // too slow
		buf.AppendQuote(target)*/
		case json.Marshaler:
			return target.MarshalJSON()
		default: // contains []byte, string
			return json.Marshal(v)
		}
	}
	b := make([]byte, buf.Len())
	copy(b, buf.Bytes())
	return b, nil
}
