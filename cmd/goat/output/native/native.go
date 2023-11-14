/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "native" output formatter.
  This format tries to produce output that is similar to the native
  tools, e.g. ping, traceroute, dig and so on.
*/

package native

import (
	"fmt"
	"strings"

	"github.com/robert-kisteleki/goat/cmd/goat/output"
	"github.com/robert-kisteleki/goat/result"
)

var verbose bool
var total uint

func init() {
	output.Register("native", supports, setup, start, process, finish)
}

func supports(outtype string) bool {
	if outtype == "ping" || outtype == "trace" {
		return true
	}
	return false
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
			nativeOutputPing(rt)
		case *result.TracerouteResult:
			nativeOutputTraceroute(rt)
		case *result.DnsResult:
			nativeOutputDns(rt)
		default:
			fmt.Printf("No output formatter defined for result type '%T'\n", rt)
		}
	default:
		fmt.Printf("No output formatter defined for object type '%T'\n", t)
	}
}

func finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}

// Create a close-to-native output for a ping result
func nativeOutputPing(res *result.PingResult) {
	fmt.Printf("PROBE %d PING %s (%v): %d data bytes\n",
		res.ProbeID,
		res.Destination(),
		res.DestinationAddr,
		res.PacketSize-8,
	)
	for i, reply := range res.Replies {
		fmt.Printf("%d bytes from %v: icmp_seq=%d ttl=%d time=%.3f ms\n",
			res.PacketSize,
			reply.Source,
			i,
			reply.Ttl,
			reply.Rtt,
		)
	}
	fmt.Printf("--- %s ping statistics ---\n", res.DestinationName)
	loss := 1.0
	if res.Received != 0 {
		loss = 100.0 - float64(res.Sent/res.Received)*100
	}
	fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n",
		res.Sent,
		res.Received,
		loss,
	)
	fmt.Printf("round-trip min/avg/med/max = %.3f/%.3f/%.3f/%.3f ms\n",
		res.Minimum,
		res.Average,
		res.Median,
		res.Maximum,
	)
	fmt.Println()
}

func nativeOutputTraceroute(res *result.TracerouteResult) {
	fmt.Printf("PROBE %d traceroute to %s (%v): %d hops max, %d byte packets\n",
		res.ProbeID,
		res.Destination(),
		res.DestinationAddr,
		255,
		res.PacketSize,
	)
	for _, hop := range res.Hops {
		last := ""
		for i, ans := range hop.Responses {
			if i == 0 {
				fmt.Printf("%3d  ", hop.HopNumber)
			}
			switch {
			case ans.Timeout:
				fmt.Printf("*")
			case ans.Error != nil:
				fmt.Printf("[ERROR: %s]", *ans.Error)
			default:
				if last != ans.From.String() {
					if last != "" {
						fmt.Printf("\n     ")
					}
					fmt.Printf("%s (%s)", ans.From, ans.From)
				}
				if ans.Late != nil {
					fmt.Printf(" LATE")
				} else {
					fmt.Printf(" %.3f ms", ans.Rtt)
				}
				last = ans.From.String()
			}
			if i != len(hop.Responses)-1 {
				fmt.Printf(" ")
			} else {
				fmt.Println()
			}
		}
	}
}

func nativeOutputDns(res *result.DnsResult) {
	fmt.Printf("; Probe %d, source %v\n", res.ProbeID, res.FromAddr)

	for _, resp := range res.Responses {
		fmt.Println(";; Got answer:")
		fmt.Printf(";; ->>HEADER<<- opcode: QUERY, status: %s, id: %d\n",
			result.DnsRcodeNames[resp.Rcode],
			resp.QueryID,
		)
		flags := "qr"
		if resp.Authoritative {
			flags += " aa"
		}
		if resp.RecursionDesired {
			flags += " rd"
		}
		flags += fmt.Sprintf("; QUERY: %d, ANSWER: %d, AUTHORITY: %d, ADDITIONAL: %d",
			resp.QueriesCount,
			resp.AnswerCount,
			resp.NameServerCount,
			resp.AdditionalCount,
		)
		fmt.Println(";;", flags)
		if resp.RecursionDesired && !resp.RecursionAvailable {
			fmt.Println(";; WARNING: recursion requested but not available")
		}
		fmt.Println()

		if resp.AdditionalCount > 0 {
			for _, ans := range resp.Extra {
				// OPT
				if ans.Type == 41 {
					fmt.Println(";; OPT PSEUDOSECTION:")
					fmt.Printf("; EDNS: version: %d; flags:; udp: %d\n", ans.Ttl, ans.Class)
					if len(resp.Edsn0Nsid) > 0 {
						hex := make([]string, len(resp.Edsn0Nsid))
						for i, c := range resp.Edsn0Nsid {
							hex[i] = fmt.Sprintf("%x", c)
						}
						fmt.Printf("; NSID: %s (\"%s\")\n", strings.Join(hex, " "), resp.Edsn0Nsid)
					}
				}
			}
		}

		nativeOutputDnsResponse(resp)

		fmt.Println()
		fmt.Printf(";; Query time: %d msec\n", int(resp.ResponseTime))
		fmt.Printf(";; SERVER: %v\n", resp.Destination)
		fmt.Printf(";; WHEN: %v\n", resp.TimeStamp)
		fmt.Printf(";; MSG SIZE  rcvd: %d\n", resp.ResponseSize)
		fmt.Println()
	}
}

func nativeOutputDnsResponse(resp result.DnsResponse) {
	printrec := func(ans result.DnsAnswer) {
		fmt.Printf("%s\t%d\t%s\t%s\t%s\n",
			ans.Name,
			ans.Ttl,
			result.DnsClassNames[ans.Class],
			result.DnsTypeNames[ans.Type],
			ans.Data,
		)
	}
	fmt.Println(";; QUESTION SECTION:")
	fmt.Printf(";%s\t%s\t%s\n",
		resp.Question.Name,
		result.DnsClassNames[resp.Question.Class],
		result.DnsTypeNames[resp.Question.Type],
	)
	fmt.Println()

	if resp.AnswerCount > 0 {
		fmt.Println(";; ANSWER SECTION:")
		for _, ans := range resp.Answer {
			printrec(ans)
		}
	}
	if resp.AdditionalCount > 0 {
		fmt.Println()
		fmt.Println(";; ADDITIONAL SECTION:")
		for _, ans := range resp.Extra {
			if ans.Type != 41 { // skip OPT
				printrec(ans)
			}
		}
	}
}
