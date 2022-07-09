/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/robert-kisteleki/goatapi"
	"golang.org/x/exp/slices"
)

// struct to receive/store command line args for anchor filtering
type findAnchorFlags struct {
	filterID     uint
	filterASN4   uint
	filterASN6   uint
	filterCC     string
	filterSearch string

	format string
	limit  uint
	count  bool
}

// Implementation of the "find anchor" subcommand. Parses command line flags
// and interacts with goatAPI to apply those filters+options to fetch results
func commandFindAnchor(args []string) {
	flags := parseFindAnchorArgs(args)
	filter, options := processFindAnchorFlags(flags)

	// counting only
	if _, ok := options["count"]; ok {
		count, err := filter.GetAnchorCount(flagVerbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(count)

		return
	}

	// most of the work is done by goatAPI
	anchors, err := filter.GetAnchors(flagVerbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	if flagVerbose {
		fmt.Printf("# %d anchors found\n", len(anchors))
	}

	// produce output; exact format depends on the "format" option
	ids := make([]string, 0)
	for _, anchor := range anchors {
		switch options["format"] {
		case "id":
			fmt.Println(anchor.ID)
		case "idcsv":
			ids = append(ids, fmt.Sprintf("%d", anchor.ID))
		case "some":
			fmt.Println(anchor.ShortString())
		case "most":
			fmt.Println(anchor.LongString())
		}
	}
	if options["format"] == "idcsv" {
		fmt.Println(strings.Join(ids, ","))
	}
}

// Process flags (filters & options), pass most of them on to goatAPI
// while doing sanity checks on values
func processFindAnchorFlags(flags *findAnchorFlags) (
	filter goatapi.AnchorFilter,
	options map[string]any,
) {
	options = make(map[string]any)
	filter = goatapi.NewAnchorFilter()

	// options

	formats := []string{"id", "idcsv", "some", "most"}
	if slices.Contains(formats, flags.format) {
		options["format"] = flags.format
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: invalid output format\n")
		os.Exit(1)
	}
	if flags.limit > 0 {
		filter.Limit(flags.limit)
	} else {
		fmt.Fprintf(os.Stderr, "ERROR: limit should be positive\n")
		os.Exit(1)
	}
	if flags.count {
		options["count"] = true
	}

	// filters

	if flags.filterID != 0 {
		filter.FilterID(flags.filterID)
		// with ID filtering other filters are irrelevant - so we can make a shortcut here
		return
	}

	if flags.filterCC != "" {
		// TODO: properly verify country code
		if len(flags.filterCC) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: invalid country code: %s\n", flags.filterCC)
			os.Exit(1)
		}
		filter.FilterCountry(strings.ToUpper(flags.filterCC))
	}

	if flags.filterASN4 > 0 {
		filter.FilterASN4(flags.filterASN4)
	}
	if flags.filterASN6 > 0 {
		filter.FilterASN6(flags.filterASN6)
	}

	if flags.filterSearch != "" {
		filter.FilterSearch(flags.filterSearch)
	}

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseFindAnchorArgs(args []string) *findAnchorFlags {
	var flags findAnchorFlags

	// filters
	flagsFindAnchor.UintVar(&flags.filterID, "id", 0, "A particular probe ID to fetch. If present, all other filters are ignored")
	flagsFindAnchor.UintVar(&flags.filterASN4, "asn4", 0, "Filter for probes with an IPv4 address announced by ths AS")
	flagsFindAnchor.UintVar(&flags.filterASN6, "asn6", 0, "Filter for probes with an IPv6 address announced by ths AS")
	flagsFindAnchor.StringVar(&flags.filterCC, "cc", "", "Filter for country code (2 letter ISO-3166 alpha-2)")
	flagsFindAnchor.StringVar(&flags.filterSearch, "search", "", "Filter for string in city, company or FQDN")

	// options
	flagsFindAnchor.BoolVar(&flags.count, "count", false, "Count only, don't show the actual results")
	flagsFindAnchor.StringVar(&flags.format, "format", "some", "Output contents: one of 'id', 'idcsv', 'some' (default), 'most'")

	// limit
	flagsFindAnchor.UintVar(&flags.limit, "limit", 100, "Maximum amount of anchors to retrieve")

	flagsFindAnchor.Parse(args)

	return &flags
}
