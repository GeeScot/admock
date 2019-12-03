package upstream

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/geescot/fastdns/pool"
	"github.com/geescot/go-common/httputil"
	"golang.org/x/net/dns/dnsmessage"
)

// QueryBaseURL base URL for HTTPS DNS query
const QueryBaseURL = "https://%s/dns-query?dns=%s"

// HTTPSUpstream HTTPS Struct
type HTTPSUpstream struct {
	Pool pool.Pool
}

// AskQuestion send DNS request to Cloudflare via HTTPS
func (u *HTTPSUpstream) AskQuestion(m *dnsmessage.Message) ([]byte, error) {
	packed, err := m.Pack()
	if err != nil {
		panic(err)
	}

	query := base64.RawURLEncoding.EncodeToString(packed)
	server := u.Pool.Next()
	url := fmt.Sprintf(QueryBaseURL, server, query)

	headers := httputil.Headers{"accept": "application/dns-message"}
	req := httputil.HTTP{
		TargetURL: httputil.FormatURL(url, query),
		Method:    http.MethodGet,
		Headers:   headers,
	}

	return req.Raw()
}
