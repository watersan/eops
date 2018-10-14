package jobs

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/cmdb"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"
)

/*
curl --request GET --header 'PRIVATE-TOKEN: 9xesKpc-AHAzrv35bwsb' 'https://git.gfan.com/api/v3/projects/46/repository/files?file_path=README.md&ref=20b2bd6a8e0f79103a43272e7608d15e349b72d1'
curl --request GET --header 'PRIVATE-TOKEN: 4X-w5kqayVKGpgtdoMsy' https://code.daledi.cn/api/v4/projects
GET /projects/:id/repository/tree

db.test.insert({name:"test1",vv:[{host:"u1",log:"aaaa"},{host:"u2",log:"bbbb"}]});
db.test.insert({
	name:"test2",
	value:{
		"pathm1":[
			{name:"aaaa"},
			{name:"bbbb"},
		],
	},
})
db.test.update({"name":"test2","vv.host":"u2"},{$set: {"vv.$.log":"cccc"}});
executor: ansible
  executor.sh -n jobsname -j jobslist -s [jobs argv] -p priority -i jobsid
ansible "gf-nginx-56-8:gf-nginx-46-140" -m shell -a "date"
*/

const (
	mgoJobs      = "jobs"
	mgoFlows     = "flow"
	mgoSchedule  = "schedule"
	mgoShortcut  = "shortcut"
	mgoRunlogs   = "runlogs"
	typeJobs     = 1
	typeJobsFlow = 2
)

