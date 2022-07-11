/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the "some" and "most" output formatters.
*/

package output

import (
	"fmt"

	"github.com/robert-kisteleki/goatapi/result"
)

func someSetup() {}

func someProcess(res result.Result) {
	fmt.Println(res.String())
}

func someFinish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}

func mostSetup() {}

func mostProcess(res result.Result) {
	fmt.Println(res.DetailString())
}

func mostFinish() {
	if verbose {
		fmt.Printf("# %d results\n", total)
	}
}
