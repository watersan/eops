package common

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/cache"
	"opsv2.cn/opsv2/api/httpsvr"
	"opsv2.cn/opsv2/api/options"
)

//RespMsg 响应给客户端的数据结构
type RespMsg struct {
	Code        int         `json:"code"`
	DataList    interface{} `json:"dataList"`
	TotalRecord int         `json:"totalRecord"`
	Message     string      `json:"message"`
}

//ErrCode 错误信息定义
type ErrCode struct {
	ErrNo map[int]string
}

//Logger 日志
type Logger struct {
	Log   *log.Logger
	Level int
}

type MgoDB struct{ *mgo.Database }

// EopsConf 全局Auth模块的结构
type EopsConf struct {
	MapCache  *cache.MapCache
	Cache     *cache.RedisCache
	Db        *mgo.Database
	Mgo       MgoDB
	Err       *ErrCode
	Route     *APIRoute
	GOPT      options.GlobalOPT
	AccessLog *log.Logger
	Mutex     sync.Mutex
	//JobsQueue chan *JobExecutor
	Logger
}

//APIRoute API的路由表
type APIRoute struct {
	API map[string]APIInfo
}

//APIInfo api信息
type APIInfo struct {
	Handler httpsvr.Handle
	Name    string
	/*Perm 权限
	0x0 no perm
	0x1 api
	0x2 jobrun
	0x4 jobread
	0x8 jobmodify
	//project 权限，包括集群和应用
	0x16 deploy
	0x32 config
	0x64 read
	0x128 ops
	*/
	Perm int
}

//OpLog  操作记录
type OpLog struct {
	Time    time.Time
	User    string
	IP      string
	Op      string
	Content string
	Out     *RespMsg
}

//HideType 隐藏属性定义
type HideType struct {
	Operator     string    `json:"operator"`
	Creater      string    `json:"creater"`
	CreatTime    time.Time `json:"createtime"`
	Version      int       `json:"version"`
	Reviewer     string    `json:"reviewer"`
	ReviewResult byte      `json:"reviewresult"`
	/*
		0: 未部署
		1: 已部署
		11: 部署失败
	*/
	DeployStatus int `json:"deploystatus"`
}

//GitFile gitlab api 获取仓库文件内容的结构
type GitFile struct {
	FileName     string `json:"file_name"`
	FilePath     string `json:"file_path"`
	Size         int    `json:"size"`
	Encoding     string `json:"encoding"`
	Content      string `json:"content"`
	Ref          string `json:"ref"`
	BlobID       string `json:"blob_id"`
	CommitID     string `json:"commit_id"`
	LastCommitID string `json:"last_commit_id"`
}

//GitAction git操作相关的结构
type GitAction struct {
	Action   string
	Method   string
	User     string
	Ref      string
	Url      string
	GitToken string
	Projects string
	Path     string
	Content  string
	Comment  string
	Result   []byte
}

const (
	//ZDModeAdd 操作模式：添加
	ZDModeAdd = 1
	//ZDModeMod 操作模式：修改
	ZDModeMod = 2
	//ZDModePwd 操作模式：修改密码
	ZDModePwd = 4
	//ZDModeLogin 操作模式：登陆
	ZDModeLogin = 8
	//MaxBodySize 允许最大的body大小
	MaxBodySize = int64(10 << 20)
	//MaxPageSize 最大页面记录数
	MaxPageSize = 2000
	//LogCritical 严重
	LogCritical = 1
	//LogWarning 警告
	LogWarning = 2
	//LogInfo 详细信息
	LogInfo = 4
	//LogDebug Debug级别
	LogDebug  = 8
	mdbprefix = "zd_"
	//Expired 已过期的回话
	Expired = "expired"
	//CookieSecret cookie加密秘钥
	CookieSecret = "qTEd5UxA6RCl34P1ZOXdeaRM0Q3tYFOzwCsBH7Je90xjStkenQ80qWK4pueJV8Ot"
	//TrueStr yes
	TrueStr    = "yes"
	ApptplName = "apptpl"
)

var (
	//Aeskey AES秘钥
	Aeskey = []byte(`Ul\^EK8#<(Fh{qaXw@*9bI]$HRpNc]Z)`)
)

