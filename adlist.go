package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gurparit/go-common/fileio"
	"github.com/gurparit/go-common/httputil"
	"github.com/patrickmn/go-cache"
)

func isWhitelisted(whitelist []string, search string) bool {
	for _, domain := range whitelist {
		if domain == search {
			return true
		}
	}

	return false
}

// Adlist adlist source file
type Adlist struct {
	External struct {
		Blacklists []string `json:"blacklists"`
	} `json:"external"`

	Blacklist []string `json:"blacklist"`
	Whitelist []string `json:"whitelist"`
}

func fetchBlacklist(c *cache.Cache, source string, whitelist []string) {
	req := httputil.HTTP{
		TargetURL: source,
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
		if isWhitelisted(whitelist, domain) {
			continue
		}

		if _, alreadyCached := c.Get(domain); alreadyCached {
			continue
		}

		canonicalDomain := domain + "."
		c.Set(canonicalDomain, canonicalDomain, cache.NoExpiration)
	}
}

// LoadBlacklists cache all blacklists
func LoadBlacklists(c *cache.Cache) {
	var adlist Adlist
	fileio.ReadJSON("adlist.json", &adlist)

	for _, source := range adlist.External.Blacklists {
		go fetchBlacklist(c, source, adlist.Whitelist)
	}

	for _, domain := range adlist.Blacklist {
		c.Set(domain, domain, cache.NoExpiration)
	}
}
