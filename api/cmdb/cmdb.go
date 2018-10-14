package cmdb

import (
	//"fmt"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"
)

const (
	//mgoApps apps的表名
	mgoApps = "apps"
	//mgoAppConf appconf的表名
	mgoAppConf = "appconf"
	//mgoAppTpl 应用模板
	mgoAppTpl = "apptpl"
	//mgoProject 项目的表名
	mgoProject = "project"
	//mgoCluster 集群的表名
	mgoCluster = "cluster"
	//mgoResource 资源配置的表名
	mgoResource    = "resource"
	mgoYunProvider = "yunprovider"
	mgoIDCProvider = "idcprovider"
	//AppSourceTypeHTTP 应用安装源的类型
	AppSourceTypeHTTP = 0x1
	AppSourceTypeGIT  = 0x2
	AppSourceTypeSVN  = 0x3
)

//AddItem 获取用户列表或用户信息
func AddItem(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	//rmsg := common.NewRespMsg()
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	cm := ctx.Req.URL.Path[i+1:]
	var body []byte
	var code int
	var err error
	curtime := time.Now()
	var zdcmdbitem interface{}
	id := bson.NewObjectId()
	switch cm {
	case mgoProject:
		zdcmdbitem = &MgoProject{ID: id, Createtime: curtime, Modifytime: curtime}
	case mgoCluster:
		zdcmdbitem = &MgoCluster{ID: id, Createtime: curtime, Modifytime: curtime}
	case mgoApps:
		zdcmdbitem = &MgoApp{
			ID:         id,
			Createtime: curtime,
			Modifytime: curtime,
			Status:     -2,
			Operator:   ctx.Keys["username"].(string),
		}
	case mgoAppConf:
		zdcmdbitem = &MgoAppConf{ID: id, Createtime: curtime, Status: 1, Modifytime: curtime}
		aconf := common.V{}
		body, code, err = ctx.ReqBody(&aconf)
		if code != 0 {
			econf.LogDebug("[CMDB] BodyFormatInvalid: %v", err)
			ctx.Response = code
			return
		}
		if err := json.Unmarshal(body, zdcmdbitem); err != nil {
			ctx.Response = 1204
			return
		}
		user := ctx.Keys["username"].(string)
		if code, err = AppConfPreUpdate(aconf, econf, user, "POST"); code != 0 {
			econf.LogWarning("[CMDB] update appconf with [%v] is false: %v", aconf, err)
			ctx.Response = code
			return
		}
	case mgoAppTpl:
		zdcmdbitem = &MgoAppTPL{ID: id, Createtime: curtime, Modifytime: curtime}
	case mgoYunProvider:
		zdcmdbitem = &MgoYunProvider{ID: id, Createtime: curtime, Modifytime: curtime}
	case mgoIDCProvider:
		zdcmdbitem = &MgoIDCProvider{ID: id, Createtime: curtime, Modifytime: curtime}
	default:
		ctx.Response = 1207
		return
	}
	// var body []byte
	// var code int
	if cm != mgoAppConf {
		if body, code, err = ctx.ReqBody(zdcmdbitem); code != 0 {
			econf.LogDebug("[CMDB] BodyFormatInvalid: %v", err)
			ctx.Response = code
			return
		}
	}
	if err = common.InputFilter(zdcmdbitem, 1, nil, nil); err != nil {
		econf.LogInfo("[CMDB] invalid input: %v\n", err)
		ctx.Response = 1101
		return
	}
	c := econf.Mgo.Coll(cm)
	if err = c.Insert(zdcmdbitem); err != nil {
		econf.LogWarning("[CMDB] insert %s failed: %v\n", cm, err)
		ctx.Response = 1305
		return
	}
	if cm == mgoApps {
		apptpl := mgoRecoder(bson.ObjectIdHex(zdcmdbitem.(*MgoApp).From), mgoAppTpl, econf).(*MgoAppTPL)
		if !apptpl.ID.Valid() {
			c.Remove(common.V{"_id": zdcmdbitem.(*MgoApp).ID})
			ctx.Response = 1309
			return
		}
		if err = apptplUsage(econf, zdcmdbitem.(*MgoApp).From, 1); err != nil {
			econf.LogWarning("[CMDB] update apptpl failed: %v\n", err)
			ctx.Response = 1309
			return
		}
		depends := []string{}
		for _, depend := range apptpl.Depend {
			app := &MgoApp{
				ID:         bson.NewObjectId(),
				Name:       depend,
				Cluster:    zdcmdbitem.(*MgoApp).Cluster,
				From:       depend,
				Createtime: time.Now(),
				Status:     -1,
				Operator:   zdcmdbitem.(*MgoApp).Operator,
			}
			depends = append(depends, app.ID.Hex())
			if err = c.Insert(app); err != nil {
				econf.LogWarning("[CMDB] insert %s failed: %v\n", cm, err)
				ctx.Response = 1305
				return
			}
			if err = apptplUsage(econf, depend, 1); err != nil {
				econf.LogWarning("[CMDB] update apptpl failed: %v\n", err)
				ctx.Response = 1309
				return
			}
		}
		skey := common.V{
			"_id": zdcmdbitem.(*MgoApp).ID,
		}
		if err = c.Update(skey, common.V{"$set": common.V{"status": -1, "depend": depends}}); err != nil {
			econf.LogWarning("[CMDB] update %s failed: %v\n", cm, err)
			ctx.Response = 1302
			return
		}
	}
	ctx.OpLog.Write(body)
	return
}

