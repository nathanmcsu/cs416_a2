package sharedData

import "net/rpc"

type StoredDFSMessage struct {
	ClientIP   string
	ClientID   int
	ClientPath string
}
type StoredDFS struct {
	ClientIP   string
	ClientID   int
	ClientPath string
	ClientRPC  *rpc.Client
}
type GetFileMessage struct {
	ClientID   int
	Fname      string
	ChunkIndex int
	ClientPath string
}

type WriterAndFile struct {
	Fname    string
	ClientID int
}

// Entry for FileMap metadata
type FileChunkMap struct {
	FName    string
	ChunkMap map[int]map[int]int
}

// Shared File Struct
type ArgFile struct {
	FName         string
	FileChunks    [256][32]byte
	ChunkVersions [256]int
}

type ReplicaEntry struct {
	ClientID       int
	VersionEntries [256]int
	Fname          string
}
