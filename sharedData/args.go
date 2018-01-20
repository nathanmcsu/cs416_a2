package sharedData

type StoredDFSMessage struct {
	ClientRPC  string
	ClientID   int
	ClientPath string
}
type FileChunkMessage struct {
	ClientID   int
	Fname      string
	ChunkIndex int
}
