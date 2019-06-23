package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gurparit/go-common/httputil"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/dns/dnsmessage"
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
	_, found := blacklistCache.Get(domain)
	if !found {
		return nil, false
	}

	fakeDNS := NewFakeDNS(dns.Header.ID, domain)
	packed, err := fakeDNS.Pack()
	if err != nil {
		panic(err)
	}

	return packed, true
}

func isCachedDomain(id uint16, domain string) ([]byte, bool) {
	cachedItem, found := inMemoryCache.Get(domain)
	if !found {
		return nil, false
	}

	cachedDNS := cachedDNSRecord(id, domain, cachedItem)
	data, _ := cachedDNS.Pack()
	return data, true
}

func cachedDNSRecord(id uint16, domain string, resouceBody interface{}) dnsmessage.Message {
	name := dnsmessage.MustNewName(domain)
	record := resouceBody.(dnsmessage.ResourceBody)

	question := dnsmessage.Question{
		Name:  name,
		Type:  dnsmessage.TypeA,
		Class: dnsmessage.ClassINET,
	}
	answer := dnsmessage.Resource{
		Header: dnsmessage.ResourceHeader{
			Name:  name,
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
			TTL:   1,
		},
		Body: record,
	}

	dnsRecord := dnsmessage.Message{
		Header:    dnsmessage.Header{Response: true, ID: id},
		Questions: []dnsmessage.Question{question},
		Answers:   []dnsmessage.Resource{answer},
	}

	return dnsRecord
}

func addToCache(record []byte) {
	var m dnsmessage.Message
	err := m.Unpack(record)
	if err != nil {
		panic(err)
	}

	domain := m.Questions[0].Name.String()
	if len(m.Answers) <= 0 {
		return
	}

	value := m.Answers[0].Body
	ttl := m.Answers[0].Header.TTL

	inMemoryCache.Set(domain, value, time.Duration(ttl)*time.Second)
}

var blacklistCache *cache.Cache
var inMemoryCache *cache.Cache

func dontPanic() {
	if r := recover(); r != nil {
		fmt.Println("[recovered] ", r)
	}
}

func main() {
	blacklistCache = cache.New(3600*time.Minute, 10*time.Minute)
	inMemoryCache = cache.New(3600*time.Minute, 10*time.Minute)

	conn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	defer conn.Close()

	LoadBlacklists(blacklistCache)

	for {
		defer dontPanic()

		dns, encodedQuery, addr, err := waitForDNS(conn)
		if err != nil {
			panic(err)
		}

		if fakeDNS, blacklisted := isBlacklistDomain(*dns); blacklisted {
			conn.WriteToUDP(fakeDNS, addr)
			continue
		}

		if cachedDNS, cached := isCachedDomain(dns.Header.ID, dns.Questions[0].Name.String()); cached {
			conn.WriteToUDP(cachedDNS, addr)
			continue
		}

		result, err := fetchDNSoverTLS(encodedQuery)
		addToCache(result)
		conn.WriteToUDP(result, addr)
	}
}
