package pdu

import "network/utils"

func CreateFrameByte(data string, macSource string, macTarget string) []byte {
	frame := CreateFrame(data, macSource, macTarget)
	return frame.AddLocator()
}

func CreateFrame(data string, macSource string, macTarget string) *Frame {
	frame := New()
	frame.FrameType = 0
	frame.SerialNum = utils.GenerateFrameSerialNum()
	frame.Data = []byte(data)
	frame.MacSource = macSource
	frame.MacTarget = macTarget
	// 获得CRC校验码
	frame.GetCRC32()

	return frame
}

func CreateAckFrameByte(SerialNum uint64, macSource string, macTarget string) []byte {
	frame := New()
	frame.FrameType = 1
	frame.SerialNum = SerialNum
	frame.Data = []byte{}
	frame.MacSource = macSource
	frame.MacTarget = macTarget

	frame.GetCRC32()
	return frame.AddLocator()
}
