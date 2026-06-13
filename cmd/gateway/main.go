// cmd/gateway – HTTP API gateway and static frontend server.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/preetDev004/gRPC-ISC/gateway"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
	"github.com/rs/cors"
)

func main() {
	productAddr := grpcutil.EnvOr("PRODUCT_ADDR", "localhost:9001")
	orderAddr := grpcutil.EnvOr("ORDER_ADDR", "localhost:9003")
	shippingAddr := grpcutil.EnvOr("SHIPPING_ADDR", "localhost:9002")
	httpAddr := grpcutil.EnvOr("HTTP_ADDR", ":8080")
	frontendDir := grpcutil.EnvOr("FRONTEND_DIR", "./frontend")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gw, err := gateway.New(ctx, productAddr, orderAddr, shippingAddr)
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
