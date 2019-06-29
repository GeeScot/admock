package upstream

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gurparit/fastdns/dns"
	"github.com/gurparit/go-common/httputil"
	"golang.org/x/net/dns/dnsmessage"
)

// QueryBaseURL base URL for HTTPS DNS query
const QueryBaseURL = "https://%s/dns-query?dns=%s"

// Pool pool of dns servers
var Pool *dns.Pool

// AskQuestion send DNS request to Cloudflare via HTTPS
func AskQuestion(m *dnsmessage.Message) ([]byte, error) {
	packed, err := m.Pack()
	if err != nil {
		panic(err)
	}

	query := base64.RawURLEncoding.EncodeToString(packed)
	server := Pool.Next()
	url := fmt.Sprintf(QueryBaseURL, server, query)

	headers := httputil.Headers{"accept": "application/dns-message"}
	req := httputil.HTTP{
		TargetURL: httputil.FormatURL(url, query),
		Method:    http.MethodGet,
		Headers:   headers,
	}

	return req.Raw()
}
