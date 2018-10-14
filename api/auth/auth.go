package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap"
	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/common"
	"opsv2.cn/opsv2/api/httpsvr"
)

const (
	mgoRoles   = "roles"
	mgoUsers   = "users"
	admin      = "admin"
	loginTimes = 3
	//APIPermNoCheck 不进行权限检查（不用登录）
	APIPermNoCheck = 0
	//APIPermVerified 判断条件：已登录用户
	APIPermVerified = 1
	//APIPermAllow 判断条件：MgoRole.API[url|req.Method] == true
	APIPermAllow = 2 //api
	//APIPermJobRun 作业执行权限，也包括了读权限。
	/*
		判断条件:
		1. 必须有APIPermCDEnv权限
		2. MgoRole.Environment[cdenv].Jobs[name] & APIPermJobRun == APIPermJobRun
		3. 根据job运行的目标Target判断：MgoRole.Environment[cdenv].Apps[name] & APIPermJobRun == APIPermJobRun
		name是作业路径或作业流名称
	*/
	APIPermJobRun = 4 //jobrun,jobread
	//APIPermPCADeploy 部署权限。需要在deploy包内检查
	/*
		判断名称为：apptpl-name的作业流的执行权限
	*/
	APIPermPCADeploy = 8 //deploy
	//APIPermPCAConfig 应用的自定义配置的修改权限，包括应用的变量
	APIPermPCAConfig = 16 //modify config
	//APIPermPCARead 应用的自定义配置的查看权限，包括应用的变量
	APIPermPCARead = 32 //read config
	//APIPermCDEnv 部署环境的权限
	APIPermCDEnv = 64
	//APIPermDenied 用于黑名单。比如允许某个目录下所有作业都可以执行，但个别作业不允许。
	APIPermDenied = 128
	//APIPermJobOps 作业运维，可以分配作业执行权限给其他用户
	APIPermJobOps = 256
	//APIPermPCAOps 业务运维，可以分配业务项的配置权限和部署权限。
	APIPermPCAOps = 512
	// APIPermJob       = APIPermAllow | APIPermJobRun | APIPermJobRead | APIPermJobModify
	// APIPermJobRuns   = APIPermAllow | APIPermJobRun | APIPermJobRead
	// APIPermPCA       = APIPermAllow | APIPermPCADeploy | APIPermPCAConfig | APIPermPCARead | APIPermPCAOps
	// APIPermPCADev    = APIPermAllow | APIPermPCADeploy | APIPermPCAConfig | APIPermPCARead
)

var Perms = []string{
	"APIPermVerified",
	"APIPermAllow",
	"APIPermJobRun",
	"APIPermPCADeploy",
	"APIPermPCAConfig",
	"APIPermPCARead",
	"APIPermCDEnv",
	"APIPermDenied",
	"APIPermJobOps",
	"APIPermPCAOps",
}

//RouteList auth的api路由表
func RouteList(route *common.APIRoute) {
	route.API["/auth/user|POST"] = common.APIInfo{Handler: AddUser, Name: "添加用户", Perm: APIPermAllow}
	route.API["/auth/user|PUT"] = common.APIInfo{Handler: UpdateUser, Name: "修改用户", Perm: APIPermAllow}
	route.API["/auth/user|GET"] = common.APIInfo{Handler: Users, Name: "查看用户", Perm: APIPermAllow}
	route.API["/auth/user|DELETE"] = common.APIInfo{Handler: DeleteUser, Name: "删除用户", Perm: APIPermAllow}
	route.API["/auth/login|GET"] = common.APIInfo{Handler: Login, Name: "用户登录", Perm: APIPermNoCheck}
	route.API["/auth/pwd|PUT"] = common.APIInfo{Handler: UpdateUserPWD, Name: "修改用户密码", Perm: APIPermAllow}
	route.API["/auth/roles|GET"] = common.APIInfo{Handler: ListRoles, Name: "获取角色列表", Perm: APIPermAllow}
	route.API["/auth/disableuser|PUT"] = common.APIInfo{Handler: DisableUser, Name: "关闭用户", Perm: APIPermAllow}
	route.API["/auth/permissions|GET"] = common.APIInfo{Handler: Permissions, Name: "权限列表", Perm: APIPermAllow}
	route.API["/auth/env|GET"] = common.APIInfo{Handler: CDEnv, Name: "持续部署环境设置", Perm: APIPermVerified}
}

