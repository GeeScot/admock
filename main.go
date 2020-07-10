package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/codescot/admock/acl"
	"github.com/codescot/admock/cache"
	"github.com/codescot/admock/dns"
	"github.com/codescot/admock/logger"
	"github.com/codescot/admock/pool"
	"github.com/codescot/admock/upstream"
	"github.com/codescot/go-common/env"
	"golang.org/x/net/dns/dnsmessage"
)

func isDomainBlacklisted(message *dnsmessage.Message) ([]byte, bool) {
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

func isDomainCached(message *dnsmessage.Message) ([]byte, bool) {
	encodedQuestion := dns.EncodedQuestion(message)

	records, found := dnsCache.Get(encodedQuestion)
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

	if len(m.Answers) <= 0 {
		return
	}

	encodedQuestion := dns.EncodedQuestion(&m)
	ttl := dns.TTL(&m)

	dnsCache.AddWithExpiry(encodedQuestion, m.Answers, time.Duration(ttl))
}

func getRecord(message *dnsmessage.Message) []byte {
	if dns, found := isDomainBlacklisted(message); found {
		return dns
	}

	if dns, found := isDomainCached(message); found {
		return dns
	}

	dns, err := u.AskQuestion(message)
	catch(err)

	addCache(dns)
	return dns
}

func setupDebug() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "-debug=true")
	flag.Parse()

	logger.PrintWithTimeStamp(fmt.Sprintf("Debug: %v", debug))

	if !debug {
		return
	}

	l = logger.Logger{
		Debug: debug,
		Log:   make(chan []byte),
	}

	go l.Start()
}

func setupUpstream() {
	blacklist = cache.Strings()
	dnsCache = cache.Resources()

	p := pool.NewRoundRobin()

	strategy := env.Optional("ADMOCK_STRATEGY", "https")
	switch strategy {
	case "https":
		u = &upstream.HTTPSUpstream{Pool: p}
		break
	case "udp":
		u = &upstream.UDPUpstream{Pool: p}
		break
	}

	logger.PrintWithTimeStamp(fmt.Sprintf("Strategy: %s", strategy))
}

var blacklist *cache.StringCache
var dnsCache *cache.ResourceCache

var wg sync.WaitGroup
var u upstream.Upstream

var l logger.Logger

func main() {
	setupDebug()
	setupUpstream()

	wg.Add(1)

	go run()
	go acl.Load(blacklist)

	wg.Wait()
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

		go func() {
			defer try()

			response := getRecord(dns)
			conn.WriteToUDP(response, addr)

			if l.Debug {
				l.Log <- response
			}
		}()
	}
}

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

func try() {
	if r := recover(); r != nil {
		timestamp := time.Now().Format(time.RFC3339)
		fmt.Printf("[%s] %v\n", timestamp, r)
	}
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
