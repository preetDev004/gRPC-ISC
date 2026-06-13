// Package product implements the ProductService gRPC server.
package product

import (
	"context"
	"time"

	pb "github.com/preetDev004/gRPC-ISC/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.ProductServiceServer.
type Server struct {
	pb.UnimplementedProductServiceServer
}

// New returns a ready-to-register ProductService server.
func New() *Server { return &Server{} }

// ListProducts streams every matching product to the client.
func (s *Server) ListProducts(req *pb.ListProductsRequest, stream pb.ProductService_ListProductsServer) error {
	for _, p := range List(req.GetCategoryFilter()) {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "client disconnected")
		default:
		}

		if err := stream.Send(&pb.ListProductsResponse{Product: p}); err != nil {
			return err
		}

		// Simulate a real DB / network delay so streaming is visible.
		time.Sleep(120 * time.Millisecond)
	}
	return nil
}

// GetProduct returns a single product by ID.
func (s *Server) GetProduct(_ context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	p, found := FindByID(req.GetProductId())
	if !found {
		return nil, status.Errorf(codes.NotFound, "product %q not found", req.GetProductId())
	}
	return &pb.GetProductResponse{Product: p}, nil
}
