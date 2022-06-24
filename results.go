/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"os"

	"github.com/robert-kisteleki/goatapi"
	"github.com/robert-kisteleki/goatapi/result"
	"golang.org/x/exp/slices"
)

// struct to receive/store command line args for downloads
type resultFlags struct {
	filterID           uint
	filterStart        string
	filterStop         string
	filterProbeIDs     string
	filterAnchors      bool
	filterPublicProbes bool

	format string
	limit  uint
}

// Implementation of the "result" subcommand. Parses command line flags
// and interacts with goatAPI to apply those filters+options to fetch results
func commandResult(args []string) {
	flags := parseResultArgs(args)
	filter, options := processResultFlags(flags)

	// this is left here intentionally as an alternative
	/*
		// most of the work is done by goatAPI
		results, err := filter.GetResults(flagVerbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}

		if flagVerbose {
			fmt.Printf("%d results\n", len(results))
		}
	*/

	// most of the work is done by goatAPI
	// we receive results as they come in, via a channel
	results := make(chan result.AsyncResult)
	go filter.GetResultsAsync(flagVerbose, results)

	total := 0
	for result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", result.Error)
		} else {
			res := result.Result
			switch options["format"] {
			case "some":
				fmt.Println(res.ShortString())
			case "most":
				fmt.Println(res.LongString())
			}
			total++
		}
	}
	if flagVerbose {
		fmt.Printf("%d results\n", total)
	}
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

	formats := []string{"some", "most"}
	if slices.Contains(formats, flags.format) {
		options["format"] = flags.format
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: invalid output format\n")
		os.Exit(1)
	}
	if flags.limit == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: limit should be positive\n")
		os.Exit(1)
	} else {
		filter.Limit(flags.limit)
	}

	// filters

	if flags.filterID != 0 {
		filter.FilterID(flags.filterID)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: measurement ID must be sepcified with --id\n")
		os.Exit(1)
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

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseResultArgs(args []string) *resultFlags {
	var flags resultFlags

	// filters
	flagsGetResult.UintVar(&flags.filterID, "id", 0, "The measurement ID to fetch results for. Mandatory.")
	flagsGetResult.StringVar(&flags.filterStart, "start", "", "Earliest timestamp for results")
	flagsGetResult.StringVar(&flags.filterStop, "stop", "", "Latest timestamp for results")
	flagsGetResult.StringVar(&flags.filterProbeIDs, "probe", "", "Filter on probe ID being on this comma separated list")
	flagsFindMsm.BoolVar(&flags.filterAnchors, "anchor", false, "Filter for achors only")
	flagsFindMsm.BoolVar(&flags.filterPublicProbes, "public", false, "Filter for public probes only")

	// options
	flagsGetResult.StringVar(&flags.format, "format", "some", "Output contents: 'some' or 'most'")

	// limit
	flagsGetResult.UintVar(&flags.limit, "limit", 1000, "Maximum amount of results to parse")

	flagsGetResult.Parse(args)

	return &flags
}
