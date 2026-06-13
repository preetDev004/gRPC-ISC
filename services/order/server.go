// Package order implements the OrderService gRPC server.
// It calls ShippingService over gRPC as a client.
package order

import (
	"context"
	"time"

	"github.com/google/uuid"
	pb "github.com/preetDev004/gRPC-ISC/gen"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.OrderServiceServer.
type Server struct {
	pb.UnimplementedOrderServiceServer

	repo           *Repository
	shippingClient pb.ShippingServiceClient
	productClient  pb.ProductServiceClient
}

// New dials downstream gRPC services and returns a ready-to-register server.
func New(ctx context.Context, shippingAddr, productAddr string) (*Server, error) {
	shippingConn, err := grpcutil.DialWithRetry(ctx, shippingAddr)
	if err != nil {
		return nil, err
	}
	productConn, err := grpcutil.DialWithRetry(ctx, productAddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		repo:           NewRepository(),
		shippingClient: pb.NewShippingServiceClient(shippingConn),
		productClient:  pb.NewProductServiceClient(productConn),
	}, nil
}

// PlaceOrder validates the request, creates an order, and calls ShippingService.
func (s *Server) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "quantity must be > 0")
	}
	if req.GetCustomer() == "" {
		return nil, status.Error(codes.InvalidArgument, "customer name is required")
	}

	prodResp, err := s.productClient.GetProduct(ctx, &pb.GetProductRequest{ProductId: req.GetProductId()})
	if err != nil {
		return nil, err
	}
	prod := prodResp.GetProduct()
	if prod.Stock < req.GetQuantity() {
		return nil, status.Errorf(codes.FailedPrecondition, "insufficient stock: have %d, want %d", prod.Stock, req.GetQuantity())
	}

	orderID := "ord-" + uuid.New().String()[:8]
	total := prod.PriceUsd * float64(req.GetQuantity())

	shipResp, err := s.shippingClient.CreateShipment(ctx, &pb.CreateShipmentRequest{
		OrderId:  orderID,
		Customer: req.GetCustomer(),
		Address:  req.GetAddress(),
		Product:  prod.Name,
		Quantity: req.GetQuantity(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "shipping service error: %v", err)
	}

	o := &order{
		id:         orderID,
		customer:   req.GetCustomer(),
		address:    req.GetAddress(),
		product:    prod,
		quantity:   req.GetQuantity(),
		totalUSD:   total,
		status:     "CONFIRMED",
		shippingID: shipResp.ShippingId,
		createdAt:  time.Now(),
	}
	s.repo.Save(o)

	return &pb.PlaceOrderResponse{
		OrderId:    orderID,
		ShippingId: shipResp.ShippingId,
		Status:     "CONFIRMED",
		TotalUsd:   total,
	}, nil
}

// GetOrder returns a single order by ID.
func (s *Server) GetOrder(_ context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	o, ok := s.repo.Get(req.GetOrderId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.GetOrderId())
	}
	return &pb.GetOrderResponse{Order: toProto(o)}, nil
}

// ListOrders streams all orders, optionally filtered by customer name.
func (s *Server) ListOrders(req *pb.ListOrdersRequest, stream pb.OrderService_ListOrdersServer) error {
	for _, o := range s.repo.List(req.GetCustomer()) {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "client disconnected")
		default:
		}
		if err := stream.Send(&pb.ListOrdersResponse{Order: toProto(o)}); err != nil {
			return err
		}
		time.Sleep(80 * time.Millisecond)
	}
	return nil
}
