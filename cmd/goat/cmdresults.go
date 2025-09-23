/*
  (C) Robert Kisteleki & RIPE NCC

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
	filterLatestDays   int

	stream bool

	saveFileName string   // save raw result to this file (name)
	saveFile     *os.File // save raw result to this file (handle after opening)
	saveAll      bool
	output       string // output formater
	outopts      multioption
	limit        uint
	timeout      uint // stream read timeout in secods
	backlog      bool // ask for stream result backlog?
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
		fmt.Printf("# Listening on the stream (at %s) for %sresults, starting at %v\n",
			goat.GetStreamBase(),
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
	filter.StreamTimeout(time.Duration(flags.timeout) * time.Second)

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
		t, err := parseTimeAlternatives(flags.filterStart)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse start time (%v)\n", err)
			os.Exit(1)
		}
		filter.FilterStart(t)
	}

	if flags.filterStop != "" {
		t, err := parseTimeAlternatives(flags.filterStop)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not parse stop time (%v)\n", err)
			os.Exit(1)
		}
		// if stop time is precisely a day boundary then make it the previous day 22:59:59 instead
		// this is so that we don't actually ask for results for the first second of the stop day
		h, m, s := t.Clock()
		if h == 0 && m == 0 && s == 0 {
			t = t.Add(time.Second * -1)
		}
		filter.FilterStop(t)
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
	if flags.filterLatestDays != -1 {
		if flags.filterLatestDays < 0 || flags.filterLatestDays > 16384 {
			fmt.Fprintf(os.Stderr, "ERROR: lookback days should be between 0 (all) and 16384\n")
			os.Exit(1)
		}
		filter.FilterLatestLookbackDays(uint(flags.filterLatestDays))
	}

	filter.Stream(flags.stream)
	filter.SendBacklog(flags.backlog)

	if flags.saveFileName != "" {
		f, err := os.Create(flags.saveFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not write to logfile (%v)\n", err)
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
	flagsGetResult.IntVar(&flags.filterLatestDays, "lookback", -1, "How many days to look back to consider a result. Default is 7.")

	// options
	flagsGetResult.StringVar(&flags.saveFileName, "save", "", "Save raw results to this file")
	flagsGetResult.BoolVar(&flags.saveAll, "saveall", false, "Save all retrieved results, not only the ones that matched filters")
	flagsGetResult.BoolVar(&flags.stream, "stream", false, "Use the result stream")
	flagsGetResult.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most' or other available plugins")
	flagsGetResult.Var(&flags.outopts, "opt", "Options to pass to the output formatter")
	flagsGetResult.UintVar(&flags.timeout, "timeout", 60, "Timeout in seconds for result streaming")
	flagsGetResult.BoolVar(&flags.backlog, "backlog", false, "Request backlog when streaming")

	// limit
	flagsGetResult.UintVar(&flags.limit, "limit", 0, "Maximum amount of results to fetch")

	_ = flagsGetResult.Parse(args)

	return &flags
}
