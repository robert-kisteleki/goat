/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Result interface {
	Parse(from string) (err error)
	TypeName() string
	GetTimeStamp() time.Time
	GetProbeID() uint
	GetFirmwareVersion() uint
}

type AsyncResult struct {
	Result *Result
	Error  error
}

func Parse(from string) (Result, error) {
	var res Result = &BaseResult{}
	err := res.Parse(from)
	if err != nil {
		return nil, err
	}
	if res.GetFirmwareVersion() <= 1 && res.TypeName() != "connection" {
		return nil, fmt.Errorf("firmware version 1 and below results are not supported")
	}
	return ParseWithTypeHint(from, res.TypeName())
}

func ParseWithTypeHint(from string, typehint string) (Result, error) {
	var res Result
	switch typehint {
	case "":
		return Parse(from)
	case "ping":
		pingres := PingResult{}
		res = &pingres
	case "traceroute":
		traceres := TracerouteResult{}
		res = &traceres
	case "dns":
		dnsres := DnsResult{}
		res = &dnsres
	case "ntp":
		ntpres := NtpResult{}
		res = &ntpres
	case "sslcert":
		certres := CertResult{}
		res = &certres
	case "http":
		httpres := HttpResult{}
		res = &httpres
	case "uptime":
		uptimeres := UptimeResult{}
		res = &uptimeres
	case "connection":
		connres := ConnectionResult{}
		res = &connres
	default:
		return nil, fmt.Errorf("unknown/unsupported result type %s", typehint)
	}
	err := res.Parse(from)
	return res, err
}

// StoreDelay calculates the difference (in seconds) between taking the
// measurement and storing it, i.e. how long it took for the result to be
// available. Be aware that The probe's clock may be inaccurate, so this
// value is indicative only and may even be negative
func (res *BaseResult) StoreDelay() int {
	stts := time.Time(res.StoreTimeStamp)
	ts := time.Time(res.TimeStamp)
	return int(stts.Sub(ts).Seconds())
}

// a datetime type that can be unmarshaled from UNIX epoch *or* ISO times
type uniTime time.Time

func (ut *uniTime) UnmarshalJSON(data []byte) error {
	// try parsing as UNIX epoch first
	epoch, err := strconv.Atoi(string(data))
	if err == nil {
		*ut = uniTime(time.Unix(int64(epoch), 0))
		return nil
	}

	// try parsing ISO8601(Z)
	layout := "2006-01-02T15:04:05"
	noquote := strings.ReplaceAll(string(data), "\"", "")
	noz := strings.ReplaceAll(noquote, "Z", "")
	unix, err := time.Parse(layout, noz)
	if err != nil {
		return err
	}
	*ut = uniTime(unix)
	return nil
}

// default output format for uniTime type is ISO8601
func (ut uniTime) String() string {
	return time.Time(ut).UTC().Format("2006-01-02T15:04:05Z")
}

// valueOrNA turns various types into a string if they have values
// (i.e. pointer is not nil) or "N/A" otherwise
func valueOrNA[T any](prefix string, quote bool, val *T) string {
	if val != nil {
		if quote {
			return fmt.Sprintf("\t\"%s%v\"", prefix, *val)
		} else {
			return fmt.Sprintf("\t%s%v", prefix, *val)
		}
	} else {
		return "\tN/A"
	}
}
