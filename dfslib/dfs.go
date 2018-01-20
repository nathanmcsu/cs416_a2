package dfslib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"
)

type ConnDFS struct {
	ServerRPC  *rpc.Client
	ClientRPC  *rpc.Client
	IsOffline  bool
	ClientID   int
	ClientPath string
}

type FileChunkMap struct {
	Name     string
	ChunkMap map[int]map[int]int
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

	// Populate dfsFile with the right data
	var dfsFile File

	// Fetch File if READ or WRITE first
	if mode == READ || mode == WRITE {
		var fileExists bool
		t.ServerRPC.Call("ClientToServer.CheckGlobalFileExists", fname, &fileExists)

		if err != nil {
			fmt.Println(err)
			// Search file error
		}
		if fileExists {
			// Retrieve file from server call
			t.ServerRPC.Call("ClientToServer.RetrieveLatestFile", fname, &dfsFile)
		} else {
			// Create file locally
			file, err := os.Create(t.ClientPath + fname + ".dfs")
			if err != nil {
				// Create file error
			}
			var blankChunk [256]Chunk
			blankByte, _ := json.Marshal(blankChunk)
			file.Write(blankByte)
			file.Sync()

			file, _ = os.Open(t.ClientPath + fname + ".dfs")
			data, _ := ioutil.ReadAll(file)
			var testChunk [256]Chunk
			json.Unmarshal(data, &testChunk)

			dfsFile.Name = fname
			dfsFile.FileChunks = testChunk

			cidVersionMap := make(map[int]int)
			cidVersionMap[t.ClientID] = 0
			chunkMap := make(map[int]map[int]int)
			for i := 0; i < 256; i++ {
				chunkMap[i] = cidVersionMap
				dfsFile.ChunkVersions[i] = 0
			}
			fileCMap := new(FileChunkMap)
			fileCMap.ChunkMap = chunkMap
			fileCMap.Name = fname

			var ok bool
			t.ServerRPC.Call("ClientToServer.AddNewFile", fileCMap, &ok)
		}
	}

	switch mode {
	case READ:

		fmt.Println("Read mode")
	case WRITE:
		// Acquire lock
		fmt.Println("Write Mode")
	case DREAD:
		// If server is up, acquire files
		// else search local for file
		fmt.Println("DREAD mode")
	}

	// Write File contents to disk after successful acquire

	return nil, FileDoesNotExistError("Really bad")
}
func (t *ConnDFS) UMountDFS() {

}
