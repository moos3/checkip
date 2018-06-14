package main

import (
	"testing"
)

func TestLookupGeoIP(t *testing.T) {
	address := "8.8.8.8"
	ip := lookupGeoIP(address)
	if ip.IP != "8.8.8.8" {
		t.Errorf("Ip address should be %s and got back %s", address, ip.IP)
	}
}

func TestLookupGeoTZ(t *testing.T) {
	var (
		lat float32
		lon float32
	)
	lat = 44.12168800000001
	lon = -69.34836389999998
	tzid := "America/New_York"
	tz := lookupGeoTz(lat, lon)
	if tz != "America/New_York" {
		t.Errorf("TimezoneID should have been %s and got back %s", tzid, tz)
	}
}
