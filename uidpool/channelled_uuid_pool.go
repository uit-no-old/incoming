/*
Incoming!! channelled uuid pool

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
	//"sync"
	//"time"
)

// Help text for ChannelledUUIDPool
type ChannelledUUIDPool struct {
	new_uuids    chan string
	remove_uuids chan string
	quit         chan bool
	uids         map[string]bool
}

func NewChannelledUUIDPool() *ChannelledUUIDPool {
	p := new(ChannelledUUIDPool)
	p.new_uuids = make(chan string)
	p.remove_uuids = make(chan string)
	p.quit = make(chan bool)
	p.uids = make(map[string]bool)

	go p.mainLoop()

	return p
}

func (p *ChannelledUUIDPool) mainLoop() error {
	next_uuid := p.makeNewUUID()

	for cont := true; cont; {
		select {
		case p.new_uuids <- next_uuid:
			next_uuid = p.makeNewUUID()
		case uuid := <-p.remove_uuids:
			_ = p.remove(uuid)
			// TODO: can't propagate the error to whoever sent uuid
		case <-p.quit:
			cont = false
		}
	}

	return nil
}

// Help text for makeNewUUID
func (p *ChannelledUUIDPool) makeNewUUID() string {
	for {
		new_id := uuid.New()
		if _, in_there := p.uids[new_id]; !in_there {
			p.uids[new_id] = true
			return new_id
		}
	}
}

// Help text for New
func (p *ChannelledUUIDPool) New() string {
	return <-p.new_uuids
}

// Help text for remove
func (p *ChannelledUUIDPool) remove(id string) error {
	if _, exists := p.uids[id]; !exists {
		return fmt.Errorf("Tried to remove non-existing uid %s", id)
	}

	delete(p.uids, id)
	return nil
}

// Help text for Remove
func (p *ChannelledUUIDPool) Remove(id string) error {
	p.remove_uuids <- id
	// TODO: can't wait for completion of that op here
	return nil
}

// Help text for Size
func (p *ChannelledUUIDPool) Size() int {
	return len(p.uids)
}
