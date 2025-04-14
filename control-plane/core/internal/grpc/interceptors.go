package lmgrpc

import (
	"context"
	"time"

	"github.com/raphaelCamblong/Luminous-Mesh/control-plane/core/init/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	if info.FullMethod != "/luminousmesh.NodeService/RegisterNode" {
		if err := s.authenticate(ctx); err != nil {
			return nil, err
		}
	}

	resp, err := handler(ctx, req)

	logger.L().Info("Unary RPC",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", time.Since(start)),
		zap.Error(err),
	)

	return resp, err
}

// streamInterceptor handles authentication and logging for streaming RPC calls
func (s *Server) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()

	// Authenticate stream
	if err := s.authenticate(ss.Context()); err != nil {
		return err
	}

	// Wrap stream to intercept messages
	wrapped := newWrappedStream(ss, info.FullMethod)

	// Handle stream
	err := handler(srv, wrapped)

	// Log stream completion
	logger.L().Info("Stream RPC completed",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", time.Since(start)),
		zap.Error(err),
	)

	return err
}

// authenticate validates the authentication token from the context
func (s *Server) authenticate(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return status.Error(codes.Unauthenticated, "missing authorization token")
	}

	nodeID := md.Get("node-id")
	if len(nodeID) == 0 {
		return status.Error(codes.Unauthenticated, "missing node ID")
	}

	if err := s.authManager.ValidateAuthToken(nodeID[0], tokens[0]); err != nil {
		return status.Error(codes.Unauthenticated, "invalid authorization token")
	}

	return nil
}

// wrappedStream wraps grpc.ServerStream to provide message interception
type wrappedStream struct {
	grpc.ServerStream
	method string
}

func newWrappedStream(s grpc.ServerStream, method string) *wrappedStream {
	return &wrappedStream{
		ServerStream: s,
		method:       method,
	}
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}
	return nil
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	err := w.ServerStream.SendMsg(m)
	if err != nil {
		return err
	}
	return nil
}
