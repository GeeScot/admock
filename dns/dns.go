package dns

import "golang.org/x/net/dns/dnsmessage"

// Record entry
type Record struct {
	Header dnsmessage.ResourceHeader
	Body   dnsmessage.ResourceBody
}

// NewAnswer creates a new DNS answer
func NewAnswer(id uint16, question dnsmessage.Question, records []dnsmessage.Resource) dnsmessage.Message {
	answers := []dnsmessage.Resource{}
	for _, record := range records {
		answers = append(answers, dnsmessage.Resource{
			Header: record.Header,
			Body:   record.Body,
		})
	}

	dnsRecord := dnsmessage.Message{
		Header:    dnsmessage.Header{Response: true, ID: id},
		Questions: []dnsmessage.Question{question},
		Answers:   answers,
	}

	return dnsRecord
}

// NewMockAnswer creates a new Mock DNS answer for blocked domains
func NewMockAnswer(id uint16, question dnsmessage.Question) dnsmessage.Message {
	header := dnsmessage.ResourceHeader{
		Name:  question.Name,
		Type:  question.Type,
		Class: question.Class,
		TTL:   7200,
	}
	body := &dnsmessage.AResource{
		A: [4]byte{0, 0, 0, 0},
	}

	return NewAnswer(id, question, []dnsmessage.Resource{
		dnsmessage.Resource{Header: header, Body: body},
	})
}
