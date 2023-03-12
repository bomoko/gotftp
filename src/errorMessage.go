package src

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

var ErrorCodes = map[int]uint16{
	NOT_DEFINED:            0x0,
	FILE_NOT_FOUND:         0x1,
	ACCESS_VIOLATION:       0x2,
	DISK_FULL:              0x3,
	ILLEGAL_TFTP_OPERATION: 0x4,
	UNKNOWN_TRANSFER_ID:    0x5,
	FILE_ALREADY_EXISTS:    0x6,
	NO_SUCH_USER:           0x7,
}

func GenerateErrorMessage(message string, code int) []byte {
	ret := []byte{
		0x0,
		0x5,
		0x0,
		byte(ErrorCodes[code]),
	}

	ret = append(append(ret, []byte(message)...), 0x0)

	return ret

}