//AddResource 获取用户列表或用户信息
func AddResource(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	//rmsg := common.NewRespMsg()
	resource := MgoResource{ID: bson.NewObjectId(), Createtime: time.Now()}
	var body []byte
	var code int
	var err error

	if body, code, err = ctx.ReqBody(&resource); code != 0 {
		econf.LogDebug("[CMDB] BodyFormatInvalid: %v", err)
		ctx.Response = code
		return
	}
	resource.Status = 0
	if err := common.InputFilter(&resource, 1, nil, nil); err != nil {
		econf.LogDebug("[CMDB] invalid input: %v\n", err)
		ctx.Response = 1101
		return
	}
	// additionkey := econf.BaseData("addition", "resource")
	// if additionkey != nil {
	// 	for _, bv := range additionkey.([]string) {
	// 		if vv, err := postArgs[bv]; err == true {
	// 			tmp := bson.DocElem{Name: bv, Value: vv}
	// 			inputd = append(inputd, tmp)
	// 		}
	// 	}
	// }
	c := econf.Mgo.Coll(mgoResource)
	if err := c.Insert(resource); err != nil {
		econf.LogWarning("[CMDB] insert %s failed: %v\n", mgoResource, err)
		ctx.Response = 1024
		errstr := err.Error()
		if errstr[0:6] == "E11000" {
			//记录已存在
			ctx.Response = 1103
		}
		return
	}
	ctx.OpLog.Write(body)
	return
}

//UpdateResource 获取用户列表或用户信息
func UpdateResource(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	var err error
	data := common.V{}
	skey := common.V{}
	if len(ctx.BodyMap) == 0 {
		ctx.Response = 1204
		return
	}
	postArgs := common.V(ctx.BodyMap)
	resource := MgoResource{}
	if err = common.InputFilter(&resource, 2, postArgs, data); err != nil {
		econf.LogInfo("[CMDB] invalid input: %v\n", err)
		ctx.Response = 1101
		return
	}
	// additionkey := econf.BaseData("addition", "resource")
	// if additionkey != nil {
	// 	for _, bv := range additionkey.([]string) {
	// 		filArgs[bv] = postArgs[bv]
	// 	}
	// }
	if id := postArgs.VString("id"); id != "" && common.StrFilter(id, "id") {
		skey["_id"] = bson.ObjectIdHex(id)
	} else if hostname := postArgs.VString("hostname"); hostname != "" && common.StrFilter(hostname, "username") {
		skey["hostname"] = hostname
	} else if hid := postArgs.VString("hostid"); hid != "" && common.StrFilter(hid, "username") {
		skey["hostid"] = hid
	}
	data["modifytime"] = time.Now()
	if err = econf.Mgo.Update(mgoResource, skey, common.V{"$set": data}); err != nil {
		econf.LogWarning("[CMDB] update %s failed: %v; skey=%v,data=%v\n", mgoResource, err, skey, data)
		ctx.Response = 1204
		return
	}
	ctx.OpLog.Write(common.V{"skey": skey, "data": data}.Json(""))
	return
}

//UpdateItem 获取用户列表或用户信息
func UpdateItem(ctx *httpsvr.Context) {
	if len(ctx.BodyMap) == 0 {
		ctx.Response = 1204
		return
	}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	cm := ctx.Req.URL.Path[i+1:]
	//econf.LogDebug("[CMDB] body: %v\n", postArgs)

	data, code := UpdateCmdb(cm, common.V(ctx.BodyMap), ctx)
	//rmsg.SetCode(code, econf.Err)
	ctx.Response = code
	if code == 0 {
		skey := data["skey"].(common.V)
		ctx.OpLog.Write(skey.Json("skey"))
		delete(data, "skey")
		ctx.OpLog.Write(data.Json("data"))
	}
	return
}

func UpdateApp(data common.V, ctx *httpsvr.Context) int {
	_, code := UpdateCmdb(mgoApps, data, ctx)
	return code
}
func UpdateAppTpl(data common.V, ctx *httpsvr.Context) int {
	_, code := UpdateCmdb(mgoAppTpl, data, ctx)
	return code
}

