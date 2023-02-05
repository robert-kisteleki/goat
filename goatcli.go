/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package main

import (
	_ "goatcli/output/dnsstat"
	_ "goatcli/output/id"
	_ "goatcli/output/idcsv"
	_ "goatcli/output/most"
	_ "goatcli/output/native"
	_ "goatcli/output/none"
	_ "goatcli/output/some"
)

func main() {
	configure()
	commandSelector()
}
