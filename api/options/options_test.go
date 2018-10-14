package options

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestOpt(t *testing.T) {
	gopt := GlobalOPT{}
	ReadConfig("opsv2.conf", &gopt)
	tmp, _ := json.Marshal(gopt)
	fmt.Printf("%s\n", string(tmp))
	// opt := GlobalOPT{}
	// myref := reflect.ValueOf(&opt).Elem()
	// f := myref.FieldByName("SessionExpire")

	// f.SetString("aaaa")
	// ips := ipsearch.IPRanges(make([]*net.IPNet, 2))
	// ips.ConvIPList([]string{"127.0.0.1/32", "192.168.0.0/16"})

	// f.Set(reflect.ValueOf(ips))
	// f.SetInt(int64(4))
	// fmt.Printf("f: %s\n", f.Type().Name())
	// for i := 0; i < myref.NumField(); i++ {
	// 	field := myref.Field(i)
	// 	fmt.Printf("field: %s\n", field.Name)
	// }
}
