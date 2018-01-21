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
	// Check if currnet File has the most updated ChunkVersion
	return nil
}
func (t File) Write(chunkNum uint8, chunk *Chunk) error {
	log.Println("Write")

	if t.Mode != WRITE {
		return BadFileModeError(t.Mode)
	}
	// TODO : DisconnectedError, check if Server is up

	// Steps:
	//		-Tell server you are writing, to block any reads  TODO
	// 		-Write locally first
	//		-Write to Log File
	//			-Write to Server
	//		-Get acknowledge of updated version (Check version number versus our own?)
	// 		-Delete our log if version update, if not resend write request

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
	var readmeta ClientMetaData
	json.Unmarshal(data, &readmeta)
	if readmeta.WriteLogs == nil {
		readmeta.WriteLogs = make(map[string]int)
	}
	readmeta.WriteLogs[t.FName] = t.ChunkVersions[chunkNum] + 1

	readmetaByte, _ := json.MarshalIndent(readmeta, "", " ")
	metadata.Write(readmetaByte)
	metadata.Sync()

	// Write to Server
	writeChunkMessage := &sharedData.WriteChunkMessage{
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
	metadata, err = os.OpenFile(t.ClientConn.ClientPath+"metadata.dfs", os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	data, _ = ioutil.ReadAll(metadata)
	json.Unmarshal(data, &readmeta)
	metadata.Close()
	os.Remove(t.ClientConn.ClientPath + "metadata.dfs")

	metadata, err = os.Create(t.ClientConn.ClientPath + "metadata.dfs")
	defer metadata.Close()
	delete(readmeta.WriteLogs, t.FName)

	readmetaByte, _ = json.MarshalIndent(readmeta, "", " ")
	metadata.Write(readmetaByte)
	metadata.Sync()

	return nil
}
func (t File) Close() error {
	return nil
}
