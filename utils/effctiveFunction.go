package utils

import (
	"fmt"
	"strconv"
)

var frameSerialNum uint64 = 0

// GenerateFrameSerialNum 生成帧序号
func GenerateFrameSerialNum() uint64 {
	frameSerialNum += 1
	return frameSerialNum
}

func ByteToBit(data []byte) string {
	var s string
	for i := 0; i < len(data); i++ {
		temp := int64(data[i])
		s += fmt.Sprintf("%08s", strconv.FormatInt(temp, 2))
	}
	return s
}

func BitToByte(s string) []byte {
	var res []byte

	// 转化为我们需要的 []byte 数组类型
	for i := 0; i < len(s); i += 8 {
		var b int
		for j := 0; j < 8 && i+j < len(s); j++ {
			if string(s[i+j]) == "1" {
				b += 1 << (7 - j)
			}
		}

		res = append(res, byte(b))
	}

	return res
}

// 将一个 []byte 转换成 uint64 类型
// 用于 CRC32 校验
func Code(dataByte []byte) uint64 {
	l := len(dataByte)
	var s string
	for i := 0; i < l; i++ {
		temp := uint64(dataByte[i])
		s += fmt.Sprintf("%08b", temp)
	}

	l = len(s)
	var dataUint uint64

	for i := 0; i < l; i++ {
		if s[i] == '1' {
			dataUint += 1 << (l - i - 1)
		}
	}

	return dataUint
}

// Hex12ToBit48 并非要真正把它们转换成二进制 而是要制定一套规则 编解码的规则
func Hex12ToBit48(hexString string) string {
	base, _ := strconv.ParseInt(hexString, 16, 64)
	return fmt.Sprintf("%48s", strconv.FormatInt(base, 2))
}

func Bit48ToHex12(bitString string) string {
	base, _ := strconv.ParseInt(bitString, 2, 64)
	return strconv.FormatInt(base, 16)
}