//Permissions 支持权限列表
func Permissions(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	perm := make(map[string]string)
	c := 0
	for k, v := range econf.Route.API {
		// if v.Perm > 0 {
		perm[k] = v.Name + "," + permShow(v.Perm)
		// }
		c++
	}
	rmsg.DataList = perm
	rmsg.TotalRecord = c
	ctx.Response = rmsg
}

func permShow(p int) string {
	var s string
	if p == 0 {
		return "APIPermNoCheck"
	}
	var i uint
	for i = 0; i < 10; i++ {
		if (1<<i)&p > 0 {
			if s != "" {
				s += "|"
			}
			s += Perms[i]
		}
	}
	return s
}

//AddUser 创建用户
func AddUser(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	userinfo := MgoUser{}
	var err error
	var code int
	_, code, err = ctx.ReqBody(&userinfo)
	if code != 0 {
		econf.LogDebug("[AUTH] URL: %s; code: %d; BodyFormatInvalid: %v", ctx.Req.URL.Path, code, err)
		ctx.Response = code
		return
	}
	/*
	   rver,ok := strconv.ParseFloat(params["rver"], 64)
	   if ok != nil {
	     res.WriteHeader(StatusNotAcceptable)
	     return
	   }
	*/
	if err := common.InputFilter(&userinfo, 1, nil, nil); err != nil {
		econf.LogInfo("info createuser input: %v\n", err)
		ctx.Response = 1101
		return
	}
	userinfo.Passwd = common.CryptSha256(econf.GOPT.PWDSecretKey + userinfo.Passwd)
	c := econf.Mgo.Coll(mgoUsers)
	userinfo.Createtime = time.Now()
	userinfo.Modifytime = time.Now()
	userinfo.KeyID = string(common.RandomStr(16))
	userinfo.Secret = string(common.RandomStr(64))
	if err := c.Insert(userinfo); err != nil {
		errstr := err.Error()
		ctx.Response = 1106
		if errstr[0:6] == "E11000" {
			//用户已存在
			ctx.Response = 1303
		}
		econf.LogCritical("info createuser input: %v\n", err)
		return
	}
	userinfo.Passwd = ""
	cnt, _ := json.Marshal(userinfo)
	ctx.OpLog.Write(cnt)
	return
}

//UpdateUser 更新用户信息
func UpdateUser(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	skey := common.V{}
	postArgs := common.V(ctx.BodyMap)
	var username string
	if username = postArgs.VString("name"); username != "" && common.StrFilter(username, "email") {
		skey["name"] = username
	} else {
		ctx.Response = 1207
		return
	}
	user := ctx.Keys["username"].(string)
	userinfo, _ := GetUserInfo(econf, user, 1)
	if username != user && userinfo.Roles[0] != admin {
		ctx.Response = 1107
		return
	}
	delete(postArgs, "name")
	delete(postArgs, "_id")
	delete(postArgs, "keyid")
	delete(postArgs, "secret")
	data := common.V{}
	userinfo1 := MgoUser{}
	if err := common.InputFilter(&userinfo1, 2, postArgs, data); err != nil {
		econf.LogInfo("info createuser input: %v\n", err)
		ctx.Response = 1101
		return
	}
	// if userinfo.Passwd != "" {
	//   stjson["passwd"] = common.CryptSha256(userinfo.Passwd)
	// }
	data["Modifytime"] = time.Now()
	if err := econf.Mgo.Update(mgoUsers, skey, common.V{"$set": data}); err != nil {
		econf.LogCritical("err mgdb user update: %v\n", err)
		ctx.Response = 1302
	}
	ctx.OpLog.Write(common.V{"skey": skey, "data": data}.Json(""))
	return
}

