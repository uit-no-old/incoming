package uidpool

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	//"log"
	"sync"
	//"time"
)

// Help text for LockedUUIDPool
type LockedUUIDPool struct {
	sync.Mutex
	uids map[string]bool
}

func NewLockedUUIDPool() *LockedUUIDPool {
	return &LockedUUIDPool{uids: make(map[string]bool)}
}

// Help text for New
func (p *LockedUUIDPool) New() string {
	p.Lock()
	defer p.Unlock()
	for {
		new_id := uuid.New()
		if _, in_there := p.uids[new_id]; !in_there {
			p.uids[new_id] = true
			return new_id
		}
	}
}

// Help text for Remove
func (p *LockedUUIDPool) Remove(id string) error {
	p.Lock()
	defer p.Unlock()
	if _, exists := p.uids[id]; !exists {
		return fmt.Errorf("Tried to remove non-existing uid %s", id)
	}

	delete(p.uids, id)
	return nil
}

// Help text for Size
func (p *LockedUUIDPool) Size() int {
	return len(p.uids)
}
