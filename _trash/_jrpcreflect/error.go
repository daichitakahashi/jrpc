package jrpcreflect

import (
	"reflect"
)

func reflectError(method string, expected, actual reflect.Type) *ReflectError {
	return &ReflectError{
		Method:   method,
		Expected: expected,
		Actual:   actual,
	}
}

// ReflectError is
type ReflectError struct {
	Method   string
	Expected reflect.Type
	Actual   reflect.Type
}

func (re *ReflectError) Error() string {
	return re.Method
}
