package dns

import (
	"encoding/base64"

	"github.com/gurparit/go-common/math"
	"golang.org/x/net/dns/dnsmessage"
)

// Record entry
type Record struct {
	Header dnsmessage.ResourceHeader
	Body   dnsmessage.ResourceBody
}

// NewAnswer creates a new DNS answer
func NewAnswer(id uint16, question dnsmessage.Question, answers []dnsmessage.Resource) dnsmessage.Message {
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

// Domain get domain from a dns question
func Domain(message *dnsmessage.Message) string {
	return message.Questions[0].Name.String()
}

// ID get dns message id from a dns message
func ID(message *dnsmessage.Message) uint16 {
	return message.Header.ID
}

// TTL return time to live
func TTL(message *dnsmessage.Message) uint16 {
	maxTTL := 0
	for _, answer := range message.Answers {
		maxTTL = math.Max(maxTTL, int(answer.Header.TTL))
	}

	return uint16(maxTTL)
}

// EncodedQuestion base64 encoded question
func EncodedQuestion(message *dnsmessage.Message) string {
	question := message.Questions[0].GoString()
	data := []byte(question)
	return base64.URLEncoding.EncodeToString(data)
}
