package deploy

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/cmdb"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"
	"opsv2.cn/opsv2/api/jobs"
	"opsv2.cn/opsv2/api/options"
)

const (
	//mgoDeployHistory 部署记录表
	mgoDeployHistory = "deployhistory"
	logKind          = "Deploy"
)

//RouteList api路由表
/*

 */
func RouteList(route *common.APIRoute) {
	route.API["/deploy/webhook|POST"] = common.APIInfo{Handler: WebHook, Name: "WebHook for Gitlab", Perm: auth.APIPermNoCheck}
	route.API["/deploy/do|GET"] = common.APIInfo{
		Handler: DoDeploy,
		Name:    "Run deploy on cdenv",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv | auth.APIPermPCADeploy,
	}
	route.API["/deploy/upload|POST"] = common.APIInfo{Handler: UpLoadFile, Name: "Upload File", Perm: auth.APIPermAllow}
	route.API["/deploy/callback|POST"] = common.APIInfo{Handler: CallBack, Name: "部署结束后的回调接口", Perm: auth.APIPermAllow}
	route.API["/deploy/history|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "获取部署记录",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv | auth.APIPermPCADeploy,
	}
	route.API["/deploy/testresult|PUT"] = common.APIInfo{
		Handler: TestResult,
		Name:    "更新部署的测试结果",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv | auth.APIPermPCADeploy,
	}
}

//WebHook 添加作业内容
/*
处理逻辑：
	# 校验token是否合法
	# 获取事件类型。目前只处理：Tag Push Hook和Push Hook
	# 获取应用信息和版本信息
	# 生成部署信息并保存
	# 判断是否有构建脚本，如果有执行构建
	# 由于gitlab并不能处理错误。因此，此接口的处理结果以信息形式发送给提交者
*/
func WebHook(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	reqdata := GitHookReq{}
	token := ctx.Req.Header.Get("X-Gitlab-Token")
	if token != econf.GOPT.HookToken {
		econf.LogWarning("[%s] Invalid HookToken\n", logKind)
		return
	}
	if _, code, err := ctx.ReqBody(&reqdata); code != 0 {
		econf.LogWarning("[%s] Invalid request to webhook: %v\n", logKind, err)
		return
	}
	//获取版本信息
	var ref string
	event := ctx.Req.Header.Get("X-Gitlab-Event")
	if event == "Tag Push Hook" {
		//refs/tags/v1.0.0 过滤掉refs/tags/
		ref = reqdata.Ref[10:]
	} else if event == "Push Hook" {
		if reqdata.Ref != "refs/heads/master" {
			econf.LogDebug("[%s] Only allow the master branch in WebHook\n", logKind)
			return
		}
		ref = reqdata.CheckoutSha
	} else {
		return
	}
	if strings.Index(reqdata.Commits[reqdata.TotalCommitsCount-1].Message, "opsv2 webhook") == -1 {
		return
	}
	userinfo, _ := auth.GetUserInfo(econf, reqdata.UserEmail, 3)
	if userinfo == nil {
		econf.LogWarning("[%s] %s PermissionDenied in WebHook\n", logKind, reqdata.UserEmail)
		return
	}
	ctx.Keys["username"] = userinfo.Name
	ctx.Keys["webhook"] = true
	//根据仓库信息，区别部署类型
	//var deployType string
	rdh := &MgoHistorys{Status: -1}
	pinfo := strings.Split(reqdata.Project.Name, "-")
	// var params map[string]string
	apptpl := &cmdb.MgoAppTPL{}
	econf.LogDebug("[Deploy] namespace: %s", reqdata.Project.NameSpace)

	if reqdata.Project.NameSpace == common.ApptplName {
		//deployType = reqdata.Project.NameSpace
		apptpl = cmdb.AppTplInfo(pinfo[0], econf)
	} else if reqdata.Project.NameSpace[:5] == "cmdb-" && len(pinfo) >= 2 {
		//deployType = "app"
		if len(pinfo) == 3 && pinfo[2] == "conf" {
			rdh.AppConfVer = ref
		} else {
			rdh.AppVer = ref
		}
		if reqdata.TotalCommitsCount > 0 && len(reqdata.Commits[reqdata.TotalCommitsCount-1].Message) > 8 {
			if strings.Index(reqdata.Commits[reqdata.TotalCommitsCount-1].Message, "Nodeploy") >= 0 {
				rdh.Status = -2
			}
		}
		// project := reqdata.Project.NameSpace[5:]
		// params = map[string]string{
		// 	"project": project,
		// 	"cluster": pinfo[0],
		// 	"app":     pinfo[1],
		// }
	} else {
		//从cmdb中查找GitHTTP路径是否存在于apptpl和APP中，包括配置文件的地址
		q := common.V{
			"$or": []common.V{
				{"source": reqdata.Project.GitHTTP},
				{"config": reqdata.Project.GitHTTP},
			},
		}
		appinfo := cmdb.AppInfo(q, econf)
		if appinfo.ID.Valid() {
			if appinfo.Config == reqdata.Project.GitHTTP {
				rdh.AppConfVer = ref
			} else {
				rdh.AppVer = ref
			}
			if reqdata.TotalCommitsCount > 0 && len(reqdata.Commits[reqdata.TotalCommitsCount-1].Message) > 8 {
				if strings.Index(reqdata.Commits[reqdata.TotalCommitsCount-1].Message, "Nodeploy") >= 0 {
					rdh.Status = -2
				}
			}
			// p, c, _ := cmdb.AppParent(appinfo.ID.Hex(), econf)
			// params = map[string]string{
			// 	"project": p.Name,
			// 	"cluster": c.Name,
			// 	"app":     appinfo.Name,
			// }
		} else {
			apptpl = cmdb.AppTplInfo(q, econf)
		}
	}
	var appinfo *cmdb.MgoApp
	appVar := make(map[string]string)
	taskattr := jobs.TaskAttr{}
	jobRun := jobs.MgoJobLogs{
		ID: bson.NewObjectId(),
		// TaskName: taskattr.Name,
		TaskAttr: &taskattr,
		Target:   "hosts",
		List:     []string{econf.GOPT.CDEnvBuild},
		Env:      appVar,
		Operator: userinfo.Name,
		CDEnv:    "build",
		Type:     "jobs",
	}
	appVar["ZDOPS_AppSource"] = econf.GOPT.AppSource
	if apptpl.ID.Valid() {
		if (len(pinfo) == 2 && pinfo[1] == "conf") || (apptpl.Config == reqdata.Project.GitHTTP) {
			syncAppTplConfVer(apptpl.ID.Hex(), ref, ctx)
		} else {
			appVar["ZDOPS_SourceType"] = apptpl.SourceType
			appVar["ZDOPS_Source"] = apptpl.Source
			// appVar["ZDOPS_CallBack"] = econf.GOPT.BaseURL + econf.GOPT.Prefix + "/deploy/callback"
			//执行器去/tmp/ZDOPS_Build-ZDOPS_Jobid读取文件，文件内容为打包的路径
			appVar["ZDOPS_Build"] = common.TrueStr
			appVar["ZDOPS_DeployType"] = common.ApptplName
			appVar["ZDOPS_AppTplName"] = apptpl.Name
			appVar["ZDOPS_AppInstallPath"] = econf.GOPT.AppInstallPath
			appVar["ZDOPS_AppTplVer"] = ref
			buildjob := apptpl.Build
			if isArgs := strings.IndexByte(apptpl.Build, ' '); isArgs >= 0 {
				buildjob = apptpl.Build[:isArgs]
				taskattr.Argv = apptpl.Build[isArgs+1:]
			}
			econf.LogDebug("[Deploy] buildjob: %s", buildjob)
			if jobtmp := jobs.GetJobByName(econf, buildjob, ""); jobtmp.ID.Valid() {
				taskattr.CommitID = jobtmp.CommitID
				taskattr.Timeout = jobtmp.Timeout
				taskattr.User = jobtmp.User
				taskattr.Priority = jobtmp.Priority
				taskattr.Name = jobtmp.Name
				taskattr.Path = jobtmp.Path
			} else {
				econf.LogDebug("[Deploy] buildjob %s is not exist\n", apptpl.Build)
				return
			}
			// appinfo = cmdb.AppInfo(common.V{"from": apptpl.ID.Hex()}, econf)
		}
	} else {

	}

	econf.LogDebug("[Deploy] job: %s,target: %v", taskattr.Name, appinfo.ID.Hex())
	err := jobRun.JobExec(ctx)
	if err != nil {
		econf.LogCritical("[Deploy] job exec failed with %s: %v\n", jobRun.ID.Hex(), err)
	}

}

