package api

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	connect "connectrpc.com/connect"
)

// RecoveryInterceptor returns a connect interceptor that recovers from panics in
// handlers and converts them into CodeInternal errors instead of tearing down the
// connection. Note: this only catches Go panics — it cannot catch fatal runtime
// errors such as SIGSEGV from memory corruption, which abort the whole process.
func RecoveryInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("recovered from panic in %s: %v\n%s", req.Spec().Procedure, r, debug.Stack())
					err = connect.NewError(connect.CodeInternal, fmt.Errorf("internal error"))
				}
			}()
			return next(ctx, req)
		}
	}
}
