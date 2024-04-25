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
	"fmt"
	"log"
	"net"
	"sync"
	"testing"
)

func Equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func startTestSever(port int, t *testing.T) {
	listen, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		log.Fatal(err)
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
		}
		t.Logf("Recieved dial from PortWalk of port: %v from host: %s\n", port, conn.RemoteAddr().String())
		conn.Close()
	}
}

func TestPortWalk(t *testing.T) {
	type TestCase struct {
		Name       string
		ScanPort   []int
		ListenPort []int
		ShouldFail bool
	}
	tc := []TestCase{
		TestCase{
			Name:       "Passing test, listening on 8080/27017 and scanned on 8080/27017.",
			ScanPort:   []int{8080, 27017},
			ListenPort: []int{8080, 27017},
			ShouldFail: false,
		},
		TestCase{
			Name:       "Failing test, listening on 8081 scanned on 6439",
			ScanPort:   []int{6439},
			ListenPort: []int{8081},
			ShouldFail: true,
		},
	}
	wg := &sync.WaitGroup{}
	for i := range tc {
		for x := range tc[i].ListenPort {
			wg.Add(1)
			go startTestSever(tc[i].ListenPort[x], t)
		}
		got := PortWalk("localhost", tc[i].ScanPort)
		for k, _ := range tc[i].ListenPort {
			wg.Done()
			if !Equal(got, tc[i].ListenPort) {
				if !tc[i].ShouldFail {
					t.Errorf("Test '%s' failed! PortWalk didnt detect the test server was listening on: %+v\n", tc[i].Name, tc[i].ListenPort)
				}
			}
			t.Logf("Test '%s' passed! Scanned port: '%v'", tc[i].Name, k)

		}

	}
}
