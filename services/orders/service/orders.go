package service

import (
	"sync"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/preetDev004/gRPC-ISC/services/common/genproto/orders"
)

type OrdersService struct {
	sync.RWMutex
	orders map[int32]*orders.Order
}

func NewOrdersService() *OrdersService {
	return &OrdersService{}
}

func (s *OrdersService) CreateOrder(ctx context.Context, order *orders.Order) error {
	s.Lock()
	defer s.Unlock()
	if _, exists := s.orders[order.OrderID]; exists {
		return status.Errorf(codes.AlreadyExists, "order with ID %d already exists", order.OrderID)
	}
	s.orders[order.OrderID] = order
	return nil
}

func (s *OrdersService) GetOrder(ctx context.Context, customerID int32) ([]*orders.Order, error) {
	s.RLock()
	defer s.RUnlock()
	orders := make([]*orders.Order, 0)
	for _, order := range s.orders {
		if order.CustomerID == customerID {
			orders = append(orders, order)
		}
	}
	if len(orders) == 0 {
		return nil, status.Errorf(codes.NotFound, "no orders found for customer ID %d", customerID)
	}
	return orders, nil
}