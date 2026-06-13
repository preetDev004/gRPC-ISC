// cmd/order-service – standalone OrderService gRPC server.
package main

import (
	"context"
	"log"
	"time"

	pb "github.com/preetDev004/gRPC-ISC/gen"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
	ordersvc "github.com/preetDev004/gRPC-ISC/services/order"
	"google.golang.org/grpc"
)

func main() {
	shippingAddr := grpcutil.EnvOr("SHIPPING_ADDR", "localhost:9002")
	productAddr := grpcutil.EnvOr("PRODUCT_ADDR", "localhost:9001")
	grpcAddr := grpcutil.EnvOr("GRPC_ADDR", ":9003")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderServer, err := ordersvc.New(ctx, shippingAddr, productAddr)
	if err != nil {
		log.Fatalf("failed to create order service: %v", err)
	}

	grpcutil.MustServeGRPC(grpcAddr, func(s *grpc.Server) {
		pb.RegisterOrderServiceServer(s, orderServer)
		log.Printf("OrderService listening on %s (shipping=%s product=%s)", grpcAddr, shippingAddr, productAddr)
	})
}
