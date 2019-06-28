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
	"golang.org/x/net/dns/dnsmessage"
)

func listen(conn *net.UDPConn) (*dnsmessage.Message, *net.UDPAddr, error) {
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
	domain := dns.Domain(message)

	found := blacklist.Contains(domain)
	if !found {
		return nil, false
	}

	fakeDNS := dns.NewMockAnswer(message.Header.ID, message.Questions[0])
	packed, err := fakeDNS.Pack()
	catch(err)

	return packed, true
}

func isCachedDomain(message *dnsmessage.Message) ([]byte, bool) {
	domain := dns.Domain(message)

	records, found := dnsCache.Get(domain)
	if !found {
		return nil, false
	}

	question := message.Questions[0]

	id := dns.ID(message)
	m := dns.NewAnswer(id, question, records)

	data, _ := m.Pack()
	return data, true
}

func addCache(record []byte) {
	var m dnsmessage.Message
	err := m.Unpack(record)
	catch(err)

	domain := dns.Domain(&m)
	if len(m.Answers) <= 0 {
		return
	}

	ttl := dns.TTL(&m)
	dnsCache.AddWithExpiry(domain, m.Answers, time.Duration(ttl))
}

func handleQuery(conn *net.UDPConn, addr *net.UDPAddr, message *dnsmessage.Message) {
	if dns, found := isBlacklistDomain(message); found {
		conn.WriteToUDP(dns, addr)
		return
	}

	if dns, found := isCachedDomain(message); found {
		conn.WriteToUDP(dns, addr)
		return
	}

	dns, err := cloudflare.AskQuestion(message)
	catch(err)

	addCache(dns)
	conn.WriteToUDP(dns, addr)
}

func run() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 53})
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	defer wg.Done()

	for {
		defer try()

		dns, addr, err := listen(conn)
		catch(err)

		go handleQuery(conn, addr, dns)
	}
}

var blacklist *cache.StringCache
var dnsCache *cache.ResourceCache

var wg sync.WaitGroup

func try() {
	if r := recover(); r != nil {
		fmt.Println("[recovered] ", r)
	}
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	blacklist = cache.Strings()
	dnsCache = cache.Resources()

	wg.Add(1)

	go run()
	go acl.Load(blacklist)

	wg.Wait()
}
