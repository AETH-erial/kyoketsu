#
#   		GNU GENERAL PUBLIC LICENSE
# 	    	Version 3, 29 June 2007
#
#
# kyoketsu, a Client-To-Client Network Enumeration System
# Copyright (C) 2024 Russell Hrubesky, ChiralWorks Software LLC
#
#  Copyright (C) 2007 Free Software Foundation, Inc. <https://fsf.org/>
#  Everyone is permitted to copy and distribute verbatim copies
#  of this license document, but changing it is not allowed.
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License,
# or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
# See the GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.
#

.PHONY: build format test install coverage coverage-html 

KYOKETSU = kyoketsu
KYOKETSU_WEB = kyoketsu-web

build:
	go build -x -v -o ./build/linux/$(KYOKETSU)/$(KYOKETSU) ./cmd/$(KYOKETSU)/$(KYOKETSU).go && \
	go build -x -v -o ./build/linux/$(KYOKETSU_WEB)/$(KYOKETSU_WEB) ./cmd/$(KYOKETSU_WEB)/$(KYOKETSU_WEB).go

format:
	go fmt ./...

install:
	sudo rm -f /usr/local/bin/$(KYOKETSU) && \
	sudo mv ./build/linux/$(KYOKETSU)/$(KYOKETSU) /usr/local/bin && sudo chmod u+x /usr/local/bin/$(KYOKETSU) && \
	sudo mv ./build/linux/$(KYOKETSU_WEB)/$(KYOKETSU_WEB) \
	/usr/local/bin && sudo chmod u+x /usr/local/bin/$(KYOKETSU_WEB)


test:
	go test -v ./...

coverage:
	go test -v ./... -cover


coverage-html:
	go test -v ./... -coverprofile=coverage.out 
	go tool cover -html=coverage.out -o coverage.html