//RouteList api路由表
/*
	type={"jobs","flows","schedule","jobpath"}
*/
func RouteList(route *common.APIRoute) {
	route.API["/jobs/jobs|POST"] = common.APIInfo{Handler: AddItem, Name: "添加作业", Perm: auth.APIPermAllow}
	route.API["/jobs/jobs|PUT"] = common.APIInfo{
		Handler: UpdateItem,
		Name:    "修改作业",
		Perm:    auth.APIPermAllow,
	}
	route.API["/jobs/jobs|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看作业",
		Perm:    auth.APIPermAllow | auth.APIPermJobRun,
	}
	route.API["/jobs/jobs|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除作业", Perm: auth.APIPermAllow}
	route.API["/jobs/flow|POST"] = common.APIInfo{Handler: AddItem, Name: "添加作业流", Perm: auth.APIPermAllow}
	route.API["/jobs/flow|PUT"] = common.APIInfo{Handler: UpdateItem, Name: "修改作业流", Perm: auth.APIPermAllow}
	route.API["/jobs/flow|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看作业流",
		Perm:    auth.APIPermAllow | auth.APIPermJobRun,
	}
	route.API["/jobs/flow|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除作业流", Perm: auth.APIPermAllow}
	route.API["/jobs/shortcut|POST"] = common.APIInfo{
		Handler: AddItem,
		Name:    "添加快捷方式",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/jobs/shortcut|PUT"] = common.APIInfo{
		Handler: UpdateItem,
		Name:    "修改快捷方式",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/jobs/shortcut|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看快捷方式",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/jobs/shortcut|DELETE"] = common.APIInfo{
		Handler: DeleteItem,
		Name:    "删除快捷方式",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	//route.API["/jobs/runlogs|GET"] = common.APIInfo{Handler: Items, Name: "查看作业日志", Perm: auth.APIPermAllow}
	route.API["/jobs/run|PUT"] = common.APIInfo{
		Handler: JobRun,
		Name:    "执行作业",
		Perm:    auth.APIPermAllow | auth.APIPermJobRun,
	}
	route.API["/jobs/runlogs|POST"] = common.APIInfo{
		Handler: JobLog,
		Name:    "提交作业日志",
		Perm:    auth.APIPermAllow,
	}
	route.API["/jobs/runlogs|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看作业日志",
		Perm:    auth.APIPermAllow,
	}
	route.API["/jobs/path|GET"] = common.APIInfo{Handler: jobsPath, Name: "作业路径", Perm: auth.APIPermAllow}
	route.API["/jobs/runshortcut|GET"] = common.APIInfo{
		Handler: JobRunShortcut,
		Name:    "快捷作业",
		Perm:    auth.APIPermAllow | auth.APIPermJobRun,
	}
}

//AddItem 添加作业内容
func AddItem(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	//filArgs := common.V{}
	var jobs interface{}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	jobtype := ctx.Req.URL.Path[i+1:]
	// jobtype := ctx.Params.ByName("type")
	// if jobtype == "" {
	// 	ctx.Response = 1204
	// 	return
	// }
	user := ctx.Keys["username"].(string)
	curtime := time.Now()
	id := bson.NewObjectId()
	switch jobtype {
	case mgoJobs:
		jobs = &MgoJobs{ID: id, Status: 1, Modifytime: curtime, Createtime: curtime}
		/*
			向Git提交脚本内容
		*/
		postArgs := common.V{}
		if body, code, _ := ctx.ReqBody(&postArgs); code == 0 {
			if err := json.Unmarshal(body, jobs); err != nil {
				econf.LogDebug("BodyFormatInvalid: %v", err)
				ctx.Response = code
				return
			}
		} else {
			ctx.Response = code
			return
		}
		if err := common.InputFilter(jobs, 1, nil, nil); err != nil {
			econf.LogInfo("invalid input: %v\n", err)
			ctx.Response = 1101
			return
		}
		jobinfo := jobs.(*MgoJobs)
		if q := postArgs.ObjType("scriptcode", "string"); q != nil {
			jobinfo.Operator = user
			path := jobinfo.Path + "/" + jobinfo.Name
			git := common.GitAction{
				Action:   "files",
				Method:   "POST",
				Url:      econf.GOPT.GitURL,
				GitToken: econf.GOPT.GitToken,
				User:     user,
				Projects: econf.GOPT.JobsRepo,
				Path:     path,
				Content:  q.(string),
			}
			if err := git.GitAPI(); err != nil {
				econf.LogCritical("POST to git falsed: %v\n", err)
				ctx.Response = 1601
				return
			}
		}
	case mgoFlows:
		jobs = &MgoFlows{
			ID:         id,
			Status:     0,
			Modifytime: time.Now(),
			Createtime: time.Now(),
		}
	case mgoSchedule:
		jobs = &Schedule{ID: id, Modifytime: curtime, Createtime: curtime}
	case mgoShortcut:
		jobs = &MgoShortcut{ID: id, Operator: user, Modifytime: curtime, Createtime: curtime}
	default:
		ctx.Response = 1204
		return
	}
	if jobtype != mgoJobs {
		if _, code, err := ctx.ReqBody(jobs); code != 0 {
			econf.LogDebug("[JOB] BodyFormatInvalid: %v", err)
			ctx.Response = code
			return
		}
		if err := common.InputFilter(jobs, 1, nil, nil); err != nil {
			econf.LogInfo("[JOB] invalid input: %v\n", err)
			ctx.Response = 1101
			return
		}
		if jobtype == mgoShortcut {
			taskattr := jobs.(*MgoShortcut).TaskAttr
			jobname := taskattr.Path + taskattr.Name
			if auth.PermChkJob(
				jobname,
				econf.Route.API[ctx.Keys["api"].(string)].Perm,
				ctx,
			) == false || jobs.(*MgoShortcut).Name[:7] == "apptpl-" {
				ctx.Response = 1107
				return
			}
		}
	}
	c := econf.Mgo.Coll(jobtype)
	if err := c.Insert(jobs); err != nil {
		econf.LogWarning("[JOB] insert %s failed: %v\n", jobtype, err)
		errstr := err.Error()
		ctx.Response = 1106
		if errstr[0:6] == "E11000" {
			//记录已存在
			ctx.Response = 1303
		}
	}
	p, _ := json.Marshal(jobs)
	ctx.OpLog.Write(p)
}

//UpdateItem 更新作业内容
func UpdateItem(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	filArgs := common.V{}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	jobtype := ctx.Req.URL.Path[i+1:]
	user := ctx.Keys["username"].(string)
	postArgs := common.V(ctx.BodyMap)
	skey := common.V{}
	skey["status"] = common.V{"$nin": []int{-1000, -2000}}
	var id string
	if id = postArgs.VString("_id"); id != "" && common.StrFilter(id, "id") {
		skey = common.V{"_id": bson.ObjectIdHex(id)}
		delete(postArgs, "_id")
	} else {
		ctx.Response = 1207
		return
	}
	// if _, ok := postArgs["createtime"]; ok {
	// 	delete(postArgs, "createtime")
	// }
	var jobs interface{}
	switch jobtype {
	case mgoJobs:
		jobs = &MgoJobs{}
		if err := common.InputFilter(jobs, 2, postArgs, filArgs); err != nil {
			econf.LogInfo("[JOB] invalid input: %v\n", err)
			ctx.Response = 1101
			return
		}
		filArgs["status"] = 0

		jobinfo := JobInfo(id, econf)
		if jobinfo == nil {
			econf.LogCritical("[JOB] not found job %s\n", postArgs["id"].(string))
			ctx.Response = 1301
			return
		}
		if q := postArgs.ObjType("scriptcode", "string"); q != nil {

			filArgs["status"] = 1
			filArgs["operator"] = user
			git := common.GitAction{
				Action:   "files",
				Method:   "PUT",
				Url:      econf.GOPT.GitURL,
				GitToken: econf.GOPT.GitToken,
				User:     user,
				Projects: econf.GOPT.JobsRepo,
				Path:     jobinfo.Path + "/" + jobinfo.Name,
				Content:  q.(string),
			}
			if err := git.GitAPI(); err != nil {
				econf.LogCritical("POST to git falsed: %v\n", err)
				ctx.Response = 1601
			}
			econf.MapCache.Del(git.Projects + git.Path)
			//return
		}

	case mgoFlows:
		jobs = &MgoFlows{}
		// if _,ok := postArgs["status"]; ok {
		// 	delete(postArgs,"status")
		// }
		if err := common.InputFilter(jobs, 2, postArgs, filArgs); err != nil {
			econf.LogInfo("[JOB] invalid input: %v\n", err)
			ctx.Response = 1101
			return
		}
	// case mgoSchedule:
	// 	jobs = &Schedule{}
	// case mgoShortcut:
	// 	jobs = &MgoShortcut{}
	case mgoShortcut:
		taskattr := postArgs["taskattr"].(map[string]interface{})
		jobname := taskattr["path"].(string) + taskattr["name"].(string)
		name := postArgs.VString("name")
		if auth.PermChkJob(
			jobname,
			econf.Route.API[ctx.Keys["api"].(string)].Perm,
			ctx,
		) == false || name[:7] == "apptpl-" {
			ctx.Response = 1107
			return
		}
		if !auth.IsAdmin(ctx) {
			skey["operator"] = user
		}
	default:
		ctx.Response = 1204
		return
	}
	//delete(filArgs, "createtime")
	filArgs["modifytime"] = time.Now()

	c := econf.Mgo.Coll(jobtype)
	if err := c.Update(skey, common.V{"$set": filArgs}); err != nil {
		econf.LogCritical("[JOB] Insert %s failed: %v\n", jobtype, err)
		errstr := err.Error()
		ctx.Response = 1106
		if errstr[0:6] == "E11000" {
			//记录已存在
			ctx.Response = 1301
		}
		return
	}
	ctx.OpLog.Write(skey.Json("skey"))
	ctx.OpLog.Write(filArgs.Json("data"))
}

//DeleteItem 删除作业
func DeleteItem(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	skey := common.V{}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	jobtype := ctx.Req.URL.Path[i+1:]
	postArgs := common.V(ctx.BodyMap)
	var id string
	if id = postArgs.VString("id"); id != "" && common.StrFilter(id, "id") {
		skey["_id"] = bson.ObjectIdHex(id)
	} else {
		ctx.Response = 1204
		return
	}
	if jobtype == mgoJobs {
		jobinfo := JobInfo(id, econf)
		if jobinfo == nil {
			econf.LogCritical("[JOB] not found job %s\n", id)
			ctx.Response = 1301
			return
		}
		//filArgs["name"] = jobinfo.Name
		//filArgs["path"] = jobinfo.Path
		git := common.GitAction{
			Action:   "files",
			Method:   "DELETE",
			Url:      econf.GOPT.GitURL,
			GitToken: econf.GOPT.GitToken,
			User:     ctx.Keys["username"].(string),
			Projects: econf.GOPT.JobsRepo,
			Path:     jobinfo.Path + "/" + jobinfo.Name,
		}
		if err := git.GitAPI(); err != nil {
			econf.LogCritical("[JOB] DELETE to git falsed: %v\n", err)
			// ctx.Response = 1601
			// return
		}
		econf.MapCache.Del(git.Projects + git.Path)
	} else if jobtype == mgoShortcut {
		user := ctx.Keys["username"].(string)
		if !auth.IsAdmin(ctx) {
			skey["operator"] = user
		}
	}
	c := econf.Mgo.Coll(jobtype)
	if err := c.Remove(skey); err != nil {
		econf.LogCritical("[JOB] Delete %s failed(%s): %v\n", jobtype, id, err)
		ctx.Response = 1301
		return
	}
	ctx.OpLog.Write(skey.Json("skey"))
}

//Items 获取作业列表
func Items(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	skey := common.V{}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	jobtype := ctx.Req.URL.Path[i+1:]
	// jobtype := ctx.Params.ByName("type")
	// if jobtype == "" {
	// 	ctx.Response = 1204
	// 	return
	// }

	pageSize := 2000
	pageNo := 1
	fields := common.V{}
	if jobtype == mgoRunlogs {
		fields["hostlogs"] = 0
	}
	releasejob := false
	for k, v := range ctx.Req.Form {
		switch k {
		case "pageNo":
			pageNo, _ = strconv.Atoi(v[0])
		case "pageSize":
			pageSize, _ = strconv.Atoi(v[0])
		case "id":
			if common.StrFilter(v[0], "id") {
				skey["_id"] = bson.ObjectIdHex(v[0])
			} else {
				ctx.Response = 1204
				return
			}
		case "logdetail":
			fields = common.V{}
		case "releasejob":
			if v[0] == "yes" {
				releasejob = true
			}
		default:
			if common.StrFilter(v[0], "safestr") {
				skey[k] = v[0]
			}
		}
	}
	var result []common.V
	recnum, _ := econf.Mgo.QueryM(jobtype, skey, &result, pageSize, pageNo, fields)
	rmsg.TotalRecord = recnum
	//econf.LogDebug("[JOB] get job: %v\n", result)
	if recnum > 0 {
		rmsg.DataList = result
		if recnum == 1 && jobtype == mgoJobs {
			path := result[0]["path"].(string)
			name := result[0]["name"].(string)

			if auth.PermChkJob(
				path+name,
				econf.Route.API[ctx.Keys["api"].(string)].Perm,
				ctx,
			) == false {
				ctx.Response = 1107
				return
			}

			curjobver := ctx.Req.Header.Get("X-Job-CommitID")
			if curjobver != "" && curjobver == result[0]["commitid"].(string) {
				ctx.Writer.Header().Set("X-Job-Version", "newest")
				rmsg.Message = "newest"
				rmsg.DataList = nil
				ctx.Response = rmsg
				return
			}
			ref := ""
			if releasejob {
				ref = result[0]["commitid"].(string)
			}
			if gitContent, err := econf.GetGitFile(path+"/"+name, econf.GOPT.JobsRepo, ref); err == nil {
				result[0]["content"] = gitContent.Content
				result[0]["commit_id"] = gitContent.LastCommitID
				// if result[0]["status"].(int) == 1 {
				// 	result[0]["commit_id"] = gitContent.LastCommitID
				// }
			} else {
				//base64_encode("access git server failed.")
				result[0]["content"] = "YWNjZXNzIGdpdCBzZXJ2ZXIgZmFpbGVkLg=="
				econf.LogCritical("script content[%s/%s]: %v", path, name, err)
			}
		}
	} else {
		rmsg.SetCode(1301, econf.Err)
	}

	ctx.Response = rmsg
	ctx.OpLog.Write(skey.Json("skey"))
}

//jobsPath 获取作业列表
func jobsPath(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	//filArgs := common.V{}
	results := []MgoJobs{}
	pathTree := []common.V{}
	var pathOne []string
	jobpath := make(map[string]*jobPathTree)
	recnum, _ := econf.Mgo.Query(mgoJobs, common.V{}, &results, 20000, 1, common.V{})
	if recnum == 0 {
		goto JUMP
	}
	//pathlist := econf.BaseData("", "jobpath")
	for _, result := range results {
		pathinfo := strings.Split(result.Path, "/")
		path1 := pathinfo[0]
		if _, ok := jobpath[path1]; !ok {
			jobpath[path1] = &jobPathTree{}
			pathOne = append(pathOne, path1)
		}
		if len(pathinfo) == 1 {
			jobpath[path1].files = append(jobpath[path1].files, result.Name+"|"+result.ID.Hex())
		} else if len(pathinfo) == 2 {
			path2 := pathinfo[1]
			if _, ok := jobpath[path1].nodes[path2]; !ok {
				jobpath[path1].nodes = make(map[string][]string)
				jobpath[path1].nodes[path2] = []string{}
				jobpath[path1].path2 = append(jobpath[path1].path2, path2)
			}
			jobpath[path1].nodes[path2] = append(jobpath[path1].nodes[path2], result.Name+"|"+result.ID.Hex())
		}
	}
	sort.Strings(pathOne)
	for _, path1 := range pathOne {
		pathSec := []common.V{}
		sort.Strings(jobpath[path1].path2)
		for _, path2 := range jobpath[path1].path2 {
			sort.Strings(jobpath[path1].nodes[path2])
			files := []common.V{}
			for _, file := range jobpath[path1].nodes[path2] {
				fileinfo := strings.Split(file, "|")
				files = append(files, common.V{
					"text":  fileinfo[0],
					"id":    fileinfo[1],
					"type":  "file",
					"path1": path1,
					"path2": path2,
				})
			}
			files = append(files, common.V{
				"text":  "添加脚本",
				"type":  "addfile",
				"path1": path1,
				"path2": path2,
			})
			pathSec = append(pathSec, common.V{
				"text":  path2,
				"type":  "path2",
				"path1": path1,
				"nodes": files,
			})
		}
		sort.Strings(jobpath[path1].files)
		for _, file := range jobpath[path1].files {
			fileinfo := strings.Split(file, "|")
			pathSec = append(pathSec, common.V{
				"text":  fileinfo[0],
				"id":    fileinfo[1],
				"type":  "file",
				"path1": path1,
			})
		}
		pathSec = append(pathSec, common.V{
			"text":  "添加脚本",
			"type":  "addfile",
			"path1": path1,
		})
		pathTree = append(pathTree, common.V{
			"text":  path1,
			"type":  "path1",
			"nodes": pathSec,
		})
	}
JUMP:
	pathTree = append(pathTree, common.V{
		"text": "添加脚本",
		"type": "addfile",
	})
	rmsg.TotalRecord = len(pathOne) + 1
	rmsg.DataList = pathTree
	ctx.Response = rmsg
}

//JobExec 作业执行
func (job *MgoJobLogs) JobExec(ctx *httpsvr.Context) (err error) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	user := ctx.Keys["username"].(string)
	job.Operator = user
	userinfo, _ := auth.GetUserInfo(econf, user, 3)
	if job.CDEnv == "" {
		job.CDEnv = userinfo.WorkEnv
	}

	hostlogs := []*HostLogs{}
	hostlist := []string{}
	job.ID = bson.NewObjectId()

	// oid := bson.NewObjectId()
	// id = oid.Hex()
	// if job.TaskAttr.CommitID == "" {
	// 	err = fmt.Errorf("job %s/%s is not release.", job.TaskAttr.Path, job.TaskAttr.Name)
	// 	return
	// }
	//err = nil
	// if job.JobRunPermissions(econf) == false {
	// 	err = errors.New("No permissions")
	// 	return
	// }
	if job.Target == "hosts" {
		hostlist = job.List
	} else {
		hostlist = job.getHostList(econf)
	}
	if len(hostlist) == 0 {
		err = errors.New("no hosts")
		return
	}
	for _, host := range hostlist {
		hostlogs = append(hostlogs, &HostLogs{HostName: host})
	}
	if job.Env == nil {
		job.Env = make(map[string]string)
	}
	job.Progress = 0
	job.HostNum = len(hostlist)
	job.Failed = job.HostNum
	job.HostLogs = hostlogs
	job.BeginTime = time.Now()
	if job.Type == "JobsFlow" {
		flowInfo := FlowTask(job.TaskName, econf)
		if !flowInfo.ID.Valid() {
			err = fmt.Errorf("The task %s is invalid.", job.TaskName)
			return
		}
		job.TaskAttr = flowInfo.Task
		job.Env["ZDOPS_Task"] = flowInfo.Task.base64Encode()
		job.Env["ZDOPS_TaskName"] = flowInfo.Name
	} else {
		job.Env["ZDOPS_Task"] = job.TaskAttr.base64Encode()
		job.Env["ZDOPS_TaskName"] = job.TaskAttr.Path + "/" + job.TaskAttr.Name
	}
	job.Env["ZDOPS_ExecDebug"] = "yes"
	job.Env["ZDOPS_ExecFront"] = ""
	job.Env["ZDOPS_NewestJob"] = ""
	job.Env["ZDOPS_GitToken"] = econf.GOPT.GitToken
	// tmp, _ := json.Marshal(job.TaskAttr)
	// econf.LogDebug("[Job] TaskAttr: %s\n", string(tmp))
	if job.TaskAttr.Latest {
		job.Env["ZDOPS_NewestJob"] = "-n"
	}
	job.Env["ZDOPS_TaskType"] = job.Type
	playbook := econf.GOPT.WorkDir + "/jobs/jobs.yml"
	if _, ok := job.Env["ZDOPS_CallBack"]; !ok {
		job.Env["ZDOPS_CallBack"] = econf.GOPT.BaseURL + econf.GOPT.Prefix + "/jobs/runlogs"
	}
	job.Env["ZDOPS_SavePath"] = econf.GOPT.JobsSavePath
	job.Env["ZDOPS_TaskLogID"] = job.ID.Hex()
	job.Env["ZDOPS_KeyID"] = userinfo.KeyID
	job.Env["ZDOPS_Secret"] = userinfo.Secret
	job.Env["ZDOPS_JobAPI"] = econf.GOPT.BaseURL + econf.GOPT.Prefix + "/jobs/jobs"
	job.Env["WorkDir"] = econf.GOPT.WorkDir
	job.Env["ExecutorPath"] = econf.GOPT.ExecutorPath
	job.Env["ansible_user"] = econf.GOPT.SSHUser
	c := econf.Mgo.Coll(mgoRunlogs)
	if err = c.Insert(job); err != nil {
		tmp, _ := json.Marshal(job)
		econf.LogDebug("[Job] create joblogs failed: %s\n", string(tmp))
		//econf.LogDebug("[Job] create joblogs content:(%v) failed: %v\n", job, err)
		return
	}
	go func() {
		for _, host := range job.List {
			if job.Target != "hosts" {
				host = job.CDEnv + "-" + host
			}
			je := JobExecutor{
				Host:      host,
				PlayBook:  playbook,
				ExtraVars: job.Env,
			}
			JobsQueue <- &je
			//econf.LogDebug("job host %s", je.Host)
		}
	}()
	return
}

