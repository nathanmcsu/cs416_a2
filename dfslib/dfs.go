package dfslib

import "net/rpc"

type ConnDFS struct {
	client *rpc.Client
}

func (t *ConnDFS) LocalFileExists(fname string) (exists bool, err error) {
	return false, nil
}

func (t *ConnDFS) GlobalFileExists(fname string) (exists bool, err error) {
	return false, nil
}

func (t *ConnDFS) Open(fname string, mode FileMode) (exists bool, err error) {
	return false, nil
}
func (t *ConnDFS) UMountDFS() error {
	return nil
}
