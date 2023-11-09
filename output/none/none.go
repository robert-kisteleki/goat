/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "some" output formatter.
*/

package none

import (
	"fmt"
	"goatcli/output"
	"slices"
)

var verbose bool
var total uint

func init() {
	output.Register("none", supports, setup, start, process, finish)
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
}

func start() {
	if verbose {
		fmt.Printf("# Not producing output\n")
	}
}

func process(res any) {
	total++
}

func finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}