//JobRun 作业执行接口
func JobRun(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	job := &MgoJobLogs{}
	if err := json.Unmarshal(ctx.Body, &job); err != nil {
		econf.LogDebug("[JOB] BodyFormatInvalid: %v", err)
		ctx.Response = 1206
		return
	}
	if err := common.InputFilter(job, 1, nil, nil); err != nil {
		econf.LogInfo("invalid input: %v\n", err)
		ctx.Response = 1101
		return
	}
	if auth.PermChkJob(
		job.TaskAttr.Path+job.TaskAttr.Name,
		econf.Route.API[ctx.Keys["api"].(string)].Perm,
		ctx,
	) == false || auth.PermChkJobTarget(job.Target, job.List, ctx) == false {
		ctx.Response = 1107
		return
	}

	if err := job.JobExec(ctx); err != nil {
		// tmp, _ := json.Marshal(job)
		// econf.LogDebug("[Job] joblogs: %s\n", string(tmp))
		econf.LogWarning("[Job] run %s failed: %s", job.TaskName, err)
		ctx.Response = 1701
		return
	}
	p, _ := json.Marshal(job)
	ctx.OpLog.Write(p)
	return
}

//JobRunShortcut 作业执行接口
func JobRunShortcut(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	//rmsg := common.NewRespMsg()
	id := ctx.Req.Form.Get("id")
	if !common.StrFilter(id, "id") {
		ctx.Response = 1207
		return
	}
	rShortcut := &MgoShortcut{}
	econf.Mgo.Query(mgoShortcut, bson.ObjectIdHex(id), rShortcut, 0, 0, nil)
	if !rShortcut.ID.Valid() {
		ctx.Response = 1301
		return
	}
	if auth.PermChkJob(
		rShortcut.TaskAttr.Path+rShortcut.TaskAttr.Name,
		econf.Route.API[ctx.Keys["api"].(string)].Perm,
		ctx,
	) == false || auth.PermChkJobTarget(rShortcut.Target, rShortcut.List, ctx) == false {
		ctx.Response = 1107
		return
	}
	user := ctx.Keys["username"].(string)
	userinfo, _ := auth.GetUserInfo(econf, user, 1)
	job := &MgoJobLogs{
		TaskAttr: rShortcut.TaskAttr,
		Target:   rShortcut.Target,
		List:     rShortcut.List,
		CDEnv:    userinfo.WorkEnv,
		Operator: user,
	}
	econf.LogDebug("[Job] Shortcut info: %v\n", job)
	// if len(job.HostList) == 0 {
	// 	job.getHostList(econf)
	// 	//jobArgs.HostList = getHostList(econf, jobArgs.Projects, jobArgs.Clusters, jobArgs.Apps, jobArgs.CDEnv)
	// }
	if err := job.JobExec(ctx); err != nil {
		econf.LogWarning("[Job] %s run '%s/%s %s' failed: %s", user, job.TaskAttr.Path, job.TaskAttr.Name, job.TaskAttr.Argv, err)
		ctx.Response = 1701
		return
	}
	p, _ := json.Marshal(job)
	ctx.OpLog.Write(p)
	return
}

