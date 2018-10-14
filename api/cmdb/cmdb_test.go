package cmdb

import
//"fmt"

(
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"
	"opsv2.cn/opsv2/api/internal"
)

func TestAddItem(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	res := httptest.NewRecorder()
	postArgs := common.V{"name": "机锋论坛",
		"describe":    "",
		"opser":       "daledi",
		"secondopser": "daledi",
	}
	postStr, _ := json.Marshal(postArgs)
	req, _ := internal.HttpReq("POST", "/eops/1/cmdb/project", strings.NewReader(string(postStr)), nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	// req.Header.Set("Content-Type", "application/json")
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	res.Body = new(bytes.Buffer)
	postArgs = common.V{"name": "论坛php",
		"describe":    "",
		"project":     "598580495081faaeac8e5ecb",
		"depend":      []string{},
		"shared":      false,
		"opser":       "daledi",
		"secondopser": "daledi",
	}
	postStr, _ = json.Marshal(postArgs)
	req, _ = internal.HttpReq("POST", "/eops/1/cmdb/cluster", strings.NewReader(string(postStr)), nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	// req.Header.Set("Content-Type", "application/json")
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
	// res.Body = new(bytes.Buffer)
	// postArgs = common.V{"name": "php",
	// 	"describe":   "",
	// 	"cluster":    "598596f25081faaeac8e5eed",
	// 	"depend":     []string{},
	// 	"appconf":    []string{},
	// 	"hosts":      []string{},
	// 	"from":       "",
	// 	"type":       "apptpl",
	// 	"hostscount": 3,
	// 	"deploy":     1,
	// }
	// postStr, _ = json.Marshal(postArgs)
	// req, _ = http.NewRequest("POST", "/eops/1/cmdb/app", strings.NewReader(string(postStr)))
	// req.Header.Set("Cookie", "uid=dale_di@126.com; tid="+tid+"; sessid="+sid)
	// req.Header.Set("X-Forwarded-For", "127.0.0.1")
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// m.ServeHTTP(res, req)
	// fmt.Printf("%v\n", res.Body)
}
func TestUpdateItem(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	res := httptest.NewRecorder()
	postArgs := common.V{"name": "机锋论坛",
		"describe":    "aaaaa",
		"opser":       "daledi",
		"secondopser": "daledi",
	}
	postStr, _ := json.Marshal(postArgs)
	req, _ := internal.HttpReq("PUT", "/eops/1/cmdb/project", strings.NewReader(string(postStr)), nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	// req.Header.Set("Content-Type", "application/json")
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
}
func TestDeleteItem(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	res := httptest.NewRecorder()
	postArgs := common.V{"id": "5985800b5081faaeac8e5ec3"}
	postStr, _ := json.Marshal(postArgs)
	req, _ := internal.HttpReq("DELETE", "/eops/cmdb/project", strings.NewReader(string(postStr)), nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	// req.Header.Set("Content-Type", "application/json")
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
}
func TestGetItem(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	res := httptest.NewRecorder()
	req, _ := internal.HttpReq("GET", "/eops/cmdb/project", nil, nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	m.ServeHTTP(res, req)
	fmt.Printf("GetItem\n")
	fmt.Printf("%v\n", res.Body)
}

func TestGetAppInfo(t *testing.T) {
	_, econf := internal.Init(RouteList, auth.Privilege)
	appinfo, apptpl := AppNameInfo("mybbs", "BBSCode", "php", econf)
	fmt.Printf("appid: %v\napptpl: %v\n", appinfo, apptpl)
	apps := []MgoApp{}
	n, _ := econf.Mgo.Query(mgoApps, &common.V{"cluster": "598596f25081faaeac8e5eed"}, &apps, 10, 1, nil)
	if n > 0 {
		fmt.Printf("Apps: %v\n", apps)
	}
	if appinfo := mgoRecoder("nginx", mgoApps, econf); appinfo != nil {
		appinfo = appinfo.(*MgoApp)
		fmt.Printf("appinfo: %v\n", appinfo)
	}

	// data := common.V{"_id": "59868d47ac91868bcb6029df", "appver": "201803081740"}
	// UpdateApp(data, econf)

}
func TestGetHostList(t *testing.T) {
	_, econf := internal.Init(RouteList, auth.Privilege)
	c := []string{"5ab45e21796a85864f186262"}
	hlist := AssignedHosts(econf, "apps", c, "release")
	fmt.Printf("hostlist: %v\n", hlist)
	// hlist = AssignedHosts(econf, []string{"598580495081faaeac8e5ecb"}, []string{}, []string{}, "release")
	// fmt.Printf("%v\n", hlist)
	// hlist = AssignedHosts(econf, []string{"598580495081faaeac8e5ecb"}, c, []string{"59868da5ac91868bcb6029e0"}, "release")
	// fmt.Printf("%v\n", hlist)
	skey := common.V{
		"name":         "nginx",
		"additionconf": common.V{"$elemMatch": common.V{"$eq": "sites-availables"}},
	}
	//判断输入的path是否是模版中允许的
	rapptpl := []common.V{}
	n, _ := econf.Mgo.Query(mgoAppTpl, &skey, &rapptpl, 10, 1, common.V{"additionconf": 1})
	fmt.Printf("Result: %d\n", n)
	fmt.Printf("Test AppVar\n")
	appvar := common.V{
		"add": common.V{
			"var4": "44444",
			"var5": "5555",
		},
		"del": common.V{
			"var3": "33333",
		},
		"change": common.V{
			"var1": "changeto1",
			"var2": "changeto3",
		},
	}
	testv := common.V{}
	if len(testv) == 0 {
		fmt.Printf("testv OK\n")
	}
	ctx := &httpsvr.Context{}
	ctx.Keys = make(map[string]interface{})
	ctx.Keys["econf"] = econf
	ctx.Keys["workenv"] = "release"
	if err := appVar(appvar, "5aab82c6796a854d755c410d", ctx); err != nil {
		fmt.Printf("update appvar failed: %v\n", err)
	} else {
		fmt.Printf("update appvar successed\n")
	}
	//oid := bson.NewObjectId()
}

func TestTree(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	res := httptest.NewRecorder()
	req, _ := internal.HttpReq("GET", "/eops/cmdb/projecttree", nil, nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")
	m.ServeHTTP(res, req)
	fmt.Printf("Test Tree\n")
	fmt.Printf("%v\n", res.Body)

}

// func TestImport(t *testing.T) {
// 	m, _ := internal.Init(RouteList, auth.Privilege)
// 	res := httptest.NewRecorder()
// 	params := url.Values{}
// 	params.Set("id", "5aeb28bfa6856d7ac6b5d394")
// 	params.Set("env", "release")
// 	req, _ := internal.HttpReq("GET", "/eops/cmdb/resource/import", nil, params)
// 	req.Header.Set("X-Real-IP", "127.0.0.1")
// 	m.ServeHTTP(res, req)
// 	fmt.Printf("Import\n")
// 	fmt.Printf("%v\n", res.Body)
// }

func TestCSV(t *testing.T) {
	f, err := os.Open("/Users/dale/Documents/test.csv")
	if err != nil {
		fmt.Printf("csv: %v\n", err)
		return
	}
	csvr := csv.NewReader(f)
	record, err := csvr.Read()
	fields := map[string]int{
		"ip":       -1,
		"hostid":   -1,
		"hostname": -1,
		"cpu":      -1,
		"memory":   -1,
		"disk":     -1,
		"os":       -1,
		"osver":    -1,
	}
	for i, field := range record {
		if _, ok := fields[field]; ok {
			fields[field] = i
		}
	}
	if _, ok := fields["hostid"]; !ok {
		fmt.Printf("no hostid\n")
	}
	for {
		record, err := csvr.Read()
		if err == io.EOF || record[0] == "" {
			break
		}
		data := csvFields(record, fields)
		// fieldnum := len(record)
		// data := common.V{
		// 	"ip":       sliceIndex(record, fields["ip"], fieldnum),
		// 	"hostid":   sliceIndex(record, fields["hostid"], fieldnum),
		// 	"hostname": sliceIndex(record, fields["hostname"], fieldnum),
		// 	"cpu":      sliceIndex(record, fields["cpu"], fieldnum),
		// 	"memory":   sliceIndex(record, fields["memory"], fieldnum),
		// 	"disk":     sliceIndex(record, fields["disk"], fieldnum),
		// 	"os":       sliceIndex(record, fields["os"], fieldnum),
		// 	"osver":    sliceIndex(record, fields["osver"], fieldnum),
		// }
		fmt.Printf("Record: %v\n", data)
	}
	// resource := MgoResource{}
	// myref := reflect.TypeOf(&resource).Elem()
	// for i := 0; i < myref.NumField(); i++ {
	// 	fmt.Printf("Field: %s-%s\n", myref.Field(i).Name, myref.Field(i).Tag.Get("json"))
	// }
}

func csvFields(rec []string, fields map[string]int) common.V {
	fieldnum := len(rec)
	data := common.V{}
	for field, i := range fields {
		data[field] = ""
		if i < fieldnum && i >= 0 {
			data[field] = rec[i]
		}
	}
	return data
}
func sliceIndex(strSlice []string, i, l int) string {
	if i >= l || i < 0 {
		return ""
	}
	return strSlice[i]
}
