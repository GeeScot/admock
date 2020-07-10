package pool

import (
	"os"
)

// Single a pool representing a single item
type Single struct {
	upstream string
}

// NewSingle pool with single dns server
func NewSingle() Pool {
	dns1 := os.Getenv("ADMOCK_DNS1")

	return &Single{
		upstream: dns1,
	}
}

// Next get next in pool
func (s *Single) Next() string {
	return s.upstream
}
