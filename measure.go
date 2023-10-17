/*
  (C) 2022, 2023 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"goatcli/output"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/robert-kisteleki/goatapi"
)

// struct to receive/store command line args for new measurements
type measureFlags struct {
	specOnly bool
	output   string
	outopts  multioption

	probetaginc string
	probetagexc string
	probecc     string
	probearea   string
	probeasn    string
	probeprefix string
	probelist   string
	probereuse  string
	ongoing     bool
	start       string
	stop        string
}

// Implementation of the "measure" subcommand. Parses command line flags
// and interacts with goatAPI to initiate new measurements
func commandMeasure(args []string) {
	flags := parseMeasureArgs(args)
	spec, options := processMeasureFlags(flags)
	formatter := options["output"].(string)

	if flags.specOnly {
		json, err := spec.GetApiJson()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(json))
		return
	}

	if !output.Verify(formatter) {
		fmt.Fprintf(os.Stderr, "ERROR: unknown output format '%s'\n", formatter)
		os.Exit(1)
	}

	if getApiKey("create_measurements") == nil {
		fmt.Fprintf(os.Stderr, "ERROR: you need to provide the API key create_measurement - please consult the config file\n")
		os.Exit(1)
	}

	// most of the work is done by goatAPI
	err := spec.Submit()
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
func processMeasureFlags(flags *measureFlags) (
	spec *goatapi.MeasurementSpec,
	options map[string]any,
) {
	options = make(map[string]any)
	spec = goatapi.NewMeasurementSpec()
	spec.Verbose(flagVerbose)

	spec.ApiKey(getApiKey("create_measurements"))

	// process probe sepcification(s)
	var probetaginc, probetagexc []string
	if flags.probetaginc != "" {
		probetaginc = strings.Split(flags.probetaginc, ",")
	}
	if flags.probetagexc != "" {
		probetagexc = strings.Split(flags.probetagexc, ",")
	}
	parseProbeSpec("cc", flags.probecc, spec, &probetaginc, &probetagexc)
	parseProbeSpec("area", flags.probearea, spec, &probetaginc, &probetagexc)
	parseProbeSpec("asn", flags.probeasn, spec, &probetaginc, &probetagexc)
	parseProbeSpec("prefix", flags.probeprefix, spec, &probetaginc, &probetagexc)
	parseProbeSpec("reuse", flags.probereuse, spec, &probetaginc, &probetagexc)
	parseProbeListSpec(flags.probelist, spec, &probetaginc, &probetagexc)

	// process timing
	spec.OneOff(!flags.ongoing)
	parseStartStop(spec, !flags.ongoing, flags.start, flags.stop)

	// process measurement specification
	spec.AddPing("ping1", "ping.ripe.net", 4, nil, nil)
	spec.AddPing("ping2", "ping.ripe.net", 6, nil, &goatapi.PingOptions{
		PacketSize:     999,
		IncludeProbeID: true,
	})

	// options
	options["output"] = flags.output

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseMeasureArgs(args []string) *measureFlags {
	var flags measureFlags

	// generic flags
	flagsMeasure.BoolVar(&flags.specOnly, "json", false, "Output the specification only, don't schedule the measurement")

	// probe selection
	flagsMeasure.StringVar(&flags.probetaginc, "probetaginc", "", "Probe tags to include (comma separated list)")
	flagsMeasure.StringVar(&flags.probetagexc, "probetagexc", "", "Probe tags to exclude (comma separated list)")
	flagsMeasure.StringVar(&flags.probecc, "probecc", "", "Probes to select from country (comma separated list of amount@CC)")
	flagsMeasure.StringVar(&flags.probearea, "probearea", "", "Probes to select from area (comma separated list of amount@area, area can be ww/west/nc/sc/ne/se)")
	flagsMeasure.StringVar(&flags.probeasn, "probeasn", "", "Probes to select from an ASN (comma separated list of amount@ASN)")
	flagsMeasure.StringVar(&flags.probeprefix, "probeprefix", "", "Probes to select from a prefix (comma separated list of amount@prefix)")
	flagsMeasure.StringVar(&flags.probelist, "probelist", "", "Probes to use provided as a comma separated list")
	flagsMeasure.StringVar(&flags.probereuse, "probereuse", "", "Probes to reuse from a previous measurement as amount@msmID")

	// timing
	flagsMeasure.BoolVar(&flags.ongoing, "ongoing", false, "Schedule an ongoing measurement instead of a one-off")
	flagsMeasure.StringVar(&flags.start, "start", "", "When to start this measurement")
	flagsMeasure.StringVar(&flags.stop, "stop", "", "When to stop this measurement (if it's ongoing)")

	// options
	flagsMeasure.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most'")
	flagsMeasure.Var(&flags.outopts, "opt", "Options to pass to the output formatter")

	flagsMeasure.Parse(args)

	return &flags
}

// parse probe spec as a list of amount@spec
func parseProbeSpec(
	spectype string,
	from string,
	spec *goatapi.MeasurementSpec,
	probetaginc, probetagexc *[]string,
) {
	if from == "" {
		return
	}

	list := strings.Split(from, ",")
	for _, item := range list {
		split := strings.Split(item, "@")
		if len(split) != 2 {
			fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe %s spec: '%s'\n", spectype, item)
			os.Exit(1)
		}
		n := 0
		var err error
		if split[1] == "all" {
			n = -1
		} else {
			n, err = strconv.Atoi(split[0])
			if err != nil || n <= 0 {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe %s amount: '%s'\n", spectype, split[0])
				os.Exit(1)
			}
		}
		switch spectype {
		case "cc":
			if len(split[1]) != 2 { // TODO: proper CC validation
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe CC spec: invalid CC (in %s)\n", split[1])
				os.Exit(1)
			}
			spec.AddProbesCountryWithTags(split[1], n, probetaginc, probetagexc)
		case "asn":
			specval, err := strconv.Atoi(split[1])
			if err != nil || specval <= 0 {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe ASN spec: '%s'\n", split[1])
				os.Exit(1)
			}
			spec.AddProbesAsnWithTags(uint(specval), n, probetaginc, probetagexc)
		case "prefix":
			prefix, err := netip.ParsePrefix(split[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe prefix spec: '%s'\n", split[1])
				os.Exit(1)
			}
			spec.AddProbesPrefixWithTags(prefix, n, probetaginc, probetagexc)
		case "reuse":
			msmid, err := strconv.Atoi(split[1])
			if err != nil || msmid <= 1000000 {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe reuse measurement ID: '%s'\n", split[1])
				os.Exit(1)
			}
			spec.AddProbesReuseWithTags(uint(msmid), n, probetaginc, probetagexc)
		case "area":
			var areas map[string]string = map[string]string{
				"ww":   "WW",
				"west": "West",
				"nc":   "North-Central",
				"sc":   "South-Central",
				"ne":   "North-East",
				"se":   "South-East",
			}
			area, ok := areas[split[1]]
			if !ok {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe area spec: unknown area '%s'\n", split[1])
				os.Exit(1)
			}
			spec.AddProbesAreaWithTags(area, n, probetaginc, probetagexc)
		}
	}
}

// parse probe list spec as a list of probe IDs
func parseProbeListSpec(
	from string,
	spec *goatapi.MeasurementSpec,
	probetaginc, probetagexc *[]string,
) {
	if from == "" {
		return
	}

	list := make([]uint, 0)
	plist := strings.Split(from, ",")
	for _, pid := range plist {
		n, err := strconv.Atoi(pid)
		if err != nil || n <= 0 {
			fmt.Fprintf(os.Stderr, "ERROR: invalid probe ID %s\n", pid)
			os.Exit(1)
		}
		list = append(list, uint(n))
	}
	spec.AddProbesListWithTags(list, probetaginc, probetagexc)
}

func parseStartStop(
	spec *goatapi.MeasurementSpec,
	oneoff bool,
	start string,
	stop string,
) {
	starttime, starterr := parseTimeAlternatives(start)
	stoptime, stoperr := parseTimeAlternatives(stop)

	if oneoff && stoperr == nil {
		fmt.Fprintf(os.Stderr, "ERROR: one-offs cannot have a stop time\n")
		os.Exit(1)
	}
	if starterr == nil && starttime.Unix() <= time.Now().Unix() {
		fmt.Fprintf(os.Stderr, "ERROR: start time cannot be in the past\n")
		os.Exit(1)
	}
	if stoperr == nil && stoptime.Unix() <= time.Now().Unix() {
		fmt.Fprintf(os.Stderr, "ERROR: stop time cannot be in the past\n")
		os.Exit(1)
	}
	if starterr == nil && stoperr == nil && starttime.Unix() >= stoptime.Unix() {
		fmt.Fprintf(os.Stderr, "ERROR: start time has to be before stop time\n")
		os.Exit(1)
	}

	if starterr == nil {
		spec.Start(starttime)
	}
	if stoperr == nil {
		spec.Stop(stoptime)
	}
}
