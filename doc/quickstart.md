# goatcli - Go (RIPE) Atlas Tools - CLI Quick Start Guide

## Finding Probes

### Count Probes Matching Some Criteria

```sh
$ ./goatcli findprobe --sort -id -cc NL --status C --count
575
```

### Search for Probes

```sh
$ ./goatcli findprobe --sort -id -cc NL --status C --limit 5
1004258	Connected	"NL" N/A	AS44103 N/A	[4.8995 52.3815]
1004216	Connected	"NL" N/A	AS204995 N/A	[4.6675 52.3905]
1004194	Connected	"NL"	"atlas02.netjoe.nl"	AS206238	AS206238	[6.0915 52.5095]
1004164	Connected	"NL"	"Netrebel"	AS38919	AS38919	[5.1805 52.0295]
1004163	Connected	"NL"	"Docker probe"	AS1136	AS1136	[5.6405 52.0385]
```


### Get a Particular Probe

```sh
$ ./goatcli findprobe --id 10001
10001	Connected	"NL" N/A	AS206238	AS206238	[4.9275 52.3475]
```

## Finding Anchors

### Count Anchors Matching Some Criteria

```sh
$ ./goatcli findanchor --asn4 3320 --count
5
```

### Search for Anchors

```sh
$ ./goatcli findanchor --asn4 3320
2899	7040	AT	Vienna	at-vie-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[16.392002 48.193092]
3042	7075	DE	Berlin	de-ber-as3320-client.anchors.atlas.ripe.net	AS3320	AS3320	[13.422857 52.54575]
1963	6724	DE	Frankfurt am Main	de-fra-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[8.682127 50.110924]
2900	7065	US	New York City	us-nyc-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[-74.00833 40.717663]
2045	6748	US	New York	us-nyc-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[-74.00831 40.717728]
```


### Get a Particular Anchor

```sh
$ ./goatcli findanchor --id 2899
2899	7040	AT	Vienna	at-vie-as3320.anchors.atlas.ripe.net	AS3320	AS3320	[16.392002 48.193092]
```

## Finding Measurements

### Count Measurements Matching Some Criteria

```sh
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --count
165
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --type ping --count
25
```

### Search for Measurements

```sh
$ ./goatcli findmsm --target 193.0.0.0/19 --status ong --type ping --limit 5
34680813	Ongoing	Periodic	IPv4	2021-12-30T14:23:04Z N/A	900	27	ping	k.root-servers.net
29493618	Ongoing	Periodic	IPv4	2021-04-06T17:13:13Z N/A	240	5	ping	ripe.net
21879480	Ongoing	Periodic	IPv4	2019-06-05T15:18:05Z N/A	240	50	ping	ripe.net
21879479	Ongoing	Periodic	IPv4	2019-06-05T15:17:35Z N/A	240	50	ping	ripe.net
19230504	Ongoing	Periodic	IPv4	2019-01-29T17:01:43Z N/A	60	50	ping	ping.ripe.net
```

### Get a Particular Measurement

```sh
$ ./goatcli findmsm --id 34680813
34680813	Ongoing	Periodic	IPv4	2021-12-30T14:23:04Z N/A	900	27	ping	k.root-servers.net
```

## Processing results

goatCLI can fetch results of exiting measurements (either from the data API, result streaming or from a local file). It's possible to choose what kind of output you want via output formatters:
* `some` and `most` echo some basic properties of the results
* `native` produces native-looking outputs (for now for ping and traceroute)
* `dnsstat` provides basic statistics of DNS results
* `id` and `idcsv` only output the ID of the results (`idcsv` does this in CSV format)

The API call variant supports setting the start time, end time, probe id(s), and a few more filters.

Getting data from the data API:

```sh
$ ./goatcli result --id 1001 --probe 10001 --start today --output most
```

Tuning in to result stream, displaying incoming results in real-time but also saving them to a local file:

```sh
$ ./goatcli result --id 1001 --stream --save output.jsonl
```

Loading data from a local file is simple as well (from a file with one result per line ("format=txt" a.k.a. JSONL a.k.a. NDJSON)):

```sh
$ cat some-results.jsonl | ./goatcli result
```

## Schedule a New Measurement

Use the `measurement` subcommand. Quick examples:

```sh
# create a one-off with default set of probes
$ ./goatcli measure --ping --target ping.ripe.net

# create an periodic measurement with default set of probes and predefined start and end times
$ ./goatcli measure --ping --target ping.ripe.net --periodic --start 2023-10-23T19 --end tomorrow

# create a one-off with 15 probes from AS33260
$ ./goatcli measure --ping --target ping.ripe.net --probeasn 15@3320

# create a one-off with 15 probes from AS33260 and 20 probes from NL
$ ./goatcli measure --trace --target ping.ripe.net --probeasn 15@3320 --probecc 20@nl
```

### Authentication

A valid API key for creating measurements is needed. It should be defined in the configuration file (`/.config/goat.ini`):

```
[apikeys]
# Create new measurement(s)
create_measurements = "12345678-xxxx-xxxx-xxxx-something"
```

