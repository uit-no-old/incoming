/*
Incoming!! locked uuid pool

Copyright (C) 2014 Lars Tiede, UiT The Arctic University of Norway


This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
package uidpool

import (
	"fmt"

	"code.google.com/p/go-uuid/uuid"
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