//Init 初始化eopsconf.db *mgo.Database, errlog *log.Logger, accesslog *log.Logger
func Init() *EopsConf {
	econf := &EopsConf{
		MapCache: cache.NewMapCache(),
		Cache:    nil,
		// Db:        db,
		// Mgo:       MgoDB{db},
		Err: InitErrCode(),
		// AccessLog: accesslog,
		// Logger.Level: 15,
		// Logger.Log:   errlog,
		GOPT:  options.GOPT(),
		Route: NewRoute(),
	}
	// econf.Level = 15
	// econf.Log = errlog
	return econf
}

//NewRoute 初始化路由表
func NewRoute() *APIRoute {
	apilist := make(map[string]APIInfo)
	route := APIRoute{API: apilist}
	return &route
}

//InitErrCode 初始化错误信息
func InitErrCode() *ErrCode {
	errcode := &ErrCode{}
	errcode.ErrNo = map[int]string{
		//用户登陆类
		1101: "NotLegalInput",
		1102: "WrongPWD",
		1103: "登录失败",
		1104: "NoUser",
		1105: "SeCookieInvalid",
		1106: "RegistUserFailed",
		1107: "PermissionDenied",
		1108: "LimitLogin",
		1109: "ChangePasswordFailed",
		1111: "UserDisabled",
		1199: "LoginTimeout",

		//请求处理类
		1201: "InvalidVer",
		1202: "BodyIsNull",
		1203: "BodyTooLong",
		1204: "BodyInvalid",
		1205: "BodyReadFailed",
		1206: "BodyFormatInvalid",
		1207: "InvalidRequest",

		//mongodb
		1301: "NoRecoder",
		1302: "UpdateFailed",
		1303: "DuplicateEntry",
		1304: "DeleteFailed",
		1305: "InsertFailed",
		1309: "OperationFailed",

		1401: "DoLoggingFailed",

		1501: "AppStatusLocked",

		1601: "GITRequestFailed",
		1610: "InvalidToken",

		1701: "JobRunFailed",

		1800: "AlreadyImplemented",
		1801: "InvalidParameter",
		1802: "AppTemplateInvalid",
		1803: "DeployHistroysFailed",
		1804: "DeployTaskDisabled",

		1901: "SystemError",
	}
	return errcode
}

//NewRespMsg 创建新的响应信息结构
func NewRespMsg() (rmsg *RespMsg) {
	rmsg = &RespMsg{}
	rmsg.Code = 0
	rmsg.Message = "Success"
	return
}

//SetCode 设置错误代码及错误信息
func (rmsg *RespMsg) SetCode(code int, errcode *ErrCode) {
	rmsg.Code = code
	if msg, ok := errcode.ErrNo[code]; ok {
		rmsg.Message = msg
	} else {
		rmsg.Message = "Unkown"
	}
	return
}

//Bytes 转换为byte
func (rmsg *RespMsg) Bytes() (rstr []byte) {
	rstr, _ = json.Marshal(rmsg)
	return
}

//Strings 转换为string
func (rmsg *RespMsg) Strings() (rstr string) {
	r, _ := json.Marshal(rmsg)
	return string(r)
}

//LogInfo 日志处理
// func (econf *EopsConf) LogInfo(format string, v ...interface{}) {
// 	if econf.Level&LogInfo == LogInfo {
// 		econf.Log.Printf(format, v...)
// 	}
// }

//LogDebug 日志处理
// func (econf *EopsConf) LogDebug(format string, v ...interface{}) {
// 	if econf.Level&LogDebug == LogDebug {
// 		econf.Log.Printf(format, v...)
// 	}
// }

//LogWarning 日志处理
// func (econf *EopsConf) LogWarning(format string, v ...interface{}) {
// 	if econf.Level&LogWarning == LogWarning {
// 		econf.Log.Printf(format, v...)
// 	}
// }

//LogCritical 日志处理
// func (econf *EopsConf) LogCritical(format string, v ...interface{}) {
// 	if econf.Level&LogCritical == LogCritical {
// 		econf.Log.Printf(format, v...)
// 	}
// }

