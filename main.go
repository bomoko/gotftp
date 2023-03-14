package main

import (
	"fmt"
	"gotftp/src"
	"net"
)

//TODO: let's get a version out without using goroutines - maybe we just treat it all as a
// single event loop, although that might not make sense given that we need to start reading on multiple
// different ports... so maybe goroutines from the beginning.
// Sure thing is, though, we'll make sure that we're doing the simplest use case first - no writing of files, only reading
// of what's already there.
// TODO: Deal with the standard binary option as a first step - netascii next

// Set up some of the globals we're going to use.
const tftpDirectory = "./files"

// SessionKey will simply give us the key used to index active sessions
func SessionKey(a *net.UDPAddr) string {
	return fmt.Sprintf("%v:%v", a.IP, a.Port)
}

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

	var sessions map[string]*src.RRQSession = map[string]*src.RRQSession{}

	for {
		buffer := make([]byte, 1024) //this needs to be reset
		n, addr, err := connection.ReadFromUDP(buffer)
		fmt.Printf("-> %v\n", buffer[0:n])

		if err != nil {
			fmt.Println(err)
			return
		}

		d := src.DatagramBuffer{
			Buffer: buffer,
			Offset: 0,
		}

		// Buffer containing whatever we're going to send back across the wire
		var data []byte

		dgo, err := src.DestructureDatagram(d)
		if err != nil {
			data = src.GenerateErrorMessage(err)
		}

		switch dgo.Opcode {
		case src.OPCODE_RRQ:
			session, err := src.SetupRRQSession(dgo, addr)
			if err != nil {
				data = src.GenerateErrorMessage(err)
				break
			}
			data, _ = src.GenerateRRQMessage(session)
			sessions[SessionKey(addr)] = session
		case src.OPCODE_ACK:
			//So we need to see _which_ block this is an acknowledgement for
			if src.AcknowledgeRRQSession(sessions[SessionKey(addr)], dgo) != nil {
				data = src.GenerateErrorMessage(err)
				break
			}

			if !sessions[SessionKey(addr)].Completed {
				data, _ = src.GenerateRRQMessage(sessions[SessionKey(addr)])
			}
		default:
			data = src.GenerateErrorMessage(src.GenerateTFTPError(src.NOT_DEFINED, "Only able to send you files right now"))
		}

		//fmt.Printf("data: %s\n", string(data))
		if len(data) > 0 {
			_, err = connection.WriteToUDP(data, addr)
		} else {
			fmt.Println("No data to send, skipping")
		}
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
