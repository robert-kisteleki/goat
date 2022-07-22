/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "dnsstat" output formatter. It is pretty naive but shows potential.
*/

package dnsstat

import (
	"fmt"
	"goatcli/output"
	"strings"

	"github.com/robert-kisteleki/goatapi/result"
	"golang.org/x/exp/slices"
)

var verbose bool
var total uint
var dnsstatcollector map[string]uint

func init() {
	output.Register("dnsstat", setup, process, finish)
}

func setup(isverbose bool) {
	verbose = isverbose
	dnsstatcollector = make(map[string]uint)
}

func process(res *result.Result) {
	total++

	dns := (*res).(*result.DnsResult)
	var key string

	if len(dns.Error) > 0 {
		key = "ERROR"
		registerResult(key)
	} else {
		for _, resp := range dns.Responses {
			switch {
			case resp.Rcode != result.DnsRcodeNOERR: // collect non-NOERR results separetely
				key = result.DnsRcodeNames[resp.Rcode]
			case len(resp.Error) > 0: // collect TIMEOUTs separetely
				key = "TIMEOUT"
			default: // for the rest: extract "useful data"
				set := make([]string, 0)
				for _, ans := range resp.Answer {
					s := strings.Split(ans.Data, "\t")
					var r string
					if len(s) > 4 {
						r = strings.ReplaceAll(s[4], "'", "")
					}
					set = append(set, r)
				}
				slices.Sort(set)
				key = strings.Join(set, " ")
			}

			// count how many of these we had
			registerResult(key)
		}
	}
}

func finish() {
	var anssum uint = 0
	for k, v := range dnsstatcollector {
		fmt.Printf("%d\t\"%s\"\n", v, k)
		anssum += v
	}
	if verbose {
		fmt.Printf("# %d results, %d answers\n", total, anssum)
	}
}

func registerResult(key string) {
	if val, ok := dnsstatcollector[key]; ok {
		dnsstatcollector[key] = val + 1
	} else {
		dnsstatcollector[key] = 1
	}
}
