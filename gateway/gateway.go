// Package gateway exposes the gRPC services over HTTP so the static
// frontend can interact with them without needing gRPC-Web or a proxy.
//
// Route map:
//
//	GET  /api/products          → ProductService.ListProducts  (SSE stream)
//	POST /api/orders            → OrderService.PlaceOrder
//	GET  /api/orders            → OrderService.ListOrders      (SSE stream)
//	GET  /api/orders/{id}       → OrderService.GetOrder
//	GET  /api/shipments/{id}/track → ShippingService.TrackShipment (SSE stream)
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	pb "github.com/preetDev004/gRPC-ISC/gen"
	"github.com/preetDev004/gRPC-ISC/internal/grpcutil"
)

// Gateway holds gRPC client stubs for all downstream services.
type Gateway struct {
	products pb.ProductServiceClient
	orders   pb.OrderServiceClient
	shipping pb.ShippingServiceClient
}

// New dials all three gRPC services and returns a configured Gateway.
func New(ctx context.Context, productAddr, orderAddr, shippingAddr string) (*Gateway, error) {
	pc, err := grpcutil.DialWithRetry(ctx, productAddr)
	if err != nil {
		return nil, fmt.Errorf("dial product: %w", err)
	}
	oc, err := grpcutil.DialWithRetry(ctx, orderAddr)
	if err != nil {
		return nil, fmt.Errorf("dial order: %w", err)
	}
	sc, err := grpcutil.DialWithRetry(ctx, shippingAddr)
	if err != nil {
		return nil, fmt.Errorf("dial shipping: %w", err)
	}

	return &Gateway{
		products: pb.NewProductServiceClient(pc),
		orders:   pb.NewOrderServiceClient(oc),
		shipping: pb.NewShippingServiceClient(sc),
	}, nil
}

// RegisterRoutes attaches all HTTP handlers to mux.
func (g *Gateway) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/products", g.handleProducts)
	mux.HandleFunc("/api/orders", g.handleOrders)
	mux.HandleFunc("/api/orders/", g.handleOrderByID)        // /api/orders/{id}
	mux.HandleFunc("/api/shipments/", g.handleShipmentTrack) // /api/shipments/{id}/track
}

// ── Products ──────────────────────────────────────────────────────────────────

func (g *Gateway) handleProducts(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	stream, err := g.products.ListProducts(r.Context(), &pb.ListProductsRequest{CategoryFilter: category})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sseStream(w, r, func() (any, error) {
		msg, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return msg.GetProduct(), nil
	})
}

// ── Orders ────────────────────────────────────────────────────────────────────

func (g *Gateway) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		g.placeOrder(w, r)
	case http.MethodGet:
		g.listOrders(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g *Gateway) placeOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int32  `json:"quantity"`
		Customer  string `json:"customer"`
		Address   string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := g.orders.PlaceOrder(r.Context(), &pb.PlaceOrderRequest{
		ProductId: body.ProductID,
		Quantity:  body.Quantity,
		Customer:  body.Customer,
		Address:   body.Address,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp)
}

func (g *Gateway) listOrders(w http.ResponseWriter, r *http.Request) {
	customer := r.URL.Query().Get("customer")
	stream, err := g.orders.ListOrders(r.Context(), &pb.ListOrdersRequest{Customer: customer})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sseStream(w, r, func() (any, error) {
		msg, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return msg.GetOrder(), nil
	})
}

func (g *Gateway) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	// Path: /api/orders/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/orders/")
	if id == "" {
		http.Error(w, "order id required", http.StatusBadRequest)
		return
	}
	resp, err := g.orders.GetOrder(r.Context(), &pb.GetOrderRequest{OrderId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, resp.GetOrder())
}

// ── Shipment tracking ─────────────────────────────────────────────────────────

func (g *Gateway) handleShipmentTrack(w http.ResponseWriter, r *http.Request) {
	// Path: /api/shipments/{id}/track
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/shipments/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "shipping id required", http.StatusBadRequest)
		return
	}
	shippingID := parts[0]

	stream, err := g.shipping.TrackShipment(r.Context(), &pb.TrackShipmentRequest{ShippingId: shippingID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sseStream(w, r, func() (any, error) {
		msg, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return msg.GetUpdate(), nil
	})
}

// sseStream writes server-sent events until recv returns io.EOF or the client
// disconnects. Each message is a JSON-encoded data: line.
func sseStream(w http.ResponseWriter, r *http.Request, recv func() (any, error)) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	for {
		msg, err := recv()
		if err == io.EOF {
			fmt.Fprintf(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			return
		}
		if err != nil {
			log.Printf("stream error: %v", err)
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", jsonErr(err))
			flusher.Flush()
			return
		}

		b, _ := json.Marshal(msg)
		fmt.Fprintf(w, "data: %s\n\n", b)
		flusher.Flush()
	}
}

// ── JSON helpers ──────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeGRPCError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(grpcStatusToHTTP(err))
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func jsonErr(err error) string {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return string(b)
}

func grpcStatusToHTTP(err error) int {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "NotFound"):
		return http.StatusNotFound
	case strings.Contains(msg, "InvalidArgument"):
		return http.StatusBadRequest
	case strings.Contains(msg, "FailedPrecondition"):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// HealthHandler returns 200 OK for container health checks.
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}