//UpdateUserPWD 更新用户信息
func UpdateUserPWD(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	postArgs := common.V(ctx.BodyMap)
	var name string
	skey := common.V{}
	data := common.V{}
	if name = postArgs.VString("name"); name != "" && common.StrFilter(name, "email") {
		skey["name"] = name
	} else {
		ctx.Response = 1207
		return
	}
	user := ctx.Keys["username"].(string)
	userinfo, _ := GetUserInfo(econf, user, 1)
	if name != user && userinfo.Roles[0] != admin {
		ctx.Response = 1107
		return
	}
	if econf.GOPT.LdapHost != "" {
		if LdapChangePW(userinfo.LdapDN, postArgs["oldpasswd"].(string), postArgs["newpasswd"].(string), econf) == false {
			ctx.Response = 1109
		}
		return
	}
	if userinfo.Roles[0] != admin {
		skey["passwd"] = common.CryptSha256(econf.GOPT.PWDSecretKey + postArgs["oldpasswd"].(string))
	}
	data["passwd"] = common.CryptSha256(econf.GOPT.PWDSecretKey + postArgs["newpasswd"].(string))
	data["modifytime"] = time.Now()
	if err := econf.Mgo.Update(mgoUsers, skey, common.V{"$set": data}); err != nil {
		econf.LogCritical("err mgdb user update: %v\n", err)
		ctx.Response = 1302
	}
	return
}

//DeleteUser 更新用户信息
func DeleteUser(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	name := ctx.Req.Form.Get("name")
	skey := common.V{}
	if name != "" && common.StrFilter(name, "email") {
		skey["name"] = name
	} else {
		ctx.Response = 1207
		return
	}
	c := econf.Mgo.Coll(mgoUsers)
	if err := c.Remove(skey); err != nil {
		econf.LogCritical("err mgdb user update: %v\n", err)
		ctx.Response = 1302
	}
	ctx.OpLog.Write(skey.Json(""))
	return
}

//DisableUser 更新用户信息
func DisableUser(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	name := ctx.Req.Form.Get("name")
	skey := common.V{}
	if name != "" && common.StrFilter(name, "email") {
		skey["name"] = name
	} else {
		ctx.Response = 1207
		return
	}
	data := common.V{"passwd": ""}
	if err := econf.Mgo.Update(mgoUsers, skey, common.V{"$set": data}); err != nil {
		econf.LogCritical("err mgdb user update: %v\n", err)
		ctx.Response = 1302
	}
	ctx.OpLog.Write(skey.Json(""))
	return
}

//Users 获取用户列表或用户信息
/*
参数格式：
{
  "pageNo":int        //当前页数
  "pageSize":int      //每页记录数
}
输出格式：
{
  "code":int           //状态码，0代表成功，1代表没有登陆，2代表没有权限，3代表其他
  "dataList":[]        //数据列表，
  "totalRecord":int    //总记录数
  "message":str        //错误信息
}
*/
func Users(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	skey := common.V{}
	rmsg := common.NewRespMsg()
	fields := common.V{"name": 1, "fullname": 1, "mobile": 1, "roles": 1}
	user := ctx.Keys["username"].(string)
	userinfo, _ := GetUserInfo(econf, user, 1)
	pageNo := 1
	pageSize := 2000
	for k, v := range ctx.Req.Form {
		value := v[0]
		switch k {
		case "id":
			if !common.StrFilter(value, "id") {
				ctx.Response = 1207
				return
			}
			skey["_id"] = bson.ObjectIdHex(value)
		case "pageNo":
			pageNo, _ = strconv.Atoi(value)
		case "pageSize":
			pageSize, _ = strconv.Atoi(value)
		case "name":
			if !common.StrFilter(value, "email") {
				ctx.Response = 1207
				return
			}
			skey["name"] = value
		case "slist":
			fields = common.V{"name": 1}
		}
		if pageSize > common.MaxPageSize {
			pageSize = common.MaxPageSize
		}
	}
	//只有管理员或具有删除用户权限的用户可以查看用户列表，否则只能查看用户自己的信息
	if userinfo.Roles[0] != admin && userinfo.PermissionList.API["/auth/user|DELETE"]&APIPermAllow != APIPermAllow {
		skey["passwd"] = common.V{"$ne": ""}
		skey["name"] = user
	}

	var result []common.V
	recnum, err := econf.Mgo.Query(mgoUsers, skey, &result, pageSize, pageNo, fields)
	rmsg.TotalRecord = recnum
	if err != nil {
		ctx.Response = 1301
	} else {
		rmsg.DataList = result
	}
	ctx.Response = rmsg
	ctx.OpLog.Write(skey.Json("skey"))
	return
}

