package upload

import (
	"log"
	"sync"

	"source.uit.no/star-apt/incoming/uidpool"
)

type UploaderPool interface {
	Get(string) (Uploader, bool)

	Put(Uploader) string

	// Remove removes an uploader, identified by its id, from the pool. No
	// problem if the given id does not exist
	Remove(string)

	Size() int
}

type LockedUploaderPool struct {
	uidPool   uidpool.UIDPool
	lock      sync.Mutex
	uploaders map[string]Uploader
}

func NewLockedUploaderPool() *LockedUploaderPool {
	p := new(LockedUploaderPool)
	p.uidPool = uidpool.NewUIDPool()
	p.uploaders = make(map[string]Uploader)
	return p
}

func (p *LockedUploaderPool) Get(id string) (res Uploader, exists bool) {
	p.lock.Lock()
	res, exists = p.uploaders[id]
	p.lock.Unlock()
	return
}

func (p *LockedUploaderPool) Put(ul Uploader) (id string) {
	id = p.uidPool.New()

	p.lock.Lock()
	p.uploaders[id] = ul
	p.lock.Unlock()

	log.Printf("put uploader %s into pool. Pool size: %d", id, p.Size())
	return
}

func (p *LockedUploaderPool) Remove(id string) {
	p.lock.Lock()
	delete(p.uploaders, id)
	p.lock.Unlock()

	p.uidPool.Remove(id)
	log.Printf("removed uploader %s from pool. Pool size: %d", id, p.Size())
	return
}

func (p *LockedUploaderPool) Size() (s int) {
	p.lock.Lock()
	s = len(p.uploaders)
	p.lock.Unlock()
	return
}
