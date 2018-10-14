package auth

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

//MgoUser 用户表结构
type MgoUser struct {
	/*
	  tag格式: display[0|1] notnull[0|1] Lenmin Lenmax filter
	  Lev定义过滤级别。
	  1: 创建
	  2: 修改
	  4: 修改密码
	  8: 登录
	  16: 删除
	*/
	Name       string    `json:"name" Mode:"1" Display:"1" Notnull:"1" Lenmin:"6" Lenmax:"128" Filter:"email"`
	Passwd     string    `json:"passwd" Mode:"5" Display:"1" Notnull:"1" Lenmin:"8" Lenmax:"24" Filter:"safestr"`
	LdapDN     string    `json:"ldapdn" Mode:"1" Display:"0" Notnull:"0"`
	Mobile     string    `json:"mobile" Mode:"3" Display:"1" Notnull:"1" Lenmin:"11" Lenmax:"12" Filter:"mobile"`
	Fullname   string    `json:"fullname" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"11" Filter:"nickname"`
	KeyID      string    `json:"keyid" Mode:"1"`
	Secret     string    `json:"secret" Mode:"1"`
	GitToken   string    `json:"gittoken" Mode:"1"`
	Createtime time.Time `Mode:"1" Display:"0" Notnull:"0"`
	Modifytime time.Time `json:"modifytime" Mode:"3" Display:"0" Notnull:"0"`
	Logintime  time.Time `Mode:"9" Display:"0" Notnull:"0"`
	Roles      []string  `json:"roles" Mode:"3" Display:"1" Notnull:"0" Lenmin:"2" Lenmax:"64"`
	WorkEnv    string    `json:"workenv" Mode:"3"`
	//PermissionList []string `json:"permissionlist" Mode:"3" Display:"1" Notnull:"0" Lenmin:"6" Lenmax:"64"`
}

//ZDUserSession 用户session信息
type ZDUserSession struct {
	MgoUser
	//PermissionList 已环境名称作为key查找权限
	PermissionList *MgoRole
}

//MgoRole 角色的存储格式
type MgoRole struct {
	ID       bson.ObjectId  `json:"id" bson:"_id,omitempty"`
	Name     string         `json:"name" Mode:"3" Display:"1" Notnull:"1" Lenmin:"2" Lenmax:"32" Filter:"username"`
	SubRole  *MgoRole       `json:"subroles"`
	Describe string         `json:"describe" Mode:"3" Display:"1" Notnull:"0" `
	API      map[string]int `json:"api"`
	/*

	 */
	Environment map[string]*ZDPermissions `json:"environment"`
}

//ZDPermissions 权限内容
type ZDPermissions struct {
	API     map[string]int `json:"api"`
	Project map[string]int `json:"project"`
	Cluster map[string]int `json:"cluster"`
	App     map[string]int `json:"app"`
	Job     map[string]int `json:"job"`
	//JobsWhite 作业白名单，排序后保存。可以是目录，也可以指定作业全路径。
	// JobsWhite    []string `json:"jobswhite"`
	// JobsWhiteLen int      `json:"jobswhitelen"`
	// JobsBlack    []string `json:"jobsblack"`
	// JobsBlackLen int      `json:"jobsblacklen"`
}