/*
	用户输入判断：debug
	io： info

*/
//LogInfo 日志处理
func (logger *Logger) LogInfo(format string, v ...interface{}) {
	if logger.Level&LogInfo == LogInfo {
		logger.Log.Printf("[Info] "+format, v...)
	}
}

//LogDebug 日志处理
func (logger *Logger) LogDebug(format string, v ...interface{}) {
	if logger.Level&LogDebug == LogDebug {
		// pc, _, line, _ := runtime.Caller(1)
		// f := runtime.FuncForPC(pc)
		// vv := []interface{}{f.Name(), line}
		// vv = append(vv, v...)
		// logger.Log.Printf("[Debug] (%s-%d) "+format, vv...)
		logger.Log.Printf("[Debug] "+format, v...)
	}
}
func (logger *Logger) LogDebug1(v interface{}) {
	if logger.Level&LogDebug == LogDebug {
		myref := reflect.TypeOf(v).Elem()
		if myref.Kind().String() == "struct" {
			tmp, _ := json.MarshalIndent(v, "", "  ")
			logger.Log.Printf("[Debug1] %s: %s\n", myref.Name(), string(tmp))
		} else {
			logger.Log.Printf("[Debug1] %s: %v\n", myref.Name(), v)
		}
	}
}

//LogWarning 日志处理
func (logger *Logger) LogWarning(format string, v ...interface{}) {
	if logger.Level&LogWarning == LogWarning {
		// pc, _, line, _ := runtime.Caller(1)
		// f := runtime.FuncForPC(pc)
		// vv := []interface{}{f.Name(), line}
		// vv = append(vv, v...)
		// logger.Log.Printf("[Warin] (%s-%d) "+format, vv...)
		logger.Log.Printf("[Warin] "+format, v...)
	}
}

//LogCritical 日志处理
func (logger *Logger) LogCritical(format string, v ...interface{}) {
	if logger.Level&LogCritical == LogCritical {
		logger.Log.Printf("[Critical] "+format, v...)
	}
}

//SetMdb 获取mdb的Collection对象
func (econf *EopsConf) SetMdb(db *mgo.Database) {
	econf.Mgo = MgoDB{db}
}

//GetColl 获取mdb的Collection对象
func (econf *EopsConf) GetColl(table string) *mgo.Collection {
	return econf.Db.C(mdbprefix + table)
}

//QueryMDB 搜索mongodb
func (econf *EopsConf) QueryMDB(t string, q *V, r *[]V, pageSize, pageNo int, selector V) (int, error) {
	skip := (pageNo - 1) * pageSize
	c := econf.GetColl(t)
	var qSeg *mgo.Query
	if orderby, ok := selector["orderby"]; ok {
		var sort []string
		switch orderby.(type) {
		case string:
			sort = append(sort, orderby.(string))
		case []string:
			sort = append(sort, orderby.([]string)...)
		}
		//sort := orderby.([]string)
		delete(selector, "orderby")
		qSeg = c.Find(*q).Sort(sort...).Select(selector).Skip(skip).Limit(pageSize)
	} else {
		qSeg = c.Find(*q).Select(selector).Skip(skip).Limit(pageSize)
	}
	row := V{}
	rnum, err := qSeg.Count()
	if err != nil {
		return 0, err
	}
	if r == nil {
		return rnum, nil
	}
	qiter := qSeg.Iter()
	for qiter.Next(row) {
		tt := V{}
		t, _ := json.Marshal(row)
		json.Unmarshal(t, &tt)
		*r = append(*r, tt)
	}
	return rnum, nil
}

func (mgodb MgoDB) Coll(table string) *mgo.Collection {
	return mgodb.C(mdbprefix + table)
}

//Update 重新封装的mongodb的Update，屏蔽了对指定status的更新
func (mgodb MgoDB) Update(mgocoll string, skey, data V) error {
	c := mgodb.Coll(mgocoll)
	skey["status"] = V{"$nin": []int{-1000, -2000}}
	err := c.Update(skey, data)
	return err
}

