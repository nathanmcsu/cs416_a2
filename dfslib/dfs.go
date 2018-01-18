package dfslib

import (
	"net/rpc"
)

type ConnDFS struct {
	serverRPC *rpc.Client
	clientRPC *rpc.Client
}

func (t *ConnDFS) LocalFileExists(fname string) (exists bool, err error) {
	return false, nil
}

func (t *ConnDFS) GlobalFileExists(fname string) (exists bool, err error) {
	var isExists bool
	t.serverRPC.Call("ClientToServer.CheckGlobalFileExists", fname, &isExists)
	return isExists, nil
}

func (t *ConnDFS) Open(fname string, mode FileMode) (f DFSFile, err error) {

	return nil, FileDoesNotExistError("Really bad")
}
func (t *ConnDFS) UMountDFS() {

}
