Working on a network observability tool. Architecting it as a peer-to-peer network, each node on the
p2p network performs host scans on addresses within its own address space, and reports back to a central client
that establishes itself as the broker for that round of scanning. Going to use this project as a way to explore
network enumeration, distributed system resilience, and concurrent programming at large scale.

Since any node on the network can act as a C2/master/broker server, I want to disperse the scanning data via a torrent
network between nodes, so that the speed of replication scales with the size of the network.

Im also attempting to solve the problem of scanning hosts on a large network with rotating IPs, so that when a network/host
sweep is executed, the amount of time that a mapping of the network is accurate is extended. I'm attempting to combat this by:
1. designing this is a scalable distributed system
2. utilizing the strong concurrent features of Golang
3. utilizing the torrent protocol to distribute data so that state changes can easily be reflected across nodes.