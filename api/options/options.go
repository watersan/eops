package options

import (
	"bytes"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"

	"opsv2.cn/opsv2/api/ipsearch"
)

//GlobalOPT 全局配置
type GlobalOPT struct {
	ConfFile      string `internal:"yes"`
	Prefix        string
	PrefixLen     int `internal:"yes"`
	WorkDir       string
	BaseURL       string
	SessionExpire int64
	// UserCollection   string
	// RolesCollection  string
	// AppTPLCollection string
	DataPath string
	//AppInstallPath 应用的安装路径。系统包管理器安装的应用，此配置无效。
	AppInstallPath    string
	AppSource         string
	SSHUser           string
	Environments      []string
	EnvironmentsLen   int `internal:"yes"`
	SessionSecretKey  string
	PWDSecretKey      string
	MaxBodySize       int64
	MDBPrefix         string
	SessionExpired    string
	HookToken         string
	ApptplConfRepo    string
	AppConfRepo       string
	ApptplRepo        string
	GitURL            string
	GitToken          string
	ExecutorPath      string
	Ansible           string
	JobsRepo          string
	JobsSavePath      string
	ParallelJobExe    int
	AnsibleFork       string
	MemExpire         int
	IPWhiteListForAPI ipsearch.IPRanges
	RunEnv            string `internal:"yes"`
	Bind              string
	Port              string
	PidFile           string
	LogDir            string
	MdbHost           string
	Mdb               string
	MdbUser           string
	MdbPW             string
	MdbTimeout        int
	LdapHost          string
	LdapPort          string
	LdapDN            string
	LdapTLS           string
	LdapBindDn        string
	LdapBindPasswd    string
	LogLevel          int
	CDEnvBuild        string
}

// type IpRange struct {
// 	Min uint32
// 	Max uint32
// }

const (
	CDEnvBuild   = "build"
	CDEnvDevelop = "develop"
	CDEnvTest    = "test"
	CDEnvPre     = "pre-release"
	CDEnvRelease = "release"
)

//GOPT 全局配置参数
func GOPT() GlobalOPT {
	GlobalOPTs := GlobalOPT{
		Prefix:        "/eops",
		PrefixLen:     5,
		WorkDir:       "/data/opsv2",
		BaseURL:       "http://console.zdops.com",
		SessionExpire: 3600,
		// UserCollection:   "users",
		// RolesCollection:  "roles",
		//AppTPLCollection: "apptpl",
		DataPath:         "/data",
		AppInstallPath:   "/data/www",
		AppSource:        "http://console.zdops.com/app",
		SSHUser:          "dale",
		Environments:     []string{CDEnvDevelop, CDEnvTest, CDEnvRelease},
		SessionSecretKey: "s5N2plaxfEnyg27JrRzWNxVIBwDA7IFWG7Hk2rlAKCNHCwysBqprEDMsXKx8hrbn",
		PWDSecretKey:     "6DFmZQD2D0kbNyzHPfYaZXtceCjdb623offPnKPXjZiqDJaYDWrAhkbTSZEMbt0R",
		MaxBodySize:      int64(10 << 20),
		MDBPrefix:        "zd_",
		SessionExpired:   "expired",
		HookToken:        "YLMuCgEWzR#k2pT9lzgSECk28uoYlNWtzhY0HwCrL9xOXnWKhP9B&tGx9EtNXLgF",
		GitURL:           "https://code.daledi.cn/",
		//GitToken:         "vim3wyxtkwHQCd7iPBzE",   //dale
		GitToken:       "4X-w5kqayVKGpgtdoMsy", //dyzops
		ExecutorPath:   "/data/scripts/Executor",
		Ansible:        "/usr/local/bin/ansible-playbook",
		JobsRepo:       "ops/jobs",
		JobsSavePath:   "/data/scripts",
		ParallelJobExe: 10,
		AnsibleFork:    "50",
		MemExpire:      3600,
		RunEnv:         "dev",
		Bind:           "127.0.0.1",
		Port:           "5000",
		PidFile:        "/data/log/opsv2/opsv2.pid",
		LogDir:         "/data/log/opsv2",
		MdbHost:        "localhost",
		Mdb:            "eops",
		MdbUser:        "eops",
		MdbPW:          "lejkkgt0uPXw6ArNGJKZ",
		MdbTimeout:     5,
		LogLevel:       7,
		CDEnvBuild:     "127.0.0.1",
	}
	//GlobalOPTs.AppConfRepo = "appconf" //url.QueryEscape("ops/appconf")
	//GlobalOPTs.ApptplConfRepo = "apptplconf"
	GlobalOPTs.ApptplRepo = "apptpl"
	GlobalOPTs.EnvironmentsLen = len(GlobalOPTs.Environments)
	GlobalOPTs.IPWhiteListForAPI = ipsearch.IPRanges(make([]*net.IPNet, 2))
	GlobalOPTs.IPWhiteListForAPI.ConvIPList([]string{"127.0.0.1/32", "192.168.0.0/16"})
	return GlobalOPTs
}

