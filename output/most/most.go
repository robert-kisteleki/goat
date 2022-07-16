/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "some" and "most" output formatters.
*/

package most

import (
	"fmt"
	"goatcli/output/some"
	"strings"

	"github.com/robert-kisteleki/goatapi/result"
)

var verbose bool
var total uint

func Setup(isverbose bool) {
	verbose = isverbose
}

func Process(res result.Result) {
	total++

	switch r := res.(type) {
	case *result.PingResult:
		fmt.Println(mostOutputPing(r))
	case *result.DnsResult:
		fmt.Println(mostOutputDns(r))
	case *result.TracerouteResult:
		fmt.Println(mostOutputTraceroute(r))
	case *result.CertResult:
		fmt.Println(mostOutputCert(r))
	case *result.HttpResult:
		fmt.Println(mostOutputHttp(r))
	case *result.NtpResult:
		fmt.Println(mostOutputNtp(r))
	case *result.ConnectionResult:
		fmt.Println(mostOutputConnection(r))
	case *result.UptimeResult:
		fmt.Println(mostOutputUptime(r))
	}
}

func Finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}

func mostOutputPing(res *result.PingResult) string {
	return some.SomeOutputPing(res) +
		fmt.Sprintf("\t%d/%d/%d\t%f/%f/%f/%f",
			res.Sent, res.Received, res.Duplicates,
			res.Minimum, res.Average, res.Median, res.Maximum,
		) +
		fmt.Sprintf("\t%s\t%v", res.Protocol, res.ReplyRtts())
}

func mostOutputDns(res *result.DnsResult) string {
	s := make([]string, 0)
	for _, detail := range res.Responses {
		s = append(s, mostOutputDnsResponse(&detail))
	}
	return some.SomeOutputDns(res) +
		"\t[" + strings.Join(s, " ") + "]"
}

func mostOutputDnsResponse(resp *result.DnsResponse) string {
	ret := fmt.Sprintf("%d\t%d\t%d\t%d",
		resp.AnswerCount,
		resp.QueriesCount,
		resp.NameServerCount,
		resp.AdditionalCount,
	)
	ret += mostOutputDnsResponseDetail(resp)
	return ret
}

func mostOutputDnsAnswer(detail *result.DnsAnswer) string {
	cl := result.DnsClassNames[detail.Class]
	if cl == "" {
		// yet unmapped class entries are represented as (Cxx)
		cl = fmt.Sprintf("(C%d)", detail.Class)
	}
	ty := result.DnsTypeNames[detail.Type]
	if ty == "" {
		// yet unmapped type entries are represented as (Txx)
		cl = fmt.Sprintf("(T%d)", detail.Type)
	}
	return fmt.Sprintf("%s %s %s '%s'", cl, ty, detail.Name, detail.Data)
}

func mostOutputDnsResponseDetail(resp *result.DnsResponse) string {
	s := make([]string, 0)
	for _, detail := range resp.AllAnswers() {
		s = append(s, mostOutputDnsAnswer(&detail))
	}
	return "\t[" + strings.Join(s, " ") + "]"
}

func mostOutputTraceroute(res *result.TracerouteResult) string {
	return some.SomeOutputTraceroute(res) +
		fmt.Sprintf("\t%v\t%d", res.DestinationReached(), res.ParisID)
}

func mostOutputCert(res *result.CertResult) string {
	ret := some.SomeOutputCert(res)
	if res.Alert == nil && res.Error == nil {
		ret += fmt.Sprintf("\t%s", res.ServerCipher)
	}
	certs := make([]string, 0)
	for _, cert := range res.Certificates {
		certs = append(certs, fmt.Sprintf("<%x %s>	",
			cert.SerialNumber,
			cert.Subject,
		))
	}
	ret += "\t[" + strings.Join(certs, "\t") + "]"
	return ret
}

func mostOutputHttp(res *result.HttpResult) string {
	return some.SomeOutputHttp(res) +
		fmt.Sprintf("\t\"%s\"\t%s\t%d\t%d\t%d",
			res.Error,
			res.Method,
			res.ResultCode,
			res.HeaderSize,
			res.BodySize,
		)
}

func mostOutputNtp(res *result.NtpResult) string {
	return some.SomeOutputNtp(res) +
		fmt.Sprintf("\t%s\t%v", res.Protocol, res.Replies)
}

func mostOutputConnection(res *result.ConnectionResult) string {
	return some.SomeOutputConnection(res) +
		fmt.Sprintf("\t%d\t%v", res.Asn, res.Prefix)
}

func mostOutputUptime(res *result.UptimeResult) string {
	return some.SomeOutputUptime(res)
}