It is also possible to supply a key on the command line (`--key KEY`).

### Probe Selection

The following probe selection options are available:
* `--probearea` select from areas. A comma separated list of `amount@area` where `area` can be `ww`, `west`, `nc`, `sc`, `ne`, `se`
* `--probeasn` select from ASNs. A comma separated list of `amount@ASN`
* `--probecc` select from countries. A comma separated list of `amount@CC` where CC is a valid country code
* `--probelist` provides an explicit list of probe IDs to include as a comma separated list.
* `--probeprefix` select from prefixes (IPv4 or IPv6). A comma separated list of `amount@prefix`
* `--probereuse` reuse probes from a precvious measurement. A comma separated list of `amount@msmID`

Multiple probe selection criteria can be sepcified; each of them add more probes to the selection.

`--probetaginc` and `--probetagexc` can be used to filter for probes that have been tagged (or not tagged) with those tags. Both are comma separated lists.

A default probe selection can be expressed in the configuration file (`/.config/goat.ini`) using entries with the above names in the `[probespec]` section, e.g.:

```
# default probe specifications for new measurements
[probespec]
probecc = ""
probearea = "9@ww"
probeasn = ""
probeprefix = ""
probelist = ""
probereuse = ""
```

### Timing

By default one-offs are scheduled, starting as soon as possible. You can specify a start time (`start`) in the future and for periodic meaasurements perhaps even an end time (`end`). Times can be specifies as:
* UNIX timestamps
* ISO8601 variants:
  * `YYYY-mm-ddTHH:MM:SS` - leaving time details from the right makes them default to 0 (e.g. `2023-10-24` is valid and becomes `2023-10-24T00:00:00`)

If you want to use a recurring measurement (with default or manually set `interval` and `spread`), use `--periodic`.

### Measurement Types and Options

You can define one measurement per invocation. This can be one of: `ping`, `trace`, `dns`, `tls`, `ntp` or `http` (the last one with restrictions by the system). Each measurement needs a `target`, except DNS that needs a `name` to look up and uses `target` as the server/resolver to use if it's specified, otherwise uses the local resolver.

Common options for measurements include `interval`, `spread`, `tags` and some more.

Each mesurement type accepts a number of options, such as `abuf`, `qbuf`, `nsid`, `rd` for DNS, `minhop` and `maxhop` for trace, etc. Check the help page for the complete list of these.

### Output

If the measurement scheduling request was successful and results are not requested immediately, the ID of the new measurement is displayed.

If results are requested to be shown immediately (e.g. the result stream is used) -- which is the defaut for one-off measurements -- then those results will be displayed as they become available. The `--output` and `--opt` flags can be used to control the output format. Also, `--save <FILENAME>` can be used to store incoming results in a file as well.

### Shortcuts

You can use shortcuts for all measurement types:

```sh
# this is equivalent to "goatcli measure -type ping -target x.y.z [other params]"
$ goatcli ping x.y.z [other params]
# this is equivalent to "goatcli measure -type trace -target x.y.z [other params]"
$ goatcli trace x.y.z [other params]
# this is equivalent to "goatcli measure -type dns -name x.y.z [other params]"
$ goatcli dns x.y.z [other params]
# this is equivalent to "goatcli measure -type ntp -target x.y.z [other params]"
$ goatcli ntp x.y.z [other params]
# this is equivalent to "goatcli measure -type tls -target x.y.z [other params]"
$ goatcli tls x.y.z [other params]
# this is equivalent to "goatcli measure -type http -target http://x.y.z/ [other params]"
$ goatcli http http://x.y.z/ [other params]
```

## Stop a Periodic Measurement

Use the `measurement` subcommand with the `--stop ID` flag. The required key can be added to the config file as `stop_measurements`.

## Adding and Removing Probes

To add probes to an existing measurement use the `measurement` subcommand with the `--add ID` flag`. You can use the same probe specification language as for a new measurement.

To remove probes from an existing measurement use the `measurement` subcommand with the `--remove ID` flag. You can use the `--probelist` similarly to scheduling a new measurement. Other probe specifications are not supported.

The required key can be added to the config file as `update_measurements`.

## Status Checks

Provide a short summary of the results for a [measurement status check](https://atlas.ripe.net/docs/apis/rest-api-manual/measurements/status-checks.html). It needs a measurement ID, and by default it summarises the results (is there an alert, how many probes are in that state out of how many total, and with the `most` output formatter the list of alerting probes).

```sh
$ ./goatcli status -id 61953517 -output most
true	1	9	[1005382]
```

## Output Formatters

The output formatters are extensible, feel free to write your own -- and contribute that back to this repo! You only need to make a new package under `output` that implements four functions:
* `setup()` to initialise the output processor
* `start()` to prepare for processing results; may be called once per batch of results
* `process()` to deal with one incoming result
* `finish()` to finish processing, make a summary, etc.; may be called once per batch of results

New output processors need to be registered in `goatcli.go`. See `some.go` or `native.go` for examples and [the complete description of formatters and their options](output.md).
