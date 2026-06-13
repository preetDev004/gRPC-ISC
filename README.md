# grpc-isc

Inter-service communication (Go + gRPC): browse a product catalogue, place orders,
and watch live shipment tracking wired together through real gRPC streams.

```
Browser (SSE) ──▶ HTTP Gateway (:8080)
                     │
          ┌──────────┼──────────────────┐
          │          │                  │
  ProductService  OrderService     ShippingService
    (:9001)         (:9003)            (:9002)
    gRPC srv        gRPC srv           gRPC srv
                    gRPC client ──────▶ (calls Shipping internally)
```

---

## Architecture

| Layer | What it does |
|---|---|
| **proto/** | Source of truth — three `.proto` files define every RPC |
| **gen/** | Auto-generated Go structs + gRPC stubs (`make gen`) |
| **services/product** | `ProductService` — streams the catalogue via server-side streaming |
| **services/shipping** | `ShippingService` — creates shipments, streams tracking updates |
| **services/order** | `OrderService` — places orders, **calls ShippingService over gRPC** as a client |
| **gateway/** | Thin HTTP layer: translates REST+SSE → gRPC for the browser |
| **frontend/** | Static HTML/JS — uses `EventSource` to consume SSE streams |
| **cmd/server** | Single binary that boots all three gRPC servers + the HTTP gateway |

### gRPC patterns demonstrated

| Pattern | Where |
|---|---|
| **Server-side streaming** | `ProductService.ListProducts` — streams products one by one |
| **Server-side streaming** | `ShippingService.TrackShipment` — streams live status updates |
| **Server-side streaming** | `OrderService.ListOrders` — streams all orders |
| **Unary RPC** | `OrderService.PlaceOrder`, `ShippingService.CreateShipment`, etc. |
| **Service-to-service gRPC** | `OrderService` dials `ShippingService` as a gRPC **client** |

---

## Prerequisites

| Tool | Version | Install |
|---|---|---|
| Go | 1.22+ | https://go.dev/dl |
| buf | latest | https://buf.build/docs/installation |
| Docker | any recent | https://docs.docker.com/get-docker |

---

## Quick start (local)

```bash
# 1. Clone and enter the project
git clone https://github.com/preetDev004/gRPC-ISC grpc-isc
cd grpc-isc

# 2. Generate protobuf / gRPC Go code
make gen            # runs buf dep update && buf generate && go mod tidy

# 3. Run the server
make run            # starts on http://localhost:8080
```

Open **http://localhost:8080** in your browser.

---

## Quick start (Docker)

```bash
# Build the image and start the container
make docker-run

# Open the app
open http://localhost:8080

# Stop
make docker-stop
```

The Dockerfile is a three-stage build:
1. **proto-gen** — uses the official `bufbuild/buf` image to run `buf generate`
2. **builder** — compiles a statically-linked Go binary
3. **runtime** — copies only the binary + frontend into a `scratch` image (~10 MB total)


## Testing the gRPC services with grpcurl

With the server running locally, you can hit the gRPC endpoints directly:

```bash
# List all products (server-streaming)
grpcurl -plaintext -d '{}' localhost:9001 shop.ProductService/ListProducts

# Get a single product
grpcurl -plaintext -d '{"product_id":"prod-003"}' localhost:9001 shop.ProductService/GetProduct

# Place an order
grpcurl -plaintext -d '{
  "product_id": "prod-001",
  "quantity": 2,
  "customer": "Preet Patel",
  "address": "123 King St W, Toronto ON"
}' localhost:9003 shop.OrderService/PlaceOrder

# Track a shipment (server-streaming) — use the shipping_id from PlaceOrder response
grpcurl -plaintext -d '{"shipping_id":"ship-XXXXXXXX"}' localhost:9002 shop.ShippingService/TrackShipment

# List all orders
grpcurl -plaintext -d '{}' localhost:9003 shop.OrderService/ListOrders
```

---

### Output
<img width="1918" height="961" alt="Screenshot 2026-06-13 at 12 45 19 PM" src="https://github.com/user-attachments/assets/b551f2d9-065f-41bb-ad8a-9e1ce95c75a4" />
<img width="1914" height="956" alt="Screenshot 2026-06-13 at 12 45 50 PM" src="https://github.com/user-attachments/assets/a08b0c50-f10e-4c58-a179-9e216a0a40d3" />


