package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"

	"./dfslib"
)

func registerFile(server *rpc.Server, file dfslib.DFSFile) {
	server.Register(file)
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("Usage: server.go [Client-incoming UDP ip:port]")
	}
	localPort := args[0]
	server := rpc.NewServer()

	file := new(dfslib.File)
	// connDFS := new(dfslib.ConnDFS)
	server.Register(file)
	tcpConn, err := net.Listen("tcp", localPort)
	if err != nil {
		log.Fatal("Failed to establish TCP:", err)
	}
	fmt.Println("Listening on: ", localPort)
	server.Accept(tcpConn)

}