//Login 用户登录
/*
参数格式：
{
  "name":str        //用户名
  "passwd":str      //密码
}
输出格式：
{
  "code":int           //状态码，0代表成功，1代表没有登陆，2代表没有权限，3代表其他
  "user":[]        //数据列表，
  "previlagelist":
  "totalRecord":int    //总记录数
  "message":str        //错误信息
}
*/
func Login(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	//获取请求参数，并转存到UserCollPwd
	rmsg := common.NewRespMsg()
	//filArgs := common.V{}
	addr := ctx.RemoteAddr(false)
	var err error
	if limit, _ := econf.MapCache.Get(addr); limit != nil {
		if limit.(uint64) > loginTimes {
			ctx.Response = 1108
			return
		}
	}
	if err = ctx.Req.ParseForm(); err != nil {
		ctx.Response = 1103
		return
	}
	//var username,passwd string
	username := ctx.Req.Form.Get("name")
	passwd := ctx.Req.Form.Get("passwd")
	if !common.StrFilter(username, "email") || !common.StrFilter(passwd, "safestr") {
		ctx.Response = 1103
		return
	}
	var userdn string
	loginOK := true
	userinfo, pwd := GetUserInfo(econf, username, 0)
	if userinfo != nil {
		userdn = userinfo.LdapDN
	}
	if econf.GOPT.LdapHost != "" {
		userdn = LdapAuth(username, userdn, passwd, econf)
		if userdn == "no" {
			loginOK = false
		} else {
			fname := ""
			if userdn[:4] == "uid=" {
				if i := strings.IndexByte(userdn, ','); i >= 0 {
					fname = userdn[4:i]
				}
			}
			// update user's dn to ldap
			if userinfo != nil && userinfo.LdapDN == "" {
				if err = econf.Mgo.Update(mgoUsers, common.V{"name": username}, common.V{"$set": common.V{"ldapdn": userdn, "fullname": fname}}); err != nil {
					econf.LogWarning("update userdn %s failed: %v\n", userdn, err)
				}
			} else if userinfo == nil {
				c := econf.Mgo.Coll(mgoUsers)
				if err = c.Insert(common.V{"name": username, "ldapdn": userdn, "roles": []string{"default"}, "fullname": fname}); err != nil {
					econf.LogWarning("insert userdn %s failed: %v\n", userdn, err)
				}
			}
		}
	} else {
		if pwd != common.CryptSha256(econf.GOPT.PWDSecretKey+passwd) {
			loginOK = false
		}
	}
	if loginOK == false {
		c, _ := econf.MapCache.Increment(addr, uint64(1))
		if c == 0 {
			econf.MapCache.Set(addr, uint64(1), time.Duration(900)*time.Second)
		}
		rmsg.Code = 1103
		rmsg.Message = fmt.Sprintf("登录失败。还允许重试%d次", loginTimes-c)
		ctx.Response = rmsg
		// ctx.Response = 1103
		return
	}

	//登陆成功，更新登陆时间
	skey := common.V{"name": username}
	// data := common.V{"logintime": time.Now()}
	if err = econf.Mgo.Update(mgoUsers, skey, common.V{"$currentDate": common.V{"logintime": true}}); err != nil {
		// if err = econf.Mgo.Update(mgoUsers, skey, common.V{"$set": data})
		econf.LogWarning("update login time failed: %v\n", err)
		// ctx.Response = 1103
		// return
	}
	ctx.Keys["username"] = username
	cookies := Secookies(ctx)
	for _, cookie := range cookies {
		http.SetCookie(ctx.Writer, cookie)
	}
	rmsg.DataList = userinfo
	rmsg.TotalRecord = 1
	ctx.Response = rmsg
	/*
	   "cmdb:manage","cmdb:devices:list","cmdb:apps:list","cmdb:business:list"}
	*/
	return
}

