package main

import (
	"flag"
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
var tftpDirectory string

// Flags
var port int
var enableWrites bool

// SessionKey will simply give us the key used to index active sessions
func SessionKey(a *net.UDPAddr) string {
	return fmt.Sprintf("%v:%v", a.IP, a.Port)
}

func main() {

	//Set up flags
	//TODO: we should do some kind of check and normalization on this directory name.
	flag.StringVar(&tftpDirectory, "d", "./files/", "Directory to read/write files to")
	flag.IntVar(&port, "p", 6999, "Port to run TFTP server on")
	flag.BoolVar(&enableWrites, "write-enabled", false, "Allow users to write to the server (potentially unsafe)")

	s, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%v", port))
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

	RRQSessions := map[string]*src.RRQSession{}
	WRQSessions := map[string]*src.WRQSession{}

	workingBuffer := make([]byte, 1024)

	for {

		n, addr, err := connection.ReadFromUDP(workingBuffer)
		buffer := workingBuffer[0:n] //only pull the data that we actually read.
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
			// we bail
			data = src.GenerateErrorMessage(err)
			_, err = connection.WriteToUDP(data, addr)
			continue
		}

		switch dgo.Opcode {
		case src.OPCODE_RRQ:
			session, errR := src.SetupRRQSession(tftpDirectory, dgo, addr)
			if errR != nil {
				data = src.GenerateErrorMessage(errR)
				break
			}
			data, _ = src.GenerateRRQMessage(session)
			RRQSessions[SessionKey(addr)] = session
		case src.OPCODE_WRQ:
			session, errW := src.SetupWRQSession(tftpDirectory, dgo, addr)
			if errW != nil {
				data = src.GenerateErrorMessage(errW)
				break
			}
			data, _ = src.GenerateWRQMessage(session)
			WRQSessions[SessionKey(addr)] = session
		case src.OPCODE_DATA: //we've got incoming data
			data, _ = src.AcknowledgeWRQSession(WRQSessions[SessionKey(addr)], dgo)
		case src.OPCODE_ACK:
			//So we need to see _which_ block this is an acknowledgement for
			if errA := src.AcknowledgeRRQSession(RRQSessions[SessionKey(addr)], dgo); errA != nil {
				data = src.GenerateErrorMessage(errA)
				break
			}
			if !RRQSessions[SessionKey(addr)].Completed {
				data, _ = src.GenerateRRQMessage(RRQSessions[SessionKey(addr)])
			}
		default:
			data = src.GenerateErrorMessage(src.GenerateTFTPError(src.NOT_DEFINED, "Only able to send you files right now"))
		}

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