//JobLog 作业执行接口
func JobLog(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	cbr := &CBResult{}
	if body, code, err := ctx.ReqBody(cbr); code != 0 {
		econf.LogDebug("[Job] BodyFormatInvalid: [%s] with %v", string(body), err)
		ctx.Response = code
		return
	}
	user := ctx.Keys["username"].(string)
	if err := cbr.Logger(user, econf); err != nil {
		ctx.Response = 1302
	}
	return
}

//Logger 记录作业返回的执行结果
func (cbr CBResult) Logger(user string, econf *common.EopsConf) error {
	skey := common.V{
		"_id":               bson.ObjectIdHex(cbr.ID),
		"hostlogs.hostname": cbr.Result.HostName,
		"operator":          user,
	}
	//jsonData, _ := json.Marshal(skey)
	//econf.LogInfo("skey: %s\n", string(jsonData))
	inc := 0
	if cbr.Status == 0 || cbr.Status == 900 {
		inc = -1
	}
	updata := common.V{
		// "$push": common.V{
		// 	"hostlogs.$.jobs": common.V{"$each": cbr.Result.Jobs},
		// },
		"$set": common.V{"endtime": time.Now(), "hostlogs.$.jobs": cbr.Result.Jobs},
		"$bit": common.V{"status": common.V{"or": cbr.Status}},
		"$inc": common.V{"failed": inc, "progress": 1},
	}

	// if cbr.Result.Jobs[0].Env["ZDOPS_JobsType"] == "jobs" {
	// updata["$set"] = common.V{"endtime": time.Now(), "hostlogs.$.jobs": cbr.Result.Jobs}
	// } else if cbr.Result.Jobs[0].Env["ZDOPS_JobsType"] == "jobflow" {
	// 	updata["$push"] = common.V{"hostlogs.$.jobs": common.V{"$each": cbr.Result.Jobs}}
	// 	updata["$set"] = common.V{"endtime": time.Now()}
	// }
	c := econf.Mgo.Coll(mgoRunlogs)
	if err := c.Update(skey, updata); err != nil {
		econf.LogWarning("[Job] job %s log failed: %v\n", cbr.ID, err)
		return err
	}
	return nil
}

