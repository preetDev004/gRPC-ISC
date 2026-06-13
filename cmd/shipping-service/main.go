// cmd/shipping-service – standalone ShippingService gRPC server.
package main

import (
	"log"

	pb "github.com/preetDev004/gRPC-ISC/gen"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
	shippingsvc "github.com/preetDev004/gRPC-ISC/services/shipping"
	"google.golang.org/grpc"
)

func main() {
	addr := grpcutil.EnvOr("GRPC_ADDR", ":9002")
	grpcutil.MustServeGRPC(addr, func(s *grpc.Server) {
		pb.RegisterShippingServiceServer(s, shippingsvc.New())
		log.Printf("ShippingService listening on %s", addr)
	})
}