//QueryM 查询可管理的内容。屏蔽status为-1000和-2000的内容
func (mgodb MgoDB) QueryM(mgocoll string, key interface{}, result interface{}, pageSize, pageNo int, selector V) (int, error) {
	var skey V
	switch key.(type) {
	case bson.ObjectId:
		skey = V{"_id": key.(bson.ObjectId)}
	case []bson.ObjectId:
		skey = V{"_id": V{"$in": key.([]bson.ObjectId)}}
	case V:
		skey = key.(V)
	case string:
		skey = V{"name": key.(string)}
	}
	skey["status"] = V{"$nin": []int{-1000, -2000}}
	rnum, err := mgodb.Query(mgocoll, skey, result, pageSize, pageNo, selector)
	return rnum, err
}

//Query 搜索mongodb
func (mgodb MgoDB) Query(mgocoll string, key interface{}, result interface{}, pageSize, pageNo int, selector V) (int, error) {
	var skey V
	switch key.(type) {
	case bson.ObjectId:
		skey = V{"_id": key.(bson.ObjectId)}
	case []bson.ObjectId:
		skey = V{"_id": V{"$in": key.([]bson.ObjectId)}}
	case V:
		skey = key.(V)
	case string:
		skey = V{"name": key.(string)}
	}
	c := mgodb.Coll(mgocoll)
	//	c := mgo.C(mdbprefix + mgocoll)
	qSeg := c.Find(skey)
	if orderby, ok := selector["orderby"]; ok {
		var sort []string
		switch orderby.(type) {
		case string:
			sort = append(sort, orderby.(string))
		case []string:
			sort = append(sort, orderby.([]string)...)
		case []interface{}:
			for _, value := range orderby.([]interface{}) {
				sort = append(sort, value.(string))
			}
		}
		//sort := orderby.([]string)
		delete(selector, "orderby")
		qSeg = qSeg.Sort(sort...).Select(selector)
	} else {
		qSeg = qSeg.Select(selector)
	}
	var rnum int
	var err error
	if pageNo > 0 && pageSize > 0 {
		skip := (pageNo - 1) * pageSize
		qSeg = qSeg.Skip(skip).Limit(pageSize)
		err = qSeg.All(result)
	} else {
		err = qSeg.One(result)
	}
	rnum, err = qSeg.Count()
	return rnum, err
}

//OpLog 记录操作日志
func (econf *EopsConf) OpLog(ctx *httpsvr.Context) {
	user := "unkown"
	if v, ok := ctx.Keys["username"]; ok {
		user = v.(string)
	} else {
		return
	}
	var op string
	if api, ok := ctx.Keys["api"]; ok {
		if r, ok := econf.Route.API[api.(string)]; ok {
			op = r.Name
		}
	}
	rmsg := NewRespMsg()
	switch ctx.Response.(type) {
	case []byte:
		rmsg.Message = string(ctx.Response.([]byte))
	case *RespMsg:
		rmsg = ctx.Response.(*RespMsg)
	case string:
		rmsg.Message = ctx.Response.(string)
	case int:
		code := ctx.Response.(int)
		if code > 0 {
			rmsg.SetCode(ctx.Response.(int), econf.Err)
		}
	default:
		tmp, _ := json.Marshal(ctx.Response)
		rmsg.Message = string(tmp)
	}
	if rmsg.TotalRecord > 0 {
		rmsg.DataList = nil
	}
	loginfo := OpLog{
		Time:    time.Now(),
		User:    user,
		IP:      ctx.RemoteAddr(false),
		Op:      op,
		Content: ctx.OpLog.String(),
		Out:     rmsg,
	}
	c := econf.Mgo.Coll("oplog")
	if err := c.Insert(loginfo); err != nil {
		econf.LogWarning("Oplog Failed with: %v\n", err)
	}
}

