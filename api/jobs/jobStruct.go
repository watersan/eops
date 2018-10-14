package jobs

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

//MgoJobLogs 作业运行的日志。
type MgoJobLogs struct {
	ID bson.ObjectId `json:"id" bson:"_id,omitempty"`
	//TaskAttr 任务属性，从TaskID获取的作业，再生成任务。此字段不记录到mongodb
	TaskAttr *TaskAttr `json:"taskattr"`
	//TaskName mgoJobs：path/name;mgoFlows:name，根据type进行判断
	TaskName string `json:"taskname"`
	//Target 执行的目标：projects,clusters,apps,hosts
	Target string `json:"target"`
	//List 目标列表。如果执行目标app，这里就是app列表
	List []string `json:"list"`
	//CDEnv 作业需要运行的环境
	CDEnv string `json:"CDenv"`
	//作业执行人
	Operator string `json:"operator"`
	//Type的值：Jobs or JobsFlow
	Type string `json:"type"`
	//Env 作业执行时需要设置的环境变量
	Env map[string]string `json:"env"`
	//作业创建的时间
	BeginTime time.Time `json:"begintime"`
	//总的运行时间：从第一个机器开始执行到最后一个机器执行完成。
	EndTime time.Time `json:"endtime"`
	//已执行完成的机器数
	Progress int `json:"progress"`
	//host数量
	HostNum int `json:"hostnum"`
	//执行失败的host数量：只根据host上最后一个作业的结果判断。
	Failed int `json:"failed"`
	//Status ansible-playbook的执行状态。
	/*
		0: 正常
		3: 执行失败。因为更新状态的方式是或运算，避免和901重合，所以选3
		900: 回滚成功
		901: 回滚失败
		911: 执行器异常
		912: 任务未执行
		其他: ansible的返回值
	*/
	Status int `json:"status"`
	//Msg ansible的输出内容
	Msg      string      `json:"msg"`
	HostLogs []*HostLogs `json:"hostlogs"`
}

//TaskAttr job运行属性
type TaskAttr struct {
	JobID    string `json:"jobid"`
	Alias    string `json:"alias"`
	Name     string `json:"name" Mode:"1" Display:"1" Notnull:"1" Lenmin:"5" Lenmax:"32" Filter:"scriptname"`
	Path     string `json:"path" Mode:"1" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"32" Filter:"path"`
	Argv     string `json:"argv"`
	CommitID string `json:"commitid"`
	Timeout  int    `json:"timeout" Mode:"3" Display:"1" Notnull:"0"`
	//job运行的优先级。
	Priority int       `json:"pri" Mode:"3" Display:"1" Notnull:"0" `
	User     string    `json:"user"`
	Latest   bool      `json:"latest"`
	Next     *TaskAttr `json:"next"`
	Rescue   *TaskAttr `json:"rescue"`
}

//MgoRunLogs 作业运行的日志。
type MgoRunLogs struct {
	ID      bson.ObjectId `json:"id" bson:"_id,omitempty"`
	JobAttr *JobAttr      `json:"jobattr"`
	//作业执行人
	Operator string `json:"operator"`
	//JobsType的值：typeJobs or typeJobsFlow
	JobsType string `json:"jobstype"`
	//作业创建的时间
	BeginTime time.Time `json:"begintime"`
	//总的运行时间：从第一个机器开始执行到最后一个机器执行完成。
	EndTime time.Time `json:"endtime"`
	//已执行完成的机器数
	Progress int `json:"progress"`
	//host数量
	HostNum int `json:"hostnum"`
	//执行失败的host数量
	Failed int `json:"failed"`
	//Code ansible-playbook的执行状态。
	Code int `json:"code"`
	//Msg ansible的输出内容
	Msg      int         `json:"msg"`
	HostLogs []*HostLogs `json:"hostlogs"`
}

//HostLogs 作业在每个主机上的运行日志
type HostLogs struct {
	HostName string `json:"hostname"`
	Jobs     []*JobResult
}

//JobResult 作业执行结果
type JobResult struct {
	Alias string `json:"alias"`
	Name  string `json:"name"`
	Argv  string `json:"argv"`
	Index int    `json:"index"`
	Env   map[string]string
	/*
		  Code值的定义：
		    -99：初始化值
				-10: 执行失败
				-9: 无效的用户
	*/
	Code    int     `json:"code"`
	RunTime float32 `json:"runtime"`
	UsedCPU float32 `json:"usedcpu"`
	//UsedMEM 单位是MB
	UsedMEM int    `json:"usedmem"`
	StdOut  string `json:"stdout"`
	ErrOut  string `json:"errout"`
}

//CBResult callback's result
type CBResult struct {
	ID     string
	Env    map[string]string
	Status int
	Result *HostLogs
}

