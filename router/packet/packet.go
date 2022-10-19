package packet

const (
	MinLengthOfPacket = 1 + 1 + 4 + 4
)

type Packet struct {
	Data     []byte
	IPSource string
	IPTarget string
	FrameType []byte
	FrameSerialNUm []byte
}

func New() *Packet {
	return &Packet{}
}
