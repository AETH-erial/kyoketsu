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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	kyoketsu "github.com/AETH-erial/kyoketsu/pkg"
	"github.com/manifoldco/promptui"
)

var licenseMsg = "\n	kyoketsu  Copyright (C) 2024  Russell Hrubesky, ChiralWorks Software LLC\n	This program comes with ABSOLUTELY NO WARRANTY; for details type `kyoketsu --license`\n	This is free software, and you are welcome to redistribute it\n	under certain conditions; type `kyoketsu --redist` for details.\n\n"

var redistMsg = "\n	This program is free software: you can redistribute it and/or modify\n	it under the terms of the GNU General Public License as published by\n	the Free Software Foundation, either version 3 of the License, or\n	(at your option) any later version.\n\n"

var licenseMsgLong = "\n		GNU GENERAL PUBLIC LICENSE\n		Version 3, 29 June 2007\n	Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>\n	Everyone is permitted to copy and distribute verbatim copies\n	of this license document, but changing it is not allowed.\n\n	kyoketsu, An HTTP Proxying framework for bypassing DNS Security\n	Copyright (C) 2024 Russell Hrubesky, ChiralWorks Software LLC\n\n	This program is free software: you can redistribute it and/or modify\n	it under the terms of the GNU General Public License as published by\n	the Free Software Foundation, either version 3 of the License, or\n	(at your option) any later version.\n\n	This program is distributed in the hope that it will be useful,\n	but WITHOUT ANY WARRANTY; without even the implied warranty of\n	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the\n	GNU General Public License for more details.\n\n	You should have received a copy of the GNU General Public License\n	along with this program.  If not, see <https://www.gnu.org/licenses/>.\n\n"

func main() {

	licenseInfo := flag.Bool("license", false, "Pass this flag to display license and warantee information.")
	redistInfo := flag.Bool("redist", false, "Pass this flag to display redistribution information.")
	flag.Parse()

	if *licenseInfo {
		fmt.Println(licenseMsgLong)
		os.Exit(0)
	}
	if *redistInfo {
		fmt.Println(redistMsg)
		os.Exit(0)
	}
	fmt.Println(licenseMsg)

	var err error
	localAddr, err := kyoketsu.RetrieveLocalAddresses()
	if err != nil {
		log.Fatal(err)
	}
	prompt := promptui.Select{
		Label: "Select the network you wish to scan",
		Items: localAddr.Choice,
	}
	choice, _, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	var addr kyoketsu.IpSubnetMapper
	targetNet := fmt.Sprintf("%s/%s", localAddr.Choice[choice].HostAddress, localAddr.Choice[choice].Cidr)
	addr, err = kyoketsu.GetNetworkAddresses(targetNet)
	if err != nil {
		log.Fatal(err)
	}
	scanned := make(chan kyoketsu.Host)
	go func() {
		for x := range scanned {
			if len(x.ListeningPorts) > 0 {
				fmt.Print(" |-|-|-| :::: HOST FOUND :::: |-|-|-|\n==================||==================\n")
				fmt.Printf("IPv4 Address: %s\nFully Qualified Domain Name: %s\nListening Ports: %v\n=====================================\n", x.IpAddress, x.Fqdn, x.ListeningPorts)
			}
		}
	}()
	kyoketsu.NetSweep(addr.Ipv4s, kyoketsu.RetrieveScanDirectives(), scanned)

}
