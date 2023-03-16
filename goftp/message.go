package goftp

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	Opcode   string
	Filename string
	Mode     string
	AckBlock uint16
	WrqData  []byte
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
	fmt.Println("OPCODE :", ret.Opcode)
	//Now we have a few options, depending on the opcode
	switch ret.Opcode {
	case OPCODE_RRQ:
		err = destructureDatagramRRQ(d, &ret)
		if err != nil {
			return ret, err
		}
	case OPCODE_WRQ:
		err = destructureDatagramWRQ(d, &ret)
		if err != nil {
			return ret, err
		}
	case OPCODE_DATA:
		err = descructureDatagramDATA(d, &ret)
		if err != nil {
			return ret, err
		}
		return ret, nil
	case OPCODE_ACK:
		fmt.Sprintf("Got an ACK with block %v\n", string(d.Buffer[2:4]))
		err = destructureDatagramACK(d, &ret)
		if err != nil {
			return ret, err
		}
		return ret, nil
	case OPCODE_ERROR:
		return ret, nil
	}

	return ret, nil
}

func destructureDatagramRRQ(d DatagramBuffer, ret *Datagram) error {
	d.Offset = 2 //we want to read the filename
	ret.Filename = string(d.ReadUntilDelimiter())
	ret.Mode = string(d.ReadUntilDelimiter())
	if ret.Mode != "octet" {
		return GenerateTFTPError(ILLEGAL_TFTP_OPERATION, "We only support `octect` for now")
	}
	return nil
}

func destructureDatagramWRQ(d DatagramBuffer, ret *Datagram) error {
	d.Offset = 2 //we want to read the filename
	ret.Filename = string(d.ReadUntilDelimiter())
	ret.Mode = string(d.ReadUntilDelimiter())
	if ret.Mode != "octet" {
		return GenerateTFTPError(ILLEGAL_TFTP_OPERATION, "We only support `octect` for now")
	}
	return nil
}

func descructureDatagramDATA(d DatagramBuffer, ret *Datagram) error {
	ret.AckBlock = binary.BigEndian.Uint16(d.Buffer[2:4])
	ret.WrqData = d.Buffer[4:]
	return nil
}

func destructureDatagramACK(d DatagramBuffer, ret *Datagram) error {
	fmt.Sprintf("Got an ACK with block %v\n", string(d.Buffer[2:4]))
	ret.AckBlock = binary.BigEndian.Uint16(d.Buffer[2:4])
	return nil
}

func constructDatagram(d Datagram) (DatagramBuffer, error) {
	return DatagramBuffer{}, nil
}

// Codify opcodes
func opcode(o uint16) (string, error) {
	switch o {
	case 0x1:
		return OPCODE_RRQ, nil
	case 0x2:
		return OPCODE_WRQ, nil
	case 0x3:
		return OPCODE_DATA, nil
	case 0x4:
		return OPCODE_ACK, nil
	case 0x5:
		return OPCODE_ERROR, nil
	}
	return "", GenerateTFTPError(ILLEGAL_TFTP_OPERATION, "Invalid Opcode")
}
