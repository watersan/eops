package deploy

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

/*
1、appvar 1 测试环境通过，等待正式环境部署 22
2、appvar 2 测试环境部署完成，等待测试结果 20
3、appvar 3 开发环境测完成，等待测试环境部署 10

添加新的部署任务条件：status != -1|10|11

测试环境部署条件：status >= 22
正式：status != |31|36
*/
//MgoHistorys 部署状态表
type MgoHistorys struct {
	ID            bson.ObjectId `json:"id" bson:"_id,omitempty"`
	AppID         string        `json:"appid" Mode:"1" Display:"1" Notnull:"1" Lenmin:"24" Lenmax:"24" Filter:"id"`
	AppVer        string        `json:"appver" Mode:"3" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"40" Filter:"ver"`
	ApptplVer     string        `json:"apptplver" Mode:"3" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"40" Filter:"ver"`
	ApptplConfVer string        `json:"apptplconfver" Mode:"3" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"40" Filter:"ver"`
	AppConfVer    string        `json:"appconfver" Mode:"3" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"40" Filter:"ver"`
	//Action 动作。
	//Action        string `json:"appconfver" Mode:"3" Display:"1" Notnull:"1" Lenmin:"1" Lenmax:"10" Filter:"username"`
	//LogID         string        `json:"logid" Mode:"3" Display:"1" Notnull:"1" Lenmin:"24" Lenmax:"24" Filter:"id"`
	/*
		-2 只记录，不部署。用于将appconf的变更合并到app的部署中。
		-1	初始值。同一个appid的初始值记录只能有1条。
		0	正式环境部署成功
		2	放弃此版本
		10|20	（构建|开发|测试）环境部署成功 (正式环境部署成功后为0)
		11|21|31	（开发|测试|正式）环境部署部署中
		12|22|32   开发|测试|正式）环境测试通过
		13|23|33   开发|测试|正式）环境测试未通过
		15|25|35 开发|测试|正式）环境部署回滚成功
		16|26|36 开发|测试|正式）环境部署回滚失败
		19|29|39  环境部署失败
	*/
	Status     int       `json:"status" Mode:"3" Display:"1" Notnull:"1"`
	CreateTime time.Time `json:"createtime" Mode:"3" Display:"1" Notnull:"1"`
	/*
			Flow 保存各个部署环境中执行的日志
		   Flow["dev"]
		   Flow["test"]
		   Flow["pre-release"]
		   ....
	*/
	Flow         []FlowInfo `json:"flow" Mode:"3" Display:"1" Notnull:"1"`
	ReleaseTime  time.Time  `json:"releasetime" Mode:"3" Display:"1" Notnull:"1"`
	RollbackTime time.Time  `json:"rollbacktime" Mode:"3" Display:"1" Notnull:"1"`
}

//FlowInfo 部署流程信息：用于记录每个部署过程中的日志信息
type FlowInfo struct {
	CDEnv string `json:"cdenv"`
	LogID string `json:"logid"`
	//Status 执行状态，等同于RunLogs.Failed
	Status       int       `json:"status"`
	CreateTime   time.Time `json:"createtime" Mode:"3" Display:"1" Notnull:"1"`
	CompleteTime time.Time `json:"completetime"`
}

//GitHookReq gitlab webhook的请求数据格式
type GitHookReq struct {
	ObjectKind        string        `json:"object_kind"`
	EventName         string        `json:"event_name"`
	Before            string        `json:"before"`
	After             string        `json:"after"`
	Ref               string        `json:"ref"`
	CheckoutSha       string        `json:"checkout_sha"`
	UserName          string        `json:"user_name"`
	UserEmail         string        `json:"user_email"`
	Project           *GitProject   `json:"project"`
	Commits           []*GitCommits `json:"commits"`
	TotalCommitsCount int           `json:"total_commits_count"`
	//repository *GitRepository `json:"repository"`
}

//GitProject gitlab项目信息
type GitProject struct {
	Name      string `json:"name"`
	GitHTTP   string `json:"git_http_url"`
	GitSSH    string `json:"git_ssh_url"`
	NameSpace string `json:"namespace"`
}

//GitCommits 提交的说明
type GitCommits struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
