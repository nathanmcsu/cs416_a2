package rpcDefs

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strconv"

	"../metadata"
)

// Client to Server RPC

type ClientToServer rpc.Client

func (t *ClientToServer) CheckGlobalFileExists(fname string, exists *bool) error {
	_, ok := metadata.FileMap[fname]
	*exists = ok
	return nil
}
func (t *ClientToServer) GetNewCID(localPath string, cid *int) error {
	newCID := len(metadata.ClientMap)
	*cid = newCID
	return nil
}

func (t *ClientToServer) MapAliveClient(cid int, total *int) error {
	metadata.ClientMap[cid] = true
	size := len(metadata.ClientMap)
	*total = size
	return nil
}

func (t *ClientToServer) CreateListenerClient(clientAddr string, clientConn *string) error {

	clientServer := rpc.NewServer()
	clientToServer := new(ClientToServer)
	clientServer.Register(clientToServer)

	testClientAddr := clientAddr + ":0"
	addr, err := net.ResolveTCPAddr("tcp", testClientAddr)
	if err != nil {
		fmt.Println(err)
	}

	tcpConn, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Println(err)
	}

	go clientServer.Accept(tcpConn)

	clientAddr = clientAddr + ":" + strconv.Itoa(tcpConn.Addr().(*net.TCPAddr).Port)
	*clientConn = clientAddr

	return nil
}
