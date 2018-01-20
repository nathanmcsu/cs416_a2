package dfslib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"

	"../metadata"
	"../sharedData"
)

// Server to Client RPC

type ServerToClient rpc.Client

func (t *ServerToClient) CheckLocalStorage(fileName string, retBool *bool) error {
	return nil
}
func (t *ServerToClient) GetFileChunk(fileChunkMsg sharedData.FileChunkMessage, chunk *Chunk) error {
	log.Println("GetFileChunk")
	conn := metadata.ClientMap[fileChunkMsg.ClientID]

	file, _ := os.Open(conn.ClientPath + fileChunkMsg.Fname + ".dfs")
	data, _ := ioutil.ReadAll(file)
	var testChunk [256]Chunk
	json.Unmarshal(data, &testChunk)

	*chunk = testChunk[fileChunkMsg.ChunkIndex]

	return nil
}
