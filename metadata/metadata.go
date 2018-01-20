package metadata

import "../sharedData"

// FileMap has two nested maps.  First map: FileName - [map Chunk number - map (cid - version number)]
var FileMap map[string]map[int]map[int]int

// FileName and CID using it
var ActiveFiles map[string]int

// CID and if active or not
var ClientMap map[int]sharedData.StoredDFSMessage
