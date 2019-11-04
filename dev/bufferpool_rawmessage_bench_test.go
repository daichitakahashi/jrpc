package dev

import (
	"encoding/json"
	"testing"

	"github.com/daichitakahashi/jrpc"
)

func BenchmarkBufferpoolRawMessage(b *testing.B) {
	bufferpool := jrpc.NewPool()

	type sample struct {
		Name string
		Raw  *json.RawMessage
	}
	src := []byte(`{"name":"samplename", "raw": 59765976078104236423542365411111111111111111111111111111111111111111111111111111111111111113315734765477754750238758750392769850840506887676}`)

	b.Run("normal", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := sample{}
			if json.Unmarshal(src, &v) != nil {
				panic("")
			}
		}
	})

	b.Run("copy and re-use", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bufferpool.Get()
			var b json.RawMessage = buf.Bytes()
			v := sample{
				Raw: &b,
			}

			if json.Unmarshal(src, &v) != nil {
				panic("")
			}
			var neu json.RawMessage = make([]byte, len(*v.Raw))
			copy(neu, *v.Raw)
			v.Raw = &neu
			buf.Free()
		}
	})

	b.Run("full re-use", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bufferpool.Get()
			var b json.RawMessage = buf.Bytes()
			v := sample{
				Raw: &b,
			}

			if json.Unmarshal(src, &v) != nil {
				panic("")
			}
			buf.Free()
		}
	})
}

func BenchmarkContainer(b *testing.B) {

	b.Run("struct", func(b *testing.B) {
		type container struct {
			result interface{}
			errPtr **jrpc.Error
		}

		storage := make(map[int]container)

		result := 999
		err := &jrpc.Error{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			storage[i] = container{
				result: result,
				errPtr: &err,
			}

			v := storage[i]
			result = v.result.(int)
			err = *v.errPtr
		}
	})

	b.Run("array", func(b *testing.B) {
		storage := make(map[int][2]interface{})

		result := 999
		err := &jrpc.Error{}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			storage[i] = [2]interface{}{
				result,
				&err,
			}

			v := storage[i]
			result = v[0].(int)
			err = *(v[1].(**jrpc.Error))
		}
	})
}
