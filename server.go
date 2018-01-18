package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"./dfslib"
	"./metadata"
	"./rpcDefs"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("Usage: server.go [Client-incoming UDP ip:port]")
	}
	localPort := args[0]
	server := rpc.NewServer()

	clientToServerRPC := new(rpcDefs.ClientToServer)

	server.Register(clientToServerRPC)
	tcpConn, err := net.Listen("tcp", localPort)
	if err != nil {
		log.Fatal("Failed to establish TCP:", err)
	}
	fmt.Println("Listening on: ", localPort)

	FileMap := make(map[string]dfslib.File)

	metadata.FileMap = FileMap
	server.Accept(tcpConn)

}