func UpdateCmdb(cm string, postArgs common.V, ctx *httpsvr.Context) (common.V, int) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	skey := common.V{}
	var id string
	if id = postArgs.VString("_id"); id != "" && common.StrFilter(id, "id") {
		skey["_id"] = bson.ObjectIdHex(id)
		delete(postArgs, "_id")
	} else {
		ctx.Response = 1206
	}
	var zdcmdbitem interface{}
	switch cm {
	case mgoProject:
		zdcmdbitem = &MgoProject{}
	case mgoCluster:
		zdcmdbitem = &MgoCluster{}
	case "appver":
		project, cluster, app := AppParent(id, econf)
		if auth.PermChkPCAConfig(
			project.ID.Hex(),
			cluster.ID.Hex(),
			app.ID.Hex(),
			econf.Route.API[ctx.Keys["api"].(string)].Perm,
			ctx,
		) == false {
			return nil, 1107
		}
		if appvar := postArgs.ObjType("variables", "map"); appvar != nil {
			if err := appVar(common.V(appvar.(map[string]interface{})), id, ctx); err != nil {
				return nil, 1206
			}
			return common.V{"skey": skey, "data": appvar}, 0
		}
		return nil, 1206
	case mgoApps:
		zdcmdbitem = &MgoApp{}
	case mgoAppConf:
		project, cluster, app := AppParent(id, econf)
		if auth.PermChkPCAConfig(
			project.ID.Hex(),
			cluster.ID.Hex(),
			app.ID.Hex(),
			econf.Route.API[ctx.Keys["api"].(string)].Perm,
			ctx,
		) == false {
			return nil, 1107
		}
		zdcmdbitem = &MgoAppConf{}
		uid := ctx.Keys["username"].(string)
		if code, err := AppConfPreUpdate(postArgs, econf, uid, "PUT"); code != 0 {
			econf.LogDebug("[CMDB] update appconf with [%v] is false: %v", postArgs, err)
			return nil, code
		}
	case mgoAppTpl:
		zdcmdbitem = &MgoAppTPL{}
		delete(postArgs, "name")
	case mgoYunProvider:
		zdcmdbitem = &MgoYunProvider{}
	case mgoIDCProvider:
		zdcmdbitem = &MgoIDCProvider{}
	default:
		return nil, 1207
	}
	data := common.V{}
	if err := common.InputFilter(zdcmdbitem, 2, postArgs, data); err != nil {
		econf.LogWarning("[CMDB] invalid input: %v\n", err)
		return nil, 1101
	}
	data["modifytime"] = time.Now()
	// if _, ok := filArgs["createtime"]; ok {
	// 	delete(filArgs, "createtime")
	// }
	econf.LogDebug("[CMDB] UpdateCmdb: %v\n", data)
	//delete(filArgs, "_id")
	if err := econf.Mgo.Update(cm, skey, common.V{"$set": data}); err != nil {
		econf.LogWarning("[CMDB] update %s: %v\n", cm, err)
		return data, 1302
	}
	//filArgs["skey"] = skey
	return common.V{"skey": skey, "data": data}, 0
}

//DeleteItem 更新用户信息
func DeleteItem(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	skey := common.V{}
	id := ctx.Req.Form.Get("id")
	if id != "" {
		skey["_id"] = bson.ObjectIdHex(id)
	} else {
		ctx.Response = 1206
		return
	}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	cm := ctx.Req.URL.Path[i+1:]
	c := econf.Mgo.Coll(cm)
	// result := []common.V{}
	switch cm {
	case mgoApps:
		appinfo := mgoRecoder(&skey, mgoApps, econf).(*MgoApp)
		if appinfo.ID.Valid() {
			apptplUsage(econf, appinfo.From, -1)
		}
	case mgoResource:
		// 确保资源被删除前，已经没有应用使用
		skey["usage"] = 0
	}
	if cm == mgoApps {
	}
	if err := c.Remove(skey); err != nil {
		econf.LogCritical("err mgdb remove: %v\n", err)
		ctx.Response = 1304
		return
	}
	ctx.OpLog.Write(skey.Json("skey"))
	return
}

