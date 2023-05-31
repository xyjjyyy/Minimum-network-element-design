package phy

import (
	"math/rand"
	"network/utils"
	"time"
)

const randNum = 1000000000
const (
	Normal = iota
	LittleHard
	ExtraHard
)

const (
	change0 = iota
	change1
	change2
)

type Phy struct {
	Prob int // 难度
}

func New(prob int) *Phy {
	return &Phy{
		Prob: prob,
	}
}

func (p *Phy) HandleFrame(frame []byte) []byte {
	var frameSend string
	s := utils.ByteToBit(frame)
	for _, f := range s {
		switch p.judge() {
		case change0:
			frameSend += string(f)
		case change1:
			if f == '1' {
				frameSend += "0"
			} else {
				frameSend += "1"
			}
		case change2:
		}
	}

	return utils.BitToByte(frameSend)
}

// 随机生成概率
func (p *Phy) judge() int {
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(randNum)
	var prob float32
	switch p.Prob {
	case Normal:
		return change0
	case LittleHard:
		prob = 0.00001 // 十万分之二的错误率(每bit)
	case ExtraHard:
		prob = 0.0005 // 十万分之十的错误率(每bit)
	}
	// fmt.Println(r)
	if r < int(randNum*prob) {
		return change1
	}
	if r < int(randNum*prob*2) {
		return change2
	}
	return change0
}
