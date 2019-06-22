package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gurparit/go-common/httputil"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"net/http"
	"time"
)

var baseURL1 = "https://1.1.1.1/dns-query?dns=%s"
var baseURL2 = "https://1.0.0.1/dns-query?dns=%s"

func waitForDNS(conn *net.UDPConn) (*dnsmessage.Message, string, *net.UDPAddr, error) {
	buf := make([]byte, 512)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, "", nil, errors.New("[err] invalid udp packet")
	}

	var m dnsmessage.Message
	err = m.Unpack(buf)
	if err != nil {
		return nil, "", nil, err
	}

	packed, err := m.Pack()
	return &m, base64.RawURLEncoding.EncodeToString(packed), addr, err
}

func fetchDNSoverTLS(query string) ([]byte, error) {
	headers := httputil.Headers{"accept": "application/dns-message"}
	req := httputil.HTTP{
		TargetURL: httputil.FormatURL(baseURL1, query),
		Method:    http.MethodGet,
		Headers:   headers,
	}

	return req.Raw()
}

func isBlacklistDomain(dns dnsmessage.Message) ([]byte, bool) {
	domain := dns.Questions[0].Name.String()
	_, found := inMemoryCache.Get(domain)
	if !found {
		return nil, false
	}

	fakeDNS := NewFakeDNS(dns.Header.ID, domain)
	packed, err := fakeDNS.Pack()
	if err != nil {
		fmt.Println(err.Error())
		return nil, false
	}

	return packed, true
}

func isCachedDomain(encodedQuery string) ([]byte, bool) {
	cachedItem, found := inMemoryCache.Get(encodedQuery)
	if !found {
		return nil, false
	}

	cachedDNS := cachedItem.([]byte)
	return cachedDNS, true
}

func addToCache(key string, record []byte) {
	// var m dnsmessage.Message
	// err := m.Unpack(record)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	// fmt.Println(m)
	inMemoryCache.Set(key, record, 3600*time.Second)
}

var inMemoryCache *cache.Cache

func main() {
	inMemoryCache = cache.New(3600*time.Minute, 10*time.Minute)

	conn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	defer conn.Close()

	LoadBlacklists(inMemoryCache)

	for {
		dns, encodedQuery, addr, err := waitForDNS(conn)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		if fakeDNS, blacklisted := isBlacklistDomain(*dns); blacklisted {
			conn.WriteToUDP(fakeDNS, addr)
			continue
		}

		if cachedDNS, cached := isCachedDomain(encodedQuery); cached {
			conn.WriteToUDP(cachedDNS, addr)
			continue
		}

		result, err := fetchDNSoverTLS(encodedQuery)
		addToCache(encodedQuery, result)
		conn.WriteToUDP(result, addr)
	}
}
