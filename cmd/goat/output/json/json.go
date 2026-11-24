/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "json" output formatter.
  It outputs the original JSON object of each result, as it was received.
*/

package json

import (
	"fmt"
	"slices"
	"strings"

	"github.com/robert-kisteleki/goat/cmd/goat/output"
	"github.com/robert-kisteleki/goat/result"
)

var verbose bool
var total uint

func init() {
	output.Register("json", supports, setup, start, process, finish)
}

func supports(outtype string) bool {
	return slices.Contains([]string{"ping", "trace", "dns", "tls", "ntp", "http", "connection", "uptime"}, outtype)
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
		fmt.Println(strings.TrimSpace((*t).GetRaw()))
	default:
		fmt.Printf("No output formatter defined for object type '%T'\n", t)
	}
}

func finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}
