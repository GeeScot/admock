package acl

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gurparit/fastdns/cache"
	"github.com/gurparit/go-common/array"
	"github.com/gurparit/go-common/fileio"
	"github.com/gurparit/go-common/httputil"
)

// AccessControlLists access control source file
type AccessControlLists struct {
	External struct {
		Blacklists []string `json:"blacklists"`
	} `json:"external"`

	Blacklist []string `json:"blacklist"`
	Whitelist []string `json:"whitelist"`
}

func fetchBlacklist(wg *sync.WaitGroup, c *cache.StringCache, source string, whitelist []string) {
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

	domains := strings.Split(data, "\n")
	for _, domain := range domains {
		if array.Contains(whitelist, domain) {
			continue
		}

		c.Add(domain + ".")
	}

	fmt.Printf("Done: %s\n", source)
}

// Load cache all blacklists
func Load(cache *cache.StringCache) {
	config := os.Getenv("FASTDNS_CONFIG")
	if len(config) <= 0 {
		return
	}

	var lists AccessControlLists
	fileio.ReadJSON(config, &lists)

	start := time.Now().Unix()

	var wg sync.WaitGroup
	for _, source := range lists.External.Blacklists {
		wg.Add(1)
		go fetchBlacklist(&wg, cache, source, lists.Whitelist)
	}

	for _, domain := range lists.Blacklist {
		cache.Add(domain)
	}

	wg.Wait()

	cache.Sort()

	end := time.Now().Unix()
	elapsed := end - start

	fmt.Printf("\nBlacklisted %d domains in %d seconds.\n", cache.Size, elapsed)
}
