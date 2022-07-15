/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "some" and "most" output formatters.
*/

package output

import (
	"fmt"

	"github.com/robert-kisteleki/goatapi/result"
)

func someSetup() {
}

func someProcess(res result.Result) {
	switch r := res.(type) {
	case *result.PingResult:
		fmt.Println(someOutputPing(r))
	case *result.DnsResult:
		fmt.Println(someOutputDns(r))
	case *result.TracerouteResult:
		fmt.Println(someOutputTraceroute(r))
	case *result.CertResult:
		fmt.Println(someOutputCert(r))
	case *result.HttpResult:
		fmt.Println(someOutputHttp(r))
	case *result.NtpResult:
		fmt.Println(someOutputNtp(r))
	case *result.ConnectionResult:
		fmt.Println(someOutputConnection(r))
	case *result.UptimeResult:
		fmt.Println(someOutputUptime(r))
	}
}

func someFinish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}

func someOutputPing(res *result.PingResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d/%d/%d\t%f/%f/%f/%f",
			res.Sent, res.Received, res.Duplicates,
			res.Minimum, res.Average, res.Median, res.Maximum,
		)
}

func someOutputDns(res *result.DnsResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d\t%d",
			len(res.Responses),
			len(res.Error),
		)
}

func someOutputTraceroute(res *result.TracerouteResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%s\t%d",
			res.Protocol,
			len(res.Hops),
		)
}

func someOutputCert(res *result.CertResult) string {
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

func someOutputHttp(res *result.HttpResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%s", res.Uri)
}

func someOutputNtp(res *result.NtpResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%s\t%d\t%d\t%d",
			res.ReferenceID,
			res.Stratum,
			len(res.Replies),
			len(res.Errors),
		)
}

func someOutputConnection(res *result.ConnectionResult) string {
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

func someOutputUptime(res *result.UptimeResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d",
			res.Uptime,
		)
}
