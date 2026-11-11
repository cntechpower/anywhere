// Package trace provides a panic-free wrapper around opentracing.
//
// The upstream github.com/cntechpower/utils/tracing package keeps its tracer in
// a package level variable that is only set by tracing.Init. Calling
// tracing.New without Init dereferences a nil tracer and panics, which made the
// server crash on the first proxied connection whenever TRACE_ADDR was not set.
// This package always goes through opentracing's global tracer, which falls
// back to a no-op tracer when tracing is not configured.
package trace

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// New starts a span. When tracing is not initialized the returned span is a
// no-op and Finish is safe to call.
func New(ctx context.Context, operationName string) (opentracing.Span, context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	return opentracing.StartSpanFromContext(ctx, operationName)
}

// Do runs fn inside a span and finishes the span afterwards.
func Do(ctx context.Context, operationName string, fn func() error) error {
	span, _ := New(ctx, operationName)
	defer span.Finish()
	return fn()
}
