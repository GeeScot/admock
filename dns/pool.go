package dns

import (
	"os"
	"sync"
)

// Pool of dns servers
type Pool struct {
	mux       sync.Mutex
	upstreams []string
	size      int
	index     int
}

// NewPool with dns servers
func NewPool() *Pool {
	dns1 := os.Getenv("FASTDNS_DNS1")
	dns2 := os.Getenv("FASTDNS_DNS2")

	list := []string{}
	if len(dns1) > 0 {
		list = append(list, dns1)
	}

	if len(dns2) > 0 {
		list = append(list, dns2)
	}

	return &Pool{
		mux:       sync.Mutex{},
		upstreams: list,
		size:      len(list),
		index:     0,
	}
}

// Next get next in pool
func (pool *Pool) Next() string {
	pool.mux.Lock()
	defer pool.mux.Unlock()

	pool.increment()
	return pool.current()
}

func (pool *Pool) current() string {
	return pool.upstreams[pool.index]
}

func (pool *Pool) increment() int {
	nextIndex := pool.index

	pool.index++
	if pool.index == pool.size {
		pool.index = 0
	}

	return nextIndex
}