//BaseData 获取基础数据
func (econf *EopsConf) BaseData(key, btype string) interface{} {

	memkey := "basedata_" + key + "_" + btype
	if value, err := econf.MapCache.Get(memkey); err == nil {
		return value
	}
	result := []V{}
	skey := V{"type": btype}
	if key != "" {
		skey["key"] = key
	}
	rnum, _ := econf.Mgo.Query("basedata", &skey, &result, 2000, 1, V{})
	if rnum == 0 {
		return nil
	}
	if key != "" {
		econf.MapCache.Set(memkey, result[0]["value"], time.Duration(econf.GOPT.MemExpire)*time.Second)
		return result[0]["value"]
	}
	for _, r := range result {
		//r, ok := tmp.(bson.M)
		econf.MapCache.Set("basedata_"+r["key"].(string)+"_"+btype, r["value"], time.Duration(econf.GOPT.MemExpire)*time.Second)
	}
	return result
}

//CDEnvIndex 指定的部署环境在配置的环境列表中的索引
func (econf *EopsConf) CDEnvIndex(env string) int {
	if env == "" {
		return 0
	}
	for i := 0; i < econf.GOPT.EnvironmentsLen; i++ {
		if econf.GOPT.Environments[i] == env {
			return i
		}
	}
	return -1
}

//InputFilter 对输入的内容进行合规性检查
func InputFilter(v interface{}, mode int, in, out V) (err error) {
	myref := reflect.ValueOf(v)
	if myref.IsValid() == false {
		err = errors.New("the interface is invalid")
		return
	}
	ucref := myref.Elem()
	uctype := ucref.Type()
	for i := 0; i < ucref.NumField(); i++ {
		fieldname := strings.ToLower(uctype.Field(i).Name)
		//跳过ID的检查
		if fieldname == "id" {
			continue
		}
		field := ucref.Field(i)
		fieldtag := uctype.Field(i).Tag
		filter := fieldtag.Get("Filter")
		lmin, _ := strconv.Atoi(fieldtag.Get("Lenmin"))
		lmax, _ := strconv.Atoi(fieldtag.Get("Lenmax"))
		notnull, _ := strconv.Atoi(fieldtag.Get("Notnull"))
		fmode, _ := strconv.Atoi(fieldtag.Get("Mode"))
		fieldtype := uctype.Field(i).Type.String()
		inValue, inOk := in[fieldname]
		//模式不匹配，模式不是Add而in中没有此字段，跳过检查
		if fmode&mode != mode || (mode != ZDModeAdd && inOk == false) {
			continue
		}
		switch fieldtype {
		case "string":
			str := ""
			//fmt.Printf("filter: %v-%v", fieldname, field.IsNil())
			// if mode != ZDModeAdd && vvok == false {
			// 	continue
			// }
			if mode == ZDModeAdd {
				if field.IsValid() == true {
					str = field.String()
				} else if notnull == 1 {
					err = errors.New(fieldname + " is null")
					break
				}
			} else {
				switch inValue.(type) {
				case string:
					str = inValue.(string)
				default:
					err = errors.New(fieldname + " is invalid")
					break
				}
			}
			strlen := len(str)
			if (strlen > 0 || notnull == 1) && lmax > 0 &&
				(strlen < lmin || strlen > lmax || !StrFilter(str, filter)) {
				err = errors.New(fieldname + " is invalid: " + str)
				return
			}
		case "[]string":
			if mode == ZDModeAdd {
				if field.Len() == 0 {
					if notnull == 1 {
						err = errors.New(fieldname + " is null")
						break
					}
					continue
				}
				for i := 0; i < field.Len(); i++ {
					str := field.Index(i).String()
					strlen := len(str)
					if (strlen > 0 || notnull == 1) && lmax > 0 &&
						(strlen < lmin || strlen > lmax || !StrFilter(str, filter)) {
						err = errors.New(fieldname + " is invalid: " + str)
						return
					}
				}
			} else {
				switch inValue.(type) {
				case []interface{}:
					for _, str := range inValue.([]interface{}) {
						strlen := len(str.(string))
						if (strlen > 0 || notnull == 1) && lmax > 0 &&
							(strlen < lmin || strlen > lmax || !StrFilter(str.(string), filter)) {
							err = fmt.Errorf("[Filter] str:%s-%d; lmax: %d; lmin: %d; filter: %s",
								str.(string), strlen, lmax, lmin, filter)
							// err = errors.New(fieldname + " is invalid: " + str.(string))
							return
						}
					}
				default:
					err = errors.New(fieldname + " is invalid")
					break
				}
			}
		}
		if in != nil {
			out[fieldname] = inValue
		}
	}
	// if id := in.ObjType("_id", "string"); id != nil && StrFilter(id.(string), "id") {
	// 	out["_id"] = id
	// }
	return
}

