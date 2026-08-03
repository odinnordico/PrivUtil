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
//
// When verbose is false (the default info log level) only the procedure name is
// logged; the raw panic value and stack — which may embed request-derived data —
// are logged only at debug level. The client always receives a generic error.
func RecoveryInterceptor(verbose bool) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					if verbose {
						log.Printf("recovered from panic in %s: %v\n%s", req.Spec().Procedure, r, debug.Stack())
					} else {
						log.Printf("recovered from panic in %s", req.Spec().Procedure)
					}
					err = connect.NewError(connect.CodeInternal, fmt.Errorf("internal error"))
				}
			}()
			return next(ctx, req)
		}
	}
}
