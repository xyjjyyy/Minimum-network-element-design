package main

import (
	"fmt"
	"math/rand"
	"network/utils"
	"strings"
	"time"
)

func main() {
	t := utils.ByteToBit([]byte{127, 1})
	fmt.Println(t)
	s := "86CE59A1239B"
	s = strings.ToLower(s)
	// s="1234567891233"
	b := utils.Hex12ToBit48(s)
	fmt.Println(b)

	h := utils.Bit48ToHex12(b)
	fmt.Println(h)
}

func test() string {
	var s string

	rand.Seed(time.Now().UnixMicro())
	for len(s) < 12 {
		randNum := rand.Intn('f' - '0')

		if randNum < 'a'-'0' && randNum > '9'-'0' {
			continue
		}
		t := rune('0' + randNum)
		s += string(t)
	}
	return s
}