func PermJobs(ctx *httpsvr.Context) {
	// econf := ctx.Keys["econf"].(*common.EopsConf)

}

func PermBusiness(ctx *httpsvr.Context) {

}

//ListRoles 获取Roles列表
func ListRoles(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	rinfo, err := econf.MapCache.Get("roles")
	if err != nil {
		MDBRoles(econf)
		rinfo, _ = econf.MapCache.Get("roles")
	}
	rmsg.DataList = rinfo.([]string)
	ctx.Response = rmsg
}

//CDEnv 获取部署环境,或设置当前工作的部署环境
func CDEnv(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	rmsg := common.NewRespMsg()
	if env := ctx.Req.Form.Get("env"); env != "" {
		if common.SearchStr(econf.GOPT.Environments, env) == false {
			ctx.Response = 1207
		} else {
			user := ctx.Keys["username"].(string)
			userinfo, _ := GetUserInfo(econf, user, 1)
			if userinfo == nil {
				ctx.Response = 1199
				return
			}
			userinfo.WorkEnv = env
			ctx.Keys["workenv"] = env
			c := econf.Mgo.Coll(mgoUsers)
			if err := c.Update(common.V{"name": user}, &common.V{"$set": common.V{"workenv": env}}); err != nil {
				ctx.Response = 1302
			}
		}
	} else {
		rmsg.DataList = econf.GOPT.Environments
		rmsg.TotalRecord = len(econf.GOPT.Environments)
	}
	ctx.Response = rmsg
}

