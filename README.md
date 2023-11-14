# goat - Go (RIPE) Atlas Tools - CLI & API Wrapper

goat is a Go package to interact with [RIPE Atlas](https://atlas.ripe.net/)
[APIs](https://atlas.ripe.net/api/v2/) using [Golang](https://go.dev/).
It also provides a CLI interface to most of the APIs. It is similar to
[Cousteau](https://github.com/RIPE-NCC/ripe-atlas-cousteau),
[Sagan](https://github.com/RIPE-NCC/ripe-atlas-sagan) and
[Magellan](https://github.com/RIPE-NCC/ripe-atlas-tools)
combined.

It supports:
* finding probes, anchors and measurements
* scheduling new measurements and immediately show their results
* stopping existing measurements
* modify participants of an existing measurement (add/remove probes)
* downloading results of measurements and turning them into Go objects, or displaying them
* tuning in to result streaming and turning them into Go objects
* loading a local file containing measurement results and turning them into Go objects
* various kinds of output formatters for displaying and aggregating measurement results

The tool needs Go 1.21 to compile.

# Context

[RIPE Atlas](https://atlas.ripe.net) is an open, community based active Internet
measurement network developed by the [RIPE NCC](https://www.ripe.net/) since 2010.
It provides a number of vantage points ("probes") run by volunteers, that allow
various kinds of network measurements (pings, traceroutes, DNS queries, ...) to
be run by any user.

# Quick Start

Check the [API Wrapper Quick Start Guide](doc/quickstart-api.md) and the [CLI Quick Start Guide](doc/quickstart-cli.md).

# Future Additions / TODO

* check credit balance, transfer credits, ...

# Copyright, Contributing

(C) 2022, 2023 [Robert Kisteleki](https://kistel.eu/) & [RIPE NCC](https://www.ripe.net)

Contribution is possible and encouraged via the [Github repo]("https://github.com/robert-kisteleki/goat/")

# License

See the LICENSE file