//Items 获取应用的部署任务
func Items(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	var appid string
	history := false
	pageNo := 1
	pageSize := 1000
	for k, v := range ctx.Req.Form {
		value := v[0]
		switch k {
		case "appid":
			if !common.StrFilter(value, "id") {
				ctx.Response = 1207
				return
			}
			appid = value
		case "pageNo":
			pageNo, _ = strconv.Atoi(value)
		case "pageSize":
			pageSize, _ = strconv.Atoi(value)
		case "history":
			history = true
		}
	}
	result := []common.V{}
	skey := common.V{
		"appid": appid,
	}
	if !history {
		skey["status"] = common.V{"$nin": []int{0, 2, -2, 15, 25, 35, 45, 55}}
	}
	orderby := common.V{"orderby": "-createtime"}
	rn, _ := econf.Mgo.Query(mgoDeployHistory, skey, &result, pageSize, pageNo, orderby)
	if rn > 0 {
		if !history {
			result[0]["queue"] = rn - 1
			rmsg.DataList = result[0]
			rmsg.TotalRecord = 1
		} else {
			rmsg.DataList = result
			rmsg.TotalRecord = rn
		}
	}
	ctx.Response = rmsg
	ctx.OpLog.Write(skey.Json("skey"))
	return
}

//TestResult 更新测试结果
func TestResult(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	data := common.V{}
	skey := common.V{}
	postArgs := common.V(ctx.BodyMap)
	ctx.Response = 1207
	if deployid := postArgs.VString("deployid"); deployid != "" {
		testResult := postArgs.VBool("tr")
		rdh := MgoHistorys{}
		rn, _ := econf.Mgo.Query(
			mgoDeployHistory,
			bson.ObjectIdHex(deployid),
			&rdh, 0, 0, common.V{"status": 1},
		)
		if rn == 1 {
			if auth.PermChkCDEnv(econf.GOPT.Environments[rdh.Status/10-1], ctx) {
				if rdh.Status%10 == 0 && rdh.Status > 0 {
					data["status"] = rdh.Status + 3
					if testResult {
						data["status"] = rdh.Status + 2
					}
					skey["_id"] = rdh.ID
					c := econf.Mgo.Coll(mgoDeployHistory)
					if err := c.Update(skey, common.V{"$set": data}); err == nil {
						ctx.Response = 0
						ctx.OpLog.Write(common.V{"skey": skey, "data": data}.Json(""))
					} else {
						econf.LogWarning("[%s] deployid %s update testresult failed: %v\n", logKind, deployid, err)
					}
				}
			}
		}
	}
	return
}

