/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package output

import (
	"fmt"
	"strings"

	"github.com/robert-kisteleki/goatapi/result"
	"golang.org/x/exp/slices"
)

func OutputDnsAnswer(detail *result.DnsAnswer) string {
	cl := result.DnsClassNames[detail.Class]
	if cl == "" {
		// yet unmapped class entries are represented as (Cxx)
		cl = fmt.Sprintf("(C%d)", detail.Class)
	}
	ty := result.DnsTypeNames[detail.Type]
	if ty == "" {
		// yet unmapped type entries are represented as (Txx)
		cl = fmt.Sprintf("(T%d)", detail.Type)
	}
	return fmt.Sprintf("[%s %s %s '%s']", cl, ty, detail.Name, detail.Data)
}

func OutputDnsResponseDetail(resp *result.DnsResponse) string {
	set := make([]string, 0)
	for _, detail := range resp.AllAnswers() {
		set = append(set, OutputDnsAnswer(&detail))
	}
	slices.Sort(set)
	return "[" + strings.Join(set, " ") + "]"
}

func OutputDnsResponseFocus(resp *result.DnsResponse, typehint []string) string {
	set := make([]string, 0)
	for _, detail := range resp.AllAnswers() {
		if len(typehint) == 0 ||
			slices.Contains(typehint, result.DnsTypeNames[detail.Type]) {
			set = append(set, detail.Data)
		}
	}
	slices.Sort(set)
	set = slices.Compact(set)
	return "[" + strings.Join(set, " ") + "]"
}
