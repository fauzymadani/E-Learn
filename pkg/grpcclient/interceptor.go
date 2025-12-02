package grpcclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"elearning/pkg/metrics"
)

// UnaryClientInterceptor returns a gRPC unary client interceptor for metrics
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start).Seconds()
		statusCode := "OK"
		if err != nil {
			statusCode = status.Code(err).String()
		}

		metrics.GrpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
		metrics.GrpcRequestDuration.WithLabelValues(method, statusCode).Observe(duration)

		return err
	}
}
