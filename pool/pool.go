package pool

// Pool of dns servers
type Pool interface {
	Next() string
}