//Privilege 校验
func Privilege(ctx *httpsvr.Context) {
	var err error
	ctx.Response = 1107
	ctx.Req.Form, err = url.ParseQuery(ctx.Req.URL.RawQuery)
	if err != nil {
		ctx.Response = 1204
		return
	}
	econf := ctx.Keys["econf"].(*common.EopsConf)
	clip := ctx.RemoteAddr(false)
	userinfo := &ZDUserSession{}
	var sessionLife int64
	sid, errSid := ctx.Req.Cookie("sessid")
	uid, erruid := ctx.Req.Cookie("uid")
	tid, errTid := ctx.Req.Cookie("tid")
	if errSid != nil || erruid != nil || errTid != nil {
		if clip != "127.0.0.1" {
			if econf.GOPT.IsRelease() {
				if !econf.GOPT.IPWhiteListForAPI.Check(clip) {
					econf.LogDebug("[CheckPrevilage] client not in whitelist: %s\n", clip)
					return
				}
			}
		}
		keyid := ctx.Req.Form.Get("keyid")
		secret := ctx.Req.Form.Get("secret")
		ctx.Req.Form.Del("keyid")
		ctx.Req.Form.Del("secret")
		if keyid == "" || secret == "" {
			return
		}
		userinfo, _ = GetUserInfo(econf, keyid, 2)
		if userinfo == nil || userinfo.Secret != secret {
			econf.LogDebug("[CheckPrevilage] invalid keyid: %s.%v\n", keyid, userinfo)
			return
		}
	} else {
		if common.CryptSha256(uid.Value+econf.GOPT.SessionSecretKey+clip+tid.Value) != sid.Value {
			ctx.Response = 1199
			return
		}
		//校验cookie是否过期
		tt := time.Now()
		oldtt, err := strconv.ParseInt(tid.Value, 10, 64)
		if err != nil {
			return
		}
		verifytype := 1
		if tt.Unix()-oldtt < econf.GOPT.SessionExpire {
			verifytype = 3
		}
		var status string
		userinfo, status = GetUserInfo(econf, uid.Value, verifytype)
		if status == common.Expired {
			ctx.Response = 1199
			return
		}
		if userinfo == nil {
			return
		}
		sessionLife = tt.Unix() - oldtt
	}
	ctx.Keys["username"] = userinfo.Name
	ctx.Keys["workenv"] = userinfo.WorkEnv
	if sessionLife > 0 && sessionLife >= econf.GOPT.SessionExpire {
		Secookies(ctx)
	}
	px := econf.GOPT.PrefixLen
	api := ctx.Req.URL.Path[px:]
	permkey := api + "|" + ctx.Req.Method
	ctx.Keys["api"] = permkey
	permList := userinfo.PermissionList
	if userinfo.MgoUser.Roles[0] == admin || econf.Route.API[permkey].Perm == APIPermVerified {
		ctx.Response = nil
		// return
	} else if econf.Route.API[permkey].Perm&APIPermAllow == APIPermAllow &&
		permList.API[permkey]&APIPermAllow == APIPermAllow {
		if econf.Route.API[permkey].Perm&APIPermCDEnv == APIPermCDEnv {
			env := userinfo.WorkEnv
			if env == "" {
				env = ctx.Req.Form.Get("cdenv")
			}
			if env == "" || permList.Environment[env].API[permkey]&APIPermAllow != APIPermAllow {
				return
			}
		}
	} else {
		return
	}
	cntType := ctx.Req.Header.Get("Content-Type")
	if (ctx.Req.Method == "PUT" || ctx.Req.Method == "DELETE") && strings.Index(cntType, "application/json") >= 0 {
		postArgs := common.V{}
		var code int
		var err error
		if ctx.Body, code, err = ctx.ReqBody(&postArgs); code != 0 && code != 1202 {
			econf.LogDebug("URL: %s; code: %d; BodyFormatInvalid: %v", ctx.Req.URL.Path, code, err)
			ctx.Response = code
			// return
		} else {
			delete(postArgs, "operator")
		}
		ctx.BodyMap = postArgs
	}
	// ctx.Response = nil
	return
}

//IsAdmin 是否是管理员
func IsAdmin(ctx *httpsvr.Context) bool {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	if userInfo, _ := GetUserInfo(econf, ctx.Keys["username"].(string), 3); userInfo != nil {
		return userInfo.Roles[0] == admin
	}
	return false
}

func PermChkCDEnv(env string, ctx *httpsvr.Context) bool {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	userInfo, _ := GetUserInfo(econf, ctx.Keys["username"].(string), 3)
	if userInfo == nil {
		return false
	}
	if userInfo.Roles[0] == admin {
		return true
	}
	if curEnv, ok := userInfo.PermissionList.Environment[env]; ok {
		api := ctx.Keys["api"].(string)
		perm := econf.Route.API[api].Perm
		if curEnv.API[api]&perm == perm {
			return true
		}
	}
	return false
}

//PermChkPCAConfig 检查应用的配置权限
func PermChkPCAConfig(pid, cid, aid string, perm int, ctx *httpsvr.Context) bool {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	userInfo, _ := GetUserInfo(econf, ctx.Keys["username"].(string), 3)
	if userInfo == nil {
		return false
	}
	if userInfo.Roles[0] == admin {
		return true
	}
	if curEnv, ok := userInfo.PermissionList.Environment[userInfo.WorkEnv]; ok {
		if userPerm, ok1 := curEnv.App[aid]; ok1 && userPerm&perm == perm {
			return true
		}
		if userPerm, ok1 := curEnv.Cluster[cid]; ok1 && userPerm&perm == perm {
			return true
		}
		if userPerm, ok1 := curEnv.Project[pid]; ok1 && userPerm&perm == perm {
			return true
		}
	}

	return false
}

