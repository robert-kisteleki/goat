/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the basics of how output formatters work. Each has:
    * a name selected with "--output"
	* a setup() function to init whatever variables it needs
	* a process() function to deal with one incoming result
	* a finish() function to do something at the end of result processing

  When adding a new output formatter, don't forget to Register() it in init()
*/

package output

import (
	"fmt"

	"goatcli/output/dnsstat"
	"goatcli/output/most"
	"goatcli/output/native"
	"goatcli/output/some"

	"github.com/robert-kisteleki/goatapi/result"
)

type outform struct {
	format  string
	setup   func(bool)
	process func(result.Result)
	finish  func()
}

var formats map[string]outform

// package init
func init() {
	formats = make(map[string]outform, 0)

	Register("some", some.Setup, some.Process, some.Finish)
	Register("most", most.Setup, most.Process, most.Finish)
	Register("native", native.Setup, native.Process, native.Finish)
	Register("dnsstat", dnsstat.Setup, dnsstat.Process, dnsstat.Finish)
}

// Register a new output formatter with a name and th needed functions
func Register(
	format string,
	setup func(bool),
	process func(result.Result),
	finish func(),
) {
	formats[format] = outform{format, setup, process, finish}
}

// Verify check is a particular formatter has been defined
func Verify(format string) bool {
	_, ok := formats[format]
	return ok
}

// Setup is called before any results are processed
func Setup(format string, isverbose bool) {
	if formatter, ok := formats[format]; ok {
		formatter.setup(isverbose)
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}

// Process one incoming result
func Process(format string, result result.Result) {
	if formatter, ok := formats[format]; ok {
		formatter.process(result)
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}

// Finish is called after all results are processed
func Finish(format string) {
	if formatter, ok := formats[format]; ok {
		formatter.finish()
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}
