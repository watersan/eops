package utils

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"opsv2.cn/opsv2/api/cache"
)

//UserColl 用户表结构
type UserColl struct {
	/*
	  tag格式: display[0|1] notnull[0|1] Lenmin Lenmax filter
	  Lev定义过滤级别。
	  1: 创建
	  2: 修改
	  4: 修改密码
	  8: 登录
	  16: 删除
	*/
	Name          string    `json:"name" Lev:"31" Display:"1" Notnull:"1" Lenmin:"4" Lenmax:"32" Filter:"email"`
	Email         string    `json:"email" Lev:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"email"`
	Passwd        string    `json:"passwd" Lev:"5" Display:"1" Notnull:"1" Lenmin:"8" Lenmax:"24" Filter:"passwd"`
	Mobile        string    `json:"mobile" Lev:"3" Display:"1" Notnull:"1" Lenmin:"11" Lenmax:"11" Filter:"number"`
	Fullname      string    `json:"fullname" Lev:"3" Display:"1" Notnull:"0" Lenmin:"6" Lenmax:"64"`
	Createtime    time.Time `Lev:"1" Display:"0" Notnull:"0"`
	Modifytime    time.Time `Lev:"3" Display:"0" Notnull:"0"`
	Logintime     time.Time `Lev:"15" Display:"0" Notnull:"0"`
	PrevilageList []string  `json:"previlagelist"`
}

//UserCollEdit 用于更新的用户结构
type UserCollEdit struct {
	//tag格式: display[0|1] notnull[0|1] Lenmin Lenmax filter
	Name       string    `json:"name" Display:"1" Notnull:"1" Lenmin:"4" Lenmax:"32" Filter:"name"`
	Email      string    `json:"email" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"email"`
	Mobile     string    `json:"mobile" Display:"1" Notnull:"1" Lenmin:"11" Lenmax:"11" Filter:"number"`
	Fullname   string    `json:"fullname" Display:"1" Notnull:"0" Lenmin:"6" Lenmax:"64"`
	Modifytime time.Time `Display:"0" Notnull:"0"`
}

//UserCollPwd 更新用户密码的结构
type UserCollPwd struct {
	//tag格式: display[0|1] notnull[0|1] Lenmin Lenmax filter
	Name       string    `json:"name" Display:"1" Notnull:"1" Lenmin:"4" Lenmax:"32" Filter:"name"`
	Passwd     string    `json:"passwd" Display:"1" Notnull:"1" Lenmin:"8" Lenmax:"24" Filter:"passwd"`
	Modifytime time.Time `Display:"0" Notnull:"0"`
}

//LoginStatus 用户登录状态
type LoginStatus struct {
	Code          int              `json:"code"`
	Message       string           `json:"message"`
	User          *UserColl        `json:"user"`
	PrevilageList []*PrevilageInfo `json:"previlagelist"`
}

//PrevilageInfo 角色信息
type PrevilageInfo struct {
	Name           string         `json:"name"`
	PermissionList map[string]int `json:"permissionlist"`
}

//ModuleList 权限定义
type ModuleList struct {
	//模块名称
	Name string `json:"name"`
	//子模块名称
	Subname string `json:"subname"`
}

// EopsConf 全局Auth模块的结构
type EopsConf struct {
	Cache cache.Cacher
	Db    *mgo.Database
	Debug int
}

const (
	//StatusFailed 操作失败
	StatusFailed = 460
	//StatusInputNotLegal 操作失败，输入的内容不合法
	StatusInputNotLegal = 461
	//StatusInvalidVer 请求的版本无效
	StatusInvalidVer = 462
	//StatusLoginWrongPWD 登录失败,密码错误
	StatusLoginWrongPWD = 463
	//StatusUserIsRegister 用户已注册
	StatusUserIsRegister = 464
	//StatusLoginNotUser 登录失败，没有此用户
	StatusLoginNotUser = 465
	//StatusSecookieInvalid 安全cookie无效
	StatusSecookieInvalid = 466
	//StatusNoBody body为空
	StatusNoBody = 467
	//StatusBodyTooLong body太大
	StatusBodyTooLong = 468
	//StatusBodyInvalid body格式无效
	StatusBodyInvalid = 469
	//MaxBodySize 允许最大的body大小
	MaxBodySize = int64(10 << 20)
)

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

//Init 初始化eopsconf
func Init(cache cache.Cacher, db *mgo.Database) (econf *EopsConf) {
	econf = &EopsConf{}
	econf.Cache = cache
	econf.Db = db
	econf.Debug = 7
	return
}

//PermissionList 获取权限定义列表
func PermissionList() map[string]string {
	Permission := map[string]string{
		"add": "添加",
		"mod": "修改",
		"del": "删除",
		"sch": "查找",
		"run": "运行",
	}
	return Permission
}

//GetClip 获取客户端IP
func GetClip(req *http.Request) string {
	addr := req.Header.Get("X-Real-IP")
	if addr == "" {
		addr = req.Header.Get("X-Forwarded-For")
		if addr == "" {
			addr = req.RemoteAddr
		}
	}
	return addr
}

//GetReqBody 获取请求中的Body
func GetReqBody(req *http.Request, out interface{}) (b []byte, ecode int, err error) {
	ecode = 0
	if req.Body == nil {
		ecode = 1202
		return
	}
	ct := req.Header.Get("Content-Type")
	// RFC 2616, section 7.2.1 - empty type
	//   SHOULD be treated as application/octet-stream
	if ct == "" {
		ct = "application/octet-stream"
	}
	ct, _, _ = mime.ParseMediaType(ct)
	switch ct {
	case "application/x-www-form-urlencoded", "application/json":
		reader := io.LimitReader(req.Body, MaxBodySize+1)
		b, err = ioutil.ReadAll(reader)
		if err != nil {
			ecode = 1205
			return
		}
		if int64(len(b)) > MaxBodySize {
			ecode = 1203
			return
		}
	default:
		ecode = 1204
	}
	if out != nil {
		if err = json.Unmarshal(b, out); err != nil {
			ecode = 1206
		}
	}
	return
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

//Array2Str 字符串数组转字符串
func Array2Str(strarr []string) (str string) {
	arrlen := len(strarr)
	for i := 0; i < arrlen; i++ {
		if i == 0 {
			str = strarr[i]
		} else {
			str += "," + strarr[i]
		}
	}
	return
}

//RawQuery 将map转换为url的query字符串
func RawQuery(q map[string]string) string {
	var str string
	first := true
	for k, v := range q {
		if first {
			str += k + "=" + v
			first = false
			continue
		}
		str += "&" + k + "=" + v
	}
	return str
}
