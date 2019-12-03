package upstream

import (
	"bufio"
	"fmt"
	"net"

	"github.com/geescot/fastdns/pool"
	"golang.org/x/net/dns/dnsmessage"
)

// UDPUpstream UDP Struct
type UDPUpstream struct {
	Pool pool.Pool
}

// AskQuestion send DNS request to Cloudflare via UDP
func (u *UDPUpstream) AskQuestion(m *dnsmessage.Message) ([]byte, error) {
	packed, err := m.Pack()
	if err != nil {
		panic(err)
	}

	server := u.Pool.Next()

	// addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:53", server))
	c, err := net.Dial("udp", fmt.Sprintf("%s:53", server))
	if err != nil {
		panic(err)
	}

	defer c.Close()
	c.Write(packed)

	rsp := make([]byte, 512)

	n, err := bufio.NewReader(c).Read(rsp)
	if err != nil {
		panic(err)
	}

	return rsp[0:n], nil
}
