/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"goatcli/output"
	"os"

	"github.com/robert-kisteleki/goatapi"
	"github.com/robert-kisteleki/goatapi/result"
)

// struct to receive/store command line args for downloads
type resultFlags struct {
	filterID           uint
	filterInfile       string
	filterStart        string
	filterStop         string
	filterProbeIDs     string
	filterAnchors      bool
	filterPublicProbes bool
	filterLatest       bool

	output string // output formater
	limit  uint
}

// Implementation of the "result" subcommand. Parses command line flags
// and interacts with goatAPI to apply those filters+options to fetch results
func commandResult(args []string) {
	flags := parseResultArgs(args)
	filter, options := processResultFlags(flags)

	formatter := options["output"].(string)

	if !output.Verify(formatter) {
		fmt.Fprintf(os.Stderr, "ERROR: unknown output format '%s'\n", formatter)
		os.Exit(1)
	}

	// this is left here intentionally as an alternative
	/*
		// most of the work is done by goatAPI
		results, err := filter.GetResults(flagVerbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}

		if flagVerbose {
			fmt.Printf("# %d results\n", len(results))
		}
	*/

	// if no ID and no file was specified then read from stdin
	if flags.filterID == 0 && flags.filterInfile == "" {
		filter.FilterFile("-")
	}

	// most of the work is done by goatAPI
	// we receive results as they come in, via a channel
	results := make(chan result.AsyncResult)
	go filter.GetResultsAsync(flagVerbose, results)

	output.Setup(formatter, flagVerbose)
	for result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", result.Error)
		} else {
			output.Process(formatter, result.Result)
		}
	}
	output.Finish(formatter)
}

// Process flags (filters & options), pass most of them on to goatAPI
// while doing sanity checks on values
func processResultFlags(flags *resultFlags) (
	filter goatapi.ResultsFilter,
	options map[string]any,
) {
	options = make(map[string]any)
	filter = goatapi.NewResultsFilter()

	// options

	options["output"] = flags.output

	if flags.limit == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: limit should be positive\n")
		os.Exit(1)
	} else {
		filter.Limit(flags.limit)
	}

	// filters

	if flags.filterID != 0 {
		filter.FilterID(flags.filterID)
	}

	if flags.filterInfile != "" {
		filter.FilterFile(flags.filterInfile)
	}

	if flags.filterProbeIDs != "" {
		list, err := makeIntList(flags.filterProbeIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: invalid ID in list: %s\n", flags.filterProbeIDs)
			os.Exit(1)
		}
		filter.FilterProbeIDs(list)
	}

	if flags.filterStart != "" {
		time, err := parseTimeAlternatives(flags.filterStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStart(time)
	}

	if flags.filterStop != "" {
		time, err := parseTimeAlternatives(flags.filterStop)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse stop time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStop(time)
	}

	if flags.filterAnchors {
		filter.FilterAnchors()
	}
	if flags.filterPublicProbes {
		filter.FilterPublicProbes()
	}
	if flags.filterLatest {
		filter.FilterLatest()
	}

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseResultArgs(args []string) *resultFlags {
	var flags resultFlags

	// filters
	flagsGetResult.UintVar(&flags.filterID, "id", 0, "The measurement ID to fetch results for.")
	flagsGetResult.StringVar(&flags.filterInfile, "file", "", "A file to fetch measurement from. Use \"-\" or \"\" for stdin")
	flagsGetResult.StringVar(&flags.filterStart, "start", "", "Earliest timestamp for results")
	flagsGetResult.StringVar(&flags.filterStop, "stop", "", "Latest timestamp for results")
	flagsGetResult.StringVar(&flags.filterProbeIDs, "probe", "", "Filter on probe ID being on this comma separated list")
	flagsGetResult.BoolVar(&flags.filterAnchors, "anchor", false, "Filter for achors only")
	flagsGetResult.BoolVar(&flags.filterPublicProbes, "public", false, "Filter for public probes only")
	flagsGetResult.BoolVar(&flags.filterLatest, "latest", false, "Filter for latest results only")

	// options
	flagsGetResult.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most' or other available plugins")

	// limit
	flagsGetResult.UintVar(&flags.limit, "limit", 1000, "Maximum amount of results to parse")

	flagsGetResult.Parse(args)

	return &flags
}
