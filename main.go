package main

import (
	"errors"
	"flag"
	"fmt"
	"gotftp/goftp"
	"log"
	"net"
	"path/filepath"
	"sync"
	"time"
)

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
	theLogger := log.Default()
	readFlags()

	s, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%v", port))
	if err != nil {
		theLogger.Fatal(err)
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		theLogger.Fatal(err)
	}

	defer connection.Close()

	RRQSessions := map[string]*goftp.RRQSession{}
	var RRQSessionsMu sync.Mutex
	WRQSessions := map[string]*goftp.WRQSession{}
	var WRQSessionsMu sync.Mutex

	workingBuffer := make([]byte, 1024)

	logSessionNumbers := func() {
		theLogger.Printf("Currently have %v number of active read sessions and %v number of active write sessions \n", len(RRQSessions), len(WRQSessions))
	}

	gc := func() {
		// Note that I'm using the mutext TryLock functionality because
		// we can simply skip over rather than wait for access in this context
		// We'll eventually clean out the session
		for {
			timeout := 10.0 // ten seconds timeout from the connection being closed to when we kill off the session
			for k, e := range WRQSessions {
				if e.Completed {
					if time.Now().Sub(e.ClosedAt).Seconds() > timeout {
						if WRQSessionsMu.TryLock() {
							delete(WRQSessions, k)
							theLogger.Printf("Write request from IP:%v for file %v complete", k, e.Filename)
							logSessionNumbers()
							WRQSessionsMu.Unlock()
						}
					}
				}
			}

			for k, e := range RRQSessions {
				if e.Completed {
					if time.Now().Sub(e.ClosedAt).Seconds() > timeout {
						if RRQSessionsMu.TryLock() {
							delete(RRQSessions, k)
							theLogger.Printf("Read request from IP:%v for file %v complete", k, e.Filename)
							logSessionNumbers()
							RRQSessionsMu.Unlock()
						}
					}
				}
			}
			time.Sleep(time.Second)
		}
	}

	go gc()

	for {
		//First things first, let's clean up any completed sessions

		n, addr, err := connection.ReadFromUDP(workingBuffer)
		buffer := workingBuffer[0:n] //only pull the data that we actually read.

		if err != nil {
			theLogger.Println(err)
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
			theLogger.Printf("Got Read request from IP:%v%v for file %v", addr.IP, addr.Port, dgo.Filename)
			session, errR := goftp.SetupRRQSession(cleanTftpDirectory(), dgo, addr)
			if errR != nil {
				data = goftp.GenerateErrorMessage(errR)
				break
			}
			data, _ = goftp.GenerateRRQMessage(session)
			RRQSessionsMu.Lock()
			RRQSessions[SessionKey(addr)] = session
			RRQSessionsMu.Unlock()
		case goftp.OPCODE_WRQ:
			theLogger.Printf("Got Write request from IP:%v:%v for file %v", addr.IP, addr.Port, dgo.Filename)
			if !enableWrites {
				data = goftp.GenerateErrorMessage(errors.New("Not accepting writes at the moment"))
				break
			}
			session, errW := goftp.SetupWRQSession(cleanTftpDirectory(), dgo, addr)
			if errW != nil {
				data = goftp.GenerateErrorMessage(errW)
				break
			}
			data, _ = goftp.GenerateWRQMessage(session)
			WRQSessionsMu.Lock()
			WRQSessions[SessionKey(addr)] = session
			WRQSessionsMu.Unlock()
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
		}

		if err != nil {
			theLogger.Println(err)
			return
		}
	}
}

func readFlags() {
	flag.StringVar(&tftpDirectory, "d", "./files/", "Directory to read/write files to")
	flag.IntVar(&port, "p", 6999, "Port to run TFTP server on")
	flag.BoolVar(&enableWrites, "w", false, "Allow users to write to the server (potentially unsafe)")
	flag.Parse()
}

func cleanTftpDirectory() string {
	cleanTftpDirectory, err := filepath.Abs(tftpDirectory)
	if err != nil {
		log.Default().Fatal(err)
	}
	return cleanTftpDirectory
}