//Items 更新用户信息
func Items(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	skey := common.V{}
	pageNo := 1
	pageSize := 2000
	fields := common.V{}
	i := strings.LastIndexByte(ctx.Req.URL.Path, '/')
	cm := ctx.Req.URL.Path[i+1:]
	if cm == "resourceall" {
		cm = mgoResource
	}
	for k, v := range ctx.Req.Form {
		switch k {
		case "pageNo":
			pageNo, _ = strconv.Atoi(v[0])
		case "pageSize":
			pageSize, _ = strconv.Atoi(v[0])
		case "name":
			if common.StrFilter(v[0], "username") {
				skey["name"] = v[0]
			} else {
				ctx.Response = 1206
				return
			}
		case "id", "appid", "_id":
			if common.StrFilter(v[0], "id") {
				if k == "appid" {
					skey["appid"] = v[0]
				} else {
					skey["_id"] = bson.ObjectIdHex(v[0])
				}
			} else {
				ctx.Response = 1206
				return
			}
		case "appvar":
			fields = common.V{"environment." + ctx.Keys["workenv"].(string) + ".variables": 1}
		case "additionconf":
			fields = common.V{"additionconf": 1}
		case "smalllist":
			fields = common.V{"name": 1}
			if cm == mgoResource {
				fields = common.V{"hostname": 1}
			} else if cm == mgoAppTpl {
				fields = common.V{"name": 1, "config": 1, "source": 1, "sourcetype": 1}
			}
		case "environment":
			if cm == mgoResource {
				if common.SearchStr(econf.GOPT.Environments, v[0]) {
					skey["environment"] = v[0]
				} else {
					ctx.Response = 1301
					return
				}
			}
		default:

		}
	}
	if cm == mgoYunProvider {
		fields["secret"] = 0
	}
	if cm == "appver" {
		appInfo := AppInfo(skey, econf)
		if appInfo.ID.Valid() {
			project, cluster, app := AppParent(appInfo.ID.Hex(), econf)
			if auth.PermChkPCAConfig(
				project.ID.Hex(),
				cluster.ID.Hex(),
				app.ID.Hex(),
				econf.Route.API[ctx.Keys["api"].(string)].Perm,
				ctx,
			) == false {
				ctx.Response = 1107
				return
			}
			userInfo, _ := auth.GetUserInfo(econf, ctx.Keys["username"].(string), 1)
			if workenv, ok := appInfo.Environment[userInfo.WorkEnv]; ok {
				rmsg.TotalRecord = len(workenv.Variables)
				rmsg.DataList = workenv.Variables
			}
			ctx.Response = rmsg
			ctx.OpLog.Write(skey.Json("skey"))
		} else {
			ctx.Response = 1301
		}
		return
	}
	if pageSize > common.MaxPageSize {
		pageSize = common.MaxPageSize
	}
	result := []common.V{}
	recnum, _ := econf.Mgo.QueryM(cm, skey, &result, pageSize, pageNo, fields)
	rmsg.TotalRecord = recnum
	rmsg.DataList = result
	if recnum > 0 {
		if _, hasid := skey["_id"]; hasid && cm == mgoAppConf {
			appid := result[0]["appid"].(string)
			project, cluster, app := AppParent(appid, econf)
			if auth.PermChkPCAConfig(
				project.ID.Hex(),
				cluster.ID.Hex(),
				app.ID.Hex(),
				econf.Route.API[ctx.Keys["api"].(string)].Perm,
				ctx,
			) == false {
				ctx.Response = 1107
				return
			}
			uid := ctx.Keys["username"].(string)
			appinfo := AppConfPath(appid, uid, econf)
			path := result[0]["path"].(string)
			name := result[0]["name"].(string)
			//econf.LogDebug("get conf: %v", name)
			gfile, err := econf.GetGitFile(path+"/"+name,
				"cmdb-"+appinfo["project"]+"/"+appinfo["cluster"]+"-"+appinfo["app"]+"-conf",
				"",
			)
			if err != nil {
				econf.LogWarning("git %s/%s failed: %v\n", path, name, err)
				ctx.Response = 1601
				return
			}
			result[0]["content"] = gfile.Content
		}
	}
	ctx.Response = rmsg
	ctx.OpLog.Write(skey.Json("skey"))
}

