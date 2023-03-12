package src

import (
	"errors"
	"fmt"
	"net"
	"os"
)

/**
This file contains all the logic for a RRQ session
*/

type RRQSession struct {
	Filename         string
	RequesterAddress *net.UDPAddr //we use this to keep track of what's what
	BlockNumber      int          //The current block we're sending
	FileData         []byte
}

func SetupRRQSession(incoming Datagram, requesterAddr *net.UDPAddr) (*RRQSession, error) {
	fmt.Sprintf("Trying to open %v\n", incoming.RrqFilename)
	if _, err := os.Stat("./files/" + incoming.RrqFilename); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New(incoming.RrqFilename + " does not exit")
	}

	dat, err := os.ReadFile("./files/" + incoming.RrqFilename)
	if err != nil {
		fmt.Println("Got an error opening file")
		return nil, errors.New("File not found")
	}

	return &RRQSession{
		Filename:         "",
		RequesterAddress: nil,
		BlockNumber:      1,
		FileData:         dat,
	}, nil
}

func AcknowledgeRRQSession(sess *RRQSession) {
	// since we have an acknowledgement, we don't need to keep track of the whole string anymore ...
	if len(sess.FileData) >= dataSize {
		sess.FileData = sess.FileData[dataSize:]
	} else { //we've actually already sent the data, so we send an empty array
		sess.FileData = []byte{}
	}
	sess.BlockNumber++
}

const dataSize = 512

func GenerateRRQMessage(session *RRQSession) ([]byte, error) {
	//We break this sucker into 512k chunks
	ret := []byte{
		0x0,
		0x3,
		0x0,
		byte(session.BlockNumber),
	}

	//now we work out whether this is the end or not
	head := session.FileData
	if len(session.FileData) >= dataSize {
		head = session.FileData[0:dataSize]
	}
	return append(ret, head...), nil
}