//PermChkJobTarget 检查作业执行的目标是否允许。执行的目标只能是app
func PermChkJobTarget(target string, list []string, ctx *httpsvr.Context) bool {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	userInfo, _ := GetUserInfo(econf, ctx.Keys["username"].(string), 3)
	if userInfo == nil {
		return false
	}
	if userInfo.Roles[0] == admin {
		return true
	}
	if target != "apps" {
		return false
	}
	if curEnv, ok := userInfo.PermissionList.Environment[userInfo.WorkEnv]; ok {
		for _, item := range list {
			if curEnv.App[item]&APIPermJobRun != APIPermJobRun {
				return false
			}
		}
		return true
	}
	return false
}

//PermChkJob 检查作业的执行权限
func PermChkJob(key string, perm int, ctx *httpsvr.Context) bool {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	if perm&APIPermJobRun != APIPermJobRun {
		return true
	}
	userInfo, _ := GetUserInfo(econf, ctx.Keys["username"].(string), 3)
	if userInfo == nil {
		return false
	}
	api := ctx.Keys["api"].(string)
	//apptpl-是应用模板自动生成的工作流，此处不判断权限，依靠业务的部署权限控制。
	if userInfo.Roles[0] == admin ||
		userInfo.PermissionList.API[api]&APIPermJobOps == APIPermJobOps ||
		key[:7] == "apptpl-" {
		return true
	}

	if curEnv, ok := userInfo.PermissionList.Environment[userInfo.WorkEnv]; ok {
		if !rPath(key, curEnv.Job, APIPermDenied) && rPath(key, curEnv.Job, APIPermJobRun) {
			return true
		}
	}
	return false
}

//LdapChangePW 修改ldap的密码
func LdapChangePW(userdn, passwd, newpasswd string, econf *common.EopsConf) bool {
	var l *ldap.Conn
	var err error
	ldaphost := fmt.Sprintf("%s:%s", econf.GOPT.LdapHost, econf.GOPT.LdapPort)
	if econf.GOPT.LdapTLS == "yes" {
		l, err = ldap.DialTLS("tcp", ldaphost, &tls.Config{InsecureSkipVerify: true})
	} else {
		l, err = ldap.Dial("tcp", ldaphost)
	}
	if err != nil {
		econf.LogCritical("Cant connect ldap: %v\n", err)
		return false
	}
	defer l.Close()
	err = l.Bind(userdn, passwd)
	if err != nil {
		econf.LogInfo("%s bind failed: %v\n", userdn, err)
		return false
	}
	passwordModifyRequest := ldap.NewPasswordModifyRequest("", passwd, newpasswd)
	_, err = l.PasswordModify(passwordModifyRequest)
	if err != nil {
		econf.LogInfo("%s change passwd failed: %v\n", userdn, err)
		return false
	}
	return true
}

