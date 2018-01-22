package dfslib

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"../sharedData"
)

type File struct {
	FName         string
	FileChunks    [256]Chunk
	ChunkVersions [256]int
	Mode          FileMode
	ClientConn    ConnDFS
}

func (t File) Read(chunkNum uint8, chunk *Chunk) error {
	// Check if Read mode or dread mode
	//		if dread, return chunk
	//	if read:
	//
	// TODO:
	// 		Check if there is a write, block until the write is done
	// 		After done, check if current version is latest chunk version
	//

	return nil
}
func (t File) Write(chunkNum uint8, chunk *Chunk) error {
	log.Println("Write")

	if t.Mode != WRITE {
		return BadFileModeError(t.Mode)
	}
	// DisconnectedError, check if Server is up
	var isConnected bool
	storedDFSMessage := &sharedData.StoredDFSMessage{
		ClientIP:   t.ClientConn.ClientIP,
		ClientID:   t.ClientConn.ClientID,
		ClientPath: t.ClientConn.ClientPath,
	}

	err := t.ClientConn.ServerRPC.Call("ClientToServer.SyncHeartBeat", storedDFSMessage, &isConnected)

	if err != nil || !isConnected {
		return DisconnectedError(t.ClientConn.ServerIP)
	}

	// Steps:
	//		-Tell server you are writing, to block any reads  TODO
	// 		-Write locally first
	//		-Write to Log File
	//			-Write to Server
	//		-Get acknowledge of updated version (Check version number versus our own?)
	// 		-Delete our log if version update, if not resend write request

	//Tell server you are writing, to block any reads
	writeChunkMessage := &sharedData.WriteChunkMessage{
		FName:        t.FName,
		ChunkIndex:   int(chunkNum),
		ChunkVersion: t.ChunkVersions[chunkNum],
		ClientID:     t.ClientConn.ClientID,
	}

	var canWrite bool
	t.ClientConn.ServerRPC.Call("ClientToServer.BlockChunk", writeChunkMessage, &canWrite)
	if !canWrite {
		return DisconnectedError(t.ClientConn.ServerIP)
	}

	//Write Locally
	file, err := os.OpenFile(t.ClientConn.ClientPath+t.FName+".dfs", os.O_WRONLY, os.ModePerm)
	defer file.Close()
	if err != nil {
		log.Println(err)
	}
	data, _ := ioutil.ReadAll(file)
	var fileChunks [256]Chunk
	json.Unmarshal(data, &fileChunks)

	fileChunks[chunkNum] = *chunk
	fileBytes, _ := json.Marshal(fileChunks)
	_, err = file.Write(fileBytes)
	if err != nil {
		log.Println(err)
	}
	file.Sync()

	// Write Log
	metadata, err := os.OpenFile(t.ClientConn.ClientPath+"metadata.dfs", os.O_WRONLY, os.ModePerm)
	defer metadata.Close()
	if err != nil {
		log.Println(err)
	}

	data, _ = ioutil.ReadAll(metadata)
	log.Println(data)
	var readmeta ClientMetaData
	json.Unmarshal(data, &readmeta)
	log.Println(readmeta)
	if readmeta.WriteLogs == nil {
		readmeta.WriteLogs = make(map[string]int)
	}
	readmeta.WriteLogs[t.FName] = t.ChunkVersions[chunkNum] + 1

	readmetaByte, _ := json.MarshalIndent(readmeta, "", " ")
	metadata.Write(readmetaByte)
	metadata.Sync()

	// Write to Server
	writeChunkMessage = &sharedData.WriteChunkMessage{
		FName:        t.FName,
		ChunkIndex:   int(chunkNum),
		ChunkVersion: t.ChunkVersions[chunkNum],
		ChunkByte:    *chunk,
		ClientID:     t.ClientConn.ClientID,
	}
	var resChunkMessage sharedData.WriteChunkMessage
	t.ClientConn.ServerRPC.Call("ClientToServer.WriteChunk", writeChunkMessage, &resChunkMessage)

	if t.ChunkVersions[chunkNum] >= resChunkMessage.ChunkVersion {
		log.Println("Error in Writing, new version: ", resChunkMessage.ChunkVersion)
	}
	t.ChunkVersions[chunkNum] = resChunkMessage.ChunkVersion

	// Get rid of log entry TODO: Clean up, its deleting and remaking, find a way to overwrite
	metadata2, err := os.OpenFile(t.ClientConn.ClientPath+"metadata.dfs", os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	data, _ = ioutil.ReadAll(metadata2)
	json.Unmarshal(data, &readmeta)
	metadata.Close()
	os.Remove(t.ClientConn.ClientPath + "metadata.dfs")

	log.Println(readmeta)
	metadata, err = os.Create(t.ClientConn.ClientPath + "metadata.dfs")
	defer metadata.Close()
	delete(readmeta.WriteLogs, t.FName)

	readmetaByte, _ = json.MarshalIndent(readmeta, "", " ")
	metadata.Write(readmetaByte)
	metadata.Sync()

	return nil
}
func (t File) Close() error {
	var isConnected bool
	storedDFSMessage := &sharedData.StoredDFSMessage{
		ClientIP:   t.ClientConn.ClientIP,
		ClientID:   t.ClientConn.ClientID,
		ClientPath: t.ClientConn.ClientPath,
	}

	err := t.ClientConn.ServerRPC.Call("ClientToServer.SyncHeartBeat", storedDFSMessage, &isConnected)

	if err != nil || !isConnected {
		return DisconnectedError(t.ClientConn.ServerIP)
	}

	writerAndFile := &sharedData.WriterAndFile{
		Fname:    t.FName,
		ClientID: t.ClientConn.ClientID,
	}

	var isOk bool
	t.ClientConn.ServerRPC.Call("ClientToServer.CloseFile", writerAndFile, &isOk)

	return nil
}
