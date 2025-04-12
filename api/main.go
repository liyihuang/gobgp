package main

import (
	"context"
	"fmt"
	"time"

	api "github.com/osrg/gobgp/v3/api"
	"github.com/osrg/gobgp/v3/pkg/apiutil"
	"github.com/osrg/gobgp/v3/pkg/packet/bgp"
	"github.com/osrg/gobgp/v3/pkg/server"
)

func main() {
	s := server.NewBgpServer(
		server.GrpcListenAddress("127.0.0.1:50051"),
	)

	go s.Serve()

	// Set up global BGP configuration
	global := &api.StartBgpRequest{
		Global: &api.Global{
			Asn:        64512,     // Local AS number
			RouterId:   "1.1.1.1", // Router ID (IP format)
			ListenPort: 179,       // Disable TCP listening (set to 179 for standard port)
		},
	}

	// Apply global configuration
	s.StartBgp(context.Background(), global)

	// Display global configuration
	displayGlobalConfig(s)

	// Configure a BGP neighbor (peer)
	peer := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: "192.168.2.1", // Peer's IP address
			PeerAsn:         64513,         // Peer's AS number
		},
	}

	// Add the neighbor to the BGP server
	s.AddPeer(context.Background(), &api.AddPeerRequest{Peer: peer})
	time.Sleep(10 * time.Second)
	displayNeighbors(s)
	/*
		// Wait for peer to be established - this will block until peer is established or error occurs
		log.Println("Waiting for peer connection to establish...")
		done := make(chan struct{})
		go func() {
			s.WatchEvent(context.Background(), &api.WatchEventRequest{
				Peer: &api.WatchEventRequest_Peer{},
			}, func(r *api.WatchEventResponse) {
				if p := r.GetPeer(); p != nil &&
					p.Peer != nil &&
					p.Peer.Conf != nil &&
					p.Peer.Conf.NeighborAddress == "192.168.2.1" {

					if p.Peer.State != nil {
						log.Printf("Peer state: %s", p.Peer.State.SessionState)

						if p.Peer.State.SessionState == api.PeerState_ESTABLISHED {
							log.Println("Peer connection established")
							close(done)
						}
					}
				}
			})
		}()

		<-done
	*/
	// Add a route
	nlri := bgp.NewIPAddrPrefix(24, "192.168.1.0")
	nlriPb, _ := apiutil.MarshalNLRI(nlri)

	attrs := []bgp.PathAttributeInterface{
		bgp.NewPathAttributeOrigin(0),
		bgp.NewPathAttributeNextHop("192.168.1.1"),
	}
	attrsPb, _ := apiutil.MarshalPathAttributes(attrs)

	path := &api.Path{
		Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
		Nlri:   nlriPb,
		Pattrs: attrsPb,
	}
	s.AddPath(context.Background(), &api.AddPathRequest{
		Path: path,
	})

	// Wait for routes to be processed
	time.Sleep(30 * time.Second)

	// Display RIB
	displayRIB(s)

	select {}
}

func displayGlobalConfig(s *server.BgpServer) {
	// Get global configuration
	global, _ := s.GetBgp(context.Background(), &api.GetBgpRequest{})

	// Display global configuration in the specified format
	fmt.Printf("AS:        %d\n", global.Global.Asn)
	fmt.Printf("Router-ID: %s\n", global.Global.RouterId)
	fmt.Printf("Listening Port: %d, Addresses: 0.0.0.0, ::\n", global.Global.ListenPort)
}

func displayNeighbors(s *server.BgpServer) {
	// Print header
	fmt.Printf("%-15s %-5s %-10s %s\n", "Peer", "AS", "State", "#Received  Accepted")
	fmt.Printf("%-15s %-5s %-10s %s\n", "----", "--", "-----", "---------  --------")

	// Get list of peers
	s.ListPeer(context.Background(), &api.ListPeerRequest{}, func(peer *api.Peer) {
		// Format state
		state := "None"
		if peer.State.SessionState == api.PeerState_ESTABLISHED {
			state = "Establ"
		}

		// Print peer information
		fmt.Printf("%-15s %-5d %-10s %9d %9d\n",
			peer.Conf.NeighborAddress,
			peer.Conf.PeerAsn,
			state,
			peer.AfiSafis[0].State.Received,
			peer.AfiSafis[0].State.Accepted)
	})
}

func displayRIB(s *server.BgpServer) {
	// Print header
	fmt.Printf("%-20s %-20s %-20s \n", "Network", "Next Hop", "AS_PATH")
	fmt.Printf("%-20s %-20s %-20s \n", "-------", "--------", "-------")

	// Get RIB
	s.ListPath(context.Background(), &api.ListPathRequest{
		TableType: api.TableType_GLOBAL,
		Family:    &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
	}, func(dest *api.Destination) {
		for _, path := range dest.Paths {
			// Decode NLRI
			nlri, _ := apiutil.UnmarshalNLRI(bgp.RF_IPv4_UC, path.Nlri)
			prefix := nlri.String()

			// Decode attributes
			attrs, _ := apiutil.UnmarshalPathAttributes(path.Pattrs)
			var nextHop string
			var asPath string

			for _, attr := range attrs {
				switch a := attr.(type) {
				case *bgp.PathAttributeNextHop:
					nextHop = a.Value.String()
				case *bgp.PathAttributeAsPath:
					asPath = a.String()
				}
			}

			// Print path information
			fmt.Printf("*> %-20s %-20s %-20s \n",
				prefix,
				nextHop,
				asPath)
		}
	})
}
