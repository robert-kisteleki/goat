/*
  (C) Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"fmt"
	"net/netip"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/robert-kisteleki/goat"
	"github.com/robert-kisteleki/goat/cmd/goat/output"
)

// struct to receive/store command line args for new measurements
type measureFlags struct {
	specOnly bool
	output   string
	outopts  multioption
	save     string
	result   bool
	noresult bool
	timeout  uint

	// probe options
	probetaginc string
	probetagexc string
	probecc     string
	probearea   string
	probeasn    string
	probeprefix string
	probelist   string
	probereuse  string

	// stop or modify probes of a measurement
	msmstop   uint
	msmadd    uint
	msmremove uint

	// timing options
	periodic  bool
	starttime string
	endtime   string

	// common measurement options
	msmdescr          string
	msmtarget         string
	msmaf             uint
	msminterval       uint
	msmspread         uint
	msmresolveonprobe bool
	msmskipdnscheck   bool
	msmtags           string
	msmdnsrelookup    uint
	msmautotopup      bool
	msmautotopupdays  uint
	msmautotopupsim   float64

	// measurement types
	msmping  bool
	msmtrace bool
	msmdns   bool
	msmtls   bool
	msmhttp  bool
	msmntp   bool

	// type specific measurement options
	msmoptname     string // DNS: name to look up
	msmoptmethod   string // HTTP: method (GET, POST, HEAD)
	msmoptparis    uint   // TRACE: paris ID
	msmoptprotocol string // TRACE: protocol (UDP, TCP, ICMP), DNS: protocol (UDP, TCP)
	msmoptminhop   uint   // TRACE: first hop
	msmoptmaxhop   uint   // TRACE: last hop
	msmoptnsid     bool   // DNS: set NSID
	msmoptqbuf     bool   // DNS: store qbuf
	msmoptabuf     bool   // DNS: store abuf
	msmoptrd       bool   // DNS: RD bit
	msmoptdo       bool   // DNS: DO bit
	msmoptcd       bool   // DNS: CD bit
	msmoptretry    uint   // DNS: retry count
	msmoptclass    string // DNS: class (IN, CHAOS)
	msmopttype     string // DNS: type (A, AAAA, NS, CNAME, ...)
	msmoptport     uint   // TLS, HTTP: port number
	msmoptsni      string // TLS: SNI
	msmoptversion  string // HTTP: version (1.0, 1.1)
	msmopttiming1  bool   // HTTP: extended timing
	msmopttiming2  bool   // HTTP: more extended timing

	totalProbes int // how many probes were asked for
}

// Implementation of the "measure" subcommand. Parses command line flags
// and interacts with goatAPI to initiate new measurements or stop or update existing ones
func commandMeasure(args []string) {
	flags := parseMeasureArgs(args)
	spec, options := processMeasureFlags(flags)

	switch {
	case flags.msmstop != 0:
		if getApiKey("stop_measurements") == nil {
			fmt.Fprintf(os.Stderr, "ERROR: you need to provide the API key stop_measurements - please consult the config file\n")
			os.Exit(1)
		}
		spec.ApiKey(getApiKey("stop_measurements"))

		err := spec.Stop(flags.msmstop)
		if err == nil {
			fmt.Printf("Measurement %d has been stopped.\n", flags.msmstop)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR while trying to stop measurement %d: %v\n", flags.msmstop, err)
			os.Exit(1)
		}

		return

	case flags.msmadd != 0 && flags.msmremove != 0:
		fmt.Fprintf(os.Stderr, "ERROR: you can't ask for adding and removing probes at the same time\n")
		os.Exit(1)

	case flags.msmremove != 0:
		if flags.probearea != "" ||
			flags.probeasn != "" ||
			flags.probecc != "" ||
			flags.probeprefix != "" ||
			flags.probereuse != "" ||
			flags.probelist == "" {
			fmt.Fprintf(os.Stderr, "ERROR: probe removal only supports an explicit list of probes (--probelist flag)\n")
			os.Exit(1)
		}
		if getApiKey("update_measurements") == nil {
			fmt.Fprintf(os.Stderr, "ERROR: you need to provide the API key update_measurements - please consult the config file\n")
			os.Exit(1)
		}
		spec.ApiKey(getApiKey("update_measurements"))
		ids, err := spec.ParticipationRequest(flags.msmremove, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while trying to update measurement %d: %v\n", flags.msmremove, err)
			os.Exit(1)
		}
		if flagVerbose {
			fmt.Printf("# Request IDs: %v\n", ids)
		}
		fmt.Println("OK")
		return

	case flags.msmadd != 0:
		if getApiKey("update_measurements") == nil {
			fmt.Fprintf(os.Stderr, "ERROR: you need to provide the API key update_measurements - please consult the config file\n")
			os.Exit(1)
		}
		spec.ApiKey(getApiKey("update_measurements"))
		ids, err := spec.ParticipationRequest(flags.msmadd, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR while trying to update measurement %d: %v\n", flags.msmadd, err)
			os.Exit(1)
		}
		if flagVerbose {
			fmt.Printf("# Request IDs: %v\n", ids)
		}
		fmt.Println("OK")
		return
	}

	formatter := options["output"].(string)

	if flags.msmaf != 4 && flags.msmaf != 6 {
		fmt.Fprintf(os.Stderr, "ERROR: invalid address family, it should be 4 or 6\n")
		os.Exit(1)
	}

	if flags.specOnly {
		json, err := spec.GetApiJson()
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(json))
		return
	}

	if !output.Verify(formatter, flagsToOutFormat(flags)) {
		fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported output format '%s' for '%s'\n", flagsToOutFormat(flags), formatter)
		os.Exit(1)
	}

	if getApiKey("create_measurements") == nil {
		fmt.Fprintf(os.Stderr, "ERROR: you need to provide the API key create_measurements - please consult the config file\n")
		os.Exit(1)
	}
	spec.ApiKey(getApiKey("create_measurements"))

	// most of the work is done by goatAPI
	msmlist, err := spec.Schedule()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	if flagVerbose {
		fmt.Println("# Measurement ID:", msmlist[0])
	} else if !flags.result {
		fmt.Println(msmlist[0])
	}

	// the measurement is ready - tune into the result stream if the user wanted that
	if flags.result {
		// prepare
		rflags := resultFlags{
			stream:       true,
			filterID:     msmlist[0],
			saveFileName: flags.save,
			saveAll:      true,
			output:       flags.output,
			outopts:      flags.outopts,
			limit:        0,
			timeout:      flags.timeout,
			backlog:      true,
		}
		if !flags.periodic {
			rflags.limit = uint(flags.totalProbes)
		}
		if flags.msmdns {
			_ = rflags.outopts.Set("type:" + flags.msmopttype)
		}

		// most of the work is done by the implementation of the result streaming feature
		commandResultFromFlags(&rflags)
	}
}

// Process flags (options), pass most of them on to goatAPI
// while doing sanity checks on values
func processMeasureFlags(flags *measureFlags) (
	spec *goat.MeasurementSpec,
	options map[string]any,
) {
	options = make(map[string]any)
	spec = goat.NewMeasurementSpec()
	spec.Verbose(flagVerbose)

	// process probe sepcification(s)
	var probetaginc, probetagexc []string
	if flags.probetaginc != "" {
		probetaginc = strings.Split(flags.probetaginc, ",")
	}
	if flags.probetagexc != "" {
		probetagexc = strings.Split(flags.probetagexc, ",")
	}

	// apply defaults from the config file if nothing was specified
	if flags.probecc == "" &&
		flags.probearea == "" &&
		flags.probeasn == "" &&
		flags.probeprefix == "" &&
		flags.probereuse == "" &&
		flags.probelist == "" {

		if flags.probecc == "" {
			flags.probecc = getProbeSpecDefault("cc")
		}
		if flags.probearea == "" {
			flags.probearea = getProbeSpecDefault("area")
		}
		if flags.probeasn == "" {
			flags.probeasn = getProbeSpecDefault("asn")
		}
		if flags.probeprefix == "" {
			flags.probeprefix = getProbeSpecDefault("prefix")
		}
		if flags.probereuse == "" {
			flags.probereuse = getProbeSpecDefault("reuse")
		}
		if flags.probelist == "" {
			flags.probelist = getProbeSpecDefault("list")
		}

		// last resort: add 10 probes world wide
		if flags.probecc == "" &&
			flags.probearea == "" &&
			flags.probeasn == "" &&
			flags.probeprefix == "" &&
			flags.probereuse == "" &&
			flags.probelist == "" {
			parseProbeSpec("area", "10@ww", spec, &probetaginc, &probetagexc, &flags.totalProbes)
		}
	}

	// parse probe specifications
	parseProbeSpec("cc", flags.probecc, spec, &probetaginc, &probetagexc, &flags.totalProbes)
	parseProbeSpec("area", flags.probearea, spec, &probetaginc, &probetagexc, &flags.totalProbes)
	parseProbeSpec("asn", flags.probeasn, spec, &probetaginc, &probetagexc, &flags.totalProbes)
	parseProbeSpec("prefix", flags.probeprefix, spec, &probetaginc, &probetagexc, &flags.totalProbes)
	parseProbeSpec("reuse", flags.probereuse, spec, &probetaginc, &probetagexc, &flags.totalProbes)
	parseProbeListSpec(flags.probelist, spec, &probetaginc, &probetagexc, &flags.totalProbes)

	// process timing
	spec.OneOff(!flags.periodic)
	parseStartStop(spec, !flags.periodic, flags.starttime, flags.endtime)

	// process measurement specification
	parseMeasurementPing(flags, spec)
	parseMeasurementTraceroute(flags, spec)
	parseMeasurementDns(flags, spec)
	parseMeasurementTls(flags, spec)
	parseMeasurementHttp(flags, spec)
	parseMeasurementNtp(flags, spec)

	// options
	options["output"] = flags.output

	return
}

// Define and parse command line args for this subcommand using the flags package
func parseMeasureArgs(args []string) *measureFlags {
	var flags measureFlags

	// special cases: stop a meeasurement, add or remove probes
	flagsMeasure.UintVar(&flags.msmstop, "stop", 0, "Stop a particular measurement")
	flagsMeasure.UintVar(&flags.msmadd, "add", 0, "Add probes to a particular measurement")
	flagsMeasure.UintVar(&flags.msmremove, "remove", 0, "Remove probes from a particular measurement")

	// generic flags
	flagsMeasure.BoolVar(&flags.specOnly, "json", false, "Output the specification only, don't schedule the measurement")
	flagsMeasure.BoolVar(&flags.result, "result", false, "Immediately tune in to the result stream. By default true for one-offs, false for periodic ones.")
	flagsMeasure.BoolVar(&flags.noresult, "noresult", false, "Don't tune in to the result stream, even for a one-off.")

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
	flagsMeasure.BoolVar(&flags.periodic, "periodic", false, "Schedule a periodic measurement instead of a one-off")
	flagsMeasure.StringVar(&flags.starttime, "start", "", "When to start this measurement")
	flagsMeasure.StringVar(&flags.endtime, "end", "", "When to end this measurement (if it's ongoing)")

	// measurement types
	flagsMeasure.BoolVar(&flags.msmping, "ping", false, "Schedule a ping measurement")
	flagsMeasure.BoolVar(&flags.msmtrace, "trace", false, "Schedule a traceroute measurement")
	flagsMeasure.BoolVar(&flags.msmdns, "dns", false, "Schedule a DNS measurement")
	flagsMeasure.BoolVar(&flags.msmtls, "tls", false, "Schedule a TLS measurement")
	flagsMeasure.BoolVar(&flags.msmhttp, "http", false, "Schedule a HTTP measurement")
	flagsMeasure.BoolVar(&flags.msmntp, "ntp", false, "Schedule a NTP measurement")

	// measurement common flags
	flagsMeasure.UintVar(&flags.msmaf, "af", 4, "Address family: 4 for IPv4 or 6 for IPv6")
	flagsMeasure.StringVar(&flags.msmtarget, "target", "", "Target of the measurement")
	flagsMeasure.BoolVar(&flags.msmresolveonprobe, "resolveonprobe", true, "The probe should do the DNS resolution")
	flagsMeasure.BoolVar(&flags.msmskipdnscheck, "skipdnscheck", false, "Skip DNS check upon creation")
	flagsMeasure.UintVar(&flags.msminterval, "interval", 0, "Interval for an ongoing measurement")
	flagsMeasure.UintVar(&flags.msmspread, "spread", 0, "Spread for an ongoing measurement")
	flagsMeasure.StringVar(&flags.msmtags, "tags", "", "Tags for a measurement")
	flagsMeasure.UintVar(&flags.msmdnsrelookup, "relookup", 0, "How often to re-lookup the IP of a DNS based target when not using resolve-on-probe (in hours, mininum 24)")
	flagsMeasure.BoolVar(&flags.msmautotopup, "topup", false, "Enable auto-topup of probes involved in the measurement")
	flagsMeasure.UintVar(&flags.msmautotopupdays, "topupdays", 0, "Try to replace a probe after these many days of disconnect (1-30, default 7)")
	flagsMeasure.Float64Var(&flags.msmautotopupsim, "topupsim", 0.0, "Similarity metric for replacement probes (0.0-1.0, default 0.5)")

	// measurement type specific options
	flagsMeasure.UintVar(&flags.msmoptparis, "paris", 16, "TRACE: paris ID")
	flagsMeasure.StringVar(&flags.msmoptprotocol, "proto", "UDP", "TRACE, DNS: protocol to use (UDP, TCP, ICMP or UDP, TCP)")
	flagsMeasure.UintVar(&flags.msmoptminhop, "minhop", 1, "TRACE: first hop")
	flagsMeasure.UintVar(&flags.msmoptmaxhop, "maxhop", 32, "TRACE: last hop")
	flagsMeasure.StringVar(&flags.msmoptname, "name", "", "DNS: name to look up")
	flagsMeasure.BoolVar(&flags.msmoptnsid, "nsid", false, "DNS: ask for NSID")
	flagsMeasure.BoolVar(&flags.msmoptqbuf, "qbuf", false, "DNS: ask for qbuf")
	flagsMeasure.BoolVar(&flags.msmoptabuf, "abuf", false, "DNS: ask for abuf")
	flagsMeasure.BoolVar(&flags.msmoptrd, "rd", false, "DNS: set RD bit")
	flagsMeasure.BoolVar(&flags.msmoptdo, "do", false, "DNS: set DO bit")
	flagsMeasure.BoolVar(&flags.msmoptcd, "cd", false, "DNS: set CD bit")
	flagsMeasure.UintVar(&flags.msmoptretry, "retry", 0, "DNS: retry count")
	flagsMeasure.StringVar(&flags.msmoptclass, "class", "IN", "DNS: query class (IN, CHAOS)")
	flagsMeasure.StringVar(&flags.msmopttype, "type", "A", "DNS: query type (A, AAAA, NS, CNAME, ...)")
	flagsMeasure.StringVar(&flags.msmoptmethod, "method", "HEAD", "HTTP: method to use (HEAD, GET, POST)")
	flagsMeasure.UintVar(&flags.msmoptport, "port", 0, "TLS: port number")
	flagsMeasure.StringVar(&flags.msmoptsni, "sni", "", "TLS: SNI to use. Defaults to target name")
	flagsMeasure.StringVar(&flags.msmoptversion, "version", "1.0", "HTTP: version to use (1.0, 1.1)")
	flagsMeasure.BoolVar(&flags.msmopttiming1, "time1", false, "HTTP: extended timing")
	flagsMeasure.BoolVar(&flags.msmopttiming2, "time2", false, "HTTP: more extended timing")

	// options
	flagsMeasure.StringVar(&flags.output, "output", "some", "Output format: 'some' or 'most'")
	flagsMeasure.Var(&flags.outopts, "opt", "Options to pass to the output formatter")
	flagsMeasure.StringVar(&flags.save, "save", "", "Save results to this file")
	flagsMeasure.UintVar(&flags.timeout, "timeout", 60, "Timeout in seconds for result streaming")

	_ = flagsMeasure.Parse(args)

	// force uppercase for some
	flags.msmoptclass = strings.ToUpper(flags.msmoptclass)
	flags.msmopttype = strings.ToUpper(flags.msmopttype)
	flags.msmoptmethod = strings.ToUpper(flags.msmoptmethod)
	flags.msmoptprotocol = strings.ToUpper(flags.msmoptprotocol)

	// one cannot have and not have results at the same time
	if flags.result && flags.noresult {
		fmt.Fprintf(os.Stderr, "ERROR: please decide if you want result streaming or not\n")
		os.Exit(1)
	}
	// for one-offs turn on result streaming unless it's explicity not wanted
	if !flags.periodic && !flags.noresult {
		flags.result = true
	}
	// for periodics only turn result streaming if it's explicitly wanted
	if flags.periodic && flags.result {
		flags.result = true
	}

	return &flags
}

// parse probe spec as a list of amount@spec
func parseProbeSpec(
	spectype string,
	from string,
	spec *goat.MeasurementSpec,
	probetaginc, probetagexc *[]string,
	totalProbes *int,
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
			err = spec.AddProbesCountryWithTags(split[1], n, probetaginc, probetagexc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid country code '%s'\n", split[1])
				os.Exit(1)
			}
		case "asn":
			specval, err := strconv.Atoi(split[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe ASN spec: '%s'\n", split[1])
				os.Exit(1)
			}
			err = spec.AddProbesAsnWithTags(uint(specval), n, probetaginc, probetagexc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: invalid ASN '%s'\n", split[1])
				os.Exit(1)
			}
		case "prefix":
			prefix, err := netip.ParsePrefix(split[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe prefix spec: '%s'\n", split[1])
				os.Exit(1)
			}
			_ = spec.AddProbesPrefixWithTags(prefix, n, probetaginc, probetagexc)
		case "reuse":
			msmid, err := strconv.Atoi(split[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: unable to parse probe reuse measurement ID: '%s'\n", split[1])
				os.Exit(1)
			}
			err = spec.AddProbesReuseWithTags(uint(msmid), n, probetaginc, probetagexc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: measurement ID '%s'\n", split[1])
				os.Exit(1)
			}
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
			_ = spec.AddProbesAreaWithTags(area, n, probetaginc, probetagexc)
		}
		*totalProbes += n
	}
}

// parse probe list spec as a list of probe IDs
func parseProbeListSpec(
	from string,
	spec *goat.MeasurementSpec,
	probetaginc, probetagexc *[]string,
	totalProbes *int,
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
	err := spec.AddProbesListWithTags(list, probetaginc, probetagexc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: unable to use probe list '%s'\n", from)
		os.Exit(1)
	}
	*totalProbes += len(list)
}

func parseStartStop(
	spec *goat.MeasurementSpec,
	oneoff bool,
	start string,
	end string,
) {
	starttime, starterr := parseTimeAlternatives(start)
	endtime, enderr := parseTimeAlternatives(end)

	if oneoff && enderr == nil {
		fmt.Fprintf(os.Stderr, "ERROR: one-offs cannot have a stop time\n")
		os.Exit(1)
	}
	if starterr == nil && starttime.Unix() <= time.Now().Unix() {
		fmt.Fprintf(os.Stderr, "ERROR: start time cannot be in the past\n")
		os.Exit(1)
	}
	if enderr == nil && endtime.Unix() <= time.Now().Unix() {
		fmt.Fprintf(os.Stderr, "ERROR: stop time cannot be in the past\n")
		os.Exit(1)
	}
	if starterr == nil && enderr == nil && starttime.Unix() >= endtime.Unix() {
		fmt.Fprintf(os.Stderr, "ERROR: start time has to be before stop time\n")
		os.Exit(1)
	}

	if starterr == nil {
		spec.StartTime(starttime)
	}
	if enderr == nil {
		spec.EndTime(endtime)
	}
}

func processBaseOptions(flags *measureFlags) *goat.BaseOptions {
	opts := goat.BaseOptions{}
	if flags.msminterval != 0 {
		opts.Interval = flags.msminterval
	}
	if flags.msmspread != 0 {
		opts.Spread = flags.msmspread
	}
	opts.ResolveOnProbe = flags.msmresolveonprobe
	opts.SkipDNSCheck = flags.msmskipdnscheck
	if flags.msmtags != "" {
		opts.Tags = strings.Split(flags.msmtags, ",")
	}
	if flags.msmdnsrelookup > 0 && flags.msmdnsrelookup < 24 {
		fmt.Fprintf(os.Stderr, "ERROR: DNS re-lookup time has to be >=24 (hours)\n")
		os.Exit(1)
	}
	opts.DnsReLookup = flags.msmdnsrelookup
	opts.AutoTopup = flags.msmautotopup
	if flags.msmautotopupdays > 30 {
		fmt.Fprintf(os.Stderr, "ERROR: auto-topup days needs to be < 30 (days)\n")
		os.Exit(1)
	}
	opts.AutoTopupDays = flags.msmautotopupdays
	if flags.msmautotopupsim < 0.0 || flags.msmautotopupsim > 1.0 {
		fmt.Fprintf(os.Stderr, "ERROR: auto-topup similartity limit needs to be between 0.0-1.0\n")
		os.Exit(1)
	}
	opts.AutoTopupSimilarity = flags.msmautotopupsim
	return &opts
}

func parseMeasurementPing(
	flags *measureFlags,
	spec *goat.MeasurementSpec,
) {
	if !flags.msmping {
		return
	}

	if flags.msmtarget == "" {
		fmt.Fprintf(os.Stderr, "ERROR: a target should be specified\n")
		os.Exit(1)
	}

	descr := flags.msmdescr
	if descr == "" {
		descr = fmt.Sprintf("Ping measurement to %s", flags.msmtarget)
	}

	baseopts := processBaseOptions(flags)
	err := spec.AddPing(descr, flags.msmtarget, flags.msmaf, baseopts, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func parseMeasurementTraceroute(
	flags *measureFlags,
	spec *goat.MeasurementSpec,
) {
	if !flags.msmtrace {
		return
	}

	if flags.msmtarget == "" {
		fmt.Fprintf(os.Stderr, "ERROR: a target should be specified\n")
		os.Exit(1)
	}

	descr := flags.msmdescr
	if descr == "" {
		descr = fmt.Sprintf("Traceroute measurement to %s", flags.msmtarget)
	}
	baseopts := processBaseOptions(flags)
	traceopts := goat.TraceOptions{}
	if flags.msmoptparis != 16 {
		traceopts.ParisId = flags.msmoptparis
	}
	if flags.msmoptprotocol != "UDP" {
		if !slices.Contains(goat.TraceProtocols, flags.msmoptprotocol) {
			fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported TRACE protocol: '%s'\n", flags.msmoptprotocol)
			os.Exit(1)
		}
		traceopts.Protocol = flags.msmoptprotocol
	}
	if flags.msmoptminhop == 0 ||
		flags.msmoptminhop > 128 ||
		flags.msmoptmaxhop == 0 ||
		flags.msmoptmaxhop > 128 {
		fmt.Fprintf(os.Stderr, "ERROR: minhop and maxhop should be between 1 and 128\n")
		os.Exit(1)
	}
	if flags.msmoptminhop > flags.msmoptmaxhop {
		fmt.Fprintf(os.Stderr, "ERROR: minhop should not be more than maxhop\n")
		os.Exit(1)
	}
	if flags.msmoptminhop != 1 {
		traceopts.FirstHop = flags.msmoptminhop
	}
	if flags.msmoptmaxhop != 32 {
		traceopts.LastHop = flags.msmoptmaxhop
	}
	err := spec.AddTrace(descr, flags.msmtarget, flags.msmaf, baseopts, &traceopts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func parseMeasurementDns(
	flags *measureFlags,
	spec *goat.MeasurementSpec,
) {
	if !flags.msmdns {
		return
	}

	if flags.msmoptname == "" {
		fmt.Fprintf(os.Stderr, "ERROR: a name to be looked up should be specified\n")
		os.Exit(1)
	}

	descr := flags.msmdescr
	if descr == "" {
		descr = fmt.Sprintf("DNS lookup of %s", flags.msmoptname)
		if flags.msmtarget != "" {
			descr += " @" + flags.msmtarget
		}
	}
	baseopts := processBaseOptions(flags)
	baseopts.ResolveOnProbe = false
	dnsopts := goat.DnsOptions{}
	dnsopts.Argument = flags.msmoptname
	dnsopts.UseResolver = flags.msmtarget == ""
	dnsopts.Nsid = flags.msmoptnsid

	dnsopts.IncludeQbuf = flags.msmoptqbuf
	dnsopts.IncludeAbuf = flags.msmoptabuf
	dnsopts.SetRd = flags.msmoptrd
	dnsopts.SetDo = flags.msmoptdo
	dnsopts.SetCd = flags.msmoptcd
	if flags.msmoptprotocol != "UDP" {
		if !slices.Contains(goat.DnsProtocols, flags.msmoptprotocol) {
			fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported DNS protocol: '%s'\n", flags.msmoptprotocol)
			os.Exit(1)
		}
		dnsopts.Protocol = flags.msmoptprotocol
	}
	if flags.msmoptclass != "IN" {
		if !slices.Contains(goat.DnsClasses, flags.msmoptclass) {
			fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported DNS class: '%s'\n", flags.msmoptclass)
			os.Exit(1)
		}
		dnsopts.Class = flags.msmoptclass
	}
	if flags.msmopttype != "A" {
		if !slices.Contains(goat.DnsTypes, flags.msmopttype) {
			fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported DNS class: '%s'\n", flags.msmopttype)
			os.Exit(1)
		}
		dnsopts.Type = flags.msmopttype
	}
	dnsopts.UseMacros = strings.Contains(flags.msmoptname, "$")
	if flags.msmoptretry != 0 {
		dnsopts.Retries = flags.msmoptretry
	}
	err := spec.AddDns(descr, flags.msmtarget, flags.msmaf, baseopts, &dnsopts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func parseMeasurementTls(
	flags *measureFlags,
	spec *goat.MeasurementSpec,
) {
	if !flags.msmtls {
		return
	}

	if flags.msmtarget == "" {
		fmt.Fprintf(os.Stderr, "ERROR: a target should be specified\n")
		os.Exit(1)
	}

	descr := flags.msmdescr
	if descr == "" {
		descr = fmt.Sprintf("TLS measurement to %s", flags.msmtarget)
	}
	baseopts := processBaseOptions(flags)
	tlsopts := goat.TlsOptions{}
	if flags.msmoptport != 0 {
		tlsopts.Port = flags.msmoptport
	}
	if flags.msmoptsni == "" {
		tlsopts.Sni = flags.msmtarget
	} else {
		tlsopts.Sni = flags.msmoptsni
	}
	err := spec.AddTls(descr, flags.msmtarget, flags.msmaf, baseopts, &tlsopts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func parseMeasurementHttp(
	flags *measureFlags,
	spec *goat.MeasurementSpec,
) {
	if !flags.msmhttp {
		return
	}

	if flags.msmtarget == "" {
		fmt.Fprintf(os.Stderr, "ERROR: a target should be specified\n")
		os.Exit(1)
	}

	target, http := strings.CutPrefix(flags.msmtarget, "http://")
	if !http {
		fmt.Fprintf(os.Stderr, "ERROR: a target should start with http://\n")
		os.Exit(1)
	}
	server, path, _ := strings.Cut(target, "/")
	path, args, _ := strings.Cut("/"+path, "?")

	descr := flags.msmdescr
	if descr == "" {
		descr = fmt.Sprintf("HTTP measurement to %s", server)
	}
	baseopts := processBaseOptions(flags)
	httpopts := goat.HttpOptions{}
	httpopts.Path = path
	httpopts.Query = args
	if flags.msmoptmethod != "" {
		if !slices.Contains(goat.HttpMethods, flags.msmoptmethod) {
			fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported HTTP method: '%s'\n", flags.msmoptmethod)
			os.Exit(1)
		}
		httpopts.Method = flags.msmoptmethod
	}
	if flags.msmoptversion != "1.0" {
		if !slices.Contains(goat.HttpVersions, flags.msmoptversion) {
			fmt.Fprintf(os.Stderr, "ERROR: unknown or unsupported HTTP version: '%s'\n", flags.msmoptversion)
			os.Exit(1)
		}
		httpopts.Version = flags.msmoptversion
	}
	if flags.msmoptport != 0 {
		httpopts.Port = flags.msmoptport
	}
	httpopts.ExtendedTiming = flags.msmopttiming1
	httpopts.MoreExtendedTiming = flags.msmopttiming2
	err := spec.AddHttp(descr, server, flags.msmaf, baseopts, &httpopts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func parseMeasurementNtp(
	flags *measureFlags,
	spec *goat.MeasurementSpec,
) {
	if !flags.msmntp {
		return
	}

	if flags.msmtarget == "" {
		fmt.Fprintf(os.Stderr, "ERROR: a target should be specified\n")
		os.Exit(1)
	}

	descr := flags.msmdescr
	if descr == "" {
		descr = fmt.Sprintf("NTP measurement to %s", flags.msmtarget)
	}
	baseopts := processBaseOptions(flags)
	err := spec.AddNtp(descr, flags.msmtarget, flags.msmaf, baseopts, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func flagsToOutFormat(flags *measureFlags) string {
	switch {
	case flags.msmdns:
		return "dns"
	case flags.msmhttp:
		return "http"
	case flags.msmping:
		return "ping"
	case flags.msmntp:
		return "ntp"
	case flags.msmtls:
		return "tls"
	case flags.msmtrace:
		return "trace"
	default:
		return ""
	}

}
