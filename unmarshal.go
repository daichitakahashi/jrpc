package jrpc

import "encoding/json"

// UnmarshalParams decodes JSON-RPC Request params.
func UnmarshalParams(params *json.RawMessage, dst interface{}) *Error {
	if params == nil {
		return ErrInvalidParams()
	}
	if err := json.Unmarshal(*params, dst); err != nil {
		return ErrInvalidParams()
	}
	return nil
}
