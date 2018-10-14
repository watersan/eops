package common

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestV(t *testing.T) {
	fmd5, err := FileMd5("/Users/dale/.profile", []byte{})
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	fmt.Printf("md5: %s\n", hex.EncodeToString(fmd5))
	// b := letter(16)
	// encryptData, _ := AesEncrypt(b, Aeskey)
	// fmt.Printf("%x\n", encryptData)
}

func TestStrFilter(t *testing.T) {
	fmt.Printf("%v,", StrFilter("dale_2笛zxvs", "nickname"))
	fmt.Printf("%v,", StrFilter("admin_di.dale", "username"))
	fmt.Printf("%v,", StrFilter("dale.di@me.com", "email"))
	fmt.Printf("%v,", StrFilter("123.13.1.123", "ip"))
	fmt.Printf("%v,", StrFilter("23452", "number"))
	fmt.Printf("%v,", StrFilter("18610116661", "mobile"))
	fmt.Printf("%v,", StrFilter("http/tcp/80,dns/udp/53", "port"))
	fmt.Printf("%v,", StrFilter("abc.sh", "scriptname"))
	fmt.Printf("%v,", StrFilter("59993bf95081faaeac8e5f50", "id"))
	fmt.Printf("%v\n", StrFilter("--help xvew_dd -a c3 --bb 3ra-w.pl", "argv"))
	str := "php"
	notnull := 0
	lmin := 1
	lmax := 128
	strlen := len(str)
	if (strlen > 0 || notnull == 1) && lmax > 0 &&
		(strlen < lmin || strlen > lmax || !StrFilter(str, "username")) {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("NO\n")
	}
	//fmt.Printf("%v\n", StrFilter("php", "username"))
	//fmt.Printf("%v",StrFilter("", ))
}
