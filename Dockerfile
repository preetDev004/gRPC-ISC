# ── Stage 1: Generate protobuf code ─────────────────────────────────────────
FROM bufbuild/buf:latest AS proto-gen

WORKDIR /workspace
COPY buf.yaml buf.gen.yaml ./
COPY proto/ proto/

RUN buf dep update && buf generate


# ── Stage 2: Build Go binaries ───────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY --from=proto-gen /workspace/gen ./gen

COPY cmd/      ./cmd/
COPY gateway/  ./gateway/
COPY internal/ ./internal/
COPY services/ ./services/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/product-service ./cmd/product-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/shipping-service ./cmd/shipping-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/order-service ./cmd/order-service && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /bin/gateway ./cmd/gateway


# ── Stage 3a: Product service runtime ───────────────────────────────────────
FROM scratch AS product-service

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/product-service /service

EXPOSE 9001
ENV GRPC_ADDR=:9001
ENTRYPOINT ["/service"]


# ── Stage 3b: Shipping service runtime ──────────────────────────────────────
FROM scratch AS shipping-service

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/shipping-service /service

EXPOSE 9002
ENV GRPC_ADDR=:9002
ENTRYPOINT ["/service"]


# ── Stage 3c: Order service runtime ─────────────────────────────────────────
FROM scratch AS order-service

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/order-service /service

EXPOSE 9003
ENV GRPC_ADDR=:9003
ENV SHIPPING_ADDR=shipping-service:9002
ENV PRODUCT_ADDR=product-service:9001
ENTRYPOINT ["/service"]


# ── Stage 3d: Gateway runtime ───────────────────────────────────────────────
FROM scratch AS gateway

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/gateway /service
COPY frontend/ /frontend/

EXPOSE 8080
ENV HTTP_ADDR=:8080
ENV FRONTEND_DIR=/frontend
ENV PRODUCT_ADDR=product-service:9001
ENV ORDER_ADDR=order-service:9003
ENV SHIPPING_ADDR=shipping-service:9002
ENTRYPOINT ["/service"]
