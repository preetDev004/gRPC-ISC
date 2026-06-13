package order

import (
	"time"

	pb "github.com/preetDev004/gRPC-ISC/gen"
)

// order is the internal domain model stored in memory.
type order struct {
	id         string
	customer   string
	address    string
	product    *pb.Product
	quantity   int32
	totalUSD   float64
	status     string
	shippingID string
	createdAt  time.Time
}
