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
	localFileRead, err := os.Open(fileChunkMsg.ClientPath + fileChunkMsg.Fname + ".dfs")
	defer localFileRead.Close()
	if err != nil {
		log.Println(err)
	}
	data, _ := ioutil.ReadAll(localFileRead)

	*chunk = data
	// file, _ := os.Open(fileChunkMsg.ClientPath + fileChunkMsg.Fname + ".dfs")
	// data, _ := ioutil.ReadAll(file)
	// *chunk = data

	// var fileChunks [256][32]byte
	// json.Unmarshal(data, &fileChunks)
	// log.Println(fileChunks[0])

	return nil
}
