package main

import (
	"fmt"
	"gotftp/src"
	"math/rand"
	"net"
	"time"
)

//TODO: let's get a version out without using goroutines - maybe we just treat it all as a
// single event loop, although that might not make sense given that we need to start reading on multiple
// different ports... so maybe goroutines from the beginning.
// Sure thing is, though, we'll make sure that we're doing the simplest use case first - no writing of files, only reading
// of what's already there.
// TODO: Deal with the standard binary option as a first step - netascii next

// Set up some of the globals we're going to use.
const tftpDirectory = "./files"

func main() {
	s, err := net.ResolveUDPAddr("udp4", ":6999")
	if err != nil {
		fmt.Println(err)
		return
	}
	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer connection.Close()
	buffer := make([]byte, 1024)
	rand.Seed(time.Now().Unix())

	//var readSessions []src.RRQSession
	var session *src.RRQSession

	for {
		n, addr, err := connection.ReadFromUDP(buffer)
		//n, _, err := connection.ReadFromUDP(buffer)
		fmt.Print("-> ", string(buffer[0:n-1]))

		if err != nil {
			fmt.Println(err)
			return
		}

		d := src.DatagramBuffer{
			Buffer: buffer,
			Offset: 0,
		}

		dgo, err := src.DestructureDatagram(d)
		fmt.Println(dgo)

		var data []byte
		switch dgo.Opcode {
		case src.OPCODE_RRQ:
			fmt.Println("Got a RRQ")
			session, err = src.SetupRRQSession(dgo, addr)
			if err != nil {
				fmt.Println("Got a RRQ error")
				data = src.GenerateErrorMessage(err.Error(), src.NOT_DEFINED)
				break
			}
			data, _ = src.GenerateRRQMessage(session)
		case src.OPCODE_ACK:
			fmt.Println("Got an ACK")
			src.AcknowledgeRRQSession(session)
			data, _ = src.GenerateRRQMessage(session)
		default:
			data = src.GenerateErrorMessage("this shit doesnt work yet", src.NOT_DEFINED)
		}

		fmt.Printf("data: %s\n", string(data))
		_, err = connection.WriteToUDP(data, addr)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
