package packet

import (
	"errors"
	"strconv"
	"strings"
)

func SplitFrameAndCreatePacket(packetByte []byte) (*Packet, error) {
	l := len(packetByte)

	if l < MinLengthOfPacket {
		return nil, errors.New("packet 长度小于最低标准")
	}

	IPSource := ByteToIP(packetByte[2:6])
	IPTarget := ByteToIP(packetByte[6:10])

	packet := New()

	packet.Data = packetByte[10:]
	packet.FrameSerialNUm = packetByte[1:2]
	packet.FrameType = packetByte[:1]
	packet.IPSource = IPSource
	packet.IPTarget = IPTarget

	return packet, nil
}


func ByteToIP(b []byte) string {
	var IP string
	for i := 0; i < len(b); i++ {
		IP += strconv.Itoa(int(b[i])) + "."
	}
	return IP[:len(IP)-1]
}

func IPToByte(IP string) []byte {
	s := strings.Split(IP, ".")
	b := make([]byte, len(s))

	for i := 0; i < len(s); i++ {
		t, _ := strconv.Atoi(s[i])
		b[i] = byte(int(t))
	}
	return b
}
