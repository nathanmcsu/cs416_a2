/*

This package specifies the application's interface to the distributed
file system (DFS) system to be used in assignment 2 of UBC CS 416
2017W2.

*/

package dfslib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"

	"../sharedData"
)

type ClientMetaData struct {
	ClientID int
	// Fname and Chunk Index
	WriteLogs map[string]int
}

type ClientConnectionData struct {
	connDFS ConnDFS
}

// A Chunk is the unit of reading/writing in DFS.
type Chunk [32]byte

// Represents a type of file access.
type FileMode int

const (
	// Read mode.
	READ FileMode = iota

	// Read/Write mode.
	WRITE

	// Disconnected read mode.
	DREAD
)

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains serverAddr
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("DFS: Not connnected to server [%s]", string(e))
}

// Contains chunkNum that is unavailable
type ChunkUnavailableError uint8

func (e ChunkUnavailableError) Error() string {
	return fmt.Sprintf("DFS: Latest verson of chunk [%s] unavailable", string(e))
}

// Contains filename
type OpenWriteConflictError string

func (e OpenWriteConflictError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] is opened for writing by another client", string(e))
}

// Contains file mode that is bad.
type BadFileModeError FileMode

func (e BadFileModeError) Error() string {
	return fmt.Sprintf("DFS: Cannot perform this operation in current file mode [%s]", string(e))
}

// Contains filename
type BadFilenameError string

func (e BadFilenameError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] includes illegal characters or has the wrong length", string(e))
}

// Contains filename
type FileUnavailableError string

func (e FileUnavailableError) Error() string {
	return fmt.Sprintf("DFS: Filename [%s] is unavailable", string(e))
}

// Contains local path
type LocalPathError string

func (e LocalPathError) Error() string {
	return fmt.Sprintf("DFS: Cannot access local path [%s]", string(e))
}

// Contains filename
type FileDoesNotExistError string

