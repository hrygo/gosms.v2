package cmpp

import (
	"encoding/binary"
	"testing"

	"github.com/hrygo/gosms/bootstrap"
	"github.com/hrygo/gosms/codec"
)

var _ = bootstrap.BasePath

func TestMessageHeader_Encode(t *testing.T) {
	header := MessageHeader{
		TotalLength: 16,
		CommandId:   CMPP_CONNECT,
		SequenceId:  uint32(codec.B32Seq.NextVal()),
	}
	t.Logf("head: %v", header)
	t.Logf("head: %#x", header.Encode())

	connect := Connect{MessageHeader: header}

	connect.Encode()

}

func TestMessageHeader_Decode(t *testing.T) {
	frame := make([]byte, 16)
	binary.BigEndian.PutUint32(frame[0:4], 16)
	binary.BigEndian.PutUint32(frame[4:8], uint32(CMPP_CONNECT))
	binary.BigEndian.PutUint32(frame[8:12], 1)
	copy(frame[12:16], "1234")

	header := MessageHeader{}
	_ = header.Decode(frame)
	t.Logf("%v", header)
	t.Logf("%s", frame[12:16])
	t.Logf("%#x", frame[12:16])
}