//DoDeploy 部署
func DoDeploy(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	deployid := ctx.Req.Form.Get("deployid")
	if !common.StrFilter(deployid, "id") {
		ctx.Response = 1207
		return
	}
	rdh := &MgoHistorys{}
	_, err := econf.Mgo.Query(mgoDeployHistory, bson.ObjectIdHex(deployid), rdh, 0, 0, nil)
	if err != nil {
		econf.LogWarning("[%s] not found history with %s: %v\n", logKind, deployid, err)
		ctx.Response = 1301
		return
	}

	result := rdh.deploy(nil, ctx)
	ctx.Response = result
	return
}

//UpLoadFile 部署
/*
Method: POST
Params: （*代表此参数必须提供）
	p: project name
	c: cluster name
	a: app name
	ver: app版本号。参数为空时，会根据日期自动生成一个版本
	apptpl: 模板名称
	md5: 文件md5值。*
Params describe：
	当apptpl为空时，p,c,a三个参数都必须提供。优先处理apptpl
逻辑：
	处理上传文件，并校验md5值
	判断是否提供版本号：
	  是：不做处理，但是要求上传的文件名遵循：name-ver.[zip|tgz]格式
		否：根据日期生成一个版本号，并将上传的文件改名为：name-ver.[zip|tgz]格式
	判断是否是模板：
		是：检查所有依赖模板的应用版本和模板的版本是否一致，如果有不一致的，放弃更新。都一致，则更新
			模板的版本，为所有依赖的应用创建部署任务
		否：代表是应用更新，为应用创建部署任务。
*/
func UpLoadFile(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	var project, cluster, app, appver, apptpl, savedir, srcmd5 string
	//econf.LogDebug("[%s] form: %v\n", logKind, req.Form)
	for k, v := range ctx.Req.Form {
		value := v[0]
		if !common.StrFilter(value, "safestr") {
			ctx.Response = 1207
			econf.LogDebug("[Deploy] %s's value %s is invalid", k, value)
			return
		}
		switch k {
		case "p":
			project = value
		case "c":
			cluster = value
		case "a":
			app = value
		case "ver":
			appver = value
		case common.ApptplName:
			apptpl = value
		case "md5":
			srcmd5 = value
		}
	}
	chunk := ctx.Req.Header.Get("Content-Range")
	econf.LogDebug("[Deploy] chunk: %s\n", chunk)
	ctx.Response = 0
	var appinfo *cmdb.MgoApp
	var apptplinfo *cmdb.MgoAppTPL
	if srcmd5 == "" {
		ctx.Response = 1207
		econf.LogDebug("[Deploy] no md5")
		return
	}
	if apptpl == "" {
		appinfo, _ = cmdb.AppNameInfo(project, cluster, app, econf)
		if appinfo.Source == "" || appinfo.SourceType == "" {
			ctx.Response = 1207
			econf.LogDebug("[Deploy] The %s/%s/%s is not have source\n", project, cluster, app)
			return
		}
		savedir = econf.GOPT.WorkDir + "/" + project + "/" + cluster + "/" + app
	} else {
		apptplinfo = cmdb.AppTplInfo(apptpl, econf)
		if !apptplinfo.ID.Valid() {
			econf.LogDebug("[Deploy] not found %s", apptpl)
			ctx.Response = 1207
			return
		}
		savedir = econf.GOPT.WorkDir + "/apptpl/" + apptpl
	}
	_, fname, err := ctx.Upload(savedir, `\.(zip|tgz)$`)
	if err == httpsvr.Chunking {
		return
	} else if err != nil {
		econf.LogWarning("[%s] upload failed: %v\n", logKind, err)
		ctx.Response = 1207
		return
	}
	fmd5, _ := common.FileMd5(savedir+"/"+fname, []byte{})
	if hex.EncodeToString(fmd5) != srcmd5 {
		ctx.Response = 1207
		econf.LogDebug("[Deploy] file md5 inconformity: %s", hex.EncodeToString(fmd5))
		return
	}
	fnamelen := len(fname)
	if appver == "" {
		//为了统一版本信息，此处重新分配，格式为日期。
		appver = time.Now().Format("200601021504")
		fnameExt := fname[fnamelen-4:]
		newfile := savedir + "/" + appver + fnameExt
		fname = savedir + "/" + fname
		err := os.Rename(fname, newfile)
		if err != nil {
			econf.LogWarning("[%s] Rename %s to %s failed: %v\n", logKind, fname, newfile, err)
			ctx.Response = 1901
			return
		}
	}

	if apptpl == "" {
		appid := appinfo.ID.Hex()
		rdh := &MgoHistorys{
			ID:         bson.NewObjectId(),
			AppID:      appid,
			AppVer:     appver,
			Status:     -1,
			CreateTime: time.Now(),
		}
		if rdh.checkStatus(econf) {
			if err := rdh.add(econf); err != nil {
				econf.LogWarning("[%s] add history failed: %v\n", logKind, err)
				ctx.Response = 1803
				return
			}
		} else {
			econf.LogWarning("[%s] Not allow create deploy\n", logKind)
			ctx.Response = 1803
			return
		}
	} else {
		//确保tpl和tplconf在所有依赖的应用中的版本都一致。否则无法更新
		syncAppTplVer(apptplinfo.ID.Hex(), appver, ctx)
	}
	ctx.OpLog.WriteString(fmt.Sprintf("%s ver=%s", savedir, appver))
	return
}

