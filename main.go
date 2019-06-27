package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gurparit/fastdns/acl"
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

var blacklist *acl.StringCache
var inMemoryCache *cache.Cache

func dontPanic() {
	if r := recover(); r != nil {
		fmt.Println("[recovered] ", r)
	}
}

func handleQuery(conn *net.UDPConn, addr *net.UDPAddr, dns *dnsmessage.Message) {
	fmt.Printf("[Ask] %s\n", dns.Questions[0].Name)

	if fakeDNS, blacklisted := isBlacklistDomain(dns); blacklisted {
		fmt.Printf("[Blocked] %s\n", dns.Questions[0].Name)
		conn.WriteToUDP(fakeDNS, addr)
		return
	}

	if cachedDNS, cached := isCachedDomain(dns.Header.ID, dns.Questions[0]); cached {
		fmt.Printf("[Cached] %s\n", dns.Questions[0].Name)
		conn.WriteToUDP(cachedDNS, addr)
		return
	}

	result, err := cloudflare.AskQuestion(dns)
	if err != nil {
		panic(err)
	}

	fmt.Printf("[Fetched] %s\n", dns.Questions[0].Name)
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

var wg sync.WaitGroup

func main() {
	defaultExpiration := 3600 * time.Second
	defaultEviction := 7200 * time.Second

	inMemoryCache = cache.New(defaultExpiration, defaultEviction)

	wg.Add(1)

	go run()

	start := time.Now().Unix()
	go func() {
		blacklist = acl.LoadBlacklists()
		blacklist.Sort()
		end := time.Now().Unix()

		elapsed := end - start

		fmt.Printf("\nBlacklisted %d domains in %d seconds.\n", blacklist.Size, elapsed)
	}()

	wg.Wait()
}
