# goatcli - Go (RIPE) Atlas Tools - CLI

goatcli provides a CLI to interact with [RIPE Atlas](https://atlas.ripe.net/) [APIs](https://atlas.ripe.net/api/v2/) using [Golang](https://go.dev/). It uses the [goatapi](https://github.com/robert-kisteleki/goatapi/) under the hood. It is similar to [Magellan](https://github.com/RIPE-NCC/ripe-atlas-tools).

It supports:
* finding probes
* finding anchors
* finding measurements
* scheduling new measurements and immediately show its results
* downloading and displaying results of measurements
* tuning in to result streaming
* loading a local file containing measurement results
* various kinds of output formatters for displaying and aggregating measurement results
* (more features to come)

The tool needs Go 1.18 to compile.

# Context

[RIPE Atlas](https://atlas.ripe.net) is an open, community based active Internet
measurement network developed by the [RIPE NCC](https://www.ripe.net/) since 2010.
It provides a number of vantage points ("probes") run by volunteers, that allow
various kinds of network measurements (pings, traceroutes, DNS queries, ...) to
be run by any user.


# Quick Start

Check the [Quick Start Guide](doc/quickstart.md)

# Future Additions / TODO

* stop existing measurements
* modify participants of an existing measurement (add/remove probes)
* check credit balance, transfer credits, ...

# Copyright, Contributing

(C) 2022, 2023 [Robert Kisteleki](https://kistel.eu/) & [RIPE NCC](https://www.ripe.net)

Contribution is possible and encouraged via the [Github repo]("https://github.com/robert-kisteleki/goatcli/")

# License

See the LICENSE file
