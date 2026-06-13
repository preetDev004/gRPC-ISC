// cmd/product-service – standalone ProductService gRPC server.
package main

import (
	"log"

	pb "github.com/preetDev004/gRPC-ISC/gen"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
	productsvc "github.com/preetDev004/gRPC-ISC/services/product"
	"google.golang.org/grpc"
)

func main() {
	addr := grpcutil.EnvOr("GRPC_ADDR", ":9001")
	grpcutil.MustServeGRPC(addr, func(s *grpc.Server) {
		pb.RegisterProductServiceServer(s, productsvc.New())
		log.Printf("ProductService listening on %s", addr)
	})
}
