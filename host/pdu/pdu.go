package pdu

const (
	fullFrameMinLength = 18

	DataFrameType = 0
	AckFrameType  = 1
	ErrFrameType  = 2
)

type Frame struct {
	Data      []byte        // 数据data
	SerialNum uint64        // 帧序号
	FrameType int           // 帧类型(0-data 1-ack 2-err)
	MacSource string        //源Mac
	MacTarget string        //目标Mac
	Locator   string        // 定位符
	CRC32     string        // CRC32校验码
	Duration  int           // 持续时间 后面单位是毫秒(ms)
	Done      chan struct{} // 传递时钟结束的信息
}
