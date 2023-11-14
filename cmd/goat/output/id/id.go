/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "id" output formatter.
*/

package id

import (
	"fmt"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
)

var verbose bool
var total uint

func init() {
	output.Register("id", supports, setup, start, process, finish)
}

func supports(outtype string) bool {
	if outtype == "probe" || outtype == "anchor" || outtype == "msm" {
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
	case goat.AsyncAnchorResult:
		fmt.Println(t.Anchor.ID)
	case goat.AsyncProbeResult:
		fmt.Println(t.Probe.ID)
	case goat.AsyncMeasurementResult:
		fmt.Println(t.Measurement.ID)
	default:
		fmt.Printf("No output formatter defined for object type '%T'\n", t)
	}
}

func finish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}
