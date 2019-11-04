package positional

import "encoding/json"

type (
	// Unmarshaler is
	Unmarshaler interface {
		UnmarshalByPosition(index int, decode DecodeFunc) error
	}

	// DecodeFunc is
	DecodeFunc func(v interface{}) error
)

// IsPositional is
func IsPositional(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	return data[0] == '['
}

// UnmarshalByPosition is
func UnmarshalByPosition(data []byte, u Unmarshaler) error {
	var arr []*json.RawMessage
	err := json.Unmarshal(data, &arr)
	if err != nil {
		return err // not valid array
	}

	for i := range arr {
		err = u.UnmarshalByPosition(i, func(v interface{}) error {
			switch u := v.(type) {
			case json.Unmarshaler:
				return u.UnmarshalJSON(*arr[i])
			default:
				return json.Unmarshal(*arr[i], v)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Unmarshal is
func Unmarshal(data []byte, v interface{}) error {
	if IsPositional(data) {
		if u, ok := v.(Unmarshaler); ok {
			return UnmarshalByPosition(data, u)
		}
	}
	switch i := v.(type) {
	case json.Unmarshaler:
		return i.UnmarshalJSON(data)
	default:
		return json.Unmarshal(data, v)
	}
}