func syncAppTplVer(id, appver string, ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	if cmdb.AppTplChkVer(id, econf) {
		if code := cmdb.UpdateAppTpl(common.V{"_id": id, "appver": appver}, ctx); code != 0 {
			ctx.Response = code
			return
		}
		applist := cmdb.AppsList(econf, common.V{"from": id}, 1000, 1, nil)
		for _, app := range *applist {
			rdh := &MgoHistorys{
				ID:         bson.NewObjectId(),
				AppID:      app.ID.Hex(),
				ApptplVer:  appver,
				Status:     -1,
				CreateTime: time.Now(),
			}
			if rdh.checkStatus(econf) {
				if err := rdh.add(econf); err != nil {
					econf.LogWarning("[%s] add history failed: %v\n", logKind, err)
					ctx.Response = 1803
					return
				}
			}
		}
	}
}

func syncAppTplConfVer(id, confver string, ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	if cmdb.AppTplConfChkVer(id, econf) {
		if code := cmdb.UpdateAppTpl(common.V{"_id": id, "configver": confver}, ctx); code != 0 {
			ctx.Response = code
			return
		}
		applist := cmdb.AppsList(econf, common.V{"from": id}, 1000, 1, nil)
		for _, app := range *applist {
			rdh := &MgoHistorys{
				ID:            bson.NewObjectId(),
				AppID:         app.ID.Hex(),
				ApptplConfVer: confver,
				Status:        -1,
				CreateTime:    time.Now(),
			}
			if rdh.checkStatus(econf) {
				if err := rdh.add(econf); err != nil {
					econf.LogWarning("[%s] add history failed: %v\n", logKind, err)
					ctx.Response = 1803
					return
				}
			}
		}
	}
}

/*CallBack 部署任务的回调接口。
1. 更新任务日志
2. 更新部署状态
*/
func CallBack(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	cbr := jobs.CBResult{}
	if body, code, err := ctx.ReqBody(&cbr); code != 0 {
		econf.LogDebug("[%s] BodyFormatInvalid: [%s] with %v", logKind, string(body), err)
		ctx.Response = code
		return
	}
	user := ctx.Keys["username"].(string)
	if err := cbr.Logger(user, econf); err != nil {
		ctx.Response = 1302
		return
	}
	if err := updateStatus(cbr, econf); err != nil {
		econf.LogWarning("[%s] update task status failed: %v", logKind, err)
	}
	// jobnum := len(cbr.Result.Jobs)
	// if cbr.Result.Jobs[jobnum-1].Name == "lastjob" {
	// 	lastjob := cbr.Result.Jobs[jobnum-2]
	// 	if err := updateStatus(lastjob.Env["ZDOPS_DeployID"], lastjob.Env["ZDOPS_JobID"], econf); err != nil {
	// 		ctx.Response = 1207
	// 		return
	// 	}
	// }
	return
}

