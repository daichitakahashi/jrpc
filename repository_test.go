package jrpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
TODO:

*/

func TestMethodRepository_Methods(t *testing.T) {
	repo := NewRepository()

	type RegisterTarget struct {
		name    string
		handler Handler
		params  interface{}
		result  interface{}
	}

	willBeRegistered := []RegisterTarget{
		{
			name: "method1",
			handler: HandlerFunc(func(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
				return "one", nil
			}),
			params: 0,
			result: "",
		}, {
			name:    "method2",
			handler: nil,
			params:  1,
			result:  "",
		}, {
			name: "",
			handler: HandlerFunc(func(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
				return "three", nil
			}),
			params: 2,
			result: "",
		}, {
			name: "method4",
			handler: HandlerFunc(func(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
				return "four", nil
			}),
			params: 3,
			result: "",
		},
	}

	registered := make(map[string]RegisterTarget)

	for _, t := range willBeRegistered {
		err := repo.Register(t.name, t.handler, t.params, t.result)
		if err != nil {
			continue
		}
		registered[t.name] = t
	}

	methods := repo.Methods()
	require.Len(t, methods, 2)

	for key := range methods {
		_, ok := registered[key]
		require.True(t, ok)
		delete(registered, key)
	}

	require.Len(t, registered, 0)
}

func TestMethodRepository_Interceptors(t *testing.T) {
	repo := NewRepository()
	firstCnt := &counter{}
	secondCnt := &counter{}
	nopHandler := HandlerFunc(func(_ context.Context, _ *json.RawMessage) (interface{}, *Error) {
		return nil, nil
	})

	require.Len(t, repo.Interceptors(), 0)

	repo.With(firstCnt.Interceptor)
	repo.With(firstCnt.Interceptor)
	repo.Register("first", nopHandler, nil, nil)

	require.Len(t, repo.Interceptors(), 2)

	repo.Namespace("second", func(repo Repository) {
		repo.With(secondCnt.Interceptor)
		repo.Register("second", nopHandler, nil, nil)

		require.Len(t, repo.Interceptors(), 1)
	})

	md, ok := repo.methods.Load("first")
	require.True(t, ok)
	md.(*Metadata).InterceptorChain(nil, nil, nil, md.(*Metadata).Handler)

	md, ok = repo.methods.Load("second.second")
	require.True(t, ok)
	md.(*Metadata).InterceptorChain(nil, nil, nil, md.(*Metadata).Handler)

	require.Equal(t, 4, firstCnt.c)
	require.Equal(t, 1, secondCnt.c)
}

func TestMethodRepository_TrimSeparator(t *testing.T) {
	type TestCase struct {
		namespace    string
		newNamespace string
		desc         string
	}
	testcases := []TestCase{
		{
			namespace:    "sample",
			newNamespace: "sample",
			desc:         "straight",
		}, {
			namespace:    ".head",
			newNamespace: "head",
			desc:         "unnecessary head sep",
		}, {
			namespace:    "...head",
			newNamespace: "head",
			desc:         "unnecessary serial head sep",
		}, {
			namespace:    "tail.",
			newNamespace: "tail",
			desc:         "unnecessary tail sep",
		}, {
			namespace:    "tail...",
			newNamespace: "tail",
			desc:         "unnecessary serial tail sep",
		}, {
			namespace:    "...both...",
			newNamespace: "both",
			desc:         "unnecessary serial head & tail sep",
		}, {
			namespace:    "inter.rupt",
			newNamespace: "inter.rupt",
			desc:         "explicit namespace",
		}, {
			namespace:    "inter...rupt",
			newNamespace: "inter...rupt",
			desc:         "explicit namespace(serial sep)",
		}, {
			namespace:    "....................",
			newNamespace: "",
			desc:         "all sep",
		}, {
			namespace:    "a",
			newNamespace: "a",
			desc:         "single rune",
		}, {
			namespace:    ".",
			newNamespace: "",
			desc:         "only single sep",
		}, {
			namespace:    "",
			newNamespace: "",
			desc:         "empty",
		},
	}

	repo := &MethodRepository{
		sep: ".",
	}
	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			namespace := repo.trimSeparator(testcase.namespace)
			require.Equal(t, testcase.newNamespace, namespace)
		})
	}
}

func TestMethodRepository_AppendNamespace(t *testing.T) {
	type TestCase struct {
		former       string
		latter       string
		newNamespace string
		desc         string
	}
	testcases := []TestCase{
		{
			former:       "aaaaa",
			latter:       "bbbbb",
			newNamespace: "aaaaa.bbbbb",
			desc:         "straight",
		}, {
			former:       "",
			latter:       "bbbbb",
			newNamespace: "bbbbb",
			desc:         "former empty",
		}, {
			former:       "aaaaa",
			latter:       "",
			newNamespace: "aaaaa",
			desc:         "latter empty",
		},
	}
	repo := &MethodRepository{
		sep: ".",
	}
	for _, testcase := range testcases {
		t.Run(testcase.desc, func(t *testing.T) {
			namespace := repo.appendNamespace(testcase.former, testcase.latter)
			require.Equal(t, testcase.newNamespace, namespace)
		})
	}
}

type counter struct {
	c int
}

func (c *counter) Interceptor(ctx context.Context, params *json.RawMessage, _ *RequestInfo, handler Handler) (interface{}, *Error) {
	c.c++
	return handler.ServeJSONRPC(ctx, params)
}
