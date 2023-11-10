# goatCLI changelog

## next

* FIX: native trace out didn't show hop errors nicely
* NEW: add shortcuts for measurement types, e.g. `goatcli ping ping.me` just works
* NEW: pre-flight checks for output format
* NEW: add `type:TYPE` option to the dnsstat output formatter

## v0.5.0

* CHANGED: renamed `--stop` to `--end` to indicate preferred end time of periodic measurements
* NEW: add support to stop existing measurements
* NEW: add support to add or remove probes to/from existing measurements (`--add` and `--remove`)

## v0.4.1

* NEW: measurement scheduling now immediately tunes in to the result stream to show results for one-offs
* CHANGED: renamed the `--ongoing` flag to `--periodic` for measurement scheduling
* CHANGED: various enhancements to output formats and accommodate goatapi result changes

## v0.4.0

* NEW: add support for measurement status checks (`status` subcommand)
* NEW: add support for measurement scheduling (`measure` subcommand)
  * support all measurement types with lots of options and sane defaults
  * support all probe selection criteria
  * support default probe selection specification in config file

## v0.3.0

* NEW: add option to stream results from the result stream instead of the data API
* NEW: add option to save all results obtained (from data API, stream API or even
  from a file) to a file
* NEW: added a dummy output formatter ("none")
* CHANGED: change default limit on downloading results to 0 (no limit)

## v0.2.2

* NEW: output formatters can now accept options (`--opt X`)
* NEW: output annotator: output formatters can get probe ASN, CC and prefix
  information. Probe metadata is cached across runs, for a week.
* CHANGED: output formatters now have a `start()` method to signal processing
  of new (batch of) results; this maybe even multiple times - just like with
  `finish()` from now on. `setup()` is only called once, before any results are
  processed.
* NEW: `--opt ccstat` and `--opt asnstat` options for the `dnsstats` formatter

## v0.2.1

* CHANGED: internal changes on how output formatters work
* CHANGED: output filters are now used for probes, anchors and measurements too
* CHANGED: adapt code to goatAPI v0.2.1 async results

## v0.2.0

* NEW: support for downloading results of a measurement
  * with start time, stop time and probe filters
  * with option to get "latest" only
* NEW: support for processing results from an already downloaded file
* NEW: preliminary support for output processors
  * some, most: basic properties of the results
  * native: a native-looking output (i.e. similar to ping, traceroute, ...)
  * dnsstat: a simple DNS result summariser
* CHANGED: minor verbose output format changes
* CHANGED: output for "some" and "most" moved here from goatAPI

## v0.1.0

* support listing probes, anchors, measurements with virtually all filtering options
* support counting items, retrieveing all matching ones or just a specific one by ID
* support for a (primitive) configuration file (~/.config/goat.ini) and command line flags
* support for "list_measurements" API key via the config file
