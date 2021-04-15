package utils

import "github.com/opentracing/opentracing-go"

type option func(span opentracing.Span)

func WithTag(k, v string) func(span opentracing.Span) {
	return func(span opentracing.Span) {
		span.SetTag(k, v)
	}
}

func WithLog(k, v string) func(span opentracing.Span) {
	return func(span opentracing.Span) {
		span.LogKV(k, v)
	}
}

func WrapTrace(span opentracing.Span, options ...option) {
	for _, o := range options {
		o(span)
	}
}