//ProjectTree 获取项目树
func ProjectTree(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	user := ctx.Keys["username"].(string)
	cdenv := ctx.Keys["workenv"].(string)
	projects := []MgoProject{}
	clusters := []MgoCluster{}
	apps := []MgoApp{}
	econf.Mgo.Query(mgoProject, common.V{}, &projects, 2000, 1, nil)
	econf.Mgo.Query(mgoCluster, common.V{}, &clusters, 2000, 1, nil)
	econf.Mgo.Query(mgoApps, common.V{}, &apps, 2000, 1, nil)
	tree := make([]common.V, len(projects)+1)
	depends := make(map[string]bool)
	for _, app := range apps {
		for _, d := range app.Depend {
			depends[d] = true
		}
	}
	aCluster := make(map[string][]common.V)
	aClusterCount := make(map[string]int)
	for _, app := range apps {
		nodeApps := common.V{
			"text": app.Name,
			"id":   app.ID.Hex(),
			"type": "apps",
		}
		tags := []string{}
		if depends[app.ID.Hex()] {
			nodeApps["color"] = "#ff6600"
		} else {
			tags = append(tags, fmt.Sprintf("%d", app.Environment[cdenv].HostsCount))
		}
		if app.Status != 0 && app.Operator != user && auth.IsAdmin(ctx) {
			tags = append(tags, "<i class=\"fa fa-lock\"></i>")
		}
		nodeApps["tags"] = tags
		aCluster[app.Cluster] = append(aCluster[app.Cluster], nodeApps)
		aClusterCount[app.Cluster] += app.Environment[cdenv].HostsCount
	}
	cProject := make(map[string][]common.V)
	cProjectCount := make(map[string]int)
	for _, cluster := range clusters {
		aCluster[cluster.ID.Hex()] = append(aCluster[cluster.ID.Hex()], common.V{
			"text":    "添加应用",
			"type":    "apps",
			"cluster": cluster.ID.Hex(),
			"nameid":  "add",
		})
		nodeCluster := common.V{
			"text":   cluster.Name,
			"id":     cluster.ID.Hex(),
			"shared": cluster.Shared,
			"depend": cluster.Depend,
			"type":   "cluster",
			"state":  common.V{"expanded": false},
			"node":   aCluster[cluster.ID.Hex()],
			"tags":   []string{fmt.Sprintf("%d", aClusterCount[cluster.ID.Hex()])},
		}
		cProject[cluster.Project] = append(cProject[cluster.Project], nodeCluster)
		cProjectCount[cluster.Project] += aClusterCount[cluster.ID.Hex()]
	}
	var n int
	for i, project := range projects {
		cProject[project.ID.Hex()] = append(cProject[project.ID.Hex()], common.V{
			"text":    "添加集群",
			"type":    "cluster",
			"project": project.ID.Hex(),
			"nameid":  "add",
		})
		tree[i] = common.V{
			"text":  project.Name,
			"id":    project.ID.Hex(),
			"state": common.V{"expanded": false},
			"type":  "project",
			"node":  cProject[project.ID.Hex()],
			"tags":  []string{fmt.Sprintf("%d", cProjectCount[project.ID.Hex()])},
		}
		n = i
	}
	tree[n+1] = common.V{
		"text":   "添加项目",
		"type":   "project",
		"nameid": "add",
	}
	rmsg := common.NewRespMsg()
	rmsg.TotalRecord = len(tree)
	rmsg.DataList = tree
	ctx.Response = rmsg
}

//AllocResource 为应用分配资源
func AllocResource(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	var err error
	if len(ctx.BodyMap) == 0 {
		ctx.Response = 1204
		return
	}
	postArgs := common.V(ctx.BodyMap)
	appid := postArgs.VString("appid")
	if appid == "" || !common.StrFilter(appid, "id") {
		ctx.Response = 1207
		return
	}
	resourcelist := []string{}
	//hostlist := []string{}
	if rlist := postArgs.ObjType("resourcelist", "[]interface"); rlist != nil {
		for _, k := range rlist.([]interface{}) {
			switch v := k.(type) {
			case string:
				resourcelist = append(resourcelist, v)
			}
		}
	} else {
		ctx.Response = 1205
		return
	}
	userinfo, _ := auth.GetUserInfo(econf, ctx.Keys["username"].(string), 1)
	econf.LogDebug("[CMDB] AllocResource list: %d\n", len(resourcelist))
	alloctype := postArgs.VString("alloctype")
	if alloctype == "" {
		alloctype = "replace"
	}
	skey := common.V{"_id": bson.ObjectIdHex(appid)}
	var data common.V
	step := 1
	switch alloctype {
	case "replace":
		data = common.V{"$set": common.V{
			"environment." + userinfo.WorkEnv + ".hosts":      resourcelist,
			"environment." + userinfo.WorkEnv + ".hostscount": len(resourcelist),
		}}
		resetResource(econf, appid, userinfo.WorkEnv, resourcelist)
	case "push":
		data = common.V{
			"$push": common.V{
				"environment." + userinfo.WorkEnv + ".hosts": common.V{"$each": resourcelist},
			},
			"$inc": common.V{"environment." + userinfo.WorkEnv + ".hostscount": len(resourcelist)},
		}
	case "pull":
		data = common.V{
			"$pull": common.V{
				"environment." + userinfo.WorkEnv + ".hosts": common.V{"$in": resourcelist},
			},
			"$inc": common.V{"environment." + userinfo.WorkEnv + ".hostscount": -len(resourcelist)},
		}
		step = -1
	}
	if err = econf.Mgo.Update(mgoApps, skey, data); err != nil {
		econf.LogWarning("[CMDB] update %s with %s failed: %v\n", mgoApps, appid, err)
		ctx.Response = 1204
	}
	appInfo := mgoRecoder(bson.ObjectIdHex(appid), mgoApps, econf).(*MgoApp)
	if !appInfo.ID.Valid() {
		ctx.Response = 1204
		return
	}
	for _, id := range appInfo.Depend {
		skey["_id"] = bson.ObjectIdHex(id)
		if err := econf.Mgo.Update(mgoApps, skey, data); err != nil {
			econf.LogWarning("[CMDB] update %s with %s failed: %v\n", mgoApps, id, err)
			ctx.Response = 1204
			return
		}
	}
	skey = common.V{"hostname": common.V{"$in": resourcelist}}
	upaction := common.V{"$inc": common.V{"usage": step}}
	c := econf.Mgo.Coll(mgoResource)
	if _, err := c.UpdateAll(skey, upaction); err != nil {
		econf.LogWarning("[CMDB] update mgo %s failed: %v\n", mgoResource, err)
		ctx.Response = 1204
		return
	}
	ctx.OpLog.Write(common.V{"skey": skey, "data": data}.Json(""))
	return
}

