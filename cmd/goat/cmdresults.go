/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
	"github.com/robert-kisteleki/goat/result"
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

	stream bool

	saveFileName string   // save raw result to this file (name)
	saveFile     *os.File // save raw result to this file (handle after opening)
	saveAll      bool
	output       string // output formater
	outopts      multioption
	limit        uint
}

// Implementation stub of the "result" subcommand, starting from command line args
func commandResult(args []string) {
	flags := parseResultArgs(args)
	commandResultFromFlags(flags)
}

// Actual implementation of the "result" subcommand. Based on flags it
// interacts with goatAPI to apply those filters+options to fetch results
func commandResultFromFlags(flags *resultFlags) {
	filter, options := processResultFlags(flags)

	formatter := options["output"].(string)

	if !output.Verify(formatter, "") {
		fmt.Fprintf(os.Stderr, "ERROR: unknown output format '%s'\n", formatter)
		os.Exit(1)
	}

	// if no ID and no file was specified then read from stdin
	if flags.filterID == 0 && flags.filterInfile == "" {
		filter.FilterFile("-")
	}

	if flags.saveFile != nil {
		defer flags.saveFile.Close()
	}

	if flagVerbose && flags.stream {
		limits := ""
		if flags.limit != 0 {
			limits = fmt.Sprintf("%d ", flags.limit)
		}
		fmt.Printf("# Listening on the stream for %sresults, starting at %v\n",
			limits,
			time.Now().UTC(),
		)
	}

	// most of the work is done by goatAPI
	// we receive results as they come in, via a channel
	results := make(chan result.AsyncResult)
	go filter.GetResults(flagVerbose, results)

	output.Setup(formatter, flagVerbose, flags.outopts)
	output.Start(formatter)
	for result := range results {
		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "EOF") {
				fmt.Fprintf(os.Stderr, "EOF\n")
			} else {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", result.Error)
			}
		} else {
			output.Process(formatter, result.Result)
		}
	}
	output.Finish(formatter)

	if flagVerbose && flags.stream {
		fmt.Printf("# Done listening to the stream at %v\n", time.Now().UTC())
	}
}

// Process flags (filters & options), pass most of them on to goatAPI
// while doing sanity checks on values
func processResultFlags(flags *resultFlags) (
	filter goat.ResultsFilter,
	options map[string]any,
) {
	options = make(map[string]any)
	filter = goat.NewResultsFilter()

	// options

	options["output"] = flags.output

	filter.Limit(flags.limit)

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

	filter.Stream(flags.stream)

	if flags.saveFileName != "" {
		f, err := os.Create(flags.saveFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: cold not write to logfile (%v)\n", err)
			os.Exit(1)
		}
		filter.Save(f)
	}

	filter.SaveAll(flags.saveAll)

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
	flagsGetResult.StringVar(&flags.saveFileName, "save", "", "Save raw results to this file")
	flagsGetResult.BoolVar(&flags.saveAll, "saveall", false, "Save all retrieved results, not only the ones that matched filters")
	flagsGetResult.BoolVar(&flags.stream, "stream", false, "Use the result stream")
	flagsGetResult.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most' or other available plugins")
	flagsGetResult.Var(&flags.outopts, "opt", "Options to pass to the output formatter")

	// limit
	flagsGetResult.UintVar(&flags.limit, "limit", 0, "Maximum amount of results to fetch")

	flagsGetResult.Parse(args)

	return &flags
}
