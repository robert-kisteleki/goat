/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "native" output formatter.
  This format tries to produce output that is similar to the native
  tools, e.g. ping, traceroute, dig and so on.
*/

package output

import (
	"fmt"

	"github.com/robert-kisteleki/goatapi/result"
)

func nativeSetup() {}

func nativeProcess(res result.Result) {
	switch r := res.(type) {
	case *result.PingResult:
		nativeOutputPing(r)
	/*
		case *result.TracerouteResult:
			nativeOutputTraceroute(r)
	*/
	/*
		case *result.DnsResult:
			nativeOutputDns(r)
	*/
	default:
		fmt.Printf("No native formatter defined for result type '%s'\n",
			res.TypeName(),
		)
	}
}

func nativeFinish() {
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
	fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n",
		res.Sent,
		res.Received,
		100.0-float32(res.Sent/res.Received)*100,
	)
	fmt.Printf("round-trip min/avg/med/max = %.3f/%.3f/%.3f/%.3f ms\n",
		res.Minimum,
		res.Average,
		res.Median,
		res.Maximum,
	)
	fmt.Println()
}

/*
func nativeOutputTraceroute(res *result.TracerouteResult) {

}
*/

/*
func nativeOutputDns(res *result.DnsResult) {

}
*/