//ISPList 更新用户信息
func ISPList(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	r := econf.BaseData("isplist", "isp")
	rmsg.DataList = r
	switch r.(type) {
	case []interface{}:
		rmsg.TotalRecord = len(r.([]interface{}))
	default:
		rmsg.TotalRecord = 1
	}
	ctx.Response = rmsg
}

//SVCProvider 服务商列表
func SVCProvider(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	rYun := []common.V{}
	yunnum, _ := econf.Mgo.Query(mgoYunProvider, common.V{}, &rYun, 2000, 1, common.V{"name": 1, "proxy": 1})
	rIDC := []common.V{}
	idcnum, _ := econf.Mgo.Query(mgoIDCProvider, common.V{}, &rIDC, 2000, 1, common.V{"name": 1, "proxy": 1})
	rYun = append(rYun, rIDC...)
	rmsg.DataList = rYun
	rmsg.TotalRecord = yunnum + idcnum
	ctx.Response = rmsg
}

//AppConfPreUpdate appconf更新前的操作，包括判断状态，更新apps，更新git
func AppConfPreUpdate(aconf common.V, econf *common.EopsConf, user, method string) (code int, err error) {
	args := make(map[string]string)
	code = 1101
	/*
		获取必要参数，并对参数进行必要的判断
	*/
	fields := map[string]string{
		"name":    "username",
		"path":    "path",
		"from":    "username",
		"appid":   "id",
		"content": "",
	}
	for k, v := range fields {
		tmp, ok := aconf[k]
		if ok == false {
			err = errors.New("havenot the key: " + v)
			return
		}
		switch tmp.(type) {
		case string:
			str := tmp.(string)
			if common.StrFilter(str, v) == false {
				err = errors.New("the " + k + "'s value " + str + " invalid")
				return
			}
			args[k] = str
		default:
			err = errors.New("the key's[" + v + "] value is not string")
			return
		}
	}
	skey := common.V{
		"_id":          bson.ObjectIdHex(args["from"]),
		"additionconf": common.V{"$elemMatch": common.V{"$eq": args["path"]}},
	}
	//判断输入的path是否是模版中允许的
	_, err = econf.Mgo.Query(mgoAppTpl, skey, nil, 10, 1, common.V{"additionconf": 1})
	if err != nil {
		//err = errors.New("apptpl have not the path: " + args["path"])
		return
	}
	appinfo := AppConfPath(args["appid"], user, econf)
	switch appinfo["status"] {
	case "false":
		err = errors.New(appinfo["msg"])
		return
	case "locked":
		err = errors.New("The AppConf is being edited by other people")
		return
	default:
		c := econf.Mgo.Coll(mgoApps)
		if err = c.Update(common.V{"_id": bson.ObjectIdHex(args["appid"])}, common.V{"$set": common.V{"operator": user, "status": 1}}); err != nil {
			code = 1204
			err = errors.New("update status falsed")
			return
		}
	}
	userinfo, _ := auth.GetUserInfo(econf, user, 3)
	token := econf.GOPT.GitToken
	if userinfo.GitToken != "" {
		token = userinfo.GitToken
	}
	git := common.GitAction{
		Action:   "files",
		Method:   method,
		Url:      econf.GOPT.GitURL,
		GitToken: token,
		User:     user,
		Projects: "cmdb-" + appinfo["project"] + "/" + appinfo["cluster"] + "-" + appinfo["app"] + "-conf",
		//Path:     appinfo["project"] + "%2F" + appinfo["cluster"] + "%2F" + appinfo["app"] + "%2F" + args["path"] + "%2F" + args["name"],
		Path:    args["path"] + "/" + args["name"],
		Content: args["content"],
	}
	code = 0
	if err = git.GitAPI(); err != nil {
		code = 1601
	}
	return
}

//AppConfPath 获取app依赖关系及判断是否有人在编辑配置
func AppConfPath(appid, user string, econf *common.EopsConf) (result map[string]string) {
	result = make(map[string]string)
	result["status"] = "false"
	appInfo := mgoRecoder(bson.ObjectIdHex(appid), mgoApps, econf).(*MgoApp)
	if !appInfo.ID.Valid() {
		result["msg"] = "Not found the apps"
		return
	}
	result["app"] = appInfo.Name
	//判断当前apps是否有没有下发的配置。如果有，只允许上次编辑的人员进行更新，避免多人同时编辑
	if appInfo.Operator != user && appInfo.Status != 0 {
		result["status"] = "locked"
		return
	}
	clsInfo := mgoRecoder(bson.ObjectIdHex(appInfo.Cluster), mgoCluster, econf).(*MgoCluster)
	if !clsInfo.ID.Valid() {
		result["msg"] = "Not found the cluster"
		return
	}
	projectInfo := mgoRecoder(bson.ObjectIdHex(clsInfo.Project), mgoProject, econf).(*MgoProject)
	if !projectInfo.ID.Valid() {
		result["msg"] = "Not found the project"
		return
	}
	result["cluster"] = clsInfo.Name
	result["project"] = projectInfo.Name
	result["status"] = "ok"
	return
}

