package dfslib

type File struct {
	FName         string
	FileChunks    [256]Chunk
	ChunkVersions [256]int
}

func (t File) Read(chunkNum uint8, chunk *Chunk) error {
	return nil
}
func (t File) Write(chunkNum uint8, chunk *Chunk) error {
	return nil
}
func (t File) Close() error {
	return nil
}
