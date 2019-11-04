package dev

import (
	"bytes"
	"strings"
	"testing"
)

func BenchmarkNamespace(b *testing.B) {

	base := "namespace"
	var sep byte = '.'
	namespace := "example"
	result := "namespace.example"

	b.Run("[]byte_append", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := make([]byte, 0, len(base)+1+len(namespace))
			v = append(v, base...)
			v = append(v, sep)
			v = append(v, namespace...)
			expected := string(v)
			if expected != result {
				panic("incorrext")
			}
		}
	})

	b.Run("[]byte_copy", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			v := make([]byte, len(base)+1+len(namespace))
			n := copy(v, base)
			v[n] = sep
			copy(v[n+1:], namespace)
			expected := string(v)
			if expected != result {
				panic("incorrext")
			}
		}
	})

	b.Run("string", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			expected := base + string(sep) + namespace
			if expected != result {
				panic("incorrext")
			}
		}
	})

	sepString := "."
	b.Run("sep_string", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			expected := base + sepString + namespace
			if expected != result {
				panic("incorrext")
			}
		}
	})

	b.Run("bytes.Buffer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			buf.WriteString(base)
			buf.WriteByte(sep)
			buf.WriteString(namespace)
			expected := buf.String()
			if expected != result {
				panic("incorrext")
			}
		}
	})

	b.Run("strings.Builder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var builder strings.Builder
			builder.Grow(len(base) + 1 + len(namespace))
			builder.WriteString(base)
			builder.WriteByte(sep)
			builder.WriteString(namespace)
			expected := builder.String()
			if expected != result {
				panic("incorrext")
			}
		}
	})
}
