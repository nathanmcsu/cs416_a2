package rpcDefs

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"

	"../metadata"
	"../sharedData"
)

// Client to Server RPC

type ClientToServer rpc.Client

func (t *ClientToServer) CheckGlobalFileExists(fname string, exists *bool) error {
	log.Println("GlobalFileExists")
	_, ok := metadata.FileMap[fname]
	*exists = ok
	return nil
}
func (t *ClientToServer) GetNewCID(localPath string, cid *int) error {
	log.Println("GetNewCID")
	newCID := len(metadata.ClientMap)
	*cid = newCID
	return nil
}

// Puts active clients in metadata and stores information
func (t *ClientToServer) MapAliveClient(storedDFSMsg sharedData.StoredDFSMessage, total *int) error {
	log.Println("MapAliveClient")
	storedDFS := &sharedData.StoredDFS{ClientID: storedDFSMsg.ClientID, ClientIP: storedDFSMsg.ClientIP, ClientPath: storedDFSMsg.ClientPath}
	conn, err := net.Dial("tcp", storedDFS.ClientIP)
	if err != nil {
		log.Println("Failed to establish connection to client: ", storedDFS.ClientIP)
	}
	rpcConn := rpc.NewClient(conn)
	storedDFS.ClientRPC = rpcConn

	metadata.ClientMap[storedDFS.ClientID] = *storedDFS
	size := len(metadata.ClientMap)

	metadata.ActiveClientMap[storedDFS.ClientID] = true

	*total = size
	return nil
}

// Retrieve the latest possible file, only checks version and if client is active, might be stale
func (t *ClientToServer) RetrieveLatestFile(fname string, argFile *sharedData.ArgFile) error {
	log.Println("RetrieveLatestFile")
	tempDFSFile := sharedData.ArgFile{FName: fname}
	chunkMap := metadata.FileMap[fname]

	tempClientFile := make(map[int][256][32]byte)

	log.Println(chunkMap)
	log.Println(metadata.ActiveClientMap)

	// Get the latest version of each Chunk
	for chunkIndex, versionMap := range chunkMap {
		highestV := -1
		for cid, v := range versionMap {
			if _, present := metadata.ActiveClientMap[cid]; present && v >= highestV {
				// client with latest chunk is present
				highestV = v
				chunks, exists := tempClientFile[cid]
				if exists {
					tempDFSFile.FileChunks[chunkIndex] = chunks[chunkIndex]
					tempDFSFile.ChunkVersions[chunkIndex] = v
				} else {
					client := metadata.ClientMap[cid]
					var fileMessage sharedData.GetFileMessage
					fileMessage.Fname = tempDFSFile.FName
					fileMessage.ChunkIndex = chunkIndex
					fileMessage.ClientID = cid
					fileMessage.ClientPath = client.ClientPath

					//fmt.Println("Calling client: ", cid, "with IP: ", client.ClientIP)
					//fmt.Println("Chunk Index: ", chunkIndex)

					var clientFile []byte
					var errOrTime bool
					client.ClientRPC.Call("ServerToClient.GetFile", fileMessage, &clientFile)

					// TIMEOUT function, might need for later: TODO
					// c := make(chan error, 1)
					// go func() {
					// 	c <- client.ClientRPC.Call("ServerToClient.GetFileChunk", fileChunkMsg, &clientFile)
					// }()
					// select {
					// case err := <-c:
					// 	// use err and result
					// 	errOrTime = true
					// 	fmt.Println(err)
					// case <-time.After(5000000):
					// 	// call timed out
					// 	errOrTime = true
					// 	fmt.Println("Timed out")
					// }
					if !errOrTime {
						var fileChunks [256][32]byte
						json.Unmarshal(clientFile, &fileChunks)
						tempClientFile[cid] = fileChunks
						tempDFSFile.FileChunks[chunkIndex] = fileChunks[chunkIndex]
						tempDFSFile.ChunkVersions[chunkIndex] = v
					}
				}
			}
		}
	}
	// log.Println(tempDFSFile.FileChunks)
	log.Println(tempDFSFile)
	*argFile = tempDFSFile
	return nil
}

// Add new file to file map
func (t *ClientToServer) AddNewFile(fileCMap sharedData.FileChunkMap, argok *bool) error {
	log.Println("AddNewFile")
	_, exists := metadata.FileMap[fileCMap.FName]
	if !exists {
		metadata.FileMap[fileCMap.FName] = fileCMap.ChunkMap
	}
	_, *argok = metadata.FileMap[fileCMap.FName]
	return nil
}

// Adds a new client to file map
func (t *ClientToServer) AddNewReplica(replicaEntryMessage sharedData.ReplicaEntry, argok *bool) error {
	log.Println("AddNewReplica")
	chunkMap := metadata.FileMap[replicaEntryMessage.Fname]
	for chunkIndex, versionMap := range chunkMap {
		versionMap[replicaEntryMessage.ClientID] = replicaEntryMessage.VersionEntries[chunkIndex]
	}
	return nil
}

