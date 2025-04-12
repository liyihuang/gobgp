package main

import (
	"github.com/osrg/gobgp/v3/pkg/server"
)

func main() {
	s := server.NewBgpServer(
		server.GrpcListenAddress("127.0.0.1:50051"),
	)
	go s.Serve()
	select {}
}
