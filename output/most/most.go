/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "most" output formatter.
*/

package most

import (
	"fmt"
	"goatcli/output"
	"goatcli/output/annotate"
	"goatcli/output/some"
	"strings"

	"github.com/robert-kisteleki/goatapi"
	"github.com/robert-kisteleki/goatapi/result"
)

var verbose bool
var total uint

func init() {
	output.Register("most", setup, start, process, finish)
}

func setup(isverbose bool, options []string) {
	verbose = isverbose
}

func start() {
}

func process(res any) {
	total++

	switch t := res.(type) {
	case *result.Result:
		switch rt := (*t).(type) {
		case *result.PingResult:
			fmt.Println(mostOutputPing(rt))
		case *result.DnsResult:
			fmt.Println(mostOutputDns(rt))
		case *result.TracerouteResult:
			fmt.Println(mostOutputTraceroute(rt))
		case *result.CertResult:
			fmt.Println(mostOutputCert(rt))
		case *result.HttpResult:
			fmt.Println(mostOutputHttp(rt))
		case *result.NtpResult:
			fmt.Println(mostOutputNtp(rt))
		case *result.ConnectionResult:
			fmt.Println(mostOutputConnection(rt))
		case *result.UptimeResult:
			fmt.Println(mostOutputUptime(rt))
		default:
			fmt.Printf("No output formatter defined for result type '%T'\n", rt)
		}
	case goatapi.AsyncAnchorResult:
		fmt.Println(t.Anchor.LongString())
	case goatapi.AsyncProbeResult:
		fmt.Println(t.Probe.LongString())
	case goatapi.AsyncMeasurementResult:
		fmt.Println(t.Measurement.LongString())
	default:
		fmt.Printf("No output formatter defined for object type '%T'\n", t)
	}
}

func finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}

func mostOutputPing(res *result.PingResult) string {
	return some.SomeOutputPing(res) +
		fmt.Sprintf("\t%s", annotate.GetProbeCountry(res.ProbeID)) +
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