func (gopt GlobalOPT) IsRelease() bool {
	return gopt.RunEnv == "release"
}

//ReadConfig 从配置文件读取配置
func ReadConfig(cfile string, gopt *GlobalOPT) error {
	// gopt := GlobalOPT{}
	f, err := os.Open(cfile)
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bytes.NewBuffer([]byte{})
	if _, err := buf.ReadFrom(f); err != nil {
		return err
	}
	myref := reflect.ValueOf(gopt).Elem()
	// envlist := []string{}
	// ips := ipsearch.IPRanges{}
	for {
		line, err := buf.ReadBytes(0x0a)
		if err != nil {
			break
		}
		if line[0] == '#' {
			continue
		}
		i := bytes.IndexByte(line, '=')
		name := string(bytes.TrimSpace(line[:i]))
		value := string(bytes.TrimSpace(line[i+1:]))
		field := myref.FieldByName(name)
		if ftype, ok := myref.Type().FieldByName(name); !ok || (ok && ftype.Tag.Get("internal") == "yes") {
			continue
		}
		switch field.Type().String() {
		case "string":
			field.SetString(value)
		case "int", "int64":
			if v, err := strconv.ParseInt(value, 10, 0); err == nil {
				field.SetInt(v)
			}
		}
		if name == "Environments" {
			gopt.Environments = strings.Split(value, ",")
			gopt.EnvironmentsLen = len(gopt.Environments)
		} else if name == "IPWhiteListForAPI" {
			iplist := strings.Split(value, ",")
			gopt.IPWhiteListForAPI = ipsearch.IPRanges(make([]*net.IPNet, len(iplist)))
			gopt.IPWhiteListForAPI.ConvIPList(iplist)
		} else if name == "Prefix" {
			gopt.PrefixLen = len(gopt.Prefix)
		}
	}
	return nil
}

/*
git目录：
	|- appconf
		|- projectname
			|- clustername
				|- appname
	|- apptplconf
		|- appname
	|- apptpl
		|- appname

WorkDir目录结构：
	|- jobs    根据作业的目录结构自动创建
		|- jobs.yml
		|- jobflowname.yml
	|- apptpl
		|- appname
			|- package
				version.zip
				...
			|- conf
			 	|- version/
				...
  |- projectname/
		|- clustername/
			|- appname/
				|- history     每次部署的软件包和配置文件
				|- release     当前运行的软件目录
				|- .version.ini  保存当前版本和上次的版本信息。软件包和配置文件
ZDOPS_DeployDir:
	|- appname/
		|- history     每次部署的软件包和配置文件
			|- app_version
			|- conf_version
		|- release     当前运行的软件目录
		|- .version.ini  保存当前版本和上次的版本信息。软件包和配置文件


*/
