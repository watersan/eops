package cmdb

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

//MgoProject 项目
type MgoProject struct {
	ID          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"32" Filter:"username"`
	Describe    string        `json:"describe" Mode:"3" Display:"1" Notnull:"0" `
	Topology    string        `json:"topology" Mode:"3"`
	Opser       string        `json:"opser" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"username"`
	SecondOpser string        `json:"secondopser" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"username"`
	Modifytime  time.Time     `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime  time.Time     `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoCluster 项目
type MgoCluster struct {
	ID          bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"32" Filter:"username"`
	Describe    string        `json:"describe" Mode:"3" Display:"1" Notnull:"0" `
	Project     string        `json:"project" Mode:"3" Display:"1" Notnull:"1" `
	Depend      []string      `json:"depend" Mode:"3" Display:"1" Notnull:"0" `
	Shared      bool          `json:"shared" Mode:"3" Display:"1" Notnull:"1" `
	Opser       string        `json:"opser" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"username"`
	SecondOpser string        `json:"secondopser" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"username"`
	Modifytime  time.Time     `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime  time.Time     `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoApp 项目
type MgoApp struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"32" Filter:"username"`
	Cluster    string        `json:"cluster" Mode:"3" Display:"1" Notnull:"1" `
	Depend     []string      `json:"depend" Mode:"3" Display:"1" Notnull:"0" `
	From       string        `json:"from" Mode:"3" Display:"1" Notnull:"1" `
	Config     string        `json:"config" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	ConfigVer  string        `json:"configver"`
	SourceType string        `json:"sourcetype" Mode:"3" Display:"1" Notnull:"0"`
	Source     string        `json:"source" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	AppVer     string        `json:"appver" Mode:"3" Display:"1" Notnull:"0" `
	//Type       string   `json:"type" Mode:"3" Display:"1" Notnull:"0"`
	//AppConf    []string  `json:"appconf" Mode:"3" Display:"1" Notnull:"0" `
	Environment map[string]*APPEnvironment `json:"environment" Mode:"2" Display:"1" Notnull:"0" `
	Variables   map[string]string          `json:"variables" Mode:"2" Display:"1" Notnull:"0"`
	/*
		-2: 新添加，但未完成：更新模板使用数和添加依赖应用。
		-1: 新添加，未执行部署
		0: 正常，已经下发到设备
		>0: 已修改，但没有进行同步或未完成同步。
	*/
	Status     int       `json:"status" Mode:"3" Display:"1" Notnull:"0"`
	Operator   string    `json:"operator" Mode:"3" Display:"1" Notnull:"0"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//APPEnvironment 应用在不同环境下的配置
type APPEnvironment struct {
	Hosts      []string          `json:"hosts" Mode:"3" Display:"1" Notnull:"0" `
	HostsCount int               `json:"hostscount" Mode:"3" Display:"1" Notnull:"0"`
	Variables  map[string]string `json:"variables" `
}

//MgoAppConf 应用配置
type MgoAppConf struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name      string        `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"32" Filter:"username"`
	Path      string        `json:"path" Mode:"3" Display:"1" Notnull:"1"`
	AppID     string        `json:"appid" Mode:"1" Display:"1" Notnull:"1"`
	Version   string        `json:"version" Mode:"3" Display:"1" Notnull:"0"`
	CommitMsg string        `json:"commitmsg" Mode:"3" Display:"1" Notnull:"0"`
	Operator  string        `json:"operator" Mode:"3" Display:"1" Notnull:"1"`
	IsEnable  bool          `json:"isenable"`
	/*
		0: 正常，完成同步
		非零: 已修改，但没有进行同步或未完成全部环境的同步。
	*/
	Status     float64   `json:"status" Mode:"3" Display:"1" Notnull:"0"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoAppTPL 项目
/*
如果增加多版本记录，就通过增加以下字段：
Version：版本号，每次递增
Latest：标记为最后的版本
Author：版本创建人
*/
type MgoAppTPL struct {
	//Name 唯一索引
	ID   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string        `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"128" Filter:"username"`
	//Config 配置文件的git地址
	Config string `json:"config" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	//Reload       string `json:"reload" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	ConfigVer    string `json:"configver"`
	ConfigVerify string `json:"configverify" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	SourceType   string `json:"sourcetype" Mode:"3" Display:"1" Notnull:"0"`
	Source       string `json:"source" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	AppVer       string `json:"appver" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	Build        string `json:"build" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	Install      string `json:"install" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	//Upgrade         string   `json:"upgrade" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	PreInstall  string `json:"preinstall" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	PostInstall string `json:"postinstall" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	RollBack    string `json:"rollback" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	//Process 进程名称
	Process string `json:"process" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	//Port 端口号
	Port            string `json:"port" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128" Filter:"port"`
	VerifyUsability string `json:"verifyusability" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	//Logpath         string   `json:"logpath" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	Depend       []string `json:"depend" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"128" Filter:"username"`
	Additionconf []string `json:"additionconf" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"128" Filter:"path"`
	//Installpath     string    `json:"installpath" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"128"`
	//Confpath 配置文件路径。
	Confpath   string    `json:"confpath" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"128"`
	Monitor    string    `json:"monitor" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"128"`
	Usage      int       `json:"usage" Mode:"3" Display:"1" Notnull:"0"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoYunProvider 服务商
type MgoYunProvider struct {
	ID bson.ObjectId `json:"id" bson:"_id,omitempty"`
	//Name aliyun,aws,ucloud
	Name      string   `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"128" Filter:"nickname"`
	Key       string   `json:"key" Mode:"3" Display:"1" Notnull:"1" Lenmin:"16" Lenmax:"64"`
	Secret    string   `json:"secret" Mode:"3" Display:"1" Notnull:"1" Lenmin:"16" Lenmax:"64"`
	Proxy     string   `json:"proxy" Mode:"3" Display:"1" Notnull:"0" Lenmin:"7" Lenmax:"15" Filter:"ip"`
	ProxyType string   `json:"proxytype" Mode:"3" Display:"1" Notnull:"0" Lenmin:"3" Lenmax:"5"`
	Regions   []string `json:"regions" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"32" Filter:"safestr"`
	//ProjectID 项目ID，只对ucloud有效
	ProjectID  string    `json:"projectid" Mode:"3" Display:"1" Notnull:"0"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoIDCProvider 服务商
type MgoIDCProvider struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"128" Filter:"nickname"`
	Company    string        `json:"company" Mode:"3" Display:"1" Notnull:"1" Lenmin:"4" Lenmax:"64"`
	Location   string        `json:"location" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"10"`
	ISP        string        `json:"isp" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"20"`
	Conects    string        `json:"conects" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"20"`
	Email      string        `json:"email" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"email"`
	Mobile     string        `json:"mobile" Mode:"3" Display:"1" Notnull:"1" Lenmin:"11" Lenmax:"12" Filter:"mobile"`
	Proxy      string        `json:"proxy" Mode:"3" Display:"1" Notnull:"0" Lenmin:"7" Lenmax:"15" Filter:"ip"`
	BW         int           `json:"bw" Mode:"3" Display:"1" Notnull:"1"`
	BeginDate  string        `json:"begindate" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"20"`
	EndDate    string        `json:"enddate" Mode:"3" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"20"`
	Modifytime time.Time     `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time     `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoResource 资源信息
type MgoResource struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	IP       string        `json:"ip" Mode:"3" Display:"1" Notnull:"1" Lenmin:"7" Lenmax:"21" Filter:"ipport"`
	EIP      string        `json:"eip" Mode:"3" Display:"1" Notnull:"0" Lenmin:"7" Lenmax:"21" Filter:"ipport"`
	HostID   string        `json:"hostid" Mode:"3" Display:"1" Notnull:"1" Lenmin:"3" Lenmax:"64" Filter:"safestr"`
	HostName string        `json:"hostname" Mode:"3" Display:"1" Notnull:"1" Lenmin:"4" Lenmax:"128" Filter:"username"`
	Location string        `json:"location" Mode:"3" Display:"1" Notnull:"1"`
	Region   string        `json:"region" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"32" Filter:"safestr"`
	//Type 资源的类型，目前只是对aws有效，
	Type   string `json:"type" Mode:"3" Display:"1" Notnull:"0"`
	CPU    int    `json:"cpu" Mode:"3" Display:"1" Notnull:"1"`
	Memory int    `json:"memory" Mode:"3" Display:"1" Notnull:"1"`
	/*
		/:20G;/data:100G
	*/
	Disk        string `json:"disk" Mode:"3" Display:"1" Notnull:"1"`
	OS          string `json:"os" Mode:"3" Display:"1" Notnull:"1"`
	OSVer       string `json:"osver" Mode:"3" Display:"1" Notnull:"1"`
	Environment string `json:"environment" Mode:"3" Display:"1" Notnull:"1"`
	/*
		{
			0: UnUsed
			1: Used
			2: maintaining
			3: stoped
		}
	*/
	Status     int               `json:"status" Mode:"3" Display:"1" Notnull:"1"`
	Usage      int               `json:"usage"`
	Addition   map[string]string `json:"addition" Mode:"3" Display:"1" Notnull:"0"`
	Modifytime time.Time         `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time         `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

/*
项目名称是唯一的，项目和集群名称联合唯一。
*/
