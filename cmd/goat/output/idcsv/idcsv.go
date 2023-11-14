/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "idcsv" output formatter.
*/

package idcsv

import (
	"fmt"
	"strings"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
)

var verbose bool
var total uint
var ids = make([]string, 0)

func init() {
	output.Register("idcsv", supports, setup, start, process, finish)
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
		ids = append(ids, fmt.Sprintf("%d", t.Anchor.ID))
	case goat.AsyncProbeResult:
		ids = append(ids, fmt.Sprintf("%d", t.Probe.ID))
	case goat.AsyncMeasurementResult:
		ids = append(ids, fmt.Sprintf("%d", t.Measurement.ID))
	default:
		fmt.Printf("No output formatter defined for object type '%T'\n", t)
	}
}

func finish() {
	if len(ids) > 0 {
		fmt.Println(strings.Join(ids, ","))
	}
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}