//AppNameInfo 获取appid
func AppNameInfo(project, cluster, app string, econf *common.EopsConf) (*MgoApp, *MgoAppTPL) {
	rProject := MgoProject{}
	if rn, _ := econf.Mgo.Query(mgoProject, project, &rProject, 0, 0, nil); rn == 0 {
		return nil, nil
	}
	rCluster := MgoCluster{}
	if rn, _ := econf.Mgo.Query(mgoCluster, common.V{"name": cluster, "project": rProject.ID.Hex()}, &rCluster, 0, 0, nil); rn == 0 {
		return nil, nil
	}
	rApp := MgoApp{}
	if rn, _ := econf.Mgo.Query(mgoApps, common.V{"name": app, "cluster": rCluster.ID.Hex()}, &rApp, 0, 0, nil); rn == 0 {
		return nil, nil
	}
	if rApp.From == "" {
		econf.LogCritical("[MGDB] app(%s) have not tpl\n", app)
		return nil, nil
	}
	rApptpl := MgoAppTPL{}
	econf.Mgo.Query(mgoAppTpl, bson.ObjectIdHex(rApp.From), &rApptpl, 0, 0, nil)
	return &rApp, &rApptpl
}

//AppInfo 获取appid
func AppInfo(key interface{}, econf *common.EopsConf) *MgoApp {
	return mgoRecoder(key, mgoApps, econf).(*MgoApp)
}

//ClusterInfo 获取appid
func ClusterInfo(key interface{}, econf *common.EopsConf) *MgoCluster {
	return mgoRecoder(key, mgoCluster, econf).(*MgoCluster)
}

//ProjectInfo 获取appid
func ProjectInfo(key interface{}, econf *common.EopsConf) *MgoProject {
	return mgoRecoder(key, mgoProject, econf).(*MgoProject)
}

//AppTplInfo 获取appid
func AppTplInfo(key interface{}, econf *common.EopsConf) *MgoAppTPL {
	return mgoRecoder(key, mgoAppTpl, econf).(*MgoAppTPL)
}

//AppTplChkVer 模板的版本同步状态。是否所有依赖此模板的应用的部署版本和模板一致。包括配置
func AppTplChkVer(id string, econf *common.EopsConf) bool {
	apptplinfo := AppTplInfo(id, econf)
	if apptplinfo.ID.Valid() {
		applist := AppsList(econf, id, 2000, 1, common.V{"appver": 1})
		for _, appinfo := range *applist {
			if appinfo.AppVer != apptplinfo.AppVer {
				return false
			}
		}
	}
	return true
}

//AppTplConfChkVer 模板的版本同步状态。是否所有依赖此模板的应用的部署版本和模板一致。包括配置
func AppTplConfChkVer(id string, econf *common.EopsConf) bool {
	apptplinfo := AppTplInfo(id, econf)
	if apptplinfo.ID.Valid() {
		applist := AppsList(econf, id, 2000, 1, common.V{"configver": 1})
		for _, appinfo := range *applist {
			if appinfo.ConfigVer != apptplinfo.ConfigVer {
				return false
			}
		}
	}
	return true
}

//AppsList 应用列表
func AppsList(econf *common.EopsConf, key interface{}, pageSize, pageNo int, selector common.V) *[]MgoApp {
	r := []MgoApp{}
	if _, err := econf.Mgo.Query(mgoApps, key, &r, pageSize, pageNo, selector); err != nil {
		econf.LogWarning("[Job] Not found job recoder %v with: %v\n", key, err)
	}
	return &r
}

//AppParent 根据Appid获取所属集群和项目
func AppParent(appid string, econf *common.EopsConf) (*MgoProject, *MgoCluster, *MgoApp) {
	if appInfo := mgoRecoder(bson.ObjectIdHex(appid), mgoApps, econf).(*MgoApp); appInfo.ID.Valid() {
		if clusterInfo := mgoRecoder(bson.ObjectIdHex(appInfo.Cluster), mgoCluster, econf).(*MgoCluster); clusterInfo.ID.Valid() {
			projectInfo := mgoRecoder(bson.ObjectIdHex(clusterInfo.Project), mgoProject, econf).(*MgoProject)
			if projectInfo.ID.Valid() {
				return projectInfo, clusterInfo, appInfo
			}
		}
	}
	return nil, nil, nil
}

