// Copyright 2025 Arcade Team
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package context

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	tracer     = otel.Tracer("go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc")
	propagator = otel.GetTextMapPropagator()
)

// metadataCarrier adapts metadata.MD to propagation.TextMapCarrier
type metadataCarrier struct {
	md *metadata.MD
}

func (c *metadataCarrier) Get(key string) string {
	values := c.md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c *metadataCarrier) Set(key, value string) {
	c.md.Set(key, value)
}

func (c *metadataCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.md))
	for k := range *c.md {
		keys = append(keys, k)
	}
	return keys
}

func spanName(method string) string {
	return method
}

// UnaryServerInterceptor unary server interceptor
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		ctx = propagator.Extract(ctx, &metadataCarrier{md: &md})

		name := spanName(info.FullMethod)
		ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		SetContext(ctx)
		defer ClearContext()

		if p, ok := peer.FromContext(ctx); ok {
			span.SetAttributes(attribute.String("net.peer.ip", p.Addr.String()))
		}

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
		)

		resp, err = handler(ctx, req)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(attribute.Int("rpc.grpc.status_code", int(s.Code())))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// StreamServerInterceptor stream server interceptor
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		ctx = propagator.Extract(ctx, &metadataCarrier{md: &md})

		name := spanName(info.FullMethod)
		ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		SetContext(ctx)
		defer ClearContext()

		if p, ok := peer.FromContext(ctx); ok {
			span.SetAttributes(attribute.String("net.peer.ip", p.Addr.String()))
		}

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", info.FullMethod),
		)

		wrapped := &wrappedServerStream{ServerStream: ss, ctx: ctx}
		err := handler(srv, wrapped)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(attribute.Int("rpc.grpc.status_code", int(s.Code())))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// UnaryClientInterceptor unary client interceptor
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = ContextWithSpan(ctx)
		name := spanName(method)
		ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		propagator.Inject(ctx, &metadataCarrier{md: &md})
		ctx = metadata.NewOutgoingContext(ctx, md)

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", method),
		)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(attribute.Int("rpc.grpc.status_code", int(s.Code())))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// StreamClientInterceptor stream client interceptor
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = ContextWithSpan(ctx)
		name := spanName(method)
		ctx, span := tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
		defer span.End()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}
		propagator.Inject(ctx, &metadataCarrier{md: &md})
		ctx = metadata.NewOutgoingContext(ctx, md)

		span.SetAttributes(
			attribute.String("rpc.system", "grpc"),
			attribute.String("rpc.service", method),
		)

		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
			span.SetAttributes(attribute.Int("rpc.grpc.status_code", int(s.Code())))
			span.End()
			return nil, err
		}

		return &wrappedClientStream{ClientStream: stream, span: span}, nil
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

type wrappedClientStream struct {
	grpc.ClientStream
	span trace.Span
}

func (w *wrappedClientStream) CloseSend() error {
	err := w.ClientStream.CloseSend()
	if err != nil {
		s, _ := status.FromError(err)
		w.span.SetStatus(codes.Error, s.Message())
		w.span.SetAttributes(attribute.Int("rpc.grpc.status_code", int(s.Code())))
	} else {
		w.span.SetStatus(codes.Ok, "")
	}
	w.span.End()
	return err
}
