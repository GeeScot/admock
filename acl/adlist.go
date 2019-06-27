package acl

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/gurparit/go-common/array"
	"github.com/gurparit/go-common/fileio"
	"github.com/gurparit/go-common/httputil"
)

// Adlist adlist source file
type Adlist struct {
	External struct {
		Blacklists []string `json:"blacklists"`
	} `json:"external"`

	Blacklist []string `json:"blacklist"`
	Whitelist []string `json:"whitelist"`
}

func fetchBlacklist(wg *sync.WaitGroup, c *StringCache, source string, whitelist []string) {
	defer wg.Done()

	fmt.Printf("Get: %s\n", source)

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
		if array.Contains(whitelist, domain) {
			continue
		}

		c.Add(domain + ".")
	}

	fmt.Printf("Done: %s\n", source)
}

// LoadBlacklists cache all blacklists
func LoadBlacklists() *StringCache {
	var adlist Adlist
	fileio.ReadJSON("adlist.json", &adlist)

	cache := New()

	var wg sync.WaitGroup
	for _, source := range adlist.External.Blacklists {
		wg.Add(1)
		go fetchBlacklist(&wg, cache, source, adlist.Whitelist)
	}

	for _, domain := range adlist.Blacklist {
		cache.Add(domain)
	}

	wg.Wait()

	return cache
}
