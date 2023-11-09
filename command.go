/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"flag"
	"fmt"
	"os"
)

// Figure out which subcommand was requested
func commandSelector() {
	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		return
	}

	switch {
	case args[0] == "help":
		commandHelp()
	case args[0] == "version":
		commandVersion()
	case args[0] == "findprobe" || args[0] == "fp":
		commandFindProbe(args[1:])
	case args[0] == "findanchor" || args[0] == "fa":
		commandFindAnchor(args[1:])
	case args[0] == "findmsm" || args[0] == "fm":
		commandFindMsm(args[1:])
	case args[0] == "result":
		commandResult(args[1:])
	case args[0] == "status":
		commandStatusCheck(args[1:])
	case args[0] == "measure":
		commandMeasure(args[1:])
	case args[0] == "ping":
		commandMeasure(append([]string{"-ping", "-target"}, args[1:]...))
	case args[0] == "trace":
		commandMeasure(append([]string{"-trace", "-target"}, args[1:]...))
	case args[0] == "dns":
		commandMeasure(append([]string{"-dns", "-name"}, args[1:]...))
	case args[0] == "tls":
		commandMeasure(append([]string{"-tls", "-target"}, args[1:]...))
	case args[0] == "ntp":
		commandMeasure(append([]string{"-ntp", "-target"}, args[1:]...))
	case args[0] == "http":
		commandMeasure(append([]string{"-http", "-target"}, args[1:]...))
	default:
		commandHelp()
	}
}

func commandHelp() {
	printUsage()
}

func commandVersion() {
	fmt.Println(CLIName)
}

// general usage text
func printUsage() {
	fmt.Printf("Usage: %s [options] <command> [flags]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("Available commands are:")
	fmt.Println("	help             this page")
	fmt.Println("	version          print version")
	fmt.Println("	fp|findprobe     search for probes")
	fmt.Println("	fa|findanchor    search for achors")
	fmt.Println("	fm|findmsm       search for measurements")
	fmt.Println("	result           download results")
	fmt.Println("	status           measurement status check")
	fmt.Println("	measure          start new measurement(s)")
	fmt.Println("	dns              shortcut to -dns -name")
	fmt.Println("	http             shortcut to -http -target")
	fmt.Println("	ntp              shortcut to -ntp -target")
	fmt.Println("	ping             shortcut to -ping -target")
	fmt.Println("	tls              shortcut to -tls -target")
	fmt.Println("	trace            shortcut to -trace -target")
	fmt.Println("")
	fmt.Println("Options:")
	flag.PrintDefaults()
}
