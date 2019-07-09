package pool

import (
	"os"
	"sync"
)

// RoundRobin a pool representing a round robin list
type RoundRobin struct {
	mux       sync.Mutex
	upstreams []string
	size      int
	index     int
}

// NewRoundRobin pool with dns servers
func NewRoundRobin() Pool {
	dns1 := os.Getenv("FASTDNS_DNS1")
	dns2 := os.Getenv("FASTDNS_DNS2")

	list := []string{}
	if len(dns1) > 0 {
		list = append(list, dns1)
	}

	if len(dns2) > 0 {
		list = append(list, dns2)
	}

	return &RoundRobin{
		mux:       sync.Mutex{},
		upstreams: list,
		size:      len(list),
		index:     0,
	}
}

// Next get next in pool
func (rr *RoundRobin) Next() string {
	rr.mux.Lock()
	defer rr.mux.Unlock()

	rr.increment()
	return rr.current()
}

func (rr *RoundRobin) current() string {
	return rr.upstreams[rr.index]
}

func (rr *RoundRobin) increment() int {
	nextIndex := rr.index

	rr.index++
	if rr.index == rr.size {
		rr.index = 0
	}

	return nextIndex
}
