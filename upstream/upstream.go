package upstream

import "golang.org/x/net/dns/dnsmessage"

// Upstream who am I asking for the actual DNS record
type Upstream interface {
	AskQuestion(m *dnsmessage.Message) ([]byte, error)
}
