/*
  (C) 2022 Robert Kisteleki & RIPE NCC

  See LICENSE file for the license.
*/

package result

import (
	"net/netip"
	"testing"
)

// Test firmware parsing as int
func TestFirmwareAsInt(t *testing.T) {
	var base BaseResult
	err := base.Parse(`
{
"fw":1,
"af":4,
"dst_addr":"193.0.14.129",
"from":"99.99.99.99",
"msm_id":10101,
"prb_id":9999,
"src_addr":"88.88.88.88",
"timestamp":1451606452,
"type":"dns"
}
`)
	if err != nil || base.GetFirmwareVersion() != 1 {
		t.Fatalf("firmware version as int should be accepted correctly")
	}
}

// Test firmware parsing as int
func TestFirmwareAsString(t *testing.T) {
	var base BaseResult
	err := base.Parse(`
{
"fw":"1234",
"af":4,
"dst_addr":"193.0.14.129",
"from":"99.99.99.99",
"msm_id":10101,
"prb_id":9999,
"src_addr":"88.88.88.88",
"timestamp":1451606452,
"type":"dns"
}
`)
	if err != nil || base.GetFirmwareVersion() != 1234 {
		t.Fatalf("firmware version as string should be accepted correctly")
	}
}

// Test if the base parser does a decent job
func TestBaseParser(t *testing.T) {
	var base BaseResult
	err := base.Parse(`
{
"fw":1234,
"mver":"1.2.3",
"lts":-1,
"dst_name":"example.com",
"ttr":1.234,
"af":4,
"dst_addr":"192.0.2.1",
"src_addr":"2001:0db8::1",
"msm_id":1234567,
"prb_id":2345678,
"timestamp":123456789,
"msm_name":"Meh",
"from":"192.0.2.2",
"type":"meh",
"group_id":34567890,
"step":10,
"stored_timestamp":123456790
}
`)
	if err != nil {
		t.Fatalf("Error parsing base result: %s", err)
	}

	assertEqual(t, base.GetFirmwareVersion(), uint(1234), "error parsing base field value for fw")
	assertEqual(t, base.LastTimeSync, -1, "error parsing base field value for lts")
	assertEqual(t, base.DestinationName, "example.com", "error parsing base field value for dst_name")
	dst, _ := netip.ParseAddr("192.0.2.1")
	assertEqual(t, *base.DestinationAddr, dst, "error parsing base field value for dst_addr")
	src, _ := netip.ParseAddr("2001:0db8::1")
	assertEqual(t, base.SourceAddr, src, "error parsing base field value for src_addr")
	from, _ := netip.ParseAddr("192.0.2.2")
	assertEqual(t, base.FromAddr, from, "error parsing base field value for from")
	assertEqual(t, *base.ResolveTime, 1.234, "error parsing base field value for ttr")
	assertEqual(t, base.AddressFamily, uint(4), "error parsing base field value for af")
	assertEqual(t, base.MeasurementID, uint(1234567), "error parsing base field value for msm_id")
	assertEqual(t, base.ProbeID, uint(2345678), "error parsing base field value for prb_id")
	assertEqual(t, base.GroupID, uint(34567890), "error parsing base field value for group_id")
	assertEqual(t, base.TimeStamp.String(), "1973-11-29T21:33:09Z", "error parsing base field value for timestamp")
	assertEqual(t, base.StoreTimeStamp.String(), "1973-11-29T21:33:10Z", "error parsing base field value for stored_timestamp")
	assertEqual(t, base.MeasurementName, "Meh", "error parsing base field value for msm_name")
	assertEqual(t, base.Type, "meh", "error parsing base field value for type")
}
