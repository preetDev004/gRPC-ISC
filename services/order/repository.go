package order

import "sync"

// Repository is the in-memory order store for this demo.
type Repository struct {
	mu     sync.RWMutex
	orders map[string]*order
}

// NewRepository returns an empty in-memory order store.
func NewRepository() *Repository {
	return &Repository{orders: make(map[string]*order)}
}

func (r *Repository) Save(o *order) {
	r.mu.Lock()
	r.orders[o.id] = o
	r.mu.Unlock()
}

func (r *Repository) Get(id string) (*order, bool) {
	r.mu.RLock()
	o, ok := r.orders[id]
	r.mu.RUnlock()
	return o, ok
}

func (r *Repository) List(customerFilter string) []*order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*order, 0, len(r.orders))
	for _, o := range r.orders {
		if customerFilter == "" || o.customer == customerFilter {
			out = append(out, o)
		}
	}
	return out
}