func saveFile(ctx *httpsvr.Context, savedir string, allowExt []string) string {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	mr, err := ctx.Req.MultipartReader()
	if err != nil {
		ctx.Response = 1206
		return ""
	}
	var part *multipart.Part
	part, err = mr.NextPart()
	if err != nil {
		ctx.Response = 1206
		return ""
	}
	//savedir := econf.GOPT.WorkDir + "/" + project + "/" + cluster + "/" + app
	if err = os.MkdirAll(savedir, 0755); err != nil {
		ctx.Response = 1901
		econf.LogDebug("[Deploy] mkdir %s failed: %v\n", savedir, err)
		return ""
	}
	if name := part.FormName(); name != "" {
		fname := part.FileName()
		econf.LogDebug("[%s] upload begin: %s %s\n", logKind, fname, name)
		fnamelen := len(fname)
		fnameExt := fname[fnamelen-4:]
		if fname != "" && (fnameExt == ".zip" || fnameExt == ".tgz") {
			// var b bytes.Buffer
			// io.CopyN(&b, part, 1048576+1)
			// econf.LogDebug("[%s] upload file %s with content: %s\n", logKind, fname, b.String())
			/*
				Content-Range:bytes 0-10485759/202286246
				Content-Type:multipart/form-data; boundary=----WebKitFormBoundaryduq7g7PEGTVOEwWh
			*/
			putname := fname
			fname = savedir + "/" + fname
			var f *os.File
			chunk := ctx.Req.Header.Get("Content-Range")
			//econf.LogDebug("[%s] chunk info: %s\n", logKind, chunk)
			var chunkStart, chunkEnd, chunkTotal int64
			var fstat os.FileInfo
			if len(chunk) < 12 {
				f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
			} else {
				i := strings.IndexByte(chunk[6:], '-')
				j := strings.IndexByte(chunk[6:], '/')
				chunkStart, _ = strconv.ParseInt(chunk[6:6+i], 10, 0)
				chunkEnd, _ = strconv.ParseInt(chunk[6+i+1:6+j], 10, 0)
				chunkTotal, _ = strconv.ParseInt(chunk[6+j+1:], 10, 0)
				econf.LogDebug("[%s] chunk info: %d-%d/%d\n", logKind, chunkStart, chunkEnd, chunkTotal)
				if i == -1 || j == -1 {
					ctx.Response = 1207
					return ""
				}
				if chunkStart == 0 {
					f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
				} else {
					fstat, err = os.Stat(fname)
					if err != nil {
						econf.LogWarning("[%s] The file %s is not exist.\n", logKind, fname)
						ctx.Response = 1207
						return ""
					}
					if fstat.Size() != chunkStart {
						ctx.Response = 1207
						return ""
					}
					f, err = os.OpenFile(fname, os.O_APPEND|os.O_RDWR, 0644)
				}
			}
			if err != nil {
				econf.LogDebug("[%s] Create file failed: %v\n", logKind, err)
				ctx.Response = 1206
				return ""
			}
			var wn int64
			wn, err = io.Copy(f, part)
			if cerr := f.Close(); err == nil {
				err = cerr
			}
			if err != nil {
				econf.LogWarning("[%s] Write file %s failed: %v\n", logKind, fname, err)
				os.Remove(f.Name())
				ctx.Response = 1207
				return ""
			}
			fstat, _ = os.Stat(fname)
			econf.LogDebug("[%s] upload file %s: size=%d; writelen: %d; chunkinfo: %d-%d/%d\n",
				logKind, fname, fstat.Size(), wn, chunkStart, chunkEnd, chunkTotal)
			if fstat.Size() != chunkEnd {
				ctx.Response = 1207
				return ""
			}
			//文件全部上传完成
			if chunkEnd == chunkTotal {
				return putname
			}
		}
	}
	return ""
}

//根据任务执行结果更新部署状态
func updateStatus(cbr jobs.CBResult, econf *common.EopsConf) error {
	rlog := jobs.GetLogger(cbr.Env["ZDOPS_TaskLogID"], econf)
	if !rlog.ID.Valid() {
		err := errors.New("Not found log recoder")
		return err
	}
	if rlog.Progress == rlog.HostNum {
		cdenv := cbr.Env["ZDOPS_CDEnv"]
		if cdenv == "" {
			err := errors.New("job not have cdenv")
			return err
		}
		i := econf.CDEnvIndex(cdenv)
		if i == -1 {
			err := errors.New("invalid cdenv")
			return err
		}
		st := (i + 1) * 10
		switch rlog.Status {
		case 0:
			if cdenv == options.CDEnvRelease {
				st = 0
			}
		case 900:
			st += 5
		case 901:
			st += 6
		default:
			st += 9
		}
		data := common.V{
			"$set": common.V{
				"status":              st,
				"flow.$.completetime": rlog.EndTime,
				"flow.$.status":       rlog.Failed,
				"flow.$.logid":        rlog.ID.Hex(),
			},
		}
		econf.LogDebug("deploy status: %v\n", data)
		c := econf.Mgo.Coll(mgoDeployHistory)
		skey := common.V{"_id": bson.ObjectIdHex(cbr.Env["ZDOPS_DeployID"]),
			// "flow.logid": rlog.ID.Hex(),
			"flow.cdenv": cdenv,
		}
		if err := c.Update(skey, data); err != nil {
			econf.LogWarning("[%s] skey:%v update history failed: %v\n", logKind, skey, err)
			return err
		}
	}
	return nil
}

/*

 */
