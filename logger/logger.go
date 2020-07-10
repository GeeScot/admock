package logger

import (
	"fmt"
	"net"
	"time"

	"github.com/codescot/admock/dns"
	"golang.org/x/net/dns/dnsmessage"
)

// Logger log class
type Logger struct {
	Debug bool
	Log   chan []byte
}

// Start start the logger routine
func (l *Logger) Start() {
	if !l.Debug {
		return
	}

	for response := range l.Log {
		var m dnsmessage.Message
		_ = m.Unpack(response)

		domain := dns.Domain(&m)
		if len(m.Answers) <= 0 {
			continue
		}

		for _, answer := range m.Answers {
			body := answer.Body

			if ip, ok := body.(*dnsmessage.AResource); ok {
				actualIP := net.IP{ip.A[0], ip.A[1], ip.A[2], ip.A[3]}

				PrintWithTimeStamp(fmt.Sprintf("%16v ~ %s", actualIP, domain))
			}
		}
	}
}

// PrintWithTimeStamp print a message with a standard time format
func PrintWithTimeStamp(message string) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s\n", timestamp, message)
}
