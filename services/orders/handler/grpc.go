package handler

import (
	"context"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/preetDev004/gRPC-ISC/services/common/genproto/orders"
	"github.com/preetDev004/gRPC-ISC/services/orders/service"
	"github.com/preetDev004/gRPC-ISC/services/orders/types"
)

type OrdersGRPCHandler struct {
	orderService types.OrderService // for dependency injection
	orders.UnimplementedOrderServiceServer // for grpc server implementation composition
}

func NewGrpcOrdersService() {
	gRPCHandler := &OrdersGRPCHandler{
		orderService: service.NewOrdersService(),
	}
	// register the OrderService server
}
func (h *OrdersGRPCHandler) CreateOrder(ctx context.Context, req *orders.CreateOrderRequest) (*orders.CreateOrderResponse, error) {
	order := &orders.Order{
		OrderID: 33,
		CustomerID: 21,
		ProductID: 431,
		Quantity: 10,
	}
	if err := h.orderService.CreateOrder(ctx, order); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create order: %v", err)
	}
	return &orders.CreateOrderResponse{
		Status: "Success",
	}, nil
}
func (h *OrdersGRPCHandler) GetOrder(ctx context.Context, req *orders.GetOrderRequest) (*orders.GetOrderResponse, error) {
	
}