// Package grpcutil provides shared helpers for gRPC service binaries.
package grpcutil

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EnvOr returns the environment variable value or fallback when unset.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// MustServeGRPC starts a gRPC server on addr and calls register to attach services.
func MustServeGRPC(addr string, register func(*grpc.Server)) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	s := grpc.NewServer()
	register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("grpc serve %s: %v", addr, err)
	}
}

// Dial opens an insecure gRPC client connection to addr.
func Dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// DialWithRetry dials addr, retrying until success or the context is cancelled.
func DialWithRetry(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	const maxAttempts = 30
	const delay = 500 * time.Millisecond

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		conn, err := Dial(addr)
		if err == nil {
			return conn, nil
		}
		lastErr = err
		log.Printf("dial %s attempt %d/%d: %v", addr, attempt, maxAttempts, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}
	return nil, fmt.Errorf("dial %s after %d attempts: %w", addr, maxAttempts, lastErr)
}
