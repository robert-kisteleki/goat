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
	"goatcli/output"

	"github.com/robert-kisteleki/goatapi"
)

var verbose bool
var total uint

func init() {
	output.Register("id", setup, process, finish)
}

func setup(isverbose bool) {
	verbose = isverbose
}

func process(res any) {
	total++

	switch t := res.(type) {
	case goatapi.AsyncAnchorResult:
		fmt.Println(t.Anchor.ID)
	case goatapi.AsyncProbeResult:
		fmt.Println(t.Probe.ID)
	case goatapi.AsyncMeasurementResult:
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
