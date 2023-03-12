package src

import (
	"encoding/binary"
	"reflect"
	"testing"
)

func Test_opcode(t *testing.T) {
	type args struct {
		o uint16
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Read request",
			want:    "RRQ",
			wantErr: false,
			args:    args{o: binary.BigEndian.Uint16(append(append([]byte{0x0, 0x1}, []byte("hello.txt")...), 0x0)[0:2])},
		},
		{
			name:    "Data request",
			want:    "DATA",
			wantErr: false,
			args:    args{o: 0x3},
		},
		{
			name:    "Junk data",
			wantErr: true,
			args:    args{o: 0x69},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := opcode(tt.args.o)
			if (err != nil) != tt.wantErr {
				t.Errorf("opcode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("opcode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatagramBuffer_ReadUntilDelimiter(t *testing.T) {
	type fields struct {
		Buffer []byte
		Offset int
	}
	tests := []struct {
		name           string
		fields         fields
		want           []byte
		offsetShouldBe int
	}{
		{
			name: "Empty case",
			fields: struct {
				Buffer []byte
				Offset int
			}{Buffer: []byte{}, Offset: 0},
			want:           []byte{},
			offsetShouldBe: 0,
		},
		{
			name: "Delimiter only",
			fields: struct {
				Buffer []byte
				Offset int
			}{Buffer: []byte{0x0}, Offset: 0},
			want:           []byte{},
			offsetShouldBe: 1,
		},
		{
			name: "Delimiter after two bytes",
			fields: struct {
				Buffer []byte
				Offset int
			}{Buffer: []byte{0x1, 0x1, 0x0}, Offset: 0},
			want:           []byte{0x1, 0x1},
			offsetShouldBe: 3,
		},
		{
			name: "Starts after delimiter",
			fields: struct {
				Buffer []byte
				Offset int
			}{Buffer: []byte{0x1, 0x1, 0x0, 0x2, 0x2}, Offset: 3},
			want:           []byte{0x2, 0x2},
			offsetShouldBe: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DatagramBuffer{
				Buffer: tt.fields.Buffer,
				Offset: tt.fields.Offset,
			}
			if got := d.ReadUntilDelimiter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadUntilDelimiter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_destructureDatagramRRQ(t *testing.T) {

	opcode01 := []byte{0x0, 0x1}
	filenameByteArray := append(append(opcode01, []byte("hello.txt")...), 0x0)
	octectPayloadByteArray := append(append(filenameByteArray, []byte("octet")...), 0x0)

	type args struct {
		d   DatagramBuffer
		ret *Datagram
	}
	tests := []struct {
		name         string
		args         args
		wantErr      bool
		wantFilename string
		wantMode     string
	}{
		{
			name: "read in filename",
			args: args{
				d: DatagramBuffer{
					Buffer: octectPayloadByteArray,
					Offset: 0,
				},
				ret: &Datagram{
					Opcode:      "",
					RrqFilename: "",
					RrqMode:     "",
				},
			},
			wantErr:      false,
			wantFilename: "hello.txt",
		},
		{
			name: "Check mode",
			args: args{
				d: DatagramBuffer{
					Buffer: octectPayloadByteArray,
					Offset: 0,
				},
				ret: &Datagram{
					Opcode:      "",
					RrqFilename: "",
					RrqMode:     "",
				},
			},
			wantErr:      false,
			wantFilename: "hello.txt",
			wantMode:     "octet",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := destructureDatagramRRQ(tt.args.d, tt.args.ret); (err != nil) != tt.wantErr {
				t.Errorf("destructureDatagramRRQ() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFilename != tt.args.ret.RrqFilename {
				t.Errorf("destructureDatagramRRQ() error = filename wrong wanted '%v' got '%v'", tt.wantFilename, tt.args.ret.RrqFilename)
			}
		})
	}
}
