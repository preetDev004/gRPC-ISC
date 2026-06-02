package types

import (
	"context"

	"github.com/preetDev004/gRPC-ISC/services/common/genproto/orders"
)

type OrderService interface {
	CreateOrder(context.Context, *orders.Order) error
	GetOrder(context.Context, int32) ([]*orders.Order, error)
}