/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "some" output formatter.
*/

package some

import (
	"fmt"
	"net/netip"
	"slices"
	"time"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
	"github.com/robert-kisteleki/goat/result"
)

var verbose bool
var total uint
var connectLastResults map[uint]*result.ConnectionResult
var connectTableOutput bool

func init() {
	output.Register("some", supports, setup, start, process, finish)
}

func supports(outtype string) bool {
	if slices.Contains([]string{"ping", "trace", "dns", "tls", "ntp", "http"}, outtype) ||
		outtype == "connection" || outtype == "uptime" ||
		outtype == "probe" || outtype == "anchor" || outtype == "msm" ||
		outtype == "status" {
		return true
	}
	return false
}

func setup(isverbose bool, options []string) {
	verbose = isverbose
	for _, opt := range options {
		if opt == "table" {
			connectTableOutput = true
			break
		}
	}
}

func start() {
	connectLastResults = make(map[uint]*result.ConnectionResult)
}

func process(res any) {
	out := ""
	switch t := res.(type) {
	case *result.Result:
		switch rt := (*t).(type) {
		case *result.PingResult:
			out = SomeOutputPing(rt)
		case *result.DnsResult:
			out = SomeOutputDns(rt)
		case *result.TracerouteResult:
			out = SomeOutputTraceroute(rt)
		case *result.CertResult:
			out = SomeOutputCert(rt)
		case *result.HttpResult:
			out = SomeOutputHttp(rt)
		case *result.NtpResult:
			out = SomeOutputNtp(rt)
		case *result.ConnectionResult:
			out = SomeOutputConnection(rt)
		case *result.UptimeResult:
			out = SomeOutputUptime(rt)
		}
	case goat.AsyncAnchorResult:
		out = t.Anchor.ShortString()
	case goat.AsyncProbeResult:
		out = t.Probe.ShortString()
	case goat.AsyncMeasurementResult:
		out = t.Measurement.ShortString()
	case goat.AsyncStatusCheckResult:
		out = t.Status.ShortString()
	default:
		out = fmt.Sprintf("No output formatter defined for object type '%T'\n", t)
	}

	if out != "" {
		fmt.Println(out)
		total++
	}
}

func finish() {
	for _, last := range connectLastResults {
		if last.Event == "connect" {
			fmt.Println(connectTableEntry(
				last.ProbeID,
				last.Asn,
				last.Prefix,
				last.GetTimeStamp(),
				time.Time{},
				last.Controller,
			))
			total++
		}
	}

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
	switch {
	case res.DnsError != "":
		ret += "\tDNSERR: " + res.DnsError
	case res.Alert == nil || res.Alert.Level == result.AlertLevelWarning:
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
	if connectTableOutput {
		return someOutputConnectionTable(res)
	} else {
		return someOutputConnectionSimple(res)
	}
}

func someOutputConnectionSimple(res *result.ConnectionResult) string {
	var e string
	switch res.Event {
	case "connect":
		e = "C"
	case "disconnect":
		e = "D"
	default:
		e = "?"
	}
	sas := "N/A"
	sprefix := "N/A"
	if res.Asn != 0 {
		sas = fmt.Sprintf("AS%d", res.Asn)
		sprefix = res.Prefix.String()
	}
	return res.BaseString() + fmt.Sprintf("\t%s\t%s\t%s\t%s", e, sas, sprefix, res.Controller)
}

func someOutputConnectionTable(res *result.ConnectionResult) string {
	ret := ""
	if res.Event == "disconnect" {
		from := time.Time{}
		if connectLastResults[res.ProbeID] != nil {
			from = connectLastResults[res.ProbeID].GetTimeStamp()
		}
		ret = connectTableEntry(
			res.ProbeID,
			res.Asn,
			res.Prefix,
			from,
			time.Time(res.TimeStamp),
			res.Controller,
		)
	}
	connectLastResults[res.ProbeID] = res
	return ret
}

func connectTableEntry(
	probe uint,
	as uint,
	prefix netip.Prefix,
	from time.Time,
	until time.Time,
	ctr string,
) string {
	asTime := func(t time.Time) string {
		if t.IsZero() {
			return "N/A"
		} else {
			return time.Time(t).UTC().Format("2006-01-02T15:04:05Z")
		}
	}
	sas := "N/A"
	sprefix := "N/A"
	if as != 0 {
		sas = fmt.Sprintf("AS%d", as)
		sprefix = prefix.String()
	}
	dur := "N/A"
	if !from.IsZero() && !until.IsZero() {
		dur = fmt.Sprint(until.Sub(from))
	}
	return fmt.Sprintf("%d\t%s\t%s\t%v\t%v\t%s\t%v", probe, sas, sprefix, asTime(from), asTime(until), dur, ctr)
}

func SomeOutputUptime(res *result.UptimeResult) string {
	return res.BaseString() +
		fmt.Sprintf("\t%d",
			res.Uptime,
		)
}
