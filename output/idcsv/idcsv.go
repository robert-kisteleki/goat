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
	"strings"

	"github.com/robert-kisteleki/goatapi"
)

var verbose bool
var total uint
var ids = make([]string, 0)

func init() {
	output.Register("idcsv", setup, start, process, finish)
}

func setup(isverbose bool, options []string) {
	verbose = isverbose
}

func start() {
}

func process(res any) {
	total++

	switch t := res.(type) {
	case goatapi.AsyncAnchorResult:
		ids = append(ids, fmt.Sprintf("%d", t.Anchor.ID))
	case goatapi.AsyncProbeResult:
		ids = append(ids, fmt.Sprintf("%d", t.Probe.ID))
	case goatapi.AsyncMeasurementResult:
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
