package dev

import (
	"testing"

	"github.com/armon/go-radix"
)

func BenchmarkRadix(b *testing.B) {

	b.Run("map", func(b *testing.B) {
		target := map[string]interface{}{
			"testing.repo1.addstring": func() {},
			"testing.repo1.substring": func() {},
			"testing.repo2.add":       func() {},
			"testing.repo2.subtract":  func() {},
			"testing.repo2.divide":    func() {},
			"testing.repo3.readdir":   func() {},
			"testing.repo3.stat":      func() {},
			"testing.repo4.listen":    func() {},
			"testing.repo4.serve":     func() {},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			f, ok := target["testing.repo3.readdir"]
			if ok {
				f.(func())()
			}
		}
	})

	b.Run("radix", func(b *testing.B) {
		r := radix.NewFromMap(map[string]interface{}{
			"testing.repo1.addstring": func() {},
			"testing.repo1.substring": func() {},
			"testing.repo2.add":       func() {},
			"testing.repo2.subtract":  func() {},
			"testing.repo2.divide":    func() {},
			"testing.repo3.readdir":   func() {},
			"testing.repo3.stat":      func() {},
			"testing.repo4.listen":    func() {},
			"testing.repo4.serve":     func() {},
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			f, ok := r.Get("testing.repo3.readdir")
			if ok {
				f.(func())()
			}
		}
	})
}
