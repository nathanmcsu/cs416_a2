package dfslib

import (
	"net/rpc"
)

type ConnDFS struct {
	ServerRPC *rpc.Client
	ClientRPC *rpc.Client
	IsOffline bool
	ClientID  int
}

func (t *ConnDFS) LocalFileExists(fname string) (exists bool, err error) {
	return false, nil
}

func (t *ConnDFS) GlobalFileExists(fname string) (exists bool, err error) {
	var isExists bool
	t.ServerRPC.Call("ClientToServer.CheckGlobalFileExists", fname, &isExists)
	return isExists, nil
}

func (t *ConnDFS) Open(fname string, mode FileMode) (f DFSFile, err error) {

	return nil, FileDoesNotExistError("Really bad")
}
func (t *ConnDFS) UMountDFS() {

}
