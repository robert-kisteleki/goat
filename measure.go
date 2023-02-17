/*
  (C) 2022, 2023 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"goatcli/output"
	"os"
	"time"

	"github.com/robert-kisteleki/goatapi"
)

// struct to receive/store command line args for new measurements
type measureFlags struct {
	output  string
	outopts multioption
}

// Implementation of the "measure" subcommand. Parses command line flags
// and interacts with goatAPI to initiate new mwasurements
func commandMeasure(args []string) {
	flags := parseMeasureArgs(args)
	spec, options := processMeasureFlags(flags)
	formatter := options["output"].(string)

	if !output.Verify(formatter) {
		fmt.Fprintf(os.Stderr, "ERROR: unknown output format '%s'\n", formatter)
		os.Exit(1)
	}

	// most of the work is done by goatAPI
	err := spec.Submit(flagVerbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	/*
		output.Setup(formatter, flagVerbose, flags.outopts)
		output.Start(formatter)
		for status := range statuses {
			if status.Error != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", status.Error)
				os.Exit(1)
			} else {
				output.Process(formatter, status)
			}
		}
		output.Finish(formatter)
	*/
}

// Process flags (options), pass most of them on to goatAPI
// while doing sanity checks on values
func processMeasureFlags(flags *statusCheckFlags) (
	spec *goatapi.MeasurementSpec,
	options map[string]any,
) {
	options = make(map[string]any)
	spec = goatapi.NewMeasurementSpec()

	spec.AddProbesCountryWithTags("HU", 50, &[]string{"i1, i2"}, &[]string{"e1", "e2"})
	spec.AddProbesArea("WW", 10)
	spec.AddProbesList([]uint{1, 2, 3})
	spec.AddProbesReuse(10000009, 9)

	spec.Start(time.Now().Add(time.Second * 40)) // TODO is this ok
	spec.Stop(time.Now().Add(time.Minute * 40))  // TODO is this ok

	spec.AddPing("ping1", "ping.ripe.net", true, 4, nil)
	spec.AddPing("ping2", "ping.ripe.net", true, 6, &goatapi.PingOptions{
		SkipDNSCheck:   true,
		PacketSize:     999,
		IncludeProbeID: true,
		Interval:       88,
		Tags:           []string{"r", "rr"},
	})

	// options

	options["output"] = flags.output

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseMeasureArgs(args []string) *statusCheckFlags {
	var flags statusCheckFlags

	// options
	flagsMeasure.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most'")
	flagsMeasure.Var(&flags.outopts, "opt", "Options to pass to the output formatter")

	flagsMeasure.Parse(args)

	return &flags
}
