package goftp

import (
	"errors"
	"fmt"
)

const (
	NOT_DEFINED = iota
	FILE_NOT_FOUND
	ACCESS_VIOLATION
	DISK_FULL
	ILLEGAL_TFTP_OPERATION
	UNKNOWN_TRANSFER_ID
	FILE_ALREADY_EXISTS
	NO_SUCH_USER
)

var ErrorCodes = map[uint16]uint16{
	NOT_DEFINED:            0x0,
	FILE_NOT_FOUND:         0x1,
	ACCESS_VIOLATION:       0x2,
	DISK_FULL:              0x3,
	ILLEGAL_TFTP_OPERATION: 0x4,
	UNKNOWN_TRANSFER_ID:    0x5,
	FILE_ALREADY_EXISTS:    0x6,
	NO_SUCH_USER:           0x7,
}

func WrapTFTPError(code uint16, err error) TFTP_error {
	return TFTP_error{
		Err:       err,
		ErrorCode: code,
	}
}

func GenerateTFTPError(code uint16, message string) TFTP_error {
	return TFTP_error{
		Err:       errors.New(message),
		ErrorCode: code,
	}
}

type TFTP_error struct {
	Err       error
	ErrorCode uint16
}

func (e TFTP_error) Error() string {
	return fmt.Sprintf("Error ID: %v - message: %v", ConvertErrorCodeToString(e.ErrorCode), e.Err.Error())
}

func ConvertErrorCodeToString(code uint16) string {
	switch code {
	case NOT_DEFINED:
		return "NOT_DEFINED"
	case FILE_NOT_FOUND:
		return "FILE_NOT_FOUND"
	case ACCESS_VIOLATION:
		return "ACCESS_VIOLATION"
	case DISK_FULL:
		return "DISK_FULL"
	case ILLEGAL_TFTP_OPERATION:
		return "ILLEGAL_TFTP_OPERATION"
	case UNKNOWN_TRANSFER_ID:
		return "UNKNOWN_TRANSFER_ID"
	case FILE_ALREADY_EXISTS:
		return "FILE_ALREADY_EXISTS"
	case NO_SUCH_USER:
		return "NO_SUCH_USER"
	}
	return "UNDEFINED"
}

func GenerateErrorMessage(err error) []byte {

	code := uint16(NOT_DEFINED)
	e, OK := err.(TFTP_error)
	if OK {
		code = e.ErrorCode
	}

	ret := []byte{
		0x0,
		0x5,
		0x0,
		byte(ErrorCodes[code]),
	}

	ret = append(append(ret, []byte(err.Error())...), 0x0)

	return ret

}
