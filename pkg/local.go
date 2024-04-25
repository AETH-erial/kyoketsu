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
	"math"
	"net"
	"net/netip"
	"strconv"
	"strings"
)

const IPV4_BITLEN = 32

type NetworkInterfaceNotFound struct{ Passed string }

// Implementing error interface
func (n *NetworkInterfaceNotFound) Error() string {
	return fmt.Sprintf("Interface: '%s' not found.", n.Passed)
}

type IpSubnetMapper struct {
	Ipv4s       []net.IP `json:"addresses"`
	NetworkAddr net.IP
	Current     net.IP
	Mask        int
}

type PromptEntry struct {
	HostAddress    string
	NetworkAddress string
	Cidr           string
	SubnetMask     string
	InterfaceName  string
	MacAddress     string
}

type TuiSelectionFeed struct {
	Choice []PromptEntry
}

/*
Get the next IPv4 address of the address specified in the 'addr' argument,

	:param addr: the address to get the next address of
*/
func getNextAddr(addr string) string {
	parsed, err := netip.ParseAddr(addr)
	if err != nil {
		log.Fatal("failed while parsing address in getNextAddr() ", err, "\n")
	}
	return parsed.Next().String()

}

/*
get the network address of the ip address in 'addr' with the subnet mask from 'cidr'

	    :param addr: the ipv4 address to get the network address of
		:param cidr: the CIDR notation of the subbet
*/
func getNetwork(addr string, cidr int) string {
	addr = fmt.Sprintf("%s/%v", addr, cidr)
	ip, net, err := net.ParseCIDR(addr)
	if err != nil {
		log.Fatal("failed whilst attempting to parse cidr in getNetwork() ", err, "\n")
	}
	return ip.Mask(net.Mask).String()

}

/*
Recursive function to get all of the IPv4 addresses for each IPv4 network that the host is on

	     :param ipmap: a pointer to an IpSubnetMapper struct which contains domain details such as
		               the subnet mask, the original network mask, and the current IP address used in the
					   recursive function
		:param max: This is safety feature to prevent stack overflows, so you can manually set the depth to
		            call the function
*/
func addressRecurse(ipmap IpSubnetMapper) IpSubnetMapper {

	next := getNextAddr(ipmap.Ipv4s[len(ipmap.Ipv4s)-1].String())

	if getNetwork(next, ipmap.Mask) != ipmap.NetworkAddr.String() {
		return ipmap
	}

	ipmap.Ipv4s = append(ipmap.Ipv4s, net.ParseIP(next))
	return addressRecurse(ipmap)
}

/*
Get all of the IPv4 addresses in the network that 'addr' belongs to. YOU MUST PASS THE ADDRESS WITH CIDR NOTATION
i.e. '192.168.50.1/24'

	:param addr: the ipv4 address to use for subnet discovery
*/
func GetNetworkAddresses(addr string) (IpSubnetMapper, error) {
	ip, ntwrk, err := net.ParseCIDR(addr)
	if err != nil {
		return IpSubnetMapper{}, err
	}
	mask, err := strconv.Atoi(strings.Split(addr, "/")[1])
	if err != nil {
		return IpSubnetMapper{}, err
	}
	ipmap := IpSubnetMapper{Ipv4s: []net.IP{ip},
		NetworkAddr: ip.Mask(ntwrk.Mask),
		Mask:        mask,
		Current:     ip.Mask(ntwrk.Mask)}
	return addressRecurse(ipmap), nil

}

/*
Turns a set of network mask bits into a valid IPv4 representation

	    :param ones: number of 1's in the netmask, i.e. 16 == 11111111 11111111 00000000 00000000
		:param bits: the number of bits that the mask consists of (need to keep this param for ipv6 support later)
*/
func bitsToMask(ones int, bits int) string {
	var bitmask []int

	for i := 0; i < ones; i++ {
		bitmask = append(bitmask, 1)
	}
	for i := ones; i < bits; i++ {
		bitmask = append(bitmask, 0)
	}

	octets := []string{
		strconv.Itoa(base2to10(bitmask[0:8])),
		strconv.Itoa(base2to10(bitmask[8:16])),
		strconv.Itoa(base2to10(bitmask[16:24])),
		strconv.Itoa(base2to10(bitmask[24:32])),
	}
	return strings.Join(octets, ".")
}

/*
convert a base 2 number (represented as an array) to a base 10 integer

	:param bits: the slice of ints split into an array, e.g. '11110000' would be [1 1 1 1 0 0 0 0]
*/
func base2to10(bits []int) int {
	var sum int
	sum = 0
	for i := range bits {
		bits[i] = bits[i] * powerInt(2, len(bits)-1-i)
	}
	for i := range bits {
		sum = sum + bits[i]
	}
	return sum

}

/*
	 Wrapper func for getting the value of x to the power of y, as int opposed to float64
	    :param x: the base number to operate on
		:param y: the exponent
*/
func powerInt(x int, y int) int {
	return int(math.Pow(float64(x), float64(y)))
}

// Needs cleanup, but this function populatest a data structure that will be used during TUI program startup
func RetrieveLocalAddresses() (TuiSelectionFeed, error) {
	var tuidata TuiSelectionFeed
	intf, err := net.Interfaces()
	if err != nil {
		return tuidata, err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return tuidata, err
	}
	for x := range addrs {
		var mac string
		var interfacename string
		ip, net, err := net.ParseCIDR(addrs[x].String())
		if err != nil {
			return tuidata, err
		}
		for i := range intf {
			intfAddrs, err := intf[i].Addrs()
			if err != nil {
				return tuidata, err
			}
			for y := range intfAddrs {
				if strings.Contains(intfAddrs[y].String(), strings.Split(ip.String(), "/")[0]) {
					interfacename = intf[i].Name
					mac = intf[i].HardwareAddr.String()
				}
			}
		}

		tuidata.Choice = append(tuidata.Choice, PromptEntry{
			HostAddress:    ip.String(),
			NetworkAddress: ip.Mask(net.Mask).String(),
			Cidr:           strings.Split(addrs[x].String(), "/")[1],
			SubnetMask:     bitsToMask(net.Mask.Size()),
			MacAddress:     mac,
			InterfaceName:  interfacename,
		})

	}

	return tuidata, nil

}
