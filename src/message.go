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
			//// We advance the offset PAST the delimiter if we hit one
			//d.Offset++
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
	RrqMode     string
}

const datagramMinimum = 3

const OPCODE_RRQ = "RRQ"
const OPCODE_WRQ = "WRQ"
const OPCODE_DATA = "DATA"
const OPCODE_ACK = "ACK"
const OPCODE_ERROR = "ERROR"

func DestructureDatagram(d DatagramBuffer) (Datagram, error) {
	ret := Datagram{}

	if len(d.Buffer) <= datagramMinimum {
		return ret, errors.New("Malformed data")
	}

	var err error
	ret.Opcode, err = opcode(binary.BigEndian.Uint16(d.Buffer[0:2]))
	if err != nil {
		return ret, err
	}

	//Now we have a few options, depending on the opcode
	switch ret.Opcode {
	case OPCODE_RRQ:
		err = destructureDatagramRRQ(d, &ret)
		if err != nil {
			return ret, err
		}
	case OPCODE_WRQ:
		return ret, nil
	case OPCODE_DATA:
		return ret, nil
	case OPCODE_ACK:
		return ret, nil
	case OPCODE_ERROR:
		return ret, nil
	}

	return ret, nil
}

func destructureDatagramRRQ(d DatagramBuffer, ret *Datagram) error {
	d.Offset = 2 //we want to read the filename
	ret.RrqFilename = string(d.ReadUntilDelimiter())
	ret.RrqMode = string(d.ReadUntilDelimiter())
	if ret.RrqMode != "octet" {
		return errors.New("We only support `octect` for now")
	}
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