//StrFilter 字符串过滤
func StrFilter(str string, filter string) bool {
	var reg *regexp.Regexp
	switch filter {
	case "nickname":
		reg = regexp.MustCompile(`^[a-z0-9A-Z\p{Han}]+[_\-a-z0-9A-Z\p{Han}]*$`)
	case "username":
		reg = regexp.MustCompile(`^[a-zA-Z]+[_\-\.a-z0-9A-Z]*$`)
	case "email":
		reg = regexp.MustCompile(`^[\w\_\-\.]+@[\w\-\.]+$`)
	case "ip":
		reg = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	case "ipport":
		reg = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:?\d*$`)
	case "number":
		reg = regexp.MustCompile(`^-?\d+$`)
	case "mobile":
		reg = regexp.MustCompile(`^0?1[0|3|4|5|7|8][0-9]\d{8}$`)
	case "port":
		reg = regexp.MustCompile(`^(,?\w*/?(tcp|udp)/\d+)+$`)
	case "scriptname":
		reg = regexp.MustCompile(`^[a-zA-Z][\w|\.|-]{2,}\.(sh|py|pl)$`)
	case "path":
		reg = regexp.MustCompile(`^[_\.\-a-z0-9A-Z\/]+$`)
	case "bytes":
		reg = regexp.MustCompile(`^[0-9A-Fa-f]+$`)
	case "argv":
		reg = regexp.MustCompile(`^(\-\-?[a-zA-Z]+\s+[a-zA-Z0-9\.\-\_]+\s*)+$`)
	case "id":
		reg = regexp.MustCompile(`^[a-f0-9]{24}$`)
	case "ver":
		reg = regexp.MustCompile(`^[a-z0-9A-Z\.]+$`)
	case "safestr":
		reg = regexp.MustCompile(`^[a-z0-9A-Z\.\-\\/_,]+$`)
	default:
		return true
	}
	return reg.MatchString(str)
}

/*GitAPI Gitlab Api
为确保脚本的版本不被意外改动（比如直接通过git），在commit git时，要获取本次commit的ID。
具体如下：
# 首先设置加密的commit信息。
# 同git api获取commit记录，并通过加密的commit信息找到本次修改。
# 将commitID保存到Logs.commitID
*/
func (git *GitAction) GitAPI() error {
	//create,update,get,del
	var apiurl *url.URL
	var err error
	if apiurl, err = url.ParseRequestURI(git.Url); err != nil {
		return err
	}
	query := url.Values{}
	switch git.Action {
	case "files":
		apiurl.Path = "/api/v4/projects/" + git.Projects + "/repository/files/" + git.Path
		apiurl.RawPath = "/api/v4/projects/" + url.QueryEscape(git.Projects) + "/repository/files/" + url.QueryEscape(git.Path)
	case "commit":
		apiurl.Path = "/api/v4/projects/" + git.Projects + "/repository/commits"
		apiurl.RawPath = "/api/v4/projects/" + url.QueryEscape(git.Projects) + "/repository/commits"
		query.Add("branch", "master")
		query.Add("author_name", git.User)
		query.Add("commit_message", git.Comment+"\n"+git.User)
	case "tree":
		apiurl.Path = "/api/v4/projects/" + git.Projects + "/repository/tree"
		apiurl.RawPath = "/api/v4/projects/" + url.QueryEscape(git.Projects) + "/repository/tree"
		query.Add("recursive", "true")
	default:
		err := errors.New("invalid action")
		return err
	}
	header := make(http.Header)
	header.Add("PRIVATE-TOKEN", git.GitToken)
	header.Add("Content-Type", "application/x-www-form-urlencoded; param=value")
	req := &http.Request{
		Method:     "POST",
		URL:        apiurl,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     header,
		Host:       apiurl.Host,
	}
	req.Method = git.Method

	switch git.Method {
	case "GET":
		header.Del("Content-Type")
		if git.Ref == "" {
			git.Ref = "master"
		}
		query.Add("ref", git.Ref)
		apiurl.RawQuery = query.Encode()
	case "POST", "PUT":
		query.Add("branch", "master")
		query.Add("author_name", git.User)
		query.Add("commit_message", git.Comment+"\n"+git.User)
		query.Add("encoding", "base64")
		query.Add("content", git.Content)
		req.Body = ioutil.NopCloser(strings.NewReader(query.Encode()))
	case "DELETE":
		query.Add("branch", "master")
		query.Add("author_name", git.User)
		query.Add("commit_message", git.Comment+"\n"+git.User)
		req.Body = ioutil.NopCloser(strings.NewReader(query.Encode()))
	}
	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 204 {
		err = errors.New(resp.Status)
		return err
	}
	if git.Method == "GET" {
		git.Result, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		// err = json.Unmarshal(content, git.Result)
		// return err
	}
	return nil
}

//GetGitFile 获取git上的文件内容
func (econf *EopsConf) GetGitFile(path, project, ref string) (*GitFile, error) {
	if gitfile, err := econf.MapCache.Get(project + path); err == nil {
		return gitfile.(*GitFile), nil
	}
	git := GitAction{
		Action:   "files",
		Method:   "GET",
		Url:      econf.GOPT.GitURL,
		GitToken: econf.GOPT.GitToken,
		Projects: project,
		Path:     path,
		Ref:      ref,
	}
	if err := git.GitAPI(); err != nil {
		return nil, err
	}

	gfile := &GitFile{}
	if err := json.Unmarshal(git.Result, gfile); err != nil {
		return gfile, err
	}
	econf.MapCache.Set(project+path, gfile, time.Duration(econf.GOPT.MemExpire)*time.Second)
	return gfile, nil
}

// func (econf *EopsConf) GetGitCommit(comment, project string) (string, error) {
// 	git := GitAction{
// 		Action:   "commit",
// 		Method:   "GET",
// 		Projects: project,
// 		Path:     "",
// 		Result:   &GitFile{},
// 	}
// 	if err := git.GitAPI(econf); err != nil {
// 		return "", err
// 	}
//
// }

//V
type V map[string]interface{}

//VString 返回字符串
func (v V) VString(k string) string {
	value, ok := v[k]
	if !ok {
		return ""
	}
	switch vv := value.(type) {
	case string:
		return vv
	}
	return ""
}

//VBool 返回bool值
func (v V) VBool(k string) bool {
	value, ok := v[k]
	if !ok {
		return false
	}
	switch vv := value.(type) {
	case bool:
		return vv
	}
	return false
}

func (v V) VMap(k string) V {
	value, ok := v[k]
	if !ok {
		return V{}
	}
	switch vv := value.(type) {
	case map[string]interface{}:
		return V(vv)
	case V:
		return vv
	}
	return V{}
}

//ObjType 获取interface的类型
func (v V) ObjType(k, vtype string) interface{} {
	value, ok := v[k]
	if !ok {
		return nil
	}
	switch value.(type) {
	case string:
		if vtype == "string" {
			return value
		}
	case []string:
		if vtype == "[]string" {
			return value
		}
	case []interface{}:
		if vtype == "[]interface" {
			return value
		}
	case int:
		if vtype == "int" {
			return value
		}
	case int64:
		if vtype == "int64" {
			return value
		}
	case float64:
		if vtype == "float64" {
			return value
		}
	case float32:
		if vtype == "float32" {
			return value
		}
	case map[string]interface{}:
		if vtype == "map" {
			return value
		}
	}

	return nil
}

//Type 返回key所对应值的类型
func (v V) Type(key string) string {
	vv := v[key]
	typeOf := reflect.TypeOf(vv)
	return typeOf.String()
}

//Json 将map转换为json
func (v V) Json(name string) []byte {
	var p []byte
	if name != "" {
		p, _ = json.MarshalIndent(V{name: v}, "", "  ")
	} else {
		p, _ = json.Marshal(v)
	}
	return p
}

func (v V) Unmarshal(out interface{}) error {
	p, err := json.Marshal(v)
	if err == nil {
		err = json.Unmarshal(p, out)
	}
	return err
}

func ResponseHandler(ctx *httpsvr.Context) {
	econf := ctx.Keys["econf"].(*EopsConf)
	if ctx.Response == nil {
		ctx.Response = 0
	}
	//econf.LogDebug("[Defer] Response: %v\n", ctx.Response)
	var out []byte
	ctype := "application/json; charset=UTF-8"
	switch ctx.Response.(type) {
	case *RespMsg:
		ctx.Writer.Header().Set("X-Code", strconv.Itoa(ctx.Response.(*RespMsg).Code))
		out = ctx.Response.(*RespMsg).Bytes()
	case []byte:
		out = ctx.Response.([]byte)
		ctype = "application/octet-stream"
	case string:
		out = []byte(ctx.Response.(string))
		ctype = "text/plain"
	case int:
		rmsg := NewRespMsg()
		code := ctx.Response.(int)
		if code > 0 {
			rmsg.SetCode(code, econf.Err)
		}
		ctx.Writer.Header().Set("X-Code", strconv.Itoa(code))
		//econf.LogDebug("[Defer] rmsg: %v\n", rmsg)
		out = rmsg.Bytes()
	default:
		if tmp, err := json.Marshal(ctx.Response); err == nil {
			ctype = "application/octet-stream"
			out = tmp
		} else {
			ctx.Writer.WriteHeader(http.StatusResetContent)
			return
		}
	}
	ctx.Writer.Header().Set("Content-Type", ctype)
	ctx.Writer.Write(out)
	if ctx.Writer.Status() != http.StatusNotFound {
		econf.OpLog(ctx)
	}
	if !ctx.NoLog {
		isArgs := "?"
		if ctx.Req.URL.RawQuery == "" {
			isArgs = ""
		}
		econf.AccessLog.Printf("%s %s %s %s%s%s %d %s %v\n",
			ctx.Begin.Format(time.RFC3339),
			ctx.RemoteAddr(false),
			ctx.Req.Method,
			ctx.Req.URL.Path,
			isArgs,
			ctx.Req.URL.RawQuery,
			ctx.Writer.Status(),
			http.StatusText(ctx.Writer.Status()),
			time.Since(ctx.Begin))
	}
}

//CryptSha256 sha256加密
func CryptSha256(str string) string {
	sha := sha256.New()
	io.WriteString(sha, str)
	shabytes := sha.Sum(nil)
	str = ""
	for i := 0; i < len(shabytes); i++ {
		str = str + strconv.FormatInt(int64(shabytes[i]), 16)
	}
	return str
}

//SearchStr 从字符串数组中查找内容
func SearchStr(a []string, s string) bool {
	for _, str := range a {
		if str == s {
			return true
		}
	}
	return false
}

func ReadCSV(fname string, fieldList []string) ([]V, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	fields := make(map[string]int)
	for _, field := range fieldList {
		fields[field] = -1
	}
	csvr := csv.NewReader(f)
	var record []string
	if record, err = csvr.Read(); err != nil {
		return nil, err
	}
	for i, field := range record {
		if _, ok := fields[field]; ok {
			fields[field] = i
		}
	}
	var data = []V{}
	for {
		record, err = csvr.Read()
		if err == io.EOF || record[0] == "" {
			break
		}
		fieldnum := len(record)
		lineData := V{}
		for field, i := range fields {
			lineData[field] = ""
			if i < fieldnum && i >= 0 {
				lineData[field] = record[i]
			}
		}
		data = append(data, lineData)
	}
	return data, nil
}

//48-57:数字 65-90:大写字母 97-122:小写字母 10+26+26=62
func RandomStr(l int) []byte {
	b := make([]byte, l)
	for i := 0; i < l; i++ {
		n := uint8(random(1000)%62 + 48)
		if n > 57 && n < 84 {
			n += 7
		} else if n >= 84 {
			n += 13
		}
		b[i] = n
	}
	return b
}

func FileMd5(name string, sercet []byte) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	h := md5.New()
	if len(sercet) > 0 {
		if _, err := h.Write(sercet); err != nil {
			return nil, err
		}
	}
	if _, err := io.Copy(h, file); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func random(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}
