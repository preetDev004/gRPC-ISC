package shipping

import "sync"

// Repository is the in-memory shipment store for this demo.
type Repository struct {
	mu        sync.RWMutex
	shipments map[string]*trackingState
}

// NewRepository returns an empty in-memory shipment store.
func NewRepository() *Repository {
	return &Repository{shipments: make(map[string]*trackingState)}
}

func (r *Repository) Save(shippingID string, state *trackingState) {
	r.mu.Lock()
	r.shipments[shippingID] = state
	r.mu.Unlock()
}

func (r *Repository) Get(shippingID string) (*trackingState, bool) {
	r.mu.RLock()
	state, ok := r.shipments[shippingID]
	r.mu.RUnlock()
	return state, ok
}
