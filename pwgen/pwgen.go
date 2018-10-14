package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"
)

var (
	pwlen = flag.Int("l", 20, "passwd length")
	pwnum = flag.Int("n", 10, "password amount")
	pwsc  = flag.Bool("s", false, "special character")
)

//48-57:数字 65-90:大写字母 97-122:小写字母 10+26+26=62
func RandomStr(l int, sc bool) []byte {
	b := make([]byte, l)
	for i := 0; i < l; i++ {
		var n uint8
		rnum := random(1000)
		if sc {
			n = uint8(rnum%94 + 32)
		} else {
			n = uint8(rnum%62 + 48)
			if n > 57 && n < 84 {
				n += 7
			} else if n >= 84 {
				n += 13
			}
		}
		b[i] = n
	}
	return b
}

func random(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

func main() {
	flag.Parse()
	for i := 0; i < *pwnum; i++ {
		fmt.Printf("%s\n", RandomStr(*pwlen, *pwsc))
	}
}