func (rdh *MgoHistorys) deploy(params map[string]string, ctx *httpsvr.Context) int {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	var appinfo *cmdb.MgoApp
	var apptpl *cmdb.MgoAppTPL
	var cdenv, curenv string
	cdenvindex := -1
	//判断是否是新的部署任务
	if rdh.ID.Valid() {
		//执行现有部署任务，获取当前部署环境。
		if n := len(rdh.Flow); n > 0 {
			if rdh.Status%10 == 2 || rdh.Status%9 == 9 {
				curenv = rdh.Flow[n-1].CDEnv
			} else if rdh.Status != -1 {
				return 1804
			}
		}
		appinfo = cmdb.AppInfo(bson.ObjectIdHex(rdh.AppID), econf)
		apptpl = cmdb.AppTplInfo(bson.ObjectIdHex(appinfo.From), econf)
		if !appinfo.ID.Valid() || !apptpl.ID.Valid() {
			return 1803
		}
		//如果是部署失败的任务，可以进行重试。
		if rdh.Status%9 == 9 {
			cdenv = curenv
		} else {
			//否则查找下一个部署环境。条件是应用在指定环境中已分配服务器。
			i := econf.CDEnvIndex(curenv)
			if i == -1 {
				econf.LogDebug("[%s] the env %s is exist\n", logKind, curenv)
				return 1801
			} else if i == econf.GOPT.EnvironmentsLen-1 {
				econf.LogDebug("[%s] the env %s is last\n", logKind, curenv)
				return 1800
			}
			cdenvindex = i + 1
			cdenv = econf.GOPT.Environments[cdenvindex]
			// econf.LogDebug("cdenv: %s\n", cdenv)
			// econf.LogDebug1(appinfo)
			for appinfo.Environment[cdenv] != nil && appinfo.Environment[cdenv].HostsCount == 0 {
				cdenvindex++
				if cdenvindex >= econf.GOPT.EnvironmentsLen {
					econf.LogDebug("[%s] No available environment\n", logKind)
					return 1801
				}
				cdenv = econf.GOPT.Environments[cdenvindex]
			}
		}
	} else {
		//新的部署任务
		rdh.ID = bson.NewObjectId()
		rdh.CreateTime = time.Now()
		if params == nil {
			return 0
		}
		// econf.LogDebug1(params)
		//获取appinfo和apptpl信息
		if appid, ok := params["appid"]; ok {
			if p, c, a := cmdb.AppParent(appid, econf); p != nil {
				appinfo = a
				params["project"] = p.Name
				params["cluster"] = c.Name
				params["app"] = a.Name
			} else {
				return 1801
			}
			apptpl = cmdb.AppTplInfo(bson.ObjectIdHex(appinfo.From), econf)
			if !apptpl.ID.Valid() {
				return 1801
			}
		} else {
			appinfo, apptpl = cmdb.AppNameInfo(params["project"], params["cluster"], params["app"], econf)
			if appinfo == nil {
				return 1801
			}
			// econf.LogDebug1(appinfo)
		}
		rdh.AppID = appinfo.ID.Hex()
		//判断应用的部署状态，是否允许新建部署任务
		if !rdh.checkStatus(econf) {
			econf.LogDebug("[%s] Not allow create deploy\n", logKind)
			return 1803
		}

		if err := rdh.add(econf); err != nil {
			econf.LogWarning("[%s] add history failed: %v\n", logKind, err)
			return 1803
		}

		//如果是通过webhook提交的部署任务，第一个环境必须是构建环境。
		if _, ok := ctx.Keys["webhook"]; ok {
			if benv, ok1 := appinfo.Environment[options.CDEnvBuild]; ok1 &&
				benv.HostsCount > 0 &&
				econf.GOPT.Environments[0] == options.CDEnvBuild &&
				apptpl.Build != "" {
				cdenv = options.CDEnvBuild
				cdenvindex = 0
			} else {
				econf.LogWarning("[%s] app %s have not build environment.\n", logKind, appinfo.Name)
				return 1803
			}
		} else {
			//此处是通过上传包的形式新建部署任务。如果开发环境已经分配资源，则自动进行部署任务，否则返回。
			if benv, ok1 := appinfo.Environment[options.CDEnvDevelop]; ok1 && benv.HostsCount > 0 {
				cdenv = options.CDEnvDevelop
				cdenvindex = econf.CDEnvIndex(cdenv)
			}
		}
		if cdenv == "" {
			return 0
		}
	}
	var userinfo *auth.ZDUserSession
	user, _ := ctx.Keys["username"].(string)
	if userinfo, _ = auth.GetUserInfo(econf, user, 3); userinfo == nil {
		econf.LogDebug("[%s] invalid user %s\n", logKind, user)
		return 1801
	}
	perm := userinfo.PermissionList.Environment[cdenv]
	if perm == nil {
		econf.LogDebug("[%s] %s deploy %s is no permission on %s\n", logKind, userinfo.Name, appinfo.Name, cdenv)
		return 1107
	}
	//判断部署权限
	if (perm.App[appinfo.ID.Hex()]&auth.APIPermPCADeploy != auth.APIPermPCADeploy) && userinfo.Roles[0] != "admin" {
		// user, _ := json.MarshalIndent(userinfo, "", "  ")
		// econf.LogDebug("[%s] userinfo: %s\n", logKind, string(user))
		return 1107
	}
	appEnv := appinfo.Environment
	appVar := appinfo.Variables
	appVar["ZDOPS_Project"] = params["project"]
	appVar["ZDOPS_Cluster"] = params["cluster"]
	appVar["ZDOPS_App"] = params["app"]
	appVar["ZDOPS_AppSource"] = econf.GOPT.AppSource
	//appVar["ZDOPS_SourceType"] = appinfo.SourceType
	if appinfo.SourceType != "" && appinfo.Source != "" {
		appVar["ZDOPS_SourceType"] = appinfo.SourceType
		appVar["ZDOPS_Source"] = appinfo.Source
	} else if apptpl.SourceType != "" && apptpl.Source != "" {
		appVar["ZDOPS_SourceType"] = apptpl.SourceType
		appVar["ZDOPS_Source"] = apptpl.Source
	} else {
		econf.LogDebug("[%s] the app %s have not source\n", logKind, appinfo.Name)
		return 1802
	}
	if appVar["ZDOPS_SourceType"] == "HTTP" {
		appVar["ZDOPS_AppSource"] = appVar["ZDOPS_Source"]
	}
	for k, v := range appEnv[options.CDEnvBuild].Variables {
		appVar["ZDOPS_"+k] = v
	}
	appVar["ZDOPS_CallBack"] = econf.GOPT.BaseURL + econf.GOPT.Prefix + "/deploy/callback"
	cdtype := "JobsFlow"
	appVar["ZDOPS_AppInstallPath"] = econf.GOPT.AppInstallPath
	if rdh.AppVer != "" {
		appVar["ZDOPS_DeployType"] = "app"
		appVar["ZDOPS_AppVer"] = rdh.AppVer
	}
	if rdh.ApptplVer != "" {
		appVar["ZDOPS_DeployType"] = common.ApptplName
		appVar["ZDOPS_AppTplVer"] = rdh.ApptplVer
		appVar["ZDOPS_AppTpl"] = apptpl.Name
	}
	if rdh.ApptplConfVer != "" {
		appVar["ZDOPS_DeployType"] += ",apptplconf"
		appVar["ZDOPS_AppTplConfVer"] = rdh.ApptplConfVer
		appVar["ZDOPS_AppTplConfPath"] = apptpl.Config
	}
	if rdh.AppConfVer != "" {
		appVar["ZDOPS_DeployType"] += ",appconf"
		appVar["ZDOPS_AppConfVer"] = rdh.AppConfVer
		appVar["ZDOPS_AppConfPath"] = appinfo.Config
	}
	var taskname string
	taskattr := jobs.TaskAttr{}
	if cdenv == options.CDEnvBuild {
		cdtype = "jobs"
		buildjob := apptpl.Build
		appVar["ZDOPS_Build"] = common.TrueStr
		if isArgs := strings.IndexByte(apptpl.Build, ' '); isArgs >= 0 {
			buildjob = apptpl.Build[:isArgs]
			taskattr.Argv = apptpl.Build[isArgs+1:]
		}
		// taskattr.Name = buildjob
		// if i := strings.LastIndexByte(buildjob, '/'); i >= 0 {
		// 	taskattr.Name = buildjob[i+1:]
		// 	taskattr.Path = buildjob[:i]
		// }
		if jobtmp := jobs.GetJobByName(econf, buildjob, ""); jobtmp.ID.Valid() {
			taskattr.CommitID = jobtmp.CommitID
			taskattr.Timeout = jobtmp.Timeout
			taskattr.User = jobtmp.User
			taskattr.Priority = jobtmp.Priority
			taskattr.Name = jobtmp.Name
			taskattr.Path = jobtmp.Path
		} else {
			econf.LogDebug("[Deploy] buildjob %s is not exist\n", apptpl.Build)
			return 1802
		}
		taskname = taskattr.Name
	} else {
		//获取模版对应的部署作业流
		if deployFlow(apptpl, econf) {
			taskname = "apptpl-" + apptpl.Name
		}
	}
	// projectInfo := cmdb.ProjectInfo(project, econf)
	// clusterInfo := cmdb.ClusterInfo(cluster, econf)
	appVar["ZDOPS_DeployID"] = rdh.ID.Hex()
	jobRun := jobs.MgoJobLogs{
		ID:       bson.NewObjectId(),
		TaskName: taskname,
		TaskAttr: &taskattr,
		Target:   "apps",
		List:     []string{appinfo.ID.Hex()},
		Env:      appVar,
		Operator: userinfo.Name,
		CDEnv:    cdenv,
		Type:     cdtype,
	}
	err := jobRun.JobExec(ctx)
	if err != nil {
		econf.LogCritical("[Deploy] job exec failed with %s: %v\n", jobRun.ID.Hex(), err)
		return 1701
	}
	skey := common.V{"_id": rdh.ID}
	deployinfo := common.V{
		"$set": common.V{"status": (cdenvindex+1)*10 + 1},
		"$push": common.V{"flow": FlowInfo{
			CDEnv:      cdenv,
			LogID:      jobRun.ID.Hex(),
			Status:     -1,
			CreateTime: time.Now(),
		}},
	}
	c := econf.Mgo.Coll(mgoDeployHistory)
	if err := c.Update(skey, deployinfo); err != nil {
		econf.LogDebug("[%s] update history failed: %v\n", logKind, err)
		return 1803
	}
	ctx.OpLog.WriteString("Begin deploy to " + cdenv)
	return 0
}

