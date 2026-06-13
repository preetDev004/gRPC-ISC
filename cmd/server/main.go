// cmd/server – all-in-one entry point for local development.
//
// For production-style deployments use the separate binaries under cmd/:
// product-service, shipping-service, order-service, and gateway.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/preetDev004/gRPC-ISC/gateway"
	pb "github.com/preetDev004/gRPC-ISC/gen"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
	ordersvc "github.com/preetDev004/gRPC-ISC/services/order"
	productsvc "github.com/preetDev004/gRPC-ISC/services/product"
	shippingsvc "github.com/preetDev004/gRPC-ISC/services/shipping"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

func main() {
	productListen := grpcutil.EnvOr("PRODUCT_GRPC_ADDR", ":9001")
	shippingListen := grpcutil.EnvOr("SHIPPING_GRPC_ADDR", ":9002")
	orderListen := grpcutil.EnvOr("ORDER_GRPC_ADDR", ":9003")
	productDial := grpcutil.EnvOr("PRODUCT_ADDR", "localhost:9001")
	shippingDial := grpcutil.EnvOr("SHIPPING_ADDR", "localhost:9002")
	orderDial := grpcutil.EnvOr("ORDER_ADDR", "localhost:9003")
	httpAddr := grpcutil.EnvOr("HTTP_ADDR", ":8080")
	frontendDir := grpcutil.EnvOr("FRONTEND_DIR", "./frontend")

	go grpcutil.MustServeGRPC(productListen, func(s *grpc.Server) {
		pb.RegisterProductServiceServer(s, productsvc.New())
		log.Printf("ProductService listening on %s", productListen)
	})

	go grpcutil.MustServeGRPC(shippingListen, func(s *grpc.Server) {
		pb.RegisterShippingServiceServer(s, shippingsvc.New())
		log.Printf("ShippingService listening on %s", shippingListen)
	})

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orderServer, err := ordersvc.New(ctx, shippingDial, productDial)
	if err != nil {
		log.Fatalf("failed to create order service: %v", err)
	}

	go grpcutil.MustServeGRPC(orderListen, func(s *grpc.Server) {
		pb.RegisterOrderServiceServer(s, orderServer)
		log.Printf("OrderService listening on %s", orderListen)
	})

	time.Sleep(100 * time.Millisecond)

	gw, err := gateway.New(ctx, productDial, orderDial, shippingDial)
	if err != nil {
		log.Fatalf("failed to create gateway: %v", err)
	}

	mux := http.NewServeMux()
	gw.RegisterRoutes(mux)
	mux.HandleFunc("/health", gateway.HealthHandler)
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}).Handler(mux)

	log.Printf("HTTP gateway + frontend on %s", httpAddr)
	if err := http.ListenAndServe(httpAddr, handler); err != nil {
		log.Fatalf("HTTP server: %v", err)
	}
}
