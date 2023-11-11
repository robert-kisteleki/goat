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
	"goatcli/output/annotate"
	"slices"
	"sort"
	"strings"

	"github.com/robert-kisteleki/goatapi/result"
)

var verbose bool
var total uint
var dnsstatcollector map[string]*collectorItem
var makeCcStats bool
var makeAsnStats bool
var showprogress bool
var typeFocus []string

type collectorItem struct {
	Total uint
	CCs   map[string]uint
	Asns  map[string]uint
}

func init() {
	output.Register("dnsstat", supports, setup, start, process, finish)
}

func supports(outtype string) bool {
	return outtype == "dns"
}

func setup(isverbose bool, options []string) {
	verbose = isverbose
	if slices.Contains(options, "ccstat") {
		if verbose {
			fmt.Println("# Enabled CC statistics")
		}
		makeCcStats = true
	}
	if slices.Contains(options, "asnstat") {
		if verbose {
			fmt.Println("# Enabled ASN statistics")
		}
		makeAsnStats = true
	}
	if slices.Contains(options, "progress") {
		if verbose {
			fmt.Println("# Enabled progress indicator")
		}
		showprogress = true
	}
	for _, opt := range options {
		if opt[:5] == "type:" {
			typeFocus = strings.Split(opt[5:], "+")
			break
		}
	}
	if len(typeFocus) != 0 {
		if verbose {
			fmt.Printf("# Focusing on %v types\n", typeFocus)
		}
	}
}

func start() {
	dnsstatcollector = make(map[string]*collectorItem)
	annotate.InitProbeCache()
}

func process(res any) {
	total++

	if verbose && showprogress {
		fmt.Printf("\r# Receiving results: %d", total)
	}

	switch t := res.(type) {
	case *result.Result:
		switch (*t).(type) {
		case *result.DnsResult:
			// ok
		default:
			fmt.Printf("This output formatter only works for DNS results\n")
			return
		}
	default:
		fmt.Printf("This output formatter only works for DNS results\n")
		return
	}

	resconv := res.(*result.Result)
	dns := (*resconv).(*result.DnsResult)
	var key string

	if len(dns.Error) > 0 {
		key = "ERROR"
		registerResult(key, dns.AddressFamily, dns.ProbeID)
	} else {
		for _, resp := range dns.Responses {
			switch {
			case resp.Rcode != result.DnsRcodeNOERR: // collect non-NOERR results separetely
				key = result.DnsRcodeNames[resp.Rcode]
			case len(resp.Error) > 0: // collect TIMEOUTs separetely
				key = "TIMEOUT"
			default: // for the rest: extract "useful data"
				key = output.OutputDnsResponseFocus(&resp, typeFocus)
			}

			// count how many of these we had
			registerResult(key, resp.AddressFamily, dns.ProbeID)
		}
	}
}

func finish() {
	type valPlusCount struct {
		val *collectorItem
		key string
	}

	if verbose && showprogress {
		fmt.Println()
	}

	vpc := make([]valPlusCount, 0)
	for key, val := range dnsstatcollector {
		vpc = append(vpc, valPlusCount{val, key})
	}
	sort.Slice(vpc, func(i, j int) bool { return vpc[i].val.Total > vpc[j].val.Total })

	var anssum uint = 0
	for _, v := range vpc {
		fmt.Printf("%d\t\"%s\"", v.val.Total, v.key)
		if makeCcStats {
			fmt.Print("\t")
			printTopN(v.val.CCs, 5, "")
		}
		if makeAsnStats {
			fmt.Print("\t")
			printTopN(v.val.Asns, 5, "AS")
		}
		fmt.Println()
		anssum += v.val.Total
	}

	if verbose {
		fmt.Printf("# %d results, %d answers\n", total, anssum)
	}
}

func registerResult(key string, af uint, pid uint) {
	if val, ok := dnsstatcollector[key]; ok {
		val.Total = val.Total + 1
		val.CCs[annotate.GetProbeCountry(pid)]++
		if af == 4 {
			val.Asns[annotate.GetProbeAsn4(pid)]++
		} else {
			val.Asns[annotate.GetProbeAsn6(pid)]++
		}
	} else {
		dnsstatcollector[key] = &collectorItem{
			1,
			make(map[string]uint),
			make(map[string]uint),
		}
		dnsstatcollector[key].CCs[annotate.GetProbeCountry(pid)] = 1
		if af == 4 {
			dnsstatcollector[key].Asns[annotate.GetProbeAsn4(pid)] = 1
		} else {
			dnsstatcollector[key].Asns[annotate.GetProbeAsn6(pid)] = 1
		}
	}
}

func printTopN(data map[string]uint, max uint, as string) {
	type valPlusCount struct {
		val   string
		count uint
	}

	vpc := make([]valPlusCount, 0)
	for key, val := range data {
		if key == "N/A" {
			vpc = append(vpc, valPlusCount{"(N/A)", val})
		} else {
			vpc = append(vpc, valPlusCount{key, val})
		}
	}
	sort.Slice(vpc, func(i, j int) bool { return vpc[i].count > vpc[j].count })
	for i := 0; i < 10 && i < len(vpc); i++ {
		fmt.Printf(" %s%s:%d", as, vpc[i].val, vpc[i].count)
	}
}
