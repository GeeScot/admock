package acl

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/codescot/admock/cache"
	"github.com/codescot/admock/logger"
	"github.com/codescot/go-common/fileio"
	"github.com/codescot/go-common/httputil"
)

// AccessControlLists access control source file
type AccessControlLists struct {
	Sources   []Source `json:"sources"`
	Blacklist []string `json:"blacklist"`
	Whitelist []string `json:"whitelist"`
}

// Source source for blacklist with md5 hash
type Source struct {
	URL string `json:"url"`
	MD5 string `json:"md5"`
}

func tail(source string) string {
	idx := strings.LastIndex(source, "/")
	return source[idx+1:]
}

func saveString(source string, data string) {
	filename := tail(source)
	os.Remove(filename)

	filePath := fmt.Sprintf("%s%s", os.TempDir(), filename)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	file.WriteString(data)
	file.Close()
}

func readStrings(source string) ([]string, string) {
	filename := tail(source)

	filePath := fmt.Sprintf("%s%s", os.TempDir(), filename)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return []string{}, ""
	}

	dataString := string(data)
	dataHash := fmt.Sprintf("%x", md5.Sum(data))

	domains := strings.Split(dataString, "\n")
	return domains, dataHash
}

func hasMD5Changed(currentMD5 string, source string) bool {
	req := httputil.HTTP{
		TargetURL: source,
		Method:    http.MethodGet,
	}

	data, err := req.String()
	if err != nil {
		fmt.Println(err.Error())
		return true
	}

	hasChanged := currentMD5 != data
	if hasChanged {
		saveString(source, data)
	}

	return hasChanged
}

func getRemoteBlacklist(source string) []string {
	logger.PrintWithTimeStamp(fmt.Sprintf("Remote: %s\n", source))

	req := httputil.HTTP{
		TargetURL: source,
		Method:    http.MethodGet,
	}

	data, err := req.String()
	if err != nil {
		logger.PrintWithTimeStamp(err.Error())
		return nil
	}

	saveString(source, data)

	domains := strings.Split(data, "\n")
	return domains
}

// Load cache all blacklists
func Load(cache *cache.StringCache) {
	var lists AccessControlLists

	config, ok := os.LookupEnv("ADMOCK_CONFIG")
	if ok {
		fileio.ReadJSON(config, &lists)
	} else {
		lists = AccessControlLists{
			Sources: []Source{
				Source{
					URL: "https://raw.githubusercontent.com/codescot/blackhole/master/default.txt",
					MD5: "https://raw.githubusercontent.com/codescot/blackhole/master/default.md5",
				},
			},
		}
	}

	start := time.Now().UnixNano()

	var wg sync.WaitGroup
	for _, source := range lists.Sources {
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()

			domains, md5Hash := readStrings(s.URL)
			if hasMD5Changed(md5Hash, s.MD5) {
				domains = getRemoteBlacklist(s.URL)
			}

			cache.Append(domains)
		}(source)
	}

	wg.Wait()

	if len(lists.Blacklist) > 1 {
		for _, domain := range lists.Blacklist {
			cache.Add(domain)
		}
	}

	if len(lists.Sources) > 1 || len(lists.Blacklist) > 1 {
		cache.Sort()
	}

	end := time.Now().UnixNano()
	elapsed := end - start

	logger.PrintWithTimeStamp(fmt.Sprintf("Blacklisted %d domains (in %dms).", cache.Size, time.Duration(elapsed)/time.Millisecond))
}
