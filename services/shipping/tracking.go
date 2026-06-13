package shipping

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	pb "github.com/preetDev004/gRPC-ISC/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var carriers = []string{"FedEx", "UPS", "DHL", "Canada Post"}

type shipmentDetails struct {
	ShippingID       string
	Carrier          string
	TrackingNumber   string
	EstimatedArrival string
}

// createShipment builds shipment metadata and the full tracking timeline.
func createShipment(req *pb.CreateShipmentRequest) (shipmentDetails, *trackingState) {
	carrier := carriers[rand.Intn(len(carriers))]
	shippingID := "ship-" + uuid.New().String()[:8]
	tracking := fmt.Sprintf("1Z%06d", rand.Intn(999999))
	eta := time.Now().Add(time.Duration(2+rand.Intn(5)) * 24 * time.Hour).Format("2006-01-02")

	now := time.Now()
	addr := req.GetAddress()
	if addr == "" {
		addr = "1750 Finch Ave East, Toronto, ON"
	}

	updates := []*pb.ShipmentUpdate{
		{
			ShippingId: shippingID, Status: "PROCESSING",
			Location:  "Warehouse – Toronto, ON",
			Note:      "Shipment received and being packed.",
			Timestamp: timestamppb.New(now),
		},
		{
			ShippingId: shippingID, Status: "DISPATCHED",
			Location:  "Sortation Centre – Mississauga, ON",
			Note:      fmt.Sprintf("Picked up by %s.", carrier),
			Timestamp: timestamppb.New(now.Add(2 * time.Second)),
		},
		{
			ShippingId: shippingID, Status: "IN_TRANSIT",
			Location:  "Distribution Hub – Ottawa, ON",
			Note:      "In transit to destination region.",
			Timestamp: timestamppb.New(now.Add(4 * time.Second)),
		},
		{
			ShippingId: shippingID, Status: "OUT_FOR_DELIVERY",
			Location:  "Local Depot – " + addr,
			Note:      "Out for delivery today.",
			Timestamp: timestamppb.New(now.Add(6 * time.Second)),
		},
		{
			ShippingId: shippingID, Status: "DELIVERED",
			Location:  addr,
			Note:      "Package delivered. Thank you for your order!",
			Timestamp: timestamppb.New(now.Add(8 * time.Second)),
		},
	}

	return shipmentDetails{
		ShippingID:       shippingID,
		Carrier:          carrier,
		TrackingNumber:   tracking,
		EstimatedArrival: eta,
	}, &trackingState{updates: updates}
}
