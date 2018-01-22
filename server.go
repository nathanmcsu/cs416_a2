package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"./metadata"
	"./rpcDefs"
	"./sharedData"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	metadata.FileMap = make(map[string]map[int]map[int]int)
	metadata.ClientMap = make(map[int]sharedData.StoredDFS)
	metadata.ActiveFiles = make(map[string]int)
	metadata.ActiveWriteChunks = make(map[string]map[int]int)
	metadata.ActiveClientMap = make(map[int]bool)
	server.Accept(tcpConn)

}
