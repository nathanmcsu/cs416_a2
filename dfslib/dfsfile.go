package dfslib

type File struct {
	Version string
}

func (t *File) Read(chunkNum uint8, chunk *Chunk) error {
	return nil
}
func (t *File) Write(chunkNum uint8, chunk *Chunk) error {
	return nil
}