//checkStatus 判断部署的状态，是否允许新建部署任务
func (rdh *MgoHistorys) checkStatus(econf *common.EopsConf) bool {
	result := []common.V{}
	econf.Mgo.Query(mgoDeployHistory, common.V{"appid": rdh.AppID, "status": common.V{"$in": []int{-2, -1, 10, 11}}}, &result, 100, 1, nil)
	for _, mdh := range result {
		// mdh := result[0]
		status := mdh["status"].(int)
		if status == -2 {
			if v, _ := mdh["appver"].(string); v != "" && (rdh.Status != -2 || rdh.AppVer == "") {
				rdh.AppVer = v
			}
			if v, _ := mdh["apptplver"].(string); v != "" && (rdh.Status != -2 || rdh.ApptplVer == "") {
				rdh.ApptplVer = v
			}
			if v, _ := mdh["appconfver"].(string); v != "" && (rdh.Status != -2 || rdh.AppConfVer == "") {
				rdh.AppConfVer = v
			}
			if v, _ := mdh["apptplconfver"].(string); v != "" && (rdh.Status != -2 || rdh.ApptplConfVer == "") {
				rdh.ApptplConfVer = v
			}
			rdh.ID = mdh["_id"].(bson.ObjectId)
			rdh.Status -= 2
		} else {
			return false
		}
	}
	return true
}

