package utils

import "sync"

type IdPool struct {
	mu      sync.Mutex
	nextID  int32
	freeIDs []int32
}

func NewIdPool(startID int32) *IdPool {
	return &IdPool{
		nextID:  startID,
		freeIDs: make([]int32, 0),
	}
}

// Get an ID from the pool
func (p *IdPool) Borrow() int32 {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If we have reclaimed IDs, use those first
	if len(p.freeIDs) > 0 {
		id := p.freeIDs[len(p.freeIDs)-1]
		p.freeIDs = p.freeIDs[:len(p.freeIDs)-1]
		return id
	}

	// Otherwise, generate a new one
	id := p.nextID
	p.nextID++
	return id
}

// Return an ID to the pool
func (p *IdPool) Return(id int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.freeIDs = append(p.freeIDs, id)
}
