package rpcDefs

import (
	"net/rpc"
)

// Server to Client RPC

type ServerToClient rpc.Client

func (t *ServerToClient) checkLocalStorage(fileName string, retBool *bool) error {
	return nil
}
