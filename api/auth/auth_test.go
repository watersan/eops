package auth

import
//"fmt"

(
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/internal"
)

func TestFunc(t *testing.T) {
	dd := "uid=dale,ou=People,dc=58coin,dc=com"
	i := strings.IndexByte(dd, ',')
	fmt.Printf("%s\n", dd[4:i])
}
func TestLogin(t *testing.T) {
	m, econf := internal.Init(RouteList, Privilege)
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", econf.GOPT.Prefix+"/auth/login?name=dale_di@126.com&passwd=OGtoadQbiX32M1egxKEl", nil)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	req, _ = http.NewRequest("GET", econf.GOPT.Prefix+"/auth/login?%d=%333", nil)
	res.Body = new(bytes.Buffer)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	req, _ = http.NewRequest("GET", econf.GOPT.Prefix+"/auth/login?name=33", nil)
	res.Body = new(bytes.Buffer)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	req, _ = http.NewRequest("GET", econf.GOPT.Prefix+"/auth/login?name=33&passwd=s", nil)
	res.Body = new(bytes.Buffer)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	req, _ = http.NewRequest("GET", econf.GOPT.Prefix+"/auth/login?name=dale_di@126.com&passwd=s", nil)
	res.Body = new(bytes.Buffer)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	req, _ = http.NewRequest("GET", econf.GOPT.Prefix+"/auth/login?name=test@126.com&passwd=123456789", nil)
	res.Body = new(bytes.Buffer)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)

	skey := bson.M{"name": "dale_di@126.com"}
	usersession := ZDUserSession{}
	c := econf.Mgo.Coll(mgoUsers)
	err := c.Find(skey).One(&usersession.MgoUser)
	if err != nil {
		fmt.Printf("query failed: %v\n", err)
		return
	}
	fmt.Printf("dale: %v\n", usersession)
}

// func TestChangepwd(t *testing.T) {
// 	m, econf := testinit()
// 	res := httptest.NewRecorder()
// 	postArgs := bson.M{"name": "test1@126.com",
// 		"oldpasswd": "12345678",
// 		"newpasswd": "12345678",
// 	}
// 	postStr, _ := json.Marshal(postArgs)
// 	req, _ := http.NewRequest("PUT", econf.GOPT.Prefix+"1/auth/pwd", strings.NewReader(string(postStr)))
// 	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	req.Header.Set("Cookie", "uid=test1@126.com")
// 	m.ServeHTTP(res, req)
// 	fmt.Printf("%v\n", res.Body)
// }

func TestPrevilage(t *testing.T) {
	_, econf := internal.Init(RouteList, Privilege)
	// res := httptest.NewRecorder()
	// req, _ := http.NewRequest("GET", econf.GOPT.Prefix+"/auth/user?keyid=Lk9UdRA7hG3Ipsls95Rd&secret=e5f17a1ca6e306411a700a36dea5e687", nil)
	// m.ServeHTTP(res, req)
	//fmt.Printf("%v\n", res.Body)
	userinfo, _ := GetUserInfo(econf, "dale_di@126.com", 3)
	ustr, _ := json.MarshalIndent(userinfo, "", "  ")
	fmt.Printf("%s\n", string(ustr))
	rRoles := common.V{}
	_, err := econf.Mgo.Query(mgoRoles, "admin", &rRoles, 0, 0, nil)
	if err == nil {
		fmt.Printf("Roles: %v\n", rRoles)
	}
	mm := common.V{}
	delete(mm, "a")
	if len(mm) == 0 {
		fmt.Print("OK\n")
	}
	key := "abcd/test.sh"
	var skey string
	var skeylen int
	for {
		if skey != "" {
			skey += "/"
		}
		skeylen = len(skey)
		i := strings.IndexByte(key[skeylen:], '/')
		if i < 0 {
			skey += key[skeylen:]
		} else {
			skey += key[skeylen : skeylen+i]
		}
		fmt.Printf("%s\n", skey)
		if i < 0 {
			break
		}
	}
	//fmt.Printf("%d:%d\n", APIPermJob, APIPermAllow|APIPermJobRun|APIPermJobRead|APIPermJobModify)
}
