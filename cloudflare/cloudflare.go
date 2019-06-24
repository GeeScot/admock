package cloudflare

import (
	"encoding/base64"
	"net/http"

	"github.com/gurparit/go-common/httputil"
	"golang.org/x/net/dns/dnsmessage"
)

// CloudflareDNS1 base url for cloudflare dns 1
const CloudflareDNS1 = "https://1.1.1.1/dns-query?dns=%s"

// CloudflareDNS2 base url for cloudflare dns 2
const CloudflareDNS2 = "https://1.0.0.1/dns-query?dns=%s"

// AskQuestion send DNS request to Cloudflare via HTTPS
func AskQuestion(m *dnsmessage.Message) ([]byte, error) {
	packed, err := m.Pack()
	if err != nil {
		panic(err)
	}

	query := base64.RawURLEncoding.EncodeToString(packed)

	headers := httputil.Headers{"accept": "application/dns-message"}
	req := httputil.HTTP{
		TargetURL: httputil.FormatURL(CloudflareDNS1, query),
		Method:    http.MethodGet,
		Headers:   headers,
	}

	return req.Raw()
}
