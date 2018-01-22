package metadata

import (
	"../sharedData"
)

// FileMap has two nested maps.  First map: FileName - [map Chunk number - map (cid - version number)]
var FileMap map[string]map[int]map[int]int

// FileName and CID using it
var ActiveFiles map[string]int

// FileName and map [Chunk #] and CID
var ActiveWriteChunks map[string]map[int]int

// Active Client Map
var ActiveClientMap map[int]bool

// CID and information
var ClientMap map[int]sharedData.StoredDFS
