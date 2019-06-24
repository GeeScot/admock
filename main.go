package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/gurparit/fastdns/cloudflare"
	"github.com/gurparit/fastdns/dns"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/dns/dnsmessage"
)

func waitForDNS(conn *net.UDPConn) (*dnsmessage.Message, *net.UDPAddr, error) {
	buf := make([]byte, 512)
	_, addr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return nil, nil, errors.New("[err] invalid udp packet")
	}

	var m dnsmessage.Message
	err = m.Unpack(buf)
	if err != nil {
		return nil, nil, err
	}

	return &m, addr, err
}

func isBlacklistDomain(message dnsmessage.Message) ([]byte, bool) {
	domain := message.Questions[0].Name.String()
	_, found := blacklistCache.Get(domain)
	if !found {
		return nil, false
	}

	fakeDNS := dns.NewMockAnswer(message.Header.ID, message.Questions[0])
	packed, err := fakeDNS.Pack()
	if err != nil {
		panic(err)
	}

	return packed, true
}

func isCachedDomain(id uint16, question dnsmessage.Question) ([]byte, bool) {
	domain := question.Name.String()
	item, found := inMemoryCache.Get(domain)
	if !found {
		return nil, false
	}

	records := item.([]dnsmessage.Resource)
	m := dns.NewAnswer(id, question, records)

	data, _ := m.Pack()
	return data, true
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

	header := m.Answers[0].Header
	ttl := header.TTL

	inMemoryCache.Set(domain, m.Answers, time.Duration(ttl)*time.Second)
}

var blacklistCache *cache.Cache
var inMemoryCache *cache.Cache

func dontPanic() {
	if r := recover(); r != nil {
		fmt.Println("[recovered] ", r)
	}
}

func main() {
	defaultExpiration := 3600 * time.Second
	defaultEviction := 7200 * time.Second

	blacklistCache = cache.New(defaultExpiration, defaultEviction)
	inMemoryCache = cache.New(defaultExpiration, defaultEviction)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	LoadBlacklists(blacklistCache)

	for {
		defer dontPanic()

		dns, addr, err := waitForDNS(conn)
		if err != nil {
			panic(err)
		}

		if fakeDNS, blacklisted := isBlacklistDomain(*dns); blacklisted {
			conn.WriteToUDP(fakeDNS, addr)
			continue
		}

		if cachedDNS, cached := isCachedDomain(dns.Header.ID, dns.Questions[0]); cached {
			conn.WriteToUDP(cachedDNS, addr)
			continue
		}

		result, err := cloudflare.AskQuestion(dns)
		if err != nil {
			panic(err)
		}
		
		addToCache(result)
		conn.WriteToUDP(result, addr)
	}
}