func (t *ClientToServer) CreateListenerClient(clientAddr string, x *bool) error {
	log.Println("CreateListenerClient")
	// TODO: clear if not needed
	return nil
}

// Check and gets lock for file to write
func (t *ClientToServer) CheckWriterExistsAndAdd(writerAndFile sharedData.WriterAndFile, writerExists *bool) error {
	log.Println("CheckWriterExistsAndAdd")
	_, exists := metadata.ActiveFiles[writerAndFile.Fname]
	if !exists {
		// Add writer to ActiveFiles
		metadata.ActiveFiles[writerAndFile.Fname] = writerAndFile.ClientID
		fmt.Println(metadata.ActiveFiles)
	}
	*writerExists = exists
	return nil
}

// Update metadata with new chunk version and unlock chunk for other readers
func (t *ClientToServer) WriteChunk(writeChunkMessage sharedData.WriteChunkMessage, resChunkMessage *sharedData.WriteChunkMessage) error {
	log.Println("WriteChunk")
	metadata.FileMap[writeChunkMessage.FName][writeChunkMessage.ChunkIndex][writeChunkMessage.ClientID] = writeChunkMessage.ChunkVersion + 1

	*resChunkMessage = sharedData.WriteChunkMessage{
		FName:        writeChunkMessage.FName,
		ChunkIndex:   writeChunkMessage.ChunkIndex,
		ChunkVersion: writeChunkMessage.ChunkVersion + 1,
		ClientID:     writeChunkMessage.ClientID,
	}
	// Unlock chunk for reads
	activeChunksMap, _ := metadata.ActiveWriteChunks[writeChunkMessage.FName]
	delete(activeChunksMap, writeChunkMessage.ChunkIndex)

	return nil
}

// Single heartbeat to check if still connected
func (t *ClientToServer) SyncHeartBeat(storedDFSMessage sharedData.StoredDFSMessage, isConnected *bool) error {
	log.Println("SyncHeartBeat")

	if _, exists := metadata.ActiveClientMap[storedDFSMessage.ClientID]; exists {
	} else {
		metadata.ActiveClientMap[storedDFSMessage.ClientID] = true
	}
	*isConnected = true
	return nil
}

//Lock chunk for a write
func (t *ClientToServer) BlockChunk(chunkMessage sharedData.WriteChunkMessage, canWrite *bool) error {
	log.Println("BlockChunk")

	_, exists := metadata.ActiveClientMap[chunkMessage.ClientID] // Check if connected
	_, hasLock := metadata.ActiveFiles[chunkMessage.FName]       // Check if it has lock on file to write
	if exists && hasLock {
		chunkMap, exists := metadata.ActiveWriteChunks[chunkMessage.FName]
		if !exists {
			chunkMap = make(map[int]int)
			chunkMap[chunkMessage.ChunkIndex] = chunkMessage.ClientID
			metadata.ActiveWriteChunks[chunkMessage.FName] = chunkMap
		} else {
			chunkMap[chunkMessage.ChunkIndex] = chunkMessage.ClientID
		}
		*canWrite = true
	} else {
		*canWrite = false
	}
	return nil
}

// Unlocks file for other writes
func (t *ClientToServer) CloseFile(writerAndFile sharedData.WriterAndFile, isOk *bool) error {
	log.Println("CloseFile")
	// There shouldn't be any in Active Write Chunks, but checking just in case:
	chunkMap, exists := metadata.ActiveWriteChunks[writerAndFile.Fname]
	if exists {
		for chunkIndex, cid := range chunkMap {
			if cid == writerAndFile.ClientID {
				delete(chunkMap, chunkIndex)
			}
		}
	}

	// Release File lock
	if cid, exists := metadata.ActiveFiles[writerAndFile.Fname]; exists {
		if cid == writerAndFile.ClientID {
			delete(metadata.ActiveFiles, writerAndFile.Fname)
		}
	}

	*isOk = true
	return nil
}

// Removes client from active client list
func (t *ClientToServer) CloseConnection(clientID int, isOk *bool) error {
	log.Println("CloseConnection")
	delete(metadata.ActiveClientMap, clientID)

	// There shouldn't be any in Active Write Chunks, but checking just in case:
	for _, chunkMap := range metadata.ActiveWriteChunks {
		for chunkIndex, cid := range chunkMap {
			if cid == clientID {
				delete(chunkMap, chunkIndex)
			}
		}
	}

	// Release File lock if any are still there
	for fname, cid := range metadata.ActiveFiles {
		if cid == clientID {
			delete(metadata.ActiveFiles, fname)
		}
	}

	*isOk = true
	return nil
}
