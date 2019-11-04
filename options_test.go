package jrpc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	t.Run("WithNamespaceSeparator", func(t *testing.T) {
		repo := NewRepository(WithNamespaceSeparator('a'))
		require.Equal(t, "a", repo.options.namespaceSeparator)

		require.Equal(t, "piyoapiyo", repo.appendNamespace("piyo", "piyo"))
	})

	t.Run("WithDisableConcurrentCall", func(t *testing.T) {
		repo := NewRepository(WithDisableConcurrentCall())
		require.True(t, repo.options.disableConcurrentCall)
	})

	t.Run("WithPanicHandler", func(t *testing.T) {
		var called bool
		repo := NewRepository(WithPanicHandler(func(_ *Request, _ interface{}) {
			called = true
		}))

		repo.options.panicHandler(nil, nil)
		require.True(t, called)
	})
}
