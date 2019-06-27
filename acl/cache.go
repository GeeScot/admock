package acl

import (
	"sort"
	"sync"
	"github.com/gurparit/go-common/array"
)

// StringCache simple sorted key cache
type StringCache struct {
	mux  sync.Mutex
	data []string
	Size int
}

// New create new cache
func New() *StringCache {
	return &StringCache{
		mux:  sync.Mutex{},
		data: []string{},
	}
}

// Get get a value
func (sc *StringCache) Get(i int) string {
	return sc.data[i]
}

// Add add a value
func (sc *StringCache) Add(value string) {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	sc.data = append(sc.data, value)
	sc.Size = len(sc.data)
}

// Remove remove a value
func (sc *StringCache) Remove(value string) {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	i := sort.SearchStrings(sc.data, value)
	sc.data = append(sc.data[:i], sc.data[i+1:]...)
	sc.Size = len(sc.data)
}

// Sort sort the cache
func (sc *StringCache) Sort() {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	sort.Strings(sc.data)
}

// Contains check if cache contains a value
func (sc *StringCache) Contains(value string) bool {
	sc.mux.Lock()
	defer sc.mux.Unlock()

	i := sort.SearchStrings(sc.data, value)
	s := sc.data[i-1:i+1]

	return array.Contains(s, value)
}
