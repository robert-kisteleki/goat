/*
  (C) Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// DnsResult holds a DNS result structure
type DnsResult struct {
	BaseResult               //
	Error      []DnsError    // ?
	Responses  []DnsResponse //
}

// DnsResponse holds one response from one server/resolver, with all associated data
// Various bits like counts and answers are stored here in a simple format which
// is likely a good fit for many use cases; one could look at all the gory details in
// the AnswerBuf (abuf) if more details are needed
type DnsResponse struct {
	TimeStamp     time.Time      //
	SourceAddr    netip.Addr     //
	Destination   netip.AddrPort //
	Error         []DnsError     // ?
	AddressFamily uint           //
	Protocol      string         //
	RetryCount    uint           //
	QueryBuf      []byte         //
	ResponseTime  float64        //
	ResponseSize  uint           //

	// overview
	QueryID         uint   //
	QueriesCount    uint   //
	AnswerCount     uint   //
	NameServerCount uint   //
	AdditionalCount uint   //
	Edsn0Nsid       []byte //

	// various bits
	Response           bool //
	Opcode             int  //
	Authoritative      bool //
	Truncated          bool //
	RecursionDesired   bool //
	RecursionAvailable bool //
	Zero               bool //
	AuthenticatedData  bool //
	CheckingDisabled   bool //
	Rcode              int  //

	// details
	AnswerBuf []byte      //
	Question  DnsQuestion //
	Answer    []DnsAnswer //
	Ns        []DnsAnswer //
	Extra     []DnsAnswer //

	Ttl6 uint //
}

// DnsQuestion is the question that was asked - as parsed from abuf
type DnsQuestion struct {
	Class int
	Type  int
	Name  string
}

// DnsAnswer is a (simplified) answer
// "simplified" means it only contains the full answer encoded in a string
type DnsAnswer struct {
	Class int
	Type  int
	Name  string
	Ttl   int
	Data  string
}

// DnsError is an error that may have been reported
type DnsError struct {
	Timeout  uint
	AddrInfo string
}

const (
	// this is not a full list!
	DnsTypeNONE   = 0 // if not filled in
	DnsTypeA      = 1
	DnsTypeNS     = 2
	DnsTypeCNAME  = 5
	DnsTypeSOA    = 6
	DnsTypePTR    = 12
	DnsTypeMX     = 15
	DnsTypeTXT    = 16
	DnsTypeSIG    = 24
	DnsTypeKEY    = 25
	DnsTypeAAAA   = 28
	DnsTypeLOC    = 29
	DnsTypeNAPTR  = 35
	DnsTypeOPT    = 41
	DnsTypeDS     = 43
	DnsTypeRRSIG  = 46
	DnsTypeNSEC   = 47
	DnsTypeDNSKEY = 48
	DnsTypeNSEC3  = 50
	DnsTypeTLSA   = 52
	DnsTypeHTTPS  = 65
	DnsTypeSPF    = 99

	// this is not a full list!
	DnsClassNONE  = 0 // if not filled in
	DnsClassINET  = 1
	DnsClassCHAOS = 3
	DnsClassANY   = 255

	// this is not a full list!
	DnsRcodeNOERR     = 0
	DnsRcodeFORMERR   = 1
	DnsRcodeSERVFAIL  = 2
	DnsRcodeNXDOMAIN  = 3
	DnsRcodeNOTIMP    = 4
	DnsRcodeREFUSED   = 5
	DnsRcodeNOTAUTH   = 9
	DnsRcodeBADVERS   = 16
	DnsRcodeBADCOOKIE = 23
)

// DnsTypeNames translates record types to their names
var DnsTypeNames = map[int]string{
	DnsTypeNONE:   "N/A",
	DnsTypeA:      "A",
	DnsTypeNS:     "NS",
	DnsTypeCNAME:  "CNAME",
	DnsTypeSOA:    "SOA",
	DnsTypePTR:    "PTR",
	DnsTypeMX:     "MX",
	DnsTypeTXT:    "TXT",
	DnsTypeSIG:    "SIG",
	DnsTypeKEY:    "KEY",
	DnsTypeAAAA:   "AAAA",
	DnsTypeLOC:    "LOC",
	DnsTypeNAPTR:  "NAPTR",
	DnsTypeOPT:    "OPT",
	DnsTypeDS:     "DS",
	DnsTypeRRSIG:  "RRSIG",
	DnsTypeNSEC:   "NSEC",
	DnsTypeDNSKEY: "DNSKEY",
	DnsTypeNSEC3:  "NSEC3",
	DnsTypeTLSA:   "TLSA",
	DnsTypeHTTPS:  "HTTPS",
	DnsTypeSPF:    "SPF",
}

// DnsClassNames translates record classes to their names
var DnsClassNames = map[int]string{
	DnsClassNONE:  "N/A",
	DnsClassINET:  "IN",
	DnsClassCHAOS: "CH",
	DnsClassANY:   "ANY",
}

var DnsRcodeNames = map[int]string{
	DnsRcodeNOERR:     "NOERR",
	DnsRcodeFORMERR:   "FORMERR",
	DnsRcodeSERVFAIL:  "SERVFAIL",
	DnsRcodeNXDOMAIN:  "NXDOMAIN",
	DnsRcodeNOTIMP:    "NOTIMP",
	DnsRcodeREFUSED:   "REFUSED",
	DnsRcodeNOTAUTH:   "NOAUTH",
	DnsRcodeBADVERS:   "BADVERS",
	DnsRcodeBADCOOKIE: "BADCOOKIE",
}

// TypeName returns the codename for this result type
func (result *DnsResult) TypeName() string {
	return "dns"
}

// Parse takes a DNS result JSON blob and turns it into a DnsResult object
func (dns *DnsResult) Parse(from string) (err error) {
	var idns dnsResult
	err = json.Unmarshal([]byte(from), &idns)
	if err != nil {
		return err
	}
	if idns.Type != "dns" {
		return fmt.Errorf("this is not a DNS result (type=%s)", idns.Type)
	}
	dns.BaseResult = idns.BaseResult

	dns.Error = make([]DnsError, 0)
	if idns.Error != nil {
		dns.Error = append(dns.Error, DnsError{idns.Error.Timeout, idns.Error.AddrInfo})
	}

	// concatenate RawResult and RawResults entries into a single list
	// some details are stored in different places of the JSON struct, deal with that here
	dns.Responses = make([]DnsResponse, 0)
	if idns.RawResult != nil {
		qbuf, err := decodeBuf(idns.RawQBuf)
		if err != nil {
			return fmt.Errorf("error decoding qbuf: %s", err.Error())
		}
		de, err := makeDnsResponse(
			time.Time(idns.TimeStamp),
			idns.SourceAddr,
			netip.AddrPortFrom(*idns.DestinationAddr, 53),
			idns.AddressFamily,
			idns.Protocol,
			idns.Error,
			idns.RetryCount,
			qbuf,
			*idns.RawResult,
		)
		if err != nil {
			return err
		}
		dns.Responses = append(dns.Responses, de)
	}
	for _, rs := range idns.RawResultSet {
		dstport, _ := strconv.ParseUint(rs.DestinationPort, 10, 0) // if error: 0
		var retrycount uint = 0
		if rs.RetryCount != nil {
			retrycount = *rs.RetryCount
		}
		qbuf, err := decodeBuf(rs.RawQBuf)
		if err != nil {
			return fmt.Errorf("error decoding qbuf: %s", err.Error())
		}
		de, err := makeDnsResponse(
			time.Time(rs.Time),
			rs.SourceAddr,
			netip.AddrPortFrom(rs.DestinationAddr, uint16(dstport)),
			rs.AddressFamily,
			rs.Protocol,
			rs.Error,
			retrycount,
			qbuf,
			rs.Answer,
		)
		if err != nil {
			return err
		}
		dns.Responses = append(dns.Responses, de)
	}
	return nil
}

// Filter filters out the desired class/type answers from all answers
func (result *DnsResult) Filter(class int, typ int) []DnsAnswer {
	answers := make([]DnsAnswer, 0)
	for _, resp := range result.Responses {
		answers = append(answers, resp.Filter(class, typ)...)
	}
	return answers
}

// Filter filters out the desired class/type answers from all answers
// in a specific response
func (resp *DnsResponse) Filter(class int, typ int) []DnsAnswer {
	answers := make([]DnsAnswer, 0)
	for _, answer := range resp.AllAnswers() {
		if answer.Class == class && answer.Type == typ {
			answers = append(answers, answer)
		}
	}
	return answers
}

// AllAnswers aggregates all answers from all responses into an array
func (result *DnsResult) AllAnswers() []DnsAnswer {
	answers := make([]DnsAnswer, 0)
	for _, resp := range result.Responses {
		answers = append(answers, resp.AllAnswers()...)
	}
	return answers
}

// AllAnswers aggregates all answers from a responses into an array
func (resp *DnsResponse) AllAnswers() []DnsAnswer {
	answers := make([]DnsAnswer, 0)
	answers = append(answers, resp.Answer...)
	answers = append(answers, resp.Ns...)
	answers = append(answers, resp.Extra...)
	return answers
}

//////////////////////////////////////////////////////
// API version of a DNS result

// this is the JSON structure as reported by the API
type dnsResult struct {
	BaseResult
	Error        *dnsError     `json:"error"`     //
	Protocol     string        `json:"proto"`     //
	RetryCount   uint          `json:"retry"`     //
	RawQBuf      *string       `json:"qbuf"`      //
	RawResult    *dnsAnswer    `json:"result"`    //
	RawResultSet []dnsResponse `json:"resultset"` //
}

type dnsResponse struct {
	Time            uniTime    `json:"time"`     //
	LastTimeSync    int        `json:"lts"`      //
	SourceAddr      netip.Addr `json:"src_addr"` //
	DestinationAddr netip.Addr `json:"dst_addr"` //
	DestinationPort string     `json:"dst_port"` //
	Error           *dnsError  `json:"error"`    //
	AddressFamily   uint       `json:"af"`       //
	Protocol        string     `json:"proto"`    //
	RetryCount      *uint      `json:"retry"`    //
	SubID           uint       `json:"subid"`    //
	SubMax          uint       `json:"submax"`   //
	RawQBuf         *string    `json:"qbuf"`     //
	Answer          dnsAnswer  `json:"result"`   //
}

type dnsAnswer struct {
	ResponseTime    float64      `json:"rt"`      //
	ResponseSize    uint         `json:"size"`    //
	Abuf            string       `json:"abuf"`    //
	QueryID         uint         `json:"id"`      //
	AnswerCount     uint         `json:"ancount"` //
	QueriesCount    uint         `json:"qdcount"` //
	NameServerCount uint         `json:"nscount"` //
	AdditionalCount uint         `json:"arcount"` //
	ResourceRecords *[]dnsRecord `json:"answers"` //
	Ttl6            *uint        `json:"ttl"`     //
}

type dnsRecord struct {
	DomainName   string   `json:"mname"`
	Name         string   `json:"name"`
	ResourceData []string `json:"rdata"`
	ResourceName string   `json:"rname"`
	Serial       uint     `json:"serial"`
	Ttl          uint     `json:"ttl"`
	Type         string   `json:"type"`
}

type dnsError struct {
	Timeout  uint   `json:"timeout"`
	AddrInfo string `json:"getaddrinfo"`
}

// decode an qbuf or an abuf (from base64 string to []byte) if possible
// Return a 0-len []byte if input was empty or non-existent
func decodeBuf(buf *string) ([]byte, error) {
	if buf == nil {
		return make([]byte, 0), nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*buf)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

// makeDnsResponse assembles a DnsResponse object
func makeDnsResponse(
	timestamp time.Time,
	srcaddr netip.Addr,
	dst netip.AddrPort,
	af uint,
	proto string,
	dnserror *dnsError,
	rc uint,
	qbuf []byte,
	ans dnsAnswer,
) (de DnsResponse, err error) {
	// basics
	de.TimeStamp = timestamp
	de.SourceAddr = srcaddr
	de.Destination = dst
	de.AddressFamily = af
	de.Protocol = proto
	de.QueryBuf = qbuf
	de.RetryCount = rc
	de.ResponseTime = ans.ResponseTime
	de.ResponseSize = ans.ResponseSize
	abuf, err := decodeBuf(&ans.Abuf)
	if err != nil {
		return de, fmt.Errorf("error decoding abuf: %s", err.Error())
	}
	de.AnswerBuf = abuf
	if ans.Ttl6 != nil {
		de.Ttl6 = *ans.Ttl6
	}

	// in case there was an error reported
	if dnserror != nil {
		de.Error = append(de.Error, DnsError{dnserror.Timeout, dnserror.AddrInfo})
		return
	}

	// these could be parsed out of the answer if we didn't trust
	// the values in the structure otherwise
	de.QueryID = ans.QueryID
	de.QueriesCount = ans.QueriesCount
	de.NameServerCount = ans.NameServerCount
	de.AnswerCount = ans.AnswerCount
	de.AdditionalCount = ans.AdditionalCount

	if len(de.AnswerBuf) == 0 {
		// don't try the rest if abuf was not present to begin with
		return de, nil
	}

	// use an (excellent) library to parse details out of abuf
	var parsed dns.Msg
	err = parsed.Unpack(de.AnswerBuf)
	if err != nil {
		return de, fmt.Errorf("error parsing abuf: %v", err.Error())
	}

	// concatenate the (simplified) answers from all categories
	makeAnswers := func(rrs []dns.RR) []DnsAnswer {
		list := make([]DnsAnswer, 0)
		for _, ans := range rrs {
			ah := ans.Header()
			rdata := ""
			switch rtype := ans.(type) {
			case *dns.A:
				rdata = rtype.A.String()
			case *dns.AAAA:
				rdata = rtype.AAAA.String()
			case *dns.CNAME:
				rdata = rtype.Target
			case *dns.DS:
				rdata = rtype.String()
			case *dns.NS:
				rdata = rtype.Ns
			case *dns.PTR:
				rdata = rtype.Ptr
			case *dns.RRSIG:
				rdata = rtype.String()
			case *dns.SOA:
				rdata = rtype.String()
			case *dns.SRV:
				rdata = rtype.String()
			case *dns.TXT:
				rdata = strings.Join(rtype.Txt, ", ")
			case *dns.OPT:
				for _, opt := range rtype.Option {
					switch o := opt.(type) {
					case *dns.EDNS0_NSID:
						de.Edsn0Nsid = decodeNsid(o.Nsid)
					}
				}
			}
			list = append(list,
				DnsAnswer{
					int(ah.Class),
					int(ah.Rrtype),
					ah.Name,
					int(ah.Ttl),
					rdata,
				},
			)
		}
		return list
	}
	de.Answer = makeAnswers(parsed.Answer)
	de.Ns = makeAnswers(parsed.Ns)
	de.Extra = makeAnswers(parsed.Extra)

	de.Response = parsed.Response
	de.Opcode = parsed.Opcode
	de.Authoritative = parsed.Authoritative
	de.Truncated = parsed.Truncated
	de.RecursionDesired = parsed.RecursionDesired
	de.RecursionAvailable = parsed.RecursionAvailable
	de.Zero = parsed.Zero
	de.AuthenticatedData = parsed.AuthenticatedData
	de.CheckingDisabled = parsed.CheckingDisabled
	de.Rcode = parsed.Rcode

	// include the question, because why not
	if len(parsed.Question) > 0 {
		q := parsed.Question[0]
		de.Question = DnsQuestion{int(q.Qclass), int(q.Qtype), q.Name}
	}

	return
}

func decodeNsid(hexin string) []byte {
	src := []byte(hexin)
	nsid := make([]byte, hex.DecodedLen(len(src)))
	n, err := hex.Decode(nsid, src)
	if err != nil {
		return nil
	}
	return nsid[:n]
}
