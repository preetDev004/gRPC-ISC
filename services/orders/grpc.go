package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

type gRPCServer struct {
	addr string
}

func NewGRPCServer(addr string) *gRPCServer {
	return &gRPCServer{addr: addr}
}

func (s *gRPCServer) Run() error {
	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return fmt.Errorf("failed to listen: %v", err)
	}

	// TODO: Register gRPC services here

	log.Printf("gRPC server listening on %s", s.addr)
	return grpcServer.Serve(lis)
}
