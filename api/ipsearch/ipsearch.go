package ipsearch

import (
	"bytes"
	"io"
	"net"
	"os"
	"sort"
)

//IPRanges IP段列表
type IPRanges []*net.IPNet

//Sort 对IpRanges进行升序排序。
func (ipr IPRanges) Sort() { sort.Sort(ipr) }

//ConvIPList 将ip的字符串列表转换为IPRanges。返回值如果为0，表明ipList为空或内容不合法
func (ipr IPRanges) ConvIPList(ipList []string) int {
	num := len(ipList)
	if num == 0 {
		ipr = make([]*net.IPNet, 1)
		_, ipr[0], _ = net.ParseCIDR("0.0.0.0/32")
		return 0
	}
	var n int
	for _, ipstr := range ipList {
		if _, ipnet, err := net.ParseCIDR(ipstr); err == nil {
			ipr[n] = ipnet
			n++
		}
	}
	ipr.Sort()
	return n
}

func (ipr IPRanges) Len() int { return len(ipr) }

func (ipr IPRanges) Less(i, j int) bool { return bytes.Compare(ipr[i].IP, ipr[j].IP) < 0 }

func (ipr IPRanges) Swap(i, j int) { ipr[i], ipr[j] = ipr[j], ipr[i] }

//Search 搜索指定的ip在IPRanges的位置
func (ipr IPRanges) Search(ip net.IP) *net.IPNet {
	ip = ip.To4()
	i := ipr.search(ip)
	if i < 0 {
		return nil
	}
	if i > 0 {
		i--
	}
	if ipr[i].Contains(ip) {
		return ipr[i]
	}
	return nil
}

//Check 检查addr是否在ipr的范围内
func (ipr IPRanges) Check(addr interface{}) bool {
	var ip net.IP
	switch iptmp := addr.(type) {
	case string:
		ip = net.ParseIP(iptmp).To4()
	case net.IP:
		ip = ip.To4()
	default:
		return false
	}
	if ipr.Search(ip) == nil {
		return false
	}
	return true
}

/*
	要注意的是：
	1、使用ParseIP进行转换时，必须跟上To4()。因为ParseIP转换的ip，长度为16，而ParseCIDR转换后ip的长度是4.
	   net.ParseIP("213.89.234.123").To4()
	2、函数返回的索引减一才是包含ip的net。
*/
func (ipr IPRanges) search(ip net.IP) int {
	return sort.Search(len(ipr), func(i int) bool {
		return bytes.Compare(ipr[i].IP, ip) > 0
	})
}

//ReadIPList 从文件读取ip列表
func ReadIPList(ipfile string) (IPRanges, error) {
	// "/Users/dale/sysadmin/ipz"
	ipz, err := os.OpenFile(ipfile, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	ipranges := IPRanges{}
	buf := make([]byte, 1024)
	var offset int64
	var n int
	for {
		n, err = ipz.ReadAt(buf, offset)
		if err != nil && err != io.EOF {
			break
		}
		offset += int64(n)
		var i int
		if buf[0] > 47 && buf[0] < 58 {
			i = 0
		} else {
			i = bytes.IndexByte(buf[:n], '\n')
			i++
		}
		for {
			j := bytes.IndexByte(buf[i:n], ';')
			if j >= 0 {
				var ipnet *net.IPNet
				if _, ipnet, err = net.ParseCIDR(string(buf[i : i+j])); err == nil {
					ipranges = append(ipranges, ipnet)
					// } else {
					// 	fmt.Printf("i: %d; j: %d; n: %d\n", i, i+j, n)
					// 	fmt.Printf("ipnet: %s\n", string(buf[i:i+j]))
				}
				i += j
				if j = bytes.IndexByte(buf[i:n], '\n'); j >= 0 {
					i += j + 1
				}
			}
			if j < 0 {
				offset -= int64(len(buf[i:n]))
				break
			}
		}
		if err == io.EOF {
			err = nil
			break
		}
	}
	sort.Sort(ipranges)
	return ipranges, err
}
