package dev

/*
func BenchmarkInterceptor(b *testing.B) {
	interceptors := []jrpc.Interceptor{
		interceptor(),
		interceptor(),
		interceptor(),
		interceptor(),
		interceptor(),
		interceptor(), //64
	}
	//var endpoint jrpc.InterceptorHandler
	checkResult := jrpc.InterceptorFunc(func(ctx jrpc.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
		if result != 64 {
			panic(fmt.Sprint("incorrect:", result))
		}
		return nil, nil
	})

	b.Run("pre-chaining", func(b *testing.B) {
		l := len(interceptors) - 1
		var endpoint jrpc.InterceptorHandler = checkResult
		for i := range interceptors {
			endpoint = interceptors[l-i](endpoint)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result = 1
			endpoint.Intercept(nil, nil)
		}
	})

	b.Run("chaining", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l := len(interceptors) - 1
			var endpoint jrpc.InterceptorHandler = checkResult
			for i := range interceptors {
				endpoint = interceptors[l-i](endpoint)
			}

			result = 1
			endpoint.Intercept(nil, nil)
		}
	})
}

var result int

func interceptor() jrpc.Interceptor {
	return func(next jrpc.InterceptorHandler) jrpc.InterceptorHandler {
		return jrpc.InterceptorFunc(func(ctx jrpc.Context, params *json.RawMessage) (interface{}, *jrpc.Error) {
			result *= 2
			return next.Intercept(ctx, params)
		})
	}
}
*/
