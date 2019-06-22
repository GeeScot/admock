package main

import (
	"fmt"
	"github.com/gurparit/go-common/fileio"
	"github.com/gurparit/go-common/httputil"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/dns/dnsmessage"
	"net/http"
	"regexp"
)

// Source entry for an adlist blacklist/whitelist
type Source struct {
	URL string `json:"url"`
}

// Adlist adlist source file
type Adlist struct {
	Blacklists []Source `json:"blacklists"`
}

// NewFakeDNS get a fake dns record
func NewFakeDNS(id uint16, domain string) dnsmessage.Message {
	name := dnsmessage.MustNewName(domain)
	fakeARecord := dnsmessage.AResource{
		A: [4]byte{0, 0, 0, 0},
	}

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
			TTL:   7200,
		},
		Body: &fakeARecord,
	}

	dnsRecord := dnsmessage.Message{
		Header:    dnsmessage.Header{Response: true, ID: id},
		Questions: []dnsmessage.Question{question},
		Answers:   []dnsmessage.Resource{answer},
	}

	return dnsRecord
}

func fetchBlacklist(c *cache.Cache, source Source) {
	req := httputil.HTTP{
		TargetURL: source.URL,
		Method:    http.MethodGet,
	}

	data, err := req.String()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	domainsRegex, _ := regexp.Compile("(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]")
	domains := domainsRegex.FindAllString(data, -1)
	for _, domain := range domains {
		canonicalDomain := domain + "."
		c.Set(canonicalDomain, canonicalDomain, cache.NoExpiration)
	}
}

// LoadBlacklists cache all blacklists
func LoadBlacklists(c *cache.Cache) {
	var adlist Adlist
	fileio.ReadJSON("adlist.json", &adlist)

	for _, source := range adlist.Blacklists {
		go fetchBlacklist(c, source)
	}
}