//add 添加部署记录
func (rdh *MgoHistorys) add(econf *common.EopsConf) error {
	c := econf.Mgo.Coll(mgoDeployHistory)
	var err error
	if rdh.Status < -2 {
		//合并部署。-3是checkStatus生成的临时状态。
		data := common.V{}
		data["appver"] = rdh.AppVer
		data["apptplver"] = rdh.ApptplVer
		data["appconfver"] = rdh.AppConfVer
		data["apptplconfver"] = rdh.ApptplConfVer
		data["status"] = rdh.Status + 2
		err = c.Update(&common.V{"_id": rdh.ID}, &common.V{"$set": data})
	} else {
		err = c.Insert(rdh)
		// if err = c.Insert(dh); err == nil {
		// err = createDeployFlow(dh.AppID, econf)
		// }
		// econf.LogDebug("[%s] insert history failed: %v\n", err)
	}
	return err
}

func deployFlow(apptpl *cmdb.MgoAppTPL, econf *common.EopsConf) bool {
	if flowInfo := jobs.FlowInfo("apptpl-"+apptpl.Name, econf); flowInfo.ID.Valid() {
		if flowInfo.Modifytime == apptpl.Modifytime {
			return true
		}
	}
	if err := createDeployFlow(apptpl, econf); err != nil {
		econf.LogWarning("[%s] apptpl %s no deploy flow: %v\n", logKind, apptpl.Name, err)
		return false
	}
	return true
}

//createDeployFlow 根据appid所使用的模板，创建或更新部署作业流。
func createDeployFlow(apptpl *cmdb.MgoAppTPL, econf *common.EopsConf) error {
	var err error
	// appInfo := cmdb.AppInfo(bson.ObjectIdHex(appid), econf)
	// apptpl := cmdb.AppTplInfo(bson.ObjectIdHex(apptplid), econf)
	dflow := &jobs.DeployFlows{
		Name:       "apptpl-" + apptpl.Name,
		Describe:   fmt.Sprintf("应用模板%s的自动部署作业流", apptpl.Name),
		Modifytime: apptpl.Modifytime,
	}
	if apptpl.ConfigVerify != "" {
		djob := &jobs.DeployJob{
			Alias: "ConfigVerify",
			Job:   apptpl.ConfigVerify,
		}
		dflow.Jobs = append(dflow.Jobs, djob)
	}
	if apptpl.PreInstall != "" {
		djob := &jobs.DeployJob{
			Alias: "PreInstall",
			Job:   apptpl.PreInstall,
		}
		dflow.Jobs = append(dflow.Jobs, djob)
	}
	if apptpl.Install != "" {
		djob := &jobs.DeployJob{
			Alias: "Install",
			Job:   apptpl.Install,
		}
		dflow.Jobs = append(dflow.Jobs, djob)
	}
	if apptpl.PostInstall != "" {
		djob := &jobs.DeployJob{
			Alias: "PostInstall",
			Job:   apptpl.PostInstall,
		}
		dflow.Jobs = append(dflow.Jobs, djob)
	}
	if apptpl.RollBack != "" {
		dflow.RollBack = apptpl.RollBack
	} else {
		err = errors.New("no rollback job")
	}
	if len(dflow.Jobs) == 0 {
		err = errors.New("no deploy job")
	}
	if err != nil {
		return err
	}
	err = dflow.DeployFlow(econf)
	return err
}
