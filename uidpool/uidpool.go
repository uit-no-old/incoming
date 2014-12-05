/*
A uid pool hands out guaranteed unique randomly generated string ids. Get a new
id with New(), and when you no longer need it, call Remove().

In this package are a few different implementations for this, mainly for shits
and giggles. I'm learning, okay?


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

// Help text for UIDPool interface
type UIDPool interface {
	// Help text for New
	New() string

	// Help text for Remove
	Remove(string) error

	Size() int
}

// Help text for NewUIDPool
func NewUIDPool() UIDPool {
	return NewLockedUUIDPool()
	//return NewChannelledUUIDPool()
}
