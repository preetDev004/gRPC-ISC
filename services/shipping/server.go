// Package shipping implements the ShippingService gRPC server.
package shipping

import (
	"context"
	"time"

	pb "github.com/preetDev004/gRPC-ISC/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements pb.ShippingServiceServer.
type Server struct {
	pb.UnimplementedShippingServiceServer
	repo *Repository
}

// New returns a ready-to-register ShippingService server.
func New() *Server {
	return &Server{repo: NewRepository()}
}

// CreateShipment books a shipment and stores its tracking timeline.
func (s *Server) CreateShipment(_ context.Context, req *pb.CreateShipmentRequest) (*pb.CreateShipmentResponse, error) {
	if req.GetOrderId() == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}

	details, state := createShipment(req)
	s.repo.Save(details.ShippingID, state)

	return &pb.CreateShipmentResponse{
		ShippingId:       details.ShippingID,
		Carrier:          details.Carrier,
		TrackingNumber:   details.TrackingNumber,
		EstimatedArrival: details.EstimatedArrival,
	}, nil
}

// TrackShipment streams pre-built status updates with simulated delays.
func (s *Server) TrackShipment(req *pb.TrackShipmentRequest, stream pb.ShippingService_TrackShipmentServer) error {
	state, ok := s.repo.Get(req.GetShippingId())
	if !ok {
		return status.Errorf(codes.NotFound, "shipment %q not found", req.GetShippingId())
	}

	for i, upd := range state.updates {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "client disconnected")
		default:
		}

		if err := stream.Send(&pb.TrackShipmentResponse{Update: upd}); err != nil {
			return err
		}

		time.Sleep(time.Duration(2*(i+1)) * time.Second)
	}
	return nil
}
