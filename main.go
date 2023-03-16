package main

import (
	"errors"
	"flag"
	"fmt"
	"gotftp/goftp"
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
	flag.BoolVar(&enableWrites, "w", false, "Allow users to write to the server (potentially unsafe)")

	flag.Parse()

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

	RRQSessions := map[string]*goftp.RRQSession{}
	WRQSessions := map[string]*goftp.WRQSession{}

	workingBuffer := make([]byte, 1024)

	for {

		//First things first, let's clean up any completed sessions
		timeout := 10.0 // ten seconds timeout from the connection being closed to when we kill off the session
		for k, e := range WRQSessions {
			if e.Completed {
				if time.Now().Sub(e.ClosedAt).Seconds() > timeout {
					delete(WRQSessions, k)
				}
			}
		}

		for k, e := range RRQSessions {
			if e.Completed {
				if time.Now().Sub(e.ClosedAt).Seconds() > timeout {
					delete(RRQSessions, k)
				}
			}
		}

		fmt.Printf("Currently have %v number of active read sessions and %v number of active write sessions \n", len(RRQSessions), len(WRQSessions))

		n, addr, err := connection.ReadFromUDP(workingBuffer)
		buffer := workingBuffer[0:n] //only pull the data that we actually read.
		fmt.Printf("-> %v\n", buffer[0:n])

		if err != nil {
			fmt.Println(err)
			return
		}

		d := goftp.DatagramBuffer{
			Buffer: buffer,
			Offset: 0,
		}

		// Buffer containing whatever we're going to send back across the wire
		var data []byte

		dgo, err := goftp.DestructureDatagram(d)
		if err != nil {
			// we bail
			data = goftp.GenerateErrorMessage(err)
			_, err = connection.WriteToUDP(data, addr)
			continue
		}

		switch dgo.Opcode {
		case goftp.OPCODE_RRQ:
			session, errR := goftp.SetupRRQSession(tftpDirectory, dgo, addr)
			if errR != nil {
				data = goftp.GenerateErrorMessage(errR)
				break
			}
			data, _ = goftp.GenerateRRQMessage(session)
			RRQSessions[SessionKey(addr)] = session
		case goftp.OPCODE_WRQ:
			if !enableWrites {
				data = goftp.GenerateErrorMessage(errors.New("Not accepting writes at the moment"))
				break
			}
			session, errW := goftp.SetupWRQSession(tftpDirectory, dgo, addr)
			if errW != nil {
				data = goftp.GenerateErrorMessage(errW)
				break
			}
			data, _ = goftp.GenerateWRQMessage(session)
			WRQSessions[SessionKey(addr)] = session
		case goftp.OPCODE_DATA: //we've got incoming data
			data, _ = goftp.AcknowledgeWRQSession(WRQSessions[SessionKey(addr)], dgo)
		case goftp.OPCODE_ACK:
			//So we need to see _which_ block this is an acknowledgement for
			if errA := goftp.AcknowledgeRRQSession(RRQSessions[SessionKey(addr)], dgo); errA != nil {
				data = goftp.GenerateErrorMessage(errA)
				break
			}
			if !RRQSessions[SessionKey(addr)].Completed {
				data, _ = goftp.GenerateRRQMessage(RRQSessions[SessionKey(addr)])
			}
		default:
			data = goftp.GenerateErrorMessage(goftp.GenerateTFTPError(goftp.NOT_DEFINED, "Only able to send you files right now"))
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
