/*
  (C) 2022, 2023 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"os"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
)

// struct to receive/store command line args for status checks
type statusCheckFlags struct {
	filterID      uint // measurement ID
	filterAllRTTs bool // all responses?

	output  string
	outopts multioption
}

// Implementation of the "status check" subcommand. Parses command line flags
// and interacts with goatAPI to apply those filters+options to fetch the result
func commandStatusCheck(args []string) {
	flags := parseStatusCheckArgs(args)
	filter, options := processStatusCheckFlags(flags)
	formatter := options["output"].(string)

	if !output.Verify(formatter, "status") {
		fmt.Fprintf(os.Stderr, "ERROR: unknown output format '%s' for status\n", formatter)
		os.Exit(1)
	}

	// most of the work is done by goatAPI
	statuses := make(chan goat.AsyncStatusCheckResult)
	go filter.StatusCheck(flagVerbose, statuses)

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
}

// Process flags (filters & options), pass most of them on to goatAPI
// while doing sanity checks on values
func processStatusCheckFlags(flags *statusCheckFlags) (
	filter goat.StatusCheckFilter,
	options map[string]any,
) {
	options = make(map[string]any)
	filter = goat.NewStatusCheckFilter()

	// options

	options["output"] = flags.output

	// filters

	if flags.filterID == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: measurement ID must be spcified\n")
		os.Exit(1)
	}
	filter.MsmID(flags.filterID)
	filter.GetAllRTTs(flags.filterAllRTTs)

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseStatusCheckArgs(args []string) *statusCheckFlags {
	var flags statusCheckFlags

	// filters
	flagsStatusCheck.UintVar(&flags.filterID, "id", 0, "Measurement ID to check status for")
	flagsStatusCheck.BoolVar(&flags.filterAllRTTs, "all", false, "Retrieve all recent RTTs")

	// options
	flagsStatusCheck.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most'")
	flagsStatusCheck.Var(&flags.outopts, "opt", "Options to pass to the output formatter")

	flagsStatusCheck.Parse(args)

	return &flags
}
