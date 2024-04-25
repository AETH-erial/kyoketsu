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
	"strings"
	"sync"
	"syscall"
	"time"
)

/*
Need to work with with a database schema in mind, and revolve functionality around that
*/

type Host struct {
	Fqdn           string // The FQDN of the address targeted as per the systems default resolver
	IpAddress      string // the IPv4 address (no ipv6 support yet)
	PingResponse   bool   // boolean value representing if the host responded to ICMP
	ListeningPorts []int  // list of maps depicting a port number -> service name
	PortString     string
	Id             int64
}

/*
Perform a concurrent TCP port dial on a host, either by domain name or IP.

	:param addr: the address of fqdn to scan
	:param ports a list of port numbers to dial the host with
*/
func PortWalk(addr string, ports []int) []int {
	out := []int{}
	for i := range ports {
		p := singlePortScan(addr, ports[i])
		if p != 0 {
			out = append(out, p)
		}
	}
	return out

}

type PortScanResult struct {
	// This is used to represent the results of a port scan against one host
	PortNumber int    `json:"port_number"` // The port number that was scanned
	Service    string `json:"service"`     // the name of the service that the port was identified/mapped to
	Protocol   string `json:"protocol"`    // The IP protocol (TCP/UDP)
	Listening  bool   `json:"listening"`   // A boolean value that depicts if the service is listening or not
}

/*
Wrapper function to dependency inject the resource for a port -> service name mapping.
May move to a database, or something.
*/
func RetrieveScanDirectives() []int {

	var portmap = []int{22, 80, 443, 8080, 4379, 445, 53, 153, 27017}
	return portmap
}

/*
Scans a single host on a single port

	:param addr: the address to dial
	:param port: the port number to dial
	:param svcs: the name of the service that the port is associate with
*/
func singlePortScan(addr string, port int) int {

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%d", addr, port), 4*time.Second)
	if err != nil {
		return 0
		//	return PortScanResult{PortNumber: port, Protocol: "tcp", Listening: false}
	}
	conn.Close()
	return port
	//return PortScanResult{PortNumber: port, Protocol: "tcp", Listening: true}
}

/*
Perform a port scan sweep across an entire subnet

	:param ip: the IPv4 address WITH CIDR notation
	:param portmap: the mapping of ports to scan with (port number mapped to protocol name)
*/
func NetSweep(ips []net.IP, ports []int, scanned chan Host) {

	wg := &sync.WaitGroup{}
	for i := range ips {
		wg.Add(1)
		go func(target string, portnum []int, wgrp *sync.WaitGroup) {
			defer wgrp.Done()
			portscanned := PortWalk(target, portnum)
			scanned <- Host{
				Fqdn:           getFqdn(target),
				IpAddress:      target,
				ListeningPorts: portscanned,
				PortString:     strings.Trim(strings.Join(strings.Fields(fmt.Sprint(portscanned)), ","), "[]"),
			}

		}(ips[i].String(), ports, wg)
	}
	wg.Wait()
	close(scanned)

}

func getFqdn(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return "not found with default resolver"
	}
	return strings.Join(names, ", ")
}

/*
Create a new TCP dialer to share in a goroutine
*/
func NewDialer() net.Dialer {
	return net.Dialer{}
}

/*
Create a new low level networking interface socket
:param intf: the name of the interface to bind the socket to
*/
func NewTCPSock(interfaceName string) *syscall.SockaddrLinklayer {
	sock, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, syscall.ETH_P_IP)
	if err != nil {
		log.Fatal(err, " Could not create raw AF_PACKET socket.\n")
	}
	defer syscall.Close(sock)
	intf, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Fatal(err, " Couldnt locate that interface. Are you sure you mean to pass ", interfaceName, " ?")
	}
	return &syscall.SockaddrLinklayer{
		Protocol: htons(syscall.ETH_P_IP),
		Ifindex:  intf.Index,
	}

}

// htons converts a uint16 from host- to network byte order.
func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | i>>8
}
