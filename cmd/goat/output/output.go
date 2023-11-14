/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

/*
  Defines the basics of how output formatters work. Each has:
	* a name selected with "--output"
	* a setup() function to init whatever variables it needs
	* a supports() function to determine if a particular result type is supported
	* a start() function to start a (new) batch of results
	* a process() function to deal with one incoming result
	* a finish() function to do something at the end of result processing

  When adding a new output formatter, don't forget to Register() it in init()
*/

package output

import (
	"fmt"
)

type outform struct {
	format   string
	supports func(string) bool
	setup    func(bool, []string)
	start    func()
	process  func(any)
	finish   func()
}

var formats map[string]outform

// package init
func init() {
	formats = make(map[string]outform, 0)
}

// Register a new output formatter with a name and th needed functions
func Register(
	format string,
	supports func(string) bool,
	setup func(bool, []string),
	start func(),
	process func(any),
	finish func(),
) {
	formats[format] = outform{format, supports, setup, start, process, finish}
}

// Verify check is a particular formatter has been defined
func Verify(format string, outtype string) bool {
	out, ok := formats[format]
	return ok && (outtype == "" || out.supports(outtype))
}

// Setup is called once, before any results are processed
func Setup(format string, isverbose bool, options []string) {
	if formatter, ok := formats[format]; ok {
		formatter.setup(isverbose, options)
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}

// Start is called before a batch of results is processed.
// It may be called more than once, i.e. once per batch.
func Start(format string) {
	if formatter, ok := formats[format]; ok {
		formatter.start()
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}

// Process one incoming result
func Process(format string, result any) {
	if formatter, ok := formats[format]; ok {
		formatter.process(result)
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}

// Finish is called after a batch of results is processed.
// It may be called more than once, i.e. once per batch.
func Finish(format string) {
	if formatter, ok := formats[format]; ok {
		formatter.finish()
	} else {
		// this should not happen - as long as VerifyFormatter was properly used
		panic(fmt.Sprintf("Unknown formatter %s was called\n", format))
	}
}
