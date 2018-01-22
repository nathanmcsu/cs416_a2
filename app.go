/*

A trivial application to illustrate how the dfslib library can be used
from an application in assignment 2 for UBC CS 416 2017W2.

Usage:
go run app.go
*/

package main

// Expects dfslib.go to be in the ./dfslib/ dir, relative to
// this app.go file
import (
	"fmt"
	"os"

	"./dfslib"
)

func main() {

	serverAddr := "127.0.0.1:8080"
	localIP := "127.0.0.1"
	localPath := "/home/nathan/Desktop/cs416/as2_app/app2/"

	// Connect to DFS.
	dfs, err := dfslib.MountDFS(serverAddr, localIP, localPath)
	if checkError(err) != nil {
		// return
	}

	// Close the DFS on exit.
	// Defers are really cool, check out: https://blog.golang.org/defer-panic-and-recover
	defer dfs.UMountDFS()

	// Check if hello.txt file exists in the global DFS.
	exists, err := dfs.GlobalFileExists("helloworld")
	if checkError(err) != nil {
		// return
	}

	f, err := dfs.Open("helloworld", dfslib.DREAD)
	if checkError(err) != nil {
		return
	}

	var chunk dfslib.Chunk
	err = f.Read(0, &chunk)
	checkError(err)

	if exists {
		fmt.Println("File already exists, mission accomplished")
		//return
	}

	// Open the file (and  create it if it does not exist) for writing.
	f, err = dfs.Open("helloworld", dfslib.WRITE)
	if checkError(err) != nil {
		return
	}
	// Close the file on exit.
	defer f.Close()

	// Create a chunk with a string message.
	const str = "Hello friends!"
	copy(chunk[:], str)

	// Write the 0th chunk of the file.
	err = f.Write(2, &chunk)
	if checkError(err) != nil {
		return
	}
	// Read the 0th chunk of the file.
	err = f.Read(0, &chunk)
	checkError(err)
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
