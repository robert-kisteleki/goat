/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/dnsstat"
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/id"
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/idcsv"
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/most"
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/native"
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/none"
	_ "github.com/robert-kisteleki/goat/cmd/goat/output/some"
)

func main() {
	configure()
	commandSelector()
}
