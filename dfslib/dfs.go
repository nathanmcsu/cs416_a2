package dfslib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"

	"../sharedData"
)

type ConnDFS struct {
	ServerRPC  *rpc.Client
	ClientRPC  *rpc.Client
	IsOffline  bool
	ClientID   int
	ClientPath string
	ServerIP   string
	ClientIP   string
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

	// TODO:
	//		-validate file name (BadFilenameError)
	// 		-if IsOffline true, only DREAD mode allowed (Done, DisconnectedError)
	// 		-if two writers, only one can succeed (Done OpenWriteConflictError)

	// Populate dfsFile with the right data, dfsFile is return
	var dfsFile File

	if t.IsOffline && (mode == WRITE || mode == READ) {
		return nil, DisconnectedError(t.ServerIP)
	}

	if !t.IsOffline {
		// Online

		// Check if File exists first
		var fileExists bool
		err := t.ServerRPC.Call("ClientToServer.CheckGlobalFileExists", fname, &fileExists)
		if err != nil {
			fmt.Println(err)
			// Search file error
		}

		if mode == WRITE {
			// Check if writer is present
			var writerExists bool
			writerAndFile := &sharedData.WriterAndFile{Fname: fname, ClientID: t.ClientID}
			err = t.ServerRPC.Call("ClientToServer.CheckWriterExistsAndAdd", writerAndFile, &writerExists)
			if err != nil {
				fmt.Println(err)
			}
			if writerExists {
				return nil, OpenWriteConflictError(fname)
			}
		}

		// Fetch File if READ or WRITE first, isOffline and READ/Write already checked
		if mode == READ || mode == WRITE {

			if fileExists {
				// Retrieve file from server call
				var argFile sharedData.ArgFile
				t.ServerRPC.Call("ClientToServer.RetrieveLatestFile", fname, &argFile)
			} else {
				// Create file locally
				file, err := os.Create(t.ClientPath + fname + ".dfs")
				if err != nil {
					// Create file error
				}

				// need to input some weird data
				var chunk Chunk
				const str = "Hello friends!"
				copy(chunk[:], str)
				var blankChunk [256]Chunk
				for i := 0; i < len(blankChunk); i += 3 {
					blankChunk[i] = chunk
				}

				blankByte, _ := json.Marshal(blankChunk)
				fmt.Println(len(blankByte))
				fmt.Println(len(blankChunk))
				file.Write(blankByte)
				file.Sync()

				file, _ = os.Open(t.ClientPath + fname + ".dfs")
				data, _ := ioutil.ReadAll(file)
				var testChunk [256]Chunk
				json.Unmarshal(data, &testChunk)

				dfsFile.FName = fname
				dfsFile.FileChunks = testChunk

				cidVersionMap := make(map[int]int)
				cidVersionMap[t.ClientID] = 0
				chunkMap := make(map[int]map[int]int)
				for i := 0; i < 256; i++ {
					chunkMap[i] = cidVersionMap
					dfsFile.ChunkVersions[i] = 0
				}

				// Create new entry for FileMap
				fileCMap := new(sharedData.FileChunkMap)
				fileCMap.ChunkMap = chunkMap
				fileCMap.FName = fname

				var ok bool
				t.ServerRPC.Call("ClientToServer.AddNewFile", fileCMap, &ok)
			}
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

	//  TODO
	//		Write File contents to disk after successful acquire

	return nil, FileDoesNotExistError("Really bad")
}
func (t *ConnDFS) UMountDFS() {

}