//MgoJobs 作业的数据结构。脚本内容保存在gitlab。
type MgoJobs struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name     string        `json:"name" Mode:"1" Display:"1" Notnull:"1" Lenmin:"5" Lenmax:"32" Filter:"scriptname"`
	Path     string        `json:"path"  Mode:"1" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"32" Filter:"path"`
	CommitID string        `json:"commitid"  Mode:"3" Display:"1" Notnull:"0" Lenmin:"40" Lenmax:"40" Filter:"bytes"`
	Timeout  int           `json:"timeout" Mode:"3" Display:"1" Notnull:"0" `
	//job运行的优先级。
	Priority int    `json:"priority" Mode:"3" Display:"1" Notnull:"0" `
	User     string `json:"user" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"40" Filter:"username"`
	/*
		0: 正常
		1: 已更新脚本，但未更新版本
	*/
	Status int `json:"status"`
	//Notes    string `json:"notes"`
	Operator   string    `json:"operator" Mode:"3" Display:"1" Notnull:"1"`
	Modifytime time.Time `json:"modifytime" Mode:"1" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//Schedule 定时的作业内容。
type Schedule struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name     string        `json:"name" Mode:"1" Display:"0" Notnull:"1" Lenmin:"1" Lenmax:"32" Filter:"username"`
	Argv     string        `json:"argv"`
	Describe string        `json:"describe" Mode:"3" Display:"1" Notnull:"0" `
	//Ture: jobs; False: jobsflows
	//Type bool `json:"type"`
	//crontab格式
	Schedule   string    `json:"schedule"`
	Modifytime time.Time `json:"modifytime" Mode:"1" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//MgoFlows 作业流程的数据结构。
type MgoFlows struct {
	ID       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name     string        `json:"name" Mode:"1" Display:"0" Notnull:"1" Lenmin:"1" Lenmax:"32" Filter:"username"`
	Describe string        `json:"describe" Mode:"3" Display:"1" Notnull:"0" `
	//Status 值为1：为自动创建，不允许通过作业流管理进行修改
	Status     int       `json:"status" Mode:"3" Display:"1" Notnull:"0"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
	//Task 为保证和job的一致性，任务属性不记录name和path，而是通过id形式关联
	Task *TaskAttr `json:"task" Mode:"3" Display:"0" Notnull:"0"`
}

//JobAttr job属性
type JobAttr struct {
	Name    string `json:"name" Mode:"1" Display:"1" Notnull:"1" Lenmin:"5" Lenmax:"32" Filter:"scriptname"`
	Path    string `json:"path" Mode:"1" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"32" Filter:"path"`
	Argv    string `json:"argv"`
	Version string `json:"version"`
	Latest  bool   `json:"latest"`
	Timeout string `json:"timeout" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"3" Filter:"number"`
	//job运行的优先级。
	Priority string `json:"pri" Mode:"3" Display:"1" Notnull:"0" Lenmin:"1" Lenmax:"3" Filter:"number"`
	SysUser  string `json:"sysuser"`
}

//JobRunArgs 作业执行参数
type JobRunArgs struct {
	JobAttr  *JobAttr `json:"jobattr"`
	Projects []string `json:"projects" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"32" Filter:"username"`
	Clusters []string `json:"clusters" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"32" Filter:"username"`
	Apps     []string `json:"apps" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"32" Filter:"username"`
	//Env 需要传递给ansible的变量，及传递给执行器的环境变量
	Env      map[string]string `json:"env"`
	Operator string            `json:"operator"`
	//CDEnv 作业需要运行的环境
	CDEnv string `json:"CDenv"`
	/*
		type定义：
		0： jobs
		1： jobflow
	*/
	Type int `json:"type"`
	//Split 将目标主机分批次执行。1代表10%，30%，60%。2代表10%，20%，30%，40%。
	Split int `json:"s" Mode:"1" Display:"1" Notnull:"0"`
	//Interval 在split模式中，每个批次执行间隔。
	Interval int      `json:"i" Mode:"1" Display:"1" Notnull:"0"`
	HostList []string `json:"hl" Mode:"1" Display:"1" Notnull:"0"`
}

//MgoShortcut 作业执行参数
type MgoShortcut struct {
	ID   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name string        `json:"name" Mode:"1" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"32" Filter:"username"`
	//JobRunArgs
	//TaskAttr 任务属性，从TaskID获取的作业，再生成任务。此字段不记录到mongodb
	TaskAttr *TaskAttr `json:"taskattr"`
	//TaskID mgoJobs或mgoFlows的id，根据type进行判断
	//TaskID string `json:"taskid"`
	//Target 执行的目标：project,cluster,app,host
	Target string `json:"target"`
	//List 目标列表。如果执行目标app，这里就是app列表
	List       []string  `json:"list"`
	Operator   string    `json:"operator" Mode:"1" Display:"1" Notnull:"1"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Createtime time.Time `json:"createtime" Mode:"1" Display:"0" Notnull:"0"`
}

//JobExecutor 作业执行器
type JobExecutor struct {
	Host      string
	PlayBook  string
	ExtraVars map[string]string
}

type jobPathTree struct {
	files []string
	path2 []string
	nodes map[string][]string
}

type DeployFlows struct {
	Name       string
	Describe   string
	RollBack   string
	Modifytime time.Time
	Jobs       []*DeployJob
}

type DeployJob struct {
	Alias string
	Job   string
}
