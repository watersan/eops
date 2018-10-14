package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/internal"
	"opsv2.cn/opsv2/api/utils"
)

func TestUploadFile(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	uploadfile := "/Users/dale/golang/bin/Executor.zip"
	file, err := os.Open(uploadfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s failed: %v\n", uploadfile, err)
		return
	}
	defer file.Close()
	fstat, err := os.Stat(uploadfile)
	totalsize := fstat.Size()
	chunksize := 1048576
	var n int
	var i int64
	for {
		res := httptest.NewRecorder()
		p := make([]byte, chunksize)
		n, err = file.ReadAt(p, i)
		req := uploadFile(p[:n], "Executor.zip")
		if req == nil {
			return
		}
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", i, i+int64(n), totalsize))
		req.Header.Set("X-APP-Ver", "1.0")
		// tmp, err := json.MarshalIndent(req.Header, "", "  ")
		// fmt.Printf("err: %v\n", err)
		// fmt.Printf("req: %s\n", string(tmp))
		m.ServeHTTP(res, req)
		reader := io.LimitReader(res.Body, utils.MaxBodySize+1)
		b, _ := ioutil.ReadAll(reader)
		rmsg := common.RespMsg{}
		json.Unmarshal(b, &rmsg)
		if rmsg.Code != 0 {
			break
		}
		i += int64(n)

		if n < chunksize {
			fmt.Printf("UploadResult: %s\n", string(b))
			break
		}
	}
}

func uploadFile(p []byte, fname string) *http.Request {
	body := &bytes.Buffer{}
	//body.Write(p)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", fname)
	if err != nil {
		log.Fatalf("Create form failed: %v\n", err)
		return nil
	}
	part.Write(p)
	//_, err = io.Copy(part, file)

	// for key, val := range params {
	//     _ = writer.WriteField(key, val)
	// }
	err = writer.Close()
	if err != nil {
		log.Fatalf("read failed: %v\n", err)
	}
	params := url.Values{}
	params.Set("build", "yes")
	params.Set("apptpl", "nginx")
	req, err := internal.HttpReq("POST", "/eops/deploy/upload", body, params)
	if err != nil {
		log.Fatalf("Create request failed: %v\n", err)
		return nil
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

//TestWebHook 测试webhook
func TestWebHook(t *testing.T) {
	m, econf := internal.Init(RouteList, auth.Privilege)
	// jobs.StartJobExecutor(econf)
	res := httptest.NewRecorder()
	githook := GitHookReq{
		UserEmail: "dale_di@126.com",
		Ref:       "refs/tags/1.0",
		Project: &GitProject{
			Name:      "monitor-zabbix",
			NameSpace: "cmdb-mybbs",
		},
		Commits:           []*GitCommits{{ID: "11", Message: "asdfasdfasdf"}},
		TotalCommitsCount: 1,
	}
	reqbody, _ := json.Marshal(githook)
	//fmt.Printf("reqbody: %s\n", string(reqbody))
	//"http://127.0.0.1:5000"+
	req, err := http.NewRequest("POST", "http://127.0.0.1:5000"+econf.GOPT.Prefix+"/deploy/webhook", bytes.NewReader(reqbody))
	if err != nil {
		fmt.Printf("WebHook init request failed: %v\n", err)
		return
	}
	req.Header.Set("X-Gitlab-Token", "YLMuCgEWzR#k2pT9lzgSECk28uoYlNWtzhY0HwCrL9xOXnWKhP9B&tGx9EtNXLgF")
	req.Header.Set("X-Gitlab-Event", "Tag Push Hook")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// m.ServeHTTP(res, req)
	// res.Body = new(bytes.Buffer)
	client := &http.Client{Timeout: time.Second * 10}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		fmt.Printf("WebHook request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("GitHook complete\n")
	params := url.Values{}
	params.Set("appid", "5ab45e21796a85864f186262")
	params.Set("keyid", "Lk9UdRA7hG3Ipsls95Rd")
	params.Set("secret", "e5f17a1ca6e306411a700a36dea5e687")
	req, _ = http.NewRequest("GET", econf.GOPT.Prefix+"/deploy/history", nil)
	req.URL.RawQuery = params.Encode()
	res.Body = new(bytes.Buffer)
	m.ServeHTTP(res, req)
	fmt.Printf("%v\n", res.Body)
}

type ABC struct {
	ID    bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name  string
	Alias []byte
}

func TestABC(t *testing.T) {
	_, econf := internal.Init(nil, nil)
	dd := "bytes 123-10485759/202286246"
	i := strings.IndexByte(dd[6:], '-')
	j := strings.IndexByte(dd[6:], '/')
	fmt.Printf("%s %s %s\n", dd[6:6+i], dd[6+i+1:6+j], dd[6+j+1:])
	fmt.Printf("Time: %s\n", time.Now().Add(time.Duration(-86400)*time.Second).Format("200601021504"))
	fstat, err := os.Stat("deploy.go")
	if err != nil {
		fmt.Printf("err: %v\n", err)
	} else {
		fmt.Printf("Size: %d\n", fstat.Size())
	}
	appvar := common.V{
		"post":    "443",
		"dir":     "/data",
		"servers": common.V{"www": "www.daledi.cn"},
	}
	svr := appvar["servers"].(common.V)
	svr["static"] = "img.daledi.cn"
	fmt.Printf("appvar: %v\n", appvar)
	//abc := ABC{ID: bson.NewObjectId(), Name: "test"}
	abc := &ABC{}
	c := econf.Mgo.Coll("abc")
	err = c.Find(&common.V{"_id": bson.ObjectIdHex("5ab760cc796a852dffde79e4")}).One(abc)
	if err != nil {
		fmt.Printf("Insert ABC failed: %v\n", err)
	}
	fmt.Printf("ABC: %v\n", abc)
}
