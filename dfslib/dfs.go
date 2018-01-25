package dfslib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"

	"../sharedData"
)

type ConnDFS struct {
	ServerRPC   *rpc.Client
	ClientRPC   *rpc.Client
	IsOffline   bool
	ClientID    int
	ClientPath  string
	ServerIP    string
	ClientIP    string
	ClientUDPIP string
}

func (t *ConnDFS) LocalFileExists(fname string) (exists bool, err error) {
	return false, nil
}

func (t *ConnDFS) GlobalFileExists(fname string) (exists bool, err error) {
	if t.IsOffline {
		return false, DisconnectedError(t.ServerIP)
	}
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
	dfsFile.Mode = mode
	dfsFile.FName = fname
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
				err = t.ServerRPC.Call("ClientToServer.RetrieveLatestFile", fname, &argFile)

				if len(argFile.FName) == 0 {
					return nil, FileUnavailableError(fname)
				}

				if err != nil {
					log.Println(err)
				} else {
					// Make client replica for file
					file, _ := os.Create(t.ClientPath + fname + ".dfs")
					defer file.Close()

					fileBytes, _ := json.Marshal(argFile.FileChunks)
					file.Write(fileBytes)
					file.Sync()

					file, _ = os.Open(t.ClientPath + fname + ".dfs")
					defer file.Close()

					data, _ := ioutil.ReadAll(file)
					var testChunk [256]Chunk
					json.Unmarshal(data, &testChunk)

					dfsFile.FName = fname
					dfsFile.FileChunks = testChunk
					dfsFile.ChunkVersions = argFile.ChunkVersions

					var versionEntries [256]int
					for i := 0; i < len(versionEntries); i++ {
						versionEntries[i] = argFile.ChunkVersions[i]
					}

					replicaEntryMessage := &sharedData.ReplicaEntry{
						ClientID:       t.ClientID,
						VersionEntries: versionEntries,
						Fname:          fname,
					}
					var ok bool
					t.ServerRPC.Call("ClientToServer.AddNewReplica", replicaEntryMessage, &ok)

				}
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
				defer file.Close()
				data, _ := ioutil.ReadAll(file)
				var testChunk [256]Chunk
				json.Unmarshal(data, &testChunk)

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
	} else {
		// Offline case should be only dread mode too
		file, err := os.Open(t.ClientPath + fname + ".dfs")
		defer file.Close()
		if err != nil {
			log.Println(err)
			return nil, FileDoesNotExistError(fname)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Println(err)
		}
		var fileChunks [256]Chunk
		json.Unmarshal(data, &fileChunks)
		dfsFile.FileChunks = fileChunks

	}
	//  TODO
	//		Write File contents to disk after successful acquire
	dfsFile.ClientConn = *t
	return dfsFile, nil
}
func (t *ConnDFS) UMountDFS() error {
	var isConnected bool
	storedDFSMessage := &sharedData.StoredDFSMessage{
		ClientIP:   t.ClientIP,
		ClientID:   t.ClientID,
		ClientPath: t.ClientPath,
	}

	if !t.IsOffline {

		err := t.ServerRPC.Call("ClientToServer.SyncHeartBeat", storedDFSMessage, &isConnected)

		if err != nil || !isConnected {
			return DisconnectedError(t.ServerIP)
		}
		var isOk bool
		t.ServerRPC.Call("ClientToServer.CloseConnection", t.ClientID, &isOk)

		ClientConnData.rpcConn.Close()
		// ClientConnData.udpConn.Close()
	}
	return nil
}
