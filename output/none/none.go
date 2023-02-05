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
)

var verbose bool
var total uint

func init() {
	output.Register("none", setup, start, process, finish)
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
