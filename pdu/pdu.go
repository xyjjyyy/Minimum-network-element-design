package pdu

import (
	"errors"
	"fmt"
	"hash/crc32"
	"network/utils"
)

const (
	fullFrameMinLength = 18

	DataFrameType = 0
	AckFrameType  = 1
)

type Frame struct {
	Data      []byte        // 数据data
	SerialNum uint64        // 帧序号
	FrameType int           // 帧类型(0-data 1-ack)
	MacSource string        // 源Mac
	MacTarget string        // 目标Mac
	Locator   string        // 定位符
	CRC32     string        // CRC32校验码
	Duration  int           // 持续时间 单位是毫秒(ms)
	Done      chan struct{} // 传递时钟结束的信息
}

// SplitFrame 将 frame 进行分割，返回我们的帧结构(*frame)和error
func SplitFrame(dataWithLocator []byte) (*Frame, error) {
	fullFrame, err := removeLocator(dataWithLocator)
	if err != nil {
		return nil, err
	}
	l := len(fullFrame)
	if l < fullFrameMinLength {
		return nil, errors.New("the frame waiting to be parsed is smaller than the minimum length")
	}

	dataWithoutCrc := fullFrame[:l-4]
	crc32Uint := utils.Code(fullFrame[l-4 : l])
	crc := crc32.ChecksumIEEE(dataWithoutCrc)
	if uint64(crc) != crc32Uint {
		// 若CRC32校验出现错误 直接丢弃 等待超时重传即可
		return nil, errors.New("CRC not pass")
	}

	frameSerialNum := utils.Code(fullFrame[1:2])
	frameType := utils.Code(fullFrame[:1])
	macSource := utils.Bit48ToHex12(utils.ByteToBit(fullFrame[2:8]))
	macTarget := utils.Bit48ToHex12(utils.ByteToBit(fullFrame[8:14]))
	data := fullFrame[14 : l-4]

	frame := New()
	frame.Data = data
	frame.FrameType = int(frameType)
	frame.SerialNum = frameSerialNum
	frame.MacSource = macSource
	frame.MacTarget = macTarget
	frame.GetCRC32()

	return frame, nil
}

// 移除定位符
func removeLocator(dataWithLocator []byte) ([]byte, error) {
	s := utils.ByteToBit(dataWithLocator)
	firstLoactor, secondLocator := -1, -1
	num := 0
	frequency := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			num++
		} else {
			num = 0
		}
		if num == 6 {
			frequency++
			i += 2
			if frequency == 1 {
				firstLoactor = i
			} else if frequency == 2 {
				secondLocator = i - 8
			}
			num = 0
		}
	}
	if frequency != 2 {
		return nil, fmt.Errorf("locator not found")
	}
	s = s[firstLoactor:secondLocator]

	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			num++
		} else {
			num = 0
		}
		if num == 5 {
			s = s[:i+1] + s[i+2:]
			num = 0
		}
	}
	// 转换成byte类型
	return utils.BitToByte(s), nil
}
