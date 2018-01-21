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

func (t *ClientToServer) MapAliveClient(storedDFSMsg sharedData.StoredDFSMessage, total *int) error {
	log.Println("MapAliveClient")
	fmt.Println("ClientID: ", storedDFSMsg.ClientID)
	storedDFS := &sharedData.StoredDFS{ClientID: storedDFSMsg.ClientID, ClientIP: storedDFSMsg.ClientIP, ClientPath: storedDFSMsg.ClientPath}
	conn, err := net.Dial("tcp", storedDFS.ClientIP)
	if err != nil {
		log.Println("Failed to establish connection to client: ", storedDFS.ClientIP)
	}
	rpcConn := rpc.NewClient(conn)
	storedDFS.ClientRPC = rpcConn

	metadata.ClientMap[storedDFS.ClientID] = *storedDFS
	size := len(metadata.ClientMap)
	fmt.Println(metadata.ClientMap)
	*total = size
	return nil
}

func (t *ClientToServer) RetrieveLatestFile(fname string, argFile *sharedData.ArgFile) error {
	log.Println("RetrieveLatestFile")
	tempDFSFile := sharedData.ArgFile{FName: fname}
	chunkMap := metadata.FileMap[fname]

	tempClientFile := make(map[int][256][32]byte)

	// Get the latest version of each Chunk
	for chunkIndex, versionMap := range chunkMap {
		highestV := -1
		for cid, v := range versionMap {
			if _, present := metadata.ClientMap[cid]; present && v >= highestV {
				// client with latest chunk is present
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
	*argFile = tempDFSFile
	return nil
}

func (t *ClientToServer) AddNewFile(fileCMap sharedData.FileChunkMap, argok *bool) error {
	log.Println("AddNewFile")
	_, exists := metadata.FileMap[fileCMap.FName]
	if !exists {
		metadata.FileMap[fileCMap.FName] = fileCMap.ChunkMap
	}
	_, *argok = metadata.FileMap[fileCMap.FName]
	return nil
}

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
