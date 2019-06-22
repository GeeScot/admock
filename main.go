package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gurparit/go-common/httpc"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"net/http"
)

var baseURL1 = "https://1.1.1.1/dns-query?dns=%s"
var baseURL2 = "https://1.0.0.1/dns-query?dns=%s"

func waitForDNS(conn *net.UDPConn) (string, *net.UDPAddr, error) {
	buf := make([]byte, 512)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", nil, errors.New("[err] invalid udp packet")
	}

	var m dnsmessage.Message
	err = m.Unpack(buf)
	if err != nil {
		return "", nil, err
	}

	packed, err := m.Pack()
	return base64.RawURLEncoding.EncodeToString(packed), addr, err
}

func fetchDNSoverTLS(query string) ([]byte, error) {
	req := httpc.HTTP{
		TargetURL: httpc.FormatURL(baseURL1, query),
		Method:    http.MethodGet,
		Headers:   httpc.Headers{"accept": "application/dns-message"},
	}

	return req.Raw()
}

func main() {
	conn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	defer conn.Close()

	for {
		encodedQuery, addr, err := waitForDNS(conn)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		result, err := fetchDNSoverTLS(encodedQuery)

		conn.WriteToUDP(result, addr)
	}
}