/*DeployFlow 创建应用部署的作业流。根据apptpl的内容自动创建。不允许通过作业流管理进行修改
 */
func (dflow *DeployFlows) DeployFlow(econf *common.EopsConf) error {
	var err error
	var jobname, jobargv string
	jobname = dflow.RollBack
	if i := strings.IndexByte(dflow.RollBack, ' '); i > 0 {
		jobname = dflow.RollBack[:i]
		jobargv = dflow.RollBack[i+1:]
	}
	jobInfo := GetJobByName(econf, jobname, "")
	rescue := &TaskAttr{}
	jobInfo.toTask(rescue)
	rescue.Alias = "RollBack"
	rescue.Argv = jobargv
	jobargv = ""
	taskInfo := &TaskAttr{}
	task := taskInfo
	for _, job := range dflow.Jobs {
		if task.Name != "" {
			task.Next = &TaskAttr{}
			task = task.Next
		}
		jobname = job.Job
		if i := strings.IndexByte(job.Job, ' '); i > 0 {
			jobname = job.Job[:i]
			jobargv = job.Job[i+1:]
		}
		jobInfo = GetJobByName(econf, jobname, "")
		if !jobInfo.ID.Valid() {
			err = errors.New("no found job: " + jobname)
			return err
		}
		jobInfo.toTask(task)
		task.Alias = job.Alias
		task.Argv = jobargv
		task.Rescue = rescue
	}
	c := econf.Mgo.Coll(mgoFlows)
	if chkflow := FlowInfo(dflow.Name, econf); chkflow.ID.Valid() {
		data := common.V{
			"$set": common.V{
				"task":       taskInfo,
				"modifytime": dflow.Modifytime,
			},
		}
		if err = c.Update(common.V{"name": dflow.Name}, data); err != nil {
			return err
		}
	} else {
		curtime := time.Now()
		flowInfo := &MgoFlows{
			Name:       dflow.Name,
			Describe:   dflow.Describe,
			Status:     -2000,
			Task:       taskInfo,
			Modifytime: dflow.Modifytime,
			Createtime: curtime,
		}
		if err = c.Insert(flowInfo); err != nil {
			return err
		}
	}

	return nil
}

