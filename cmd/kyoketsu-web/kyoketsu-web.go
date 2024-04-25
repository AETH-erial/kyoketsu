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
	"database/sql"
	"flag"
	"log"
	"net/http"
	"sync"

	kyoketsu "github.com/AETH-erial/kyoketsu/pkg"
)

const dbfile = "sqlite.db"

func main() {

	port := flag.Int("port", 8080, "Select the port to run the server on")
	debug := flag.Bool("debug", false, "Pass this to start the pprof server")
	flag.Parse()

	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	if *debug {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}
	hostsRepo := kyoketsu.NewSQLiteRepo(db)
	var wg sync.WaitGroup
	wg.Add(1)
	go kyoketsu.RunHttpServer(*port, hostsRepo, kyoketsu.RetrieveScanDirectives())

	if err = hostsRepo.Migrate(); err != nil {
		log.Fatal(err)
	}
	log.Println("SUCCESS ::: SQLite database initiated, and open for writing.")

	wg.Wait()

}
