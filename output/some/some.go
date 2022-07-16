/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "some" and "most" output formatters.
*/

package some

import (
	"fmt"

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
		fmt.Println(SomeOutputPing(r))
	case *result.DnsResult:
		fmt.Println(SomeOutputDns(r))
	case *result.TracerouteResult:
		fmt.Println(SomeOutputTraceroute(r))
	case *result.CertResult:
		fmt.Println(SomeOutputCert(r))
	case *result.HttpResult:
		fmt.Println(SomeOutputHttp(r))
	case *result.NtpResult:
		fmt.Println(SomeOutputNtp(r))
	case *result.ConnectionResult:
		fmt.Println(SomeOutputConnection(r))
	case *result.UptimeResult:
		fmt.Println(SomeOutputUptime(r))
	}
}

func Finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}

func SomeOutputPing(res *result.PingResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d/%d/%d\t%f/%f/%f/%f",
			res.Sent, res.Received, res.Duplicates,
			res.Minimum, res.Average, res.Median, res.Maximum,
		)
}

func SomeOutputDns(res *result.DnsResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d\t%d",
			len(res.Responses),
			len(res.Error),
		)
}

func SomeOutputTraceroute(res *result.TracerouteResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%s\t%d",
			res.Protocol,
			len(res.Hops),
		)
}

func SomeOutputCert(res *result.CertResult) string {
	ret := res.BaseString()
	if res.Error != nil {
		return ret + fmt.Sprintf("\tERROR: %s", *res.Error)
	}
	// TODO: test this
	if res.Alert != nil {
		ret += fmt.Sprintf("\tALERT: %d %d", res.Alert.Level, res.Alert.Description)
	}
	if res.Alert == nil || res.Alert.Level == result.AlertLevelWarning {
		ret += fmt.Sprintf("\t%s\t%s\t%f\t%d",
			res.Method,
			res.ProtocolVersion,
			res.ReplyTime,
			len(res.Certificates),
		)
	}
	return ret
}

func SomeOutputHttp(res *result.HttpResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%s", res.Uri)
}

func SomeOutputNtp(res *result.NtpResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%s\t%d\t%d\t%d",
			res.ReferenceID,
			res.Stratum,
			len(res.Replies),
			len(res.Errors),
		)
}

func SomeOutputConnection(res *result.ConnectionResult) string {
	var e string
	switch res.Event {
	case "connect":
		e = "C"
	case "disconnect":
		e = "D"
	default:
		e = "?"
	}
	return res.BaseString() + fmt.Sprintf("\t%s\t%s", e, res.Controller)
}

func SomeOutputUptime(res *result.UptimeResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d",
			res.Uptime,
		)
}
