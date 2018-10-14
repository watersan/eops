package ipsearch

import (
	"fmt"
	"log"
	"net"
	"testing"
)

//func BenchmarkIpRange(b *testing.B) {
func TestIpRange(t *testing.T) {
	ipranges, _ := ReadIPList("/Users/dale/sysadmin/ipz")
	ip := net.ParseIP("1.2.99.1").To4()
	// ip1 := net.ParseIP("94.127.120.0")
	if ipranges.Len() == 0 {
		log.Fatalf("ip ranges is null\n")
	}
	r := ipranges.Search(ip)
	fmt.Printf("IPnum: %d\n", ipranges.Len())
	if r != nil {
		fmt.Printf("%s\n", r.IP.String())
	} else {
		fmt.Printf("no found\n")
	}
	// fmt.Printf("IPnum: %d; ip: %s\n", i, ipranges[i-1].IP.String())

	// for i := 0; i < 10000; i++ {
	// 	ipranges.Search(ip)
	// }
}
