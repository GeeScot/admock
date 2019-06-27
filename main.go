package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gurparit/fastdns/acl"
	"github.com/gurparit/fastdns/cache"
	"github.com/gurparit/fastdns/cloudflare"
	"github.com/gurparit/fastdns/dns"
	gocache "github.com/patrickmn/go-cache"
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

func isBlacklistDomain(message *dnsmessage.Message) ([]byte, bool) {
	domain := message.Questions[0].Name.String()
	found := blacklist.Contains(domain)
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
	item, found := dnsCache.Get(domain)
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

	dnsCache.Set(domain, m.Answers, time.Duration(ttl)*time.Second)
}

func dontPanic() {
	if r := recover(); r != nil {
		fmt.Println("[recovered] ", r)
	}
}

func handleQuery(conn *net.UDPConn, addr *net.UDPAddr, dns *dnsmessage.Message) {
	if fakeDNS, blacklisted := isBlacklistDomain(dns); blacklisted {
		conn.WriteToUDP(fakeDNS, addr)
		return
	}

	if cachedDNS, cached := isCachedDomain(dns.Header.ID, dns.Questions[0]); cached {
		conn.WriteToUDP(cachedDNS, addr)
		return
	}

	result, err := cloudflare.AskQuestion(dns)
	if err != nil {
		panic(err)
	}

	addToCache(result)
	conn.WriteToUDP(result, addr)
}

func run() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	defer wg.Done()

	for {
		defer dontPanic()

		dns, addr, err := waitForDNS(conn)
		if err != nil {
			panic(err)
		}

		go handleQuery(conn, addr, dns)
	}
}

var blacklist *cache.StringCache
var dnsCache *gocache.Cache

var wg sync.WaitGroup

func main() {
	defaultExpiration := 3600 * time.Second
	defaultEviction := 7200 * time.Second

	dnsCache = gocache.New(defaultExpiration, defaultEviction)
	blacklist = cache.New()

	wg.Add(1)

	go run()

	go func() {
		acl.LoadBlacklists(blacklist)
		blacklist.Sort()
	}()

	wg.Wait()
}
