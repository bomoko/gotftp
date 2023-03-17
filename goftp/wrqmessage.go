package goftp

import (
	"errors"
	"net"
	"os"
	"time"
)

/**
This file contains all the logic for a WRQ session
*/

type WRQSession struct {
	Filename         string
	FilePointer      *os.File
	RequesterAddress *net.UDPAddr //we use this to keep track of what's what
	BlockNumber      uint16       //The current block we're sending
	FileData         []byte
	Completed        bool
	ClosedAt         time.Time
}

// The two following functions work in combination with one another
// Only once a session is ACKed will we move forward the write buffer
// Until that point, GenerateRRQMessage could potentially keep pushing out the same chunk of data

func SetupWRQSession(filesDirectory string, incoming Datagram, requesterAddr *net.UDPAddr) (*WRQSession, error) {

	fullyQualifiedFilename := filesDirectory + "/" + incoming.Filename

	if _, err := os.Stat(fullyQualifiedFilename); errors.Is(err, os.ErrNotExist) {
		// Here we are able to actually write
		f, err := os.Create(fullyQualifiedFilename)
		if err != nil {
			return nil, err
		}

		return &WRQSession{
			Filename:         fullyQualifiedFilename,
			RequesterAddress: nil,
			FilePointer:      f,
			BlockNumber:      0, //We begin by acknowledging the very first element
		}, nil
	} else {
		return nil, GenerateTFTPError(FILE_ALREADY_EXISTS, "File already exists")
	}
}

func AcknowledgeWRQSession(sess *WRQSession, datagram Datagram) ([]byte, error) {
	// since we have an acknowledgement, we don't need to keep track of the whole string anymore ...
	// TODO: actually deal with the writes ....
	if datagram.AckBlock == sess.BlockNumber { //we've already seen this one, let's just resent the ack
		return GenerateWRQMessage(sess)
	}
	//else, we push forward with the write, and ack the incoming data
	if len(datagram.WrqData) > 0 && sess.Completed == false { // we write any _actual_ data - this could signal a session completion below
		_, err := sess.FilePointer.Write(datagram.WrqData) //might need to keep track of the n
		if err != nil {
			return nil, err
		}
	}

	if len(datagram.WrqData) < dataSize {
		//this is the last stuff we need to write, so lets close this sucker up and mark the session as done
		sess.Completed = true
		sess.ClosedAt = time.Now()
		err := sess.FilePointer.Close()
		if err != nil {
			return nil, err
		}
	}

	retData, err := GenerateWRQMessage(sess)
	if err != nil {
		return nil, err
	}
	sess.BlockNumber = datagram.AckBlock
	return retData, nil
}

func GenerateWRQMessage(session *WRQSession) ([]byte, error) {
	//We break this sucker into 512k chunks
	ret := []byte{
		0x0,
		0x4, //This is an ack - so we drop this in here
		byte(session.BlockNumber >> 8),
		byte(session.BlockNumber),
	}

	return ret, nil
}
