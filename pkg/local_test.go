/*
	GNU GENERAL PUBLIC LICENSE
	Version 3, 29 June 2007

kyoketsu, a Client-To-Client Network Enumeration System
Copyright (C) 2024 Russell Hrubesky, ChiralWorks Software LLC

	Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
	Everyone is permitted to copy and distribute verbatim copies
	of this license document, but changing it is not allowed.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package kyoketsu

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
)

type IpAddresses struct {
	Addrs []string `json:"addresses"`
}

func LoadTestAddresses(loc string) map[string]struct{} {
	b, err := os.ReadFile(loc)
	if err != nil {
		log.Fatal("Test setup failed.\n", err)
	}
	var addr IpAddresses
	addrmap := map[string]struct{}{}
	err = json.Unmarshal(b, &addr)
	if err != nil {
		log.Fatal("test setup failed.\n", err)
	}
	for i := range addr.Addrs {
		addrmap[addr.Addrs[i]] = struct{}{}
	}
	return addrmap

}

// Testing the addres recursion function to return all IPs in the target address subnet
// All test cases use a select sort to assert that all addresses in the test data are in the return
func TestAddressRecurse(t *testing.T) {
	type TestCase struct {
		Name       string
		TestData   string
		InputAddr  string
		InputMask  int
		ShouldFail bool
	}

	tc := []TestCase{
		TestCase{
			Name:      "Passing testcase with valid IP address, returns all addresses.",
			TestData:  "../test/local_ips.json",
			InputAddr: "192.168.50.50",
			InputMask: 24,
		},
		TestCase{
			Name:      "Passing testcase with valid IP address that belongs to a /16 subnet",
			TestData:  "../test/slash16_ips.json",
			InputAddr: "10.252.1.1",
			InputMask: 16,
		},
	}
	for i := range tc {
		addr, network, err := net.ParseCIDR(fmt.Sprintf("%s/%v", tc[i].InputAddr, tc[i].InputMask))
		if err != nil {
			t.Errorf("Test case: '%s' failed! Reason: %s", tc[i].Name, err)
		}
		got := IpSubnetMapper{Mask: tc[i].InputMask, NetworkAddr: addr.Mask(network.Mask), Ipv4s: []net.IP{addr}}
		addressRecurse(got)
		want := LoadTestAddresses(tc[i].TestData)
		for x := range got.Ipv4s {
			t.Logf("%s\n", got.Ipv4s[x])
			gotip := got.Ipv4s[x]
			_, ok := want[gotip.String()]
			if !ok {
				t.Errorf("Test '%s' failed! Address: %s was not found in the test data: %s\n", tc[i].Name, gotip.String(), tc[i].TestData)
			}

		}
		t.Logf("Nice! Test: '%s' passed!\n", tc[i].Name)
	}

}

// Testing the function to retrieve the next network address
func TestGetNextAddr(t *testing.T) {
	type TestCase struct {
		Name       string
		Input      string
		Wants      string
		ShouldFail bool
	}

	tc := []TestCase{
		TestCase{
			Name:       "Passing test case, function returns the next address",
			Input:      "10.252.1.1",
			Wants:      "10.252.1.2",
			ShouldFail: false,
		},
		TestCase{
			Name:       "Failing test case, function returns the wrong address",
			Input:      "10.252.1.1",
			Wants:      "10.252.1.4",
			ShouldFail: true,
		},
	}
	for i := range tc {
		got := getNextAddr(tc[i].Input)
		if got != tc[i].Wants {
			if !tc[i].ShouldFail {
				t.Errorf("Test: '%s' failed! Return: %s\nTest expected: %s\nTest Should fail: %v\n", tc[i].Name, got, tc[i].Wants, tc[i].ShouldFail)
			}

		}
		t.Logf("Test: '%s' passed!\n", tc[i].Name)
	}

}

func TestGetNetwork(t *testing.T) {
	type TestCase struct {
		Name       string
		InputAddr  string
		InputMask  int
		Expects    string
		ShouldFail bool
	}
	tc := []TestCase{
		TestCase{
			Name:       "Passing test, function returns the correct network given the CIDR mask",
			InputAddr:  "192.168.50.35",
			InputMask:  24,
			Expects:    "192.168.50.0",
			ShouldFail: false,
		},
		TestCase{
			Name:       "Passing test, function returns the correct network given the CIDR mask (Larger network, /16 CIDR)",
			InputAddr:  "10.252.47.200",
			InputMask:  16,
			Expects:    "10.252.0.0",
			ShouldFail: false,
		},
	}
	for i := range tc {
		got := getNetwork(tc[i].InputAddr, tc[i].InputMask)
		if got != tc[i].Expects {
			if !tc[i].ShouldFail {
				t.Errorf("Test: '%s' failed! Returned: %s\nExpected: %s\nShould fail: %v", tc[i].Name, got, tc[i].Expects, tc[i].ShouldFail)
			}

		}

		t.Logf("Test: '%s' passed!\n", tc[i].Name)

	}

}

func TestGetNetworkAddresses(t *testing.T) {
	type TestCase struct {
		Name       string
		TestData   string
		InputAddr  string
		ShouldFail bool
	}

	tc := []TestCase{
		TestCase{
			Name:       "Passing testcase with valid IP address, returns all addresses.",
			TestData:   "../test/local_ips.json",
			InputAddr:  "192.168.50.50/24",
			ShouldFail: false,
		},
		TestCase{
			Name:       "Passing testcase with valid IP address that belongs to a /16 subnet",
			TestData:   "../test/slash16_ips.json",
			InputAddr:  "10.252.1.1/16",
			ShouldFail: false,
		},
		TestCase{
			Name:       "Failing testcase with invalid IP address.",
			TestData:   "../test/slash16_ips.json",
			InputAddr:  "abcdefgh1234455667",
			ShouldFail: true,
		},
		TestCase{
			Name:       "Failing testcase with valid IP address, but bad CIDR mask",
			TestData:   "../test/slash16_ips.json",
			InputAddr:  "192.168.50.1/deez",
			ShouldFail: true,
		},
	}
	for i := range tc {

		got, err := GetNetworkAddresses(tc[i].InputAddr)
		if err != nil {
			if !tc[i].ShouldFail {
				t.Errorf("Test: '%s' failed! Error: %s", tc[i].Name, err)
			}
			continue
		}

		want := LoadTestAddresses(tc[i].TestData)
		for x := range got.Ipv4s {
			gotip := got.Ipv4s[x]
			_, ok := want[gotip.String()]
			if !ok {
				if !tc[i].ShouldFail {
					t.Errorf("Test '%s' failed! Address: %s was not found in the test data: %s\n", tc[i].Name, gotip.String(), tc[i].TestData)
				}
			}

		}
		t.Logf("Test: '%s' passed!", tc[i].Name)

	}

}

func TestBitsToMask(t *testing.T) {
	type TestCase struct {
		Name  string
		Gets  string
		Wants string
	}

	tc := []TestCase{
		TestCase{
			Name:  "Function gets valid IP with cidr",
			Gets:  "192.168.50.1/28",
			Wants: "255.255.255.240",
		},
	}

	for i := range tc {
		_, net, err := net.ParseCIDR(tc[i].Gets)
		if err != nil {
			log.Fatal(err)
		}
		want := tc[i].Wants
		got := bitsToMask(net.Mask.Size())
		if got != want {
			t.Errorf("test '%s' failed! got: %s\nwanted: %s\n", tc[i].Name, got, want)
		}
	}
}
