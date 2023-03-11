package src

import (
	"encoding/binary"
	"errors"
)

// This is going to be stolen straight from https://github.com/vcabbage/go-tftp/blob/master/datagram.go:97
// It's a straightforward way of knowing where we're currently sitting in a byte stream
// - we use this to read incoming datagrams and construct new outgoing ones
type DatagramBuffer struct {
	Buffer []byte
	Offset int
}

// I'm wondering if this isn't better implmented as some kind of tokenization?
func (d *DatagramBuffer) ReadUntilDelimiter() []byte {
	r := []byte{}
	for _, data := range d.Buffer[d.Offset:] {
		d.Offset++
		if data == 0x0 {
			// We advance the offset PAST the delimiter if we hit one
			d.Offset++
			return r
		}
		r = append(r, data)
	}
	return r
}

// I'm going to make a single desctructured datagram type that, depending on its type, will boil down _to_ a
// DatagramBuffer, or be loaded from one

type Datagram struct {
	Opcode      string
	RrqFilename string
}

func destructureDatagram(d DatagramBuffer) (Datagram, error) {
	ret := Datagram{}
	//TODO: check total size of the datagram - it's it's less than the minimum, then error out
	var err error
	ret.Opcode, err = opcode(binary.BigEndian.Uint16(d.Buffer[0:2]))
	if err != nil {
		return ret, err
	}

	//Now we have a few options, depending on the opcode
	switch ret.Opcode {
	case "RRQ":
		err = destructureDatagramRRQ(d, &ret)
		if err != nil {
			return ret, err
		}
	case "WRQ":
		return ret, nil
	case "DATA":
		return ret, nil
	case "ACK":
		return ret, nil
	case "ERROR":
		return ret, nil
	}

	return ret, nil
}

func destructureDatagramRRQ(d DatagramBuffer, ret *Datagram) error {
	d.Offset = 2 //we want to read the filename
	ret.RrqFilename = string(d.ReadUntilDelimiter())
	return nil
}

func constructDatagram(d Datagram) (DatagramBuffer, error) {
	return DatagramBuffer{}, nil
}

// Codify opcodes
func opcode(o uint16) (string, error) {
	switch o {
	case 0x1:
		return "RRQ", nil
	case 0x2:
		return "WRQ", nil
	case 0x3:
		return "DATA", nil
	case 0x4:
		return "ACK", nil
	case 0x5:
		return "ERROR", nil
	}
	return "", errors.New("INVALID OPCODE") //TODO: this should actually codify a proper TFTP error
}

// Codify error codes