func AssignedHosts(econf *common.EopsConf, target string, list []string, cdenv string) (hostlist []string) {
	// key := make([]bson.ObjectId,len(list))
	// for i, id := range list {
	// 	key[i] = bson.ObjectIdHex(id)
	// }
	skey := common.V{}
	switch target {
	case "projects":
		clusters := []MgoCluster{}
		key := common.V{"project": common.V{"$in": list}}
		if n, _ := econf.Mgo.Query(mgoCluster, key, &clusters, 2000, 1, nil); n == 0 {
			return
		}
		clist := []string{}
		for _, cluster := range clusters {
			clist = append(clist, cluster.ID.Hex())
		}
		skey["cluster"] = common.V{"$in": clist}
	case "clusters":
		skey["cluster"] = common.V{"$in": list}
	case "apps":
		key := make([]bson.ObjectId, len(list))
		for i, id := range list {
			key[i] = bson.ObjectIdHex(id)
		}
		skey["_id"] = common.V{"$in": key}
	}
	rApps := []MgoApp{}
	rnum, _ := econf.Mgo.Query(mgoApps, skey, &rApps, 2000, 1, common.V{"environment": 1})
	if rnum == 0 {
		return
	}
	for _, r := range rApps {
		if curenv, ok := r.Environment[cdenv]; ok {
			for _, h := range curenv.Hosts {
				hostlist = append(hostlist, h)
			}
		}
	}
	//对hostlist进行去重操作。
	return
}

func mgoRecoder(key interface{}, mgocoll string, econf *common.EopsConf) interface{} {
	var r interface{}
	switch mgocoll {
	case mgoApps:
		r = &MgoApp{}
	case mgoAppTpl:
		r = &MgoAppTPL{}
	case mgoProject:
		r = &MgoProject{}
	case mgoCluster:
		r = &MgoCluster{}
	case mgoResource:
		r = &MgoResource{}
	case mgoAppConf:
		r = &MgoAppConf{}
	}
	if _, err := econf.Mgo.Query(mgocoll, key, r, 0, 0, nil); err != nil {
		econf.LogWarning("[Job] Not found job recoder %v with: %v\n", key, err)
	}
	return r
	//return nil
}

func apptplUsage(econf *common.EopsConf, tplid string, step int) error {
	skey := common.V{"_id": bson.ObjectIdHex(tplid)}
	upop := common.V{"$inc": common.V{"usage": step}}
	c := econf.Mgo.Coll(mgoAppTpl)
	err := c.Update(skey, upop)
	return err
}

func resetResource(econf *common.EopsConf, appid, env string, rlist []string) error {
	appinfo := mgoRecoder(bson.ObjectIdHex(appid), mgoApps, econf).(*MgoApp)
	if appinfo.ID.Valid() {
		idlist := []bson.ObjectId{}
		if envInfo := appinfo.Environment[env]; envInfo != nil {
			if len(appinfo.Environment[env].Hosts) == 0 {
				return nil
			}
			for _, k := range appinfo.Environment[env].Hosts {
				idlist = append(idlist, bson.ObjectIdHex(k))
			}
			skey := common.V{"_id": common.V{"$in": idlist}, "usage": common.V{"$gt": 0}}
			upaction := common.V{"$inc": common.V{"usage": -1}}
			c := econf.Mgo.Coll(mgoResource)
			if _, err := c.UpdateAll(skey, upaction); err != nil {
				return err
			}
		}
	}
	return nil
}

func appVar(appvar common.V, id string, ctx *httpsvr.Context) error {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	workenv := ctx.Keys["workenv"].(string)
	data := common.V{}
	unsetvar := common.V{}
	vars := appvar.VMap("change")
	for k, v := range vars {
		key := fmt.Sprintf("environment.%s.variables.%s", workenv, k)
		data[key] = v
	}
	//只有在正式环境中，才可以添加或删除环境变量。同时将变动同步到其他环境
	if workenv == econf.GOPT.Environments[econf.GOPT.EnvironmentsLen-1] {
		vars = appvar.VMap("add")
		for _, env := range econf.GOPT.Environments {
			for k, v := range vars {
				key := fmt.Sprintf("environment.%s.variables.%s", env, k)
				data[key] = v
			}
		}
		vars = appvar.VMap("del")
		for _, env := range econf.GOPT.Environments {
			for k := range vars {
				key := fmt.Sprintf("environment.%s.variables.%s", env, k)
				unsetvar[key] = ""
			}
		}
	}
	skey := common.V{"_id": bson.ObjectIdHex(id)}
	c := econf.Mgo.Coll(mgoApps)
	err := c.Update(skey, common.V{"$set": data})
	if err != nil {
		econf.LogCritical("[CMDB] update appvar failed: %v\n", err)
		return err
	}
	if len(unsetvar) > 0 {
		err = c.Update(skey, common.V{"$unset": unsetvar})
		if err != nil {
			econf.LogCritical("[CMDB] update appvar failed: %v\n", err)
			return err
		}
	}

	return err
}
