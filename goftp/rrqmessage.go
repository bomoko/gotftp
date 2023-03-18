package goftp

import (
	"errors"
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
	FilePointer      *os.File
	Completed        bool
	ClosedAt         time.Time
}

// The two following functions work in combination with one another
// Only once a session is ACKed will we move forward the write buffer
// Until that point, GenerateRRQMessage could potentially keep pushing out the same chunk of data

func SetupRRQSession(filesDirectory string, incoming Datagram, requesterAddr *net.UDPAddr) (*RRQSession, error) {
	fullyQualifiedFilename := filesDirectory + "/" + incoming.Filename
	if _, err := os.Stat(fullyQualifiedFilename); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New(incoming.Filename + " does not exit")
	}

	dat, err := os.ReadFile(fullyQualifiedFilename)
	fp, err := os.Open(fullyQualifiedFilename)
	if err != nil {
		return nil, err
	}

	return &RRQSession{
		Filename:         fullyQualifiedFilename,
		RequesterAddress: nil,
		BlockNumber:      1,
		FileData:         dat,
		FilePointer:      fp,
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

	dataBuff := make([]byte, dataSize)
	n, err := session.FilePointer.Read(dataBuff) //we read in 512 bytes, if possible

	if err != nil {
		return dataBuff, err
	}

	if n < dataSize { //we've likely still got places to go
		err = session.FilePointer.Close() //we can close this because we're going to save the data in the buffer
		if err != nil {
			return dataBuff, err
		}
		session.Completed = true
		session.ClosedAt = time.Now()
	}

	//We'll keep track of the last packet's data, though
	session.FileData = dataBuff[:n]

	return append(ret, session.FileData...), nil
}
