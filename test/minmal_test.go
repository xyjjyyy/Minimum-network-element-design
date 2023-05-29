package main

import (
	"fmt"
	"testing"
	"time"
)

func TestString(t *testing.T) {
	s1 := "0000000000011101100001101100111001011001001001000010001110011011010000001001000111100010001111011011000000100100010101000110010101110011011101000011100011111111110101010100101001011010"
	s2 := "0000000000011101100001101100111001011001001001000010001110011011010000001001000111100010001111011011000000100100010101000110010101110011011101000011100011111111110101010100101001011010"
	fmt.Println(len(s1), len(s2))
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			fmt.Println(i)
			break
		}
	}
}

func TestAdd(t *testing.T) {
	add([]string{"13", "324", "43"}...)
}

func add(strList ...string) {
	fmt.Println(strList)
}

func TestChan(t *testing.T) {
	ch := make(chan string, 10)
	ch <- "fdaf"
	ch <- "fadsf"

	go func() {
		for s := range ch {
			fmt.Println(s)
		}
	}()
	time.Sleep(time.Second)
	close(ch)
}

func TestScan(t *testing.T) {

	a := "fasdfas\n"
	fmt.Println(a[len(a)-1])
}