func (job *MgoJobs) toTask(task *TaskAttr) {
	// task.JobID = job.ID.Hex()
	task.Name = job.Name
	task.Path = job.Path
	task.Timeout = job.Timeout
	task.Priority = job.Priority
	task.User = job.User
}

//GetLogger 获取日志
func GetLogger(id string, econf *common.EopsConf) *MgoJobLogs {
	rRunLog := &MgoJobLogs{}
	econf.Mgo.Query(mgoRunlogs, bson.ObjectIdHex(id), rRunLog, 0, 0, nil)
	return rRunLog
}

//JobInfo 获取作业信息
func JobInfo(id string, econf *common.EopsConf) *MgoJobs {
	rJob := &MgoJobs{}
	econf.Mgo.Query(mgoJobs, bson.ObjectIdHex(id), rJob, 0, 0, nil)
	return rJob
}

//FlowInfo 获取作业信息
func FlowInfo(key interface{}, econf *common.EopsConf) *MgoFlows {
	rFlow := &MgoFlows{}
	econf.Mgo.Query(mgoFlows, key, rFlow, 0, 0, nil)
	return rFlow
}

//GetJobByName 获取Job的版本
func GetJobByName(econf *common.EopsConf, name, path string) *MgoJobs {
	// econf.LogDebug("JOB: %s\n", name)
	if path == "" {
		if i := strings.LastIndexByte(name, '/'); i >= 0 {
			path = name[:i]
			tmp := name[i+1:]
			name = tmp
		}
	}
	rJob := &MgoJobs{}
	econf.Mgo.Query(mgoJobs, common.V{"name": name, "path": path}, rJob, 0, 0, nil)
	return rJob
}

