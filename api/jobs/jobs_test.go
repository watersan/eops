package jobs

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"testing"

	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/internal"

	"gopkg.in/mgo.v2/bson"
)

//"encoding/json"
//"opsv2.cn/opsv2/api/cache"

type ABC struct {
	ID   bson.ObjectId `_id`
	Name string        `json:"name"`
}

func TestExec(t *testing.T) {
	executor := exec.Command(
		"/Users/dale/dd.sh",
		// 		// "/usr/local/bin/ansible",
		// 		// "dev.daledi.cn",
		// 		// "-b",
		// 		// "-m command",
		// 		// "-a",
		// 		// "dd.sh",
	)
	//
	// 	fmt.Printf("cmd: %s %v\n", executor.Path, executor.Args)
	executor.Env = []string{"ABC=111"}
	executor.Start()
	//fmt.Printf("sys: %v\n", os.Environ())
	executor.Wait()
	// 	//out, _ := executor.Output()
	// 	//fmt.Printf("cmd: %v", string(out))
}

func TestGetJobs(t *testing.T) {
	m, _ := internal.Init(RouteList, auth.Privilege)
	res := httptest.NewRecorder()
	//name=build.sh&path=deploy%2Fzabbix&releasejob=yes
	query := url.Values{}
	query.Set("name", "build.sh")
	query.Set("path", "deploy/zabbix")
	query.Set("releasejob", "yes")
	req, _ := internal.HttpReq("GET", "/eops/jobs/jobs", nil, query)
	// eq.URL.RawQuery = query.Encode()
	m.ServeHTTP(res, req)
	fmt.Printf("GetJobs: %v\n", res.Body)
}

func TestFlowsRun(t *testing.T) {
	_, econf := internal.Init(RouteList, auth.Privilege)
	task := MgoJobLogs{
		TaskName: "testFlow",
		Target:   "apps",
		List:     []string{"5ab45e21796a85864f186262"},
		CDEnv:    "test",
		Type:     "JobsFlow",
		Operator: "dale_di@126.com",
	}
	hlist := task.getHostList(econf)
	fmt.Printf("FlowRun: hlist: %v\n", hlist)
	job := GetJobByName(econf, "deploy/zabbix/build.sh", "")
	econf.LogDebug1(job)
	//测试作业流执行
	// res := httptest.NewRecorder()
	// JobsQueue = make(chan *JobExecutor, econf.GOPT.ParallelJobExe)
	// reqbody, _ := json.Marshal(task)
	// req, _ := internal.HttpReq("PUT", "/eops/jobrun", bytes.NewReader(reqbody), nil)
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// m.ServeHTTP(res, req)
	// fmt.Printf("FlowRun: %v\n", res.Body)
	// je := <-JobsQueue
	// je.CallExecutor(econf)

}

func TestABC(t *testing.T) {
	fmt.Printf("ABC\n")
	jobrun := JobRunArgs{
		Projects: []string{"asdfas", "xcv@ses"},
	}
	v := common.V{"projects": []string{"asdfas", "xcvses"}}
	o := common.V{}
	if err := common.InputFilter(&jobrun, 3, v, o); err != nil {
		fmt.Printf("invalid input: %v\n", err)
	}
	fmt.Printf("%v\n", o)
	var body []byte
	abc := ABC{}
	err := json.Unmarshal(body, &abc)
	if err != nil {
		fmt.Printf("ABC: %v\n", err)
	}
	//err := errors.New("text")
}

func ddd(v interface{}) {
	var cnt string
	switch v.(type) {
	case string:
		cnt = v.(string)
	case error:
		cnt = v.(error).Error()
	}
	fmt.Printf("ddd: %s\n", cnt)
}