func (e FileDoesNotExistError) Error() string {
	return fmt.Sprintf("DFS: Cannot open file [%s] in D mode as it does not exist locally", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a file in the DFS system.
type DFSFile interface {
	// Reads chunk number chunkNum into storage pointed to by
	// chunk. Returns a non-nil error if the read was unsuccessful.
	//
	// Can return the following errors:
	// - DisconnectedError (in READ,WRITE modes)
	// - ChunkUnavailableError (in READ,WRITE modes)
	Read(chunkNum uint8, chunk *Chunk) (err error)

	// Writes chunk number chunkNum from storage pointed to by
	// chunk. Returns a non-nil error if the write was unsuccessful.
	//
	// Can return the following errors:
	// - BadFileModeError (in READ,DREAD modes)
	// - DisconnectedError (in WRITE mode)
	Write(chunkNum uint8, chunk *Chunk) (err error)

	// Closes the file/cleans up. Can return the following errors:
	// - DisconnectedError
	Close() (err error)
}

// Represents a connection to the DFS system.
type DFS interface {
	// Check if a file with filename fname exists locally (i.e.,
	// available for DREAD reads).
	//
	// Can return the following errors:
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	LocalFileExists(fname string) (exists bool, err error)

	// Check if a file with filename fname exists globally.
	//
	// Can return the following errors:
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	// - DisconnectedError
	GlobalFileExists(fname string) (exists bool, err error)

	// Opens a filename with name fname using mode. Creates the file
	// in READ/WRITE modes if it does not exist. Returns a handle to
	// the file through which other operations on this file can be
	// made.
	//
	// Can return the following errors:
	// - OpenWriteConflictError (in WRITE mode)
	// - DisconnectedError (in READ,WRITE modes)
	// - FileUnavailableError (in READ,WRITE modes)
	// - FileDoesNotExistError (in DREAD mode)
	// - BadFilenameError (if filename contains non alpha-numeric chars or is not 1-16 chars long)
	Open(fname string, mode FileMode) (f DFSFile, err error)

	// Disconnects from the server. Can return the following errors:
	// - DisconnectedError
	UMountDFS() (err error)
}

// The constructor for a new DFS object instance. Takes the server's
// IP:port address string as parameter, the localIP to use to
// establish the connection to the server, and a localPath path on the
// local filesystem where the client has allocated storage (and
// possibly existing state) for this DFS.
//
// The returned dfs instance is singleton: an application is expected
// to interact with just one dfs at a time.
//
// This call should succeed regardless of whether the server is
// reachable. Otherwise, applications cannot access (local) files
// while disconnected.
//
// Can return the following errors:
// - LocalPathError
// - Networking errors related to localIP or serverAddr
func MountDFS(serverAddr string, localIP string, localPath string) (dfs DFS, err error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// TODO

	// Establish connection to server:
	var isOffline bool
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Println("Failed to establish connection to server: ", serverAddr)
		isOffline = true
	}
	connDFS := &ConnDFS{IsOffline: isOffline, ServerIP: serverAddr}
	connDFS.ClientPath = localPath
	if !isOffline {
		connDFS.ServerRPC = rpc.NewClient(conn)
	}

	// TODO: Check localPath (LocalPathError)

	// First check if localPath has existing clientID:
	metadata, err := os.Open(localPath + "metadata.dfs")
	if err != nil && !isOffline {
		// Not an existing connection, create a clientID
		metadata, err := os.Create(localPath + "metadata.dfs")
		if err != nil {
			return nil, LocalPathError(localPath)
		}
		var cid int
		err = connDFS.ServerRPC.Call("ClientToServer.GetNewCID", localPath, &cid)

		var clientmeta = new(ClientMetaData)
		clientmeta.ClientID = cid
		clientmeta.WriteLogs = make(map[string]int)
		connDFS.ClientID = cid

		clientMetaByte, _ := json.MarshalIndent(clientmeta, "", " ")

		metadata.Write(clientMetaByte)
		metadata.Sync()

	} else if err != nil && isOffline {
		log.Println("Can't connect to server to establish a new client")
	}

	// Re-open if there was a new file created
	metadata, _ = os.Open(localPath + "metadata.dfs")
	defer metadata.Close()
	// Read metadata
	data, _ := ioutil.ReadAll(metadata)
	var readmeta ClientMetaData
	json.Unmarshal(data, &readmeta)
	connDFS.ClientID = readmeta.ClientID

	if !isOffline {
		// Establish Server -> Client RPC
		clientIP := localIP + ":0"
		clientAddr, err := net.ResolveTCPAddr("tcp", clientIP)
		if err != nil {
			log.Println(err)
		}
		tcpConn, err := net.ListenTCP("tcp", clientAddr)
		if err != nil {
			log.Println(err)
		}
		clientIP = localIP + ":" + strconv.Itoa(tcpConn.Addr().(*net.TCPAddr).Port)
		log.Println("Client RCP IP: ", clientIP)
		clientServer := rpc.NewServer()
		serverToClient := new(ServerToClient)
		clientServer.Register(serverToClient)

		go clientServer.Accept(tcpConn)

		// var listenOk bool
		// err = connDFS.ServerRPC.Call("ClientToServer.CreateListenerClient", clientIP, &listenOk)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		clientConn, err := net.Dial("tcp", clientIP)
		if err != nil {
			log.Println(err)
		}
		connDFS.ClientRPC = rpc.NewClient(clientConn)
		var totalClients int

		storedDFS := &sharedData.StoredDFSMessage{ClientID: connDFS.ClientID, ClientIP: clientIP, ClientPath: localPath}
		connDFS.ClientIP = clientIP

		connDFS.ServerRPC.Call("ClientToServer.MapAliveClient", storedDFS, &totalClients)

		//TODO: Start Heartbeats to Client
		//Steps:
		//		Create another go routine to listen for UDP packets
		//		stored heartbeat port to MapAliveClient
	}

	return connDFS, nil
}
