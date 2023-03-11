package main

import "fmt"

//TODO: let's get a version out without using goroutines - maybe we just treat it all as a
// single event loop, although that might not make sense given that we need to start reading on multiple
// different ports... so maybe goroutines from the beginning.
// Sure thing is, though, we'll make sure that we're doing the simplest use case first - no writing of files, only reading
// of what's already there.
// TODO: Deal with the standard binary option as a first step - netascii next

// Set up some of the globals we're going to use.
const tftpDirectory = "./files"

func main() {
	fmt.Print("Hello world")
}
