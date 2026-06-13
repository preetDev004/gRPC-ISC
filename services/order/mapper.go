package order

import (
	pb "github.com/preetDev004/gRPC-ISC/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProto(o *order) *pb.Order {
	return &pb.Order{
		Id:         o.id,
		Customer:   o.customer,
		Address:    o.address,
		Product:    o.product,
		Quantity:   o.quantity,
		TotalUsd:   o.totalUSD,
		Status:     o.status,
		ShippingId: o.shippingID,
		CreatedAt:  timestamppb.New(o.createdAt),
	}
}
