package cache

import (
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// ResourceRecord type alias for []dnsmessage.Resource
type ResourceRecord []dnsmessage.Resource

// ResourceCache simple sorted key cache
type ResourceCache struct {
	mux  sync.Mutex
	data map[string]ResourceRecord
}

// Resources create new string cache
func Resources() *ResourceCache {
	return &ResourceCache{
		mux:  sync.Mutex{},
		data: make(map[string]ResourceRecord),
	}
}

// Get get a value
func (sc *ResourceCache) Get(key string) (ResourceRecord, bool) {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	value, found := sc.data[key]
	return value, found
}

// Add add a value
func (sc *ResourceCache) Add(key string, value ResourceRecord) {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	sc.data[key] = value
}

// AddWithExpiry add a value with a timer set to expire the record
func (sc *ResourceCache) AddWithExpiry(key string, value ResourceRecord, expiry time.Duration) {
	sc.Add(key, value)
	sc.onExpiry(key, expiry)
}

// Remove remove a value
func (sc *ResourceCache) Remove(key string) {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	delete(sc.data, key)
}

func (sc *ResourceCache) onExpiry(key string, expiry time.Duration) {
	timer := time.NewTimer(expiry * time.Second)
	go func() {
		<-timer.C
		sc.Remove(key)
	}()
}
