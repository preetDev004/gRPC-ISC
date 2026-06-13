package shipping

import pb "github.com/preetDev004/gRPC-ISC/gen"

// trackingState holds the pre-computed update sequence for a shipment.
type trackingState struct {
	updates []*pb.ShipmentUpdate
}
