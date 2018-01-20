package rpcDefs

import (
	"log"
	"net"
	"net/rpc"

	"../dfslib"
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

func (t *ClientToServer) MapAliveClient(storedDFS sharedData.StoredDFSMessage, total *int) error {
	log.Println("MapAliveClient")
	metadata.ClientMap[storedDFS.ClientID] = storedDFS
	size := len(metadata.ClientMap)
	*total = size
	return nil
}

func (t *ClientToServer) RetrieveLatestFile(fname string, dfsFile *dfslib.File) error {
	log.Println("RetrieveLatestFile")
	tempDFSFile := dfslib.File{Name: fname}

	chunkMap := metadata.FileMap[fname]

	// Get the latest version of each Chunk
	for chunkIndex, versionMap := range chunkMap {
		var highestV int
		for cid, v := range versionMap {
			if _, present := metadata.ClientMap[cid]; present && v >= highestV {
				// client with latest chunk is present
				clientIP := metadata.ClientMap[cid]
				var fileChunkMsg sharedData.FileChunkMessage
				fileChunkMsg.Fname = tempDFSFile.Name
				fileChunkMsg.ChunkIndex = chunkIndex
				fileChunkMsg.ClientID = cid

				conn, err := net.Dial("tcp", clientIP.ClientRPC)
				if err != nil {
					log.Println("Failed to establish connection to client: ", clientIP.ClientRPC)
				}
				rpc.NewClient(conn).Call("ServerToClient.GetFileChunk", fileChunkMsg, &tempDFSFile.FileChunks[chunkIndex])

			}
		}
	}

	*dfsFile = tempDFSFile
	return nil
}

func (t *ClientToServer) AddNewFile(fileCMap dfslib.FileChunkMap, argok *bool) error {
	log.Println("AddNewFile")
	_, exists := metadata.FileMap[fileCMap.Name]
	if !exists {
		metadata.FileMap[fileCMap.Name] = fileCMap.ChunkMap
	}
	_, *argok = metadata.FileMap[fileCMap.Name]
	return nil
}

func (t *ClientToServer) CreateListenerClient(clientAddr string, x *bool) error {
	log.Println("CreateListenerClient")
	//TODO
	return nil
}
