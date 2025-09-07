//go:build !windows

// for some reason, windows github actions runners don't resolve FQDNs to IPv6

package utils_test

import (
	"fmt"
	"testing"

	"github.com/magneticstain/ip-2-cloudresource/utils"
)

func TestReverseDNSLookup(t *testing.T) {
	var tests = []struct {
		ipAddr, fqdn    string
		expectedVerdict bool
	}{
		{"1.1.1.1", "one.one.one.one.", true},
		{"74.82.42.42", "ordns.he.net.", true},
		{"1.1.1.1", "google.com.", false},
	}

	for _, td := range tests {
		testName := fmt.Sprintf("%s_%s", td.fqdn, td.ipAddr)
		t.Run(testName, func(t *testing.T) {
			fqdns, _ := utils.ReverseDNSLookup(td.ipAddr)

			fqdnFound := false
			var receivedFqdn string
			for _, fqdn := range fqdns {
				if fqdn == td.fqdn {
					fqdnFound = true
					receivedFqdn = fqdn
					break
				}
			}

			if fqdnFound != td.expectedVerdict {
				t.Errorf("reverse IP lookup failed; expected %s to be %t FQDN for %s, received %s", td.fqdn, td.expectedVerdict, td.ipAddr, receivedFqdn)
			}
		})
	}
}

func TestLookupFQDN(t *testing.T) {
	var tests = []struct {
		fqdn, ipAddr    string
		expectedVerdict bool
	}{
		{"example.com", "23.192.228.80", true},                       // accurate IPv4 lookup
		{"example.com", "1.1.1.1", false},                            // inaccurate IPv4 lookup
		{"example.com", "2600:1406:5e00:6::17ce:bc12", true},         // accurate IPv6 lookup
		{"example.com", "2600:9000:24eb:3a00:1:3b80:4f00:21", false}, // inaccurate IPv6 lookup
	}

	for _, td := range tests {
		testName := fmt.Sprintf("%s_%s", td.fqdn, td.ipAddr)
		t.Run(testName, func(t *testing.T) {
			ipAddrs, _ := utils.LookupFQDN(td.fqdn)

			ipFound := false
			for _, ipAddr := range ipAddrs {
				if ipAddr.String() == td.ipAddr {
					ipFound = true
					break
				}
			}

			if ipFound != td.expectedVerdict {
				t.Errorf("FQDN lookup failed; expected %s to be %t IP address for %s; received %v", td.ipAddr, td.expectedVerdict, td.fqdn, ipAddrs)
			}
		})
	}
}

func TestDetermineIpAddrVersion(t *testing.T) {
	var tests = []struct {
		ipAddr string
		ipVer  int
		valid  bool
	}{
		{"1.1.1.1", 4, true},
		{"74.82.42.42", 4, true},
		{"93.184.216.34", 4, true},
		{"123.456.789.10", 4, false},
		{"2606:2800:220:1:248:1893:25c8:1946", 6, true},
		{"xxxx:2800:220:1:ggg:1893:25c8:nww", 6, false},
	}

	for _, td := range tests {
		testName := td.ipAddr

		t.Run(testName, func(t *testing.T) {
			calculatedIpVer, err := utils.DetermineIpAddrVersion(td.ipAddr)

			if !td.valid && err == nil {
				t.Error("expected error when trying to determine bogus IP address version, but none was returned")
			}

			if td.valid && calculatedIpVer != td.ipVer {
				t.Errorf("mismatched versions found for IP version test; IP: %s, Expected Version: %d, Calculated Version: %d, Valid IP?: %t", td.ipAddr, td.ipVer, calculatedIpVer, td.valid)
			}
		})
	}
}
