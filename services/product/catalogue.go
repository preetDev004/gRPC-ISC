package product

import (
	"strings"

	pb "github.com/preetDev004/gRPC-ISC/gen"
)

// catalogue is the in-memory product source for this demo.
var catalogue = []*pb.Product{
	{Id: "prod-001", Name: "Mechanical Keyboard", Description: "Tactile 75% layout, PBT keycaps, hot-swap switches.", Category: "Tech", PriceUsd: 149.99, Stock: 42, ImageEmoji: "⌨️"},
	{Id: "prod-002", Name: "Ergonomic Mouse", Description: "Wireless, 4000 DPI, sculpted right-hand grip.", Category: "Tech", PriceUsd: 79.99, Stock: 128, ImageEmoji: "🖱️"},
	{Id: "prod-003", Name: "4K Monitor", Description: "27-inch IPS panel, 144 Hz, USB-C power delivery.", Category: "Tech", PriceUsd: 549.00, Stock: 15, ImageEmoji: "🖥️"},
	{Id: "prod-004", Name: "Standing Desk", Description: "Electric sit-stand, dual motor, memory presets.", Category: "Furniture", PriceUsd: 699.00, Stock: 8, ImageEmoji: "🪑"},
	{Id: "prod-005", Name: "Noise-Cancelling Headphones", Description: "40 hr battery, adaptive ANC, USB-C charging.", Category: "Tech", PriceUsd: 299.99, Stock: 55, ImageEmoji: "🎧"},
	{Id: "prod-006", Name: "Desk Lamp", Description: "LED, 5 colour temps, wireless charging base.", Category: "Furniture", PriceUsd: 59.99, Stock: 200, ImageEmoji: "💡"},
	{Id: "prod-007", Name: "Webcam 4K", Description: "Auto-focus, built-in mic, privacy shutter.", Category: "Tech", PriceUsd: 119.99, Stock: 73, ImageEmoji: "📷"},
	{Id: "prod-008", Name: "Cable Management Kit", Description: "Velcro ties, desk grommet, under-desk tray.", Category: "Accessories", PriceUsd: 24.99, Stock: 500, ImageEmoji: "🔌"},
	{Id: "prod-009", Name: "USB-C Hub 10-in-1", Description: "HDMI 4K, SD card, 100W PD, Ethernet.", Category: "Accessories", PriceUsd: 64.99, Stock: 90, ImageEmoji: "🔋"},
	{Id: "prod-010", Name: "Laptop Stand", Description: "Aluminium, foldable, six height settings.", Category: "Accessories", PriceUsd: 44.99, Stock: 160, ImageEmoji: "💻"},
}

// FindByID returns a product by ID. Used by the order service internally.
func FindByID(id string) (*pb.Product, bool) {
	for _, p := range catalogue {
		if p.Id == id {
			return p, true
		}
	}
	return nil, false
}

// List returns all products, optionally filtered by category (case-insensitive).
func List(categoryFilter string) []*pb.Product {
	if categoryFilter == "" {
		return catalogue
	}

	out := make([]*pb.Product, 0, len(catalogue))
	for _, p := range catalogue {
		if strings.EqualFold(p.Category, categoryFilter) {
			out = append(out, p)
		}
	}
	return out
}
