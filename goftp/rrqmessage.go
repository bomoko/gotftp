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
	if datagram.AckBlock == sess.BlockNumber && !sess.Completed {
		// if the session is completed, we don't want to increment the block number
		// because we may need to resend this block numerous times
		if !sess.Completed {
			sess.BlockNumber++
		} else { //they've acknowledged the last block, we can make things complete now
			sess.ClosedAt = time.Now()
		}
	}
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

	if session.Completed { //if we get here, we're just acknowledging the very last item, it's empty
		return nil, nil
	}

	dataBuff := make([]byte, dataSize)
	session.FilePointer.Seek((int64(session.BlockNumber)-1)*dataSize, 0)
	n, err := session.FilePointer.Read(dataBuff) //we read in 512 bytes, if possible

	if err != nil {
		return dataBuff, err
	}

	if n < dataSize {
		session.Completed = true
	}

	return append(ret, dataBuff[:n]...), nil
}
