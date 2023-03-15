package src

import "testing"

func TestAcknowledgeRRQSession(t *testing.T) {
	type args struct {
		sess     *RRQSession
		datagram Datagram
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		wantBlockNumber uint16
		wantCompleted   bool
	}{
		{
			name: "Test normal incrementing ack",
			args: args{
				sess: &RRQSession{
					Filename:         "",
					RequesterAddress: nil,
					BlockNumber:      1,
					FileData:         []byte("this is the last bit"),
					Completed:        false,
				},
				datagram: Datagram{
					Opcode:   OPCODE_ACK,
					Filename: "",
					Mode:     "",
					AckBlock: 1,
				},
			},
			wantCompleted:   false,
			wantBlockNumber: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := AcknowledgeRRQSession(tt.args.sess, tt.args.datagram); (err != nil) != tt.wantErr {
				t.Errorf("AcknowledgeRRQSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.args.sess.BlockNumber != tt.wantBlockNumber {
				t.Errorf("AcknowledgeRRQSession() wantBlockNumber %v, got %v", tt.wantBlockNumber, tt.args.sess.BlockNumber)
			}
		})
	}
}