//LdapAuth ldap认证登陆
func LdapAuth(name, userdn, passwd string, econf *common.EopsConf) string {
	var l *ldap.Conn
	var err error
	result := "no"
	ldaphost := fmt.Sprintf("%s:%s", econf.GOPT.LdapHost, econf.GOPT.LdapPort)
	if econf.GOPT.LdapTLS == "yes" {
		l, err = ldap.DialTLS("tcp", ldaphost, &tls.Config{InsecureSkipVerify: true})
	} else {
		l, err = ldap.Dial("tcp", ldaphost)
	}
	if err != nil {
		econf.LogCritical("Cant connect ldap: %v\n", err)
		return result
	}
	defer l.Close()
	if userdn != "" {
		err = l.Bind(userdn, passwd)
		if err != nil {
			econf.LogInfo("%s login failed: %v\n", userdn, err)
			return result
		}
		return userdn
	}
	// Reconnect with TLS
	// err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// First bind with a read only user
	err = l.Bind(econf.GOPT.LdapBindDn, econf.GOPT.LdapBindPasswd)
	if err != nil {
		econf.LogCritical("Cant bind ldap: %v\n", err)
		return result
	}
	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		econf.GOPT.LdapDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(mail=%s))", name),
		[]string{"dn", "cn", "mail"},
		nil,
	)
	var sr *ldap.SearchResult
	sr, err = l.Search(searchRequest)
	if err != nil {
		econf.LogInfo("search ladp failed: %s\n", err)
		return result
	}

	if len(sr.Entries) != 1 {
		econf.LogInfo("User does not exist or too many entries returned: %s\n", name)
		return result
	}
	err = l.Bind(sr.Entries[0].DN, passwd)
	if err != nil {
		econf.LogInfo("login failed: %v\n", err)
		return result
	}
	return sr.Entries[0].DN
}

func rPath(key string, jobPerms map[string]int, perm int) bool {
	var skey string
	var skeylen int
	if jobPerms["all"]&perm == perm {
		return true
	}
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
		if jobPerms[skey]&perm == perm {
			return true
		}
		if i < 0 {
			break
		}
	}
	return false
}

//SSHAuth
// func SSHAuth() martini.Handler {
// 	return func(res http.ResponseWriter, req *http.Request,
// 		logger *log.Logger, econf *common.EopsConf, params martini.Params) {
// 		authobj := make(map[string]string)
// 		tnow := time.Now()
// 		tnowSecond := tnow.Unix() * 1000
// 		tnowMilli := int64(tnow.Nanosecond() / 1000000)
// 		userSecret := "OTFkYTMzZWY1OTE5NDEzOGI3ZjcyMDczOGYwYjkzMWRkY"
// 		authobj["upn"] = "dale"
// 		authobj["signature_method"] = "HMAC-SHA256"
// 		authobj["timestamp"] = strconv.FormatInt(tnowSecond+tnowMilli, 10)
// 		//authobj["timestamp"] = "1495683487432"
// 		authobj["api_key"] = "NGZkZWY1YzY0NDU4NGUyNzg2YmUwMzhiYjQ4MGQ1NzA0Y"
// 		authobj["api_version"] = "1.0"
// 		mac := hmac.New(sha256.New, []byte(userSecret))
// 		mac.Write([]byte(authobj["api_key"] + authobj["upn"] + authobj["timestamp"]))
// 		secstr := mac.Sum(nil)
// 		//将byte数组转换为字符串
// 		signature := ""
// 		for _, v := range secstr {
// 			vv := int64(v)
// 			z := ""
// 			if vv < 16 {
// 				z = "0"
// 			}
// 			signature = signature + z + strconv.FormatInt(int64(v), 16)
// 		}
// 		authobj["signature"] = signature
// 		authstr, _ := json.Marshal(authobj)
// 		res.Write(authstr)
// 	}
// }

//Secookies  生成安全cookie，作为sessionkey使用
func Secookies(ctx *httpsvr.Context) (cookies []*http.Cookie) {
	econf := ctx.Keys["econf"].(*common.EopsConf)
	clip := ctx.RemoteAddr(false)
	tt := time.Now()
	ckmap := make(map[string]string)
	tid := strconv.FormatInt(tt.Unix(), 10)
	name := ctx.Keys["username"].(string)
	ckmap["tid"] = tid
	ckmap["uid"] = name
	ckmap["sessid"] = common.CryptSha256(name + econf.GOPT.SessionSecretKey + clip + tid)
	for k, v := range ckmap {
		cookies = append(cookies, &http.Cookie{
			Name:     k,
			Value:    v,
			Domain:   ctx.Req.Host,
			Path:     "/",
			Secure:   false,
			HttpOnly: false,
		})
	}
	return
}
