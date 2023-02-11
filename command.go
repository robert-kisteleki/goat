/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	"flag"
	"fmt"
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
