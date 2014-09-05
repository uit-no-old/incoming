/*
A uid pool hands out guaranteed unique randomly generated string ids. Get a new
id with New(), and when you no longer need it, call Remove().

In this package are a few different implementations for this, mainly for shits
and giggles. I'm learning, okay?
*/
package uidpool

import (
//"fmt"
//"log"
)

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
