package dfslib

import (
	"io/ioutil"
	"log"
	"net/rpc"
	"os"

	"../sharedData"
)

// Server to Client RPC

type ServerToClient rpc.Client

func (t *ServerToClient) CheckLocalStorage(fileName string, retBool *bool) error {
	return nil
}
func (t *ServerToClient) GetFile(fileChunkMsg sharedData.GetFileMessage, chunk *[]byte) error {
	log.Println("GetFileChunk, Chunk Index: ", fileChunkMsg.ChunkIndex)
	file, _ := os.Open(fileChunkMsg.ClientPath + fileChunkMsg.Fname + ".dfs")
	data, _ := ioutil.ReadAll(file)
	*chunk = data
	return nil
}
