//
// run_machine_server.go
// Copyright (C) 2020 Toran Sahu <toran.sahu@yahoo.com>
//
// Distributed under terms of the MIT license.
//

package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	sqls "github.com/lighttiger2505/sqls/pkg/app"
	"github.com/toransahu/grpc-eg-go/machine"
	"github.com/toransahu/grpc-eg-go/server"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 9111, "Port on which gRPC server should listen TCP conn.")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	LSP := sqls.Serve("log.txt", "config.yml", false)
	machine.RegisterMachineServer(grpcServer, &server.MachineServer{})
	machine.RegisterLspServer(grpcServer, &server.LspServer{LSP: LSP})

	grpcServer.Serve(lis)

	log.Printf("Initializing gRPC server on port %d", *port)
}
