package pdu

import (
	"fmt"
	"hash/crc32"
	"network/utils"
	"strconv"
)

func New() *Frame {
	return &Frame{
		Locator: "01111110",
		// TODO：有机会可以改为动态调整
		Duration: 500, //设置超时时间500ms
		Done:     make(chan struct{}),
	}
}

// GetCRC32  返回CRC32校验码
func (f *Frame) GetCRC32() {
	dataNeedCRC := utils.BitToByte(f.getDataNeedCRC())
	crc := crc32.ChecksumIEEE(dataNeedCRC)
	f.CRC32 = fmt.Sprintf("%032s", strconv.FormatUint(uint64(crc), 2))
}

// 获得需要CRC32校验的部分（包含 帧类型 帧序号 数据 ）的对应二进制字符串
func (f *Frame) getDataNeedCRC() string {
	// 添加帧类型(1)
	s := fmt.Sprintf("%08s", strconv.FormatInt(int64(f.FrameType), 2))
	// 添加帧序号(1)
	s += fmt.Sprintf("%08s", strconv.FormatUint(f.SerialNum, 2))
	// 添加mac地址(12=6+6)
	s += utils.Hex12ToBit48(f.MacSource)
	s += utils.Hex12ToBit48(f.MacTarget)
	// 添加data数据部分(>=0)
	s += utils.ByteToBit(f.Data)
	return s
}

// AddLocator 加上定位符 并返回最终的生成的帧(byte类型)
func (f *Frame) AddLocator() []byte {
	// 获得需要CRC32校验的部分 包含 帧类型(1) 帧序号(1) mac地址(12=6+6) 数据
	s := f.getDataNeedCRC()
	// 添加 CRC32 校验码(4)
	s += f.CRC32
	num := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '1' {
			num++
		} else {
			num = 0
		}

		if num == 5 {
			s = s[:i+1] + "0" + s[i+1:]
			num = 0
			i++
		}
	}
	// 添加定位符(2)
	s = f.Locator + s + f.Locator
	return utils.BitToByte(s)
}