//FlowTask 获取作业流的内容
func FlowTask(name string, econf *common.EopsConf) *MgoFlows {
	flowInfo := &MgoFlows{}
	if n, _ := econf.Mgo.Query(mgoFlows, name, flowInfo, 0, 0, nil); n == 0 {
		return &MgoFlows{}
	}
	task := flowInfo.Task
	for {
		if err := task.sync(econf); err != nil {
			tmp, _ := json.Marshal(task)
			econf.LogDebug("[Job] Flowtask:%s not found job %s\n", string(tmp))
			econf.LogWarning("[Job] Flowtask:%s not found job %s/%s, err: %v\n", flowInfo.Name, task.Path, task.Name, err)
			return &MgoFlows{}
		}
		if task.Rescue != nil {
			if err := task.Rescue.sync(econf); err != nil {
				econf.LogWarning("[Job] Flowtask:%s not found job %s/%s, err: %v\n", flowInfo.Name, task.Path, task.Name, err)
				return &MgoFlows{}
			}
		}
		if task.Next != nil {
			task = task.Next
		} else {
			break
		}
	}
	return flowInfo
}

func (task *TaskAttr) sync(econf *common.EopsConf) error {
	jobInfo := &MgoJobs{}
	skey := common.V{
		"name": task.Name,
		"path": task.Path,
	}
	if n, err := econf.Mgo.Query(mgoJobs, skey, jobInfo, 0, 0, nil); n == 0 {
		//econf.LogWarning("[Job] task sync: not found job %s\n", task.JobID.Hex())
		return err
	}
	task.Name = jobInfo.Name
	task.Path = jobInfo.Path
	task.CommitID = jobInfo.CommitID
	return nil
}

func (task *TaskAttr) base64Encode() string {
	tmp, err := json.Marshal(task)
	if err != nil {
		return ""
	}
	b64 := base64.StdEncoding
	b64buf := make([]byte, b64.EncodedLen(len(tmp)))
	b64.Encode(b64buf, tmp)
	return string(b64buf)
}

func (job *MgoJobLogs) getHostList(econf *common.EopsConf) []string {
	//return cmdb.AssignedHosts(econf, job.Projects, job.Clusters, job.Apps, job.CDEnv)
	return cmdb.AssignedHosts(econf, job.Target, job.List, job.CDEnv)
}

// func jobRun(econf *common.EopsConf, jrArgs *JobRunArgs) error {
//
// 	return nil
// }
//
// func JobLog() {
//
// }
