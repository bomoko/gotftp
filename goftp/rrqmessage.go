package goftp

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

/**
This file contains all the logic for a RRQ session
*/

const dataSize = 512

type RRQSession struct {
	Filename         string
	RequesterAddress *net.UDPAddr //we use this to keep track of what's what
	BlockNumber      uint16       //The current block we're sending
	FileData         []byte
	Completed        bool
	ClosedAt         time.Time
}

// The two following functions work in combination with one another
// Only once a session is ACKed will we move forward the write buffer
// Until that point, GenerateRRQMessage could potentially keep pushing out the same chunk of data

func SetupRRQSession(filesDirectory string, incoming Datagram, requesterAddr *net.UDPAddr) (*RRQSession, error) {
	fmt.Sprintf("Trying to open %v\n", incoming.Filename)

	fullyQualifiedFilename := filesDirectory + "/" + incoming.Filename
	if _, err := os.Stat(fullyQualifiedFilename); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New(incoming.Filename + " does not exit")
	}

	dat, err := os.ReadFile(fullyQualifiedFilename)
	if err != nil {
		fmt.Println("Got an error opening file")
		return nil, errors.New("File not found")
	}

	return &RRQSession{
		Filename:         fullyQualifiedFilename,
		RequesterAddress: nil,
		BlockNumber:      1,
		FileData:         dat,
	}, nil
}

func AcknowledgeRRQSession(sess *RRQSession, datagram Datagram) error {

	//We might actually be resending here - so we only increment if the acknowledgement is the next block
	//else we don't actually advance through to the next
	if datagram.AckBlock == sess.BlockNumber {
		// if the session is completed, we don't want to increment the block number
		// because we may need to resend this block numerous times
		if !sess.Completed {
			sess.BlockNumber++
		}
		if len(sess.FileData) >= dataSize {
			sess.FileData = sess.FileData[dataSize:]
		} else { //we've actually already sent the data, so we send an empty array
			sess.FileData = []byte{}
		}
	} //TODO: should we be throwing an error here if we get a datagram with a Block that's not equal to the current?
	// TODO: check spec for above case

	return nil
}

func GenerateRRQMessage(session *RRQSession) ([]byte, error) {
	//We break this sucker into 512k chunks

	ret := []byte{
		0x0,
		0x3,
		byte(session.BlockNumber >> 8),
		byte(session.BlockNumber),
	}

	//now we work out whether this is the end or not
	head := session.FileData
	if len(session.FileData) >= dataSize {
		// There's more data to send, so only send part of it
		head = session.FileData[0:dataSize]
	} else {
		//we've reached the end of the line - anything less that 512 bytes will signal the end
		//And wso we mark the session as being complete
		session.Completed = true
		session.ClosedAt = time.Now()
	}
	return append(ret, head...), nil
}
