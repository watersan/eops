package auth

import (
	"encoding/json"
	"time"

	"gopkg.in/mgo.v2/bson"
	"opsv2.cn/opsv2/api/common"
)

//GetUserInfo 获取用户信息
/*
	verifytype定义：
	0: 登录login
	1: 根据用户名获取用户信息
	2: 根据keyid获取用户信息
*/
func GetUserInfo(econf *common.EopsConf, name string, verifytype int) (*ZDUserSession, string) {
	if verifytype != 0 {
		prefix := "user_"
		if verifytype == 2 {
			prefix = "keyid_"
		}
		if userinfo, err := econf.MapCache.Get(prefix + name); err == nil {
			econf.MapCache.SetExpire(prefix+name, time.Duration(econf.GOPT.SessionExpire)*time.Second)
			return userinfo.(*ZDUserSession), ""
		} else if verifytype == 1 {
			return nil, common.Expired
		}
	}
	pwd := ""
	skey := bson.M{}
	if verifytype == 2 {
		skey["keyid"] = name
	} else {
		skey["name"] = name
	}
	usersession := &ZDUserSession{}
	c := econf.Mgo.Coll(mgoUsers)
	err := c.Find(skey).One(&usersession.MgoUser)
	if err != nil {
		econf.LogCritical("mgdb[%s] query [%v]: %v\n", mgoUsers, skey, err)
		return nil, ""
	}
	usersession.PermissionList = &MgoRole{}
	permList := usersession.PermissionList
	//获取角色信息
	first := true
	//合并用户分配的角色权限
	for _, role := range usersession.MgoUser.Roles {
		result := RolesInfo(econf, role)
		if first {
			tmpbyte, _ := json.Marshal(result)
			//econf.LogDebug("[AUTH] role %s: %s", role, string(tmpbyte))
			json.Unmarshal(tmpbyte, permList)
			permList.Name = "permList"
			first = false
		} else {
			for k, v := range result.API {
				if vv := permList.API[k]; vv&v != v {
					permList.API[k] += v
				}
			}
			for env, perms := range result.Environment {
				if _, ok := permList.Environment[env]; !ok {
					tmpbyte, _ := json.Marshal(perms)
					permList.Environment[env] = &ZDPermissions{}
					json.Unmarshal(tmpbyte, permList.Environment[env])
				} else {
					for k, v := range perms.API {
						if vv := permList.Environment[env].API[k]; vv&v != v {
							permList.Environment[env].API[k] += v
						}
					}
					for k, v := range perms.Project {
						if vv, ok := permList.Environment[env].Project[k]; !ok {
							permList.Environment[env].Project[k] = v ^ vv
						}
					}
					for k, v := range perms.Cluster {
						if vv, ok := permList.Environment[env].Cluster[k]; !ok {
							permList.Environment[env].Cluster[k] = v ^ vv
						}
					}
					for k, v := range perms.App {
						if vv, ok := permList.Environment[env].App[k]; !ok {
							permList.Environment[env].App[k] = v ^ vv
						}
					}
					for k, v := range perms.Job {
						if vv, ok := permList.Environment[env].Job[k]; !ok {
							permList.Environment[env].Job[k] = v ^ vv
						}
					}
					// for _, v := range perms.JobsWhite {
					// 	if i := sort.SearchStrings(permList.Environment[env].JobsWhite, v); i == len(permList.Environment[env].JobsWhite) ||
					// 		permList.Environment[env].JobsWhite[i] != v {
					// 		permList.Environment[env].JobsWhite = append(permList.Environment[env].JobsWhite, v)
					// 		permList.Environment[env].JobsWhiteLen++
					// 	}
					// }
					// for _, v := range perms.JobsBlack {
					// 	if i := sort.SearchStrings(permList.Environment[env].JobsBlack, v); i == len(permList.Environment[env].JobsBlack) ||
					// 		permList.Environment[env].JobsBlack[i] != v {
					// 		permList.Environment[env].JobsBlack = append(permList.Environment[env].JobsBlack, v)
					// 		permList.Environment[env].JobsBlackLen++
					// 	}
					// }
					// sort.Strings(permList.Environment[env].JobsWhite)
					// sort.Strings(permList.Environment[env].JobsBlack)
				}
			}
		}
	}
	if permList.Name != "permList" {
		return nil, ""
	}
	if usersession.MgoUser.WorkEnv == "" {
		for _, env := range econf.GOPT.Environments {
			if _, ok := permList.Environment[env]; ok {
				usersession.MgoUser.WorkEnv = env
			}
		}
	}
	//登录接口校验密码
	if verifytype == 0 {
		pwd = usersession.MgoUser.Passwd
	}
	usersession.MgoUser.Passwd = ""
	econf.MapCache.Set("user_"+usersession.Name, usersession, time.Duration(econf.GOPT.SessionExpire)*time.Second)
	if usersession.KeyID != "" {
		econf.MapCache.Set("keyid_"+usersession.Name, usersession, time.Duration(econf.GOPT.SessionExpire)*time.Second)
	}
	return usersession, pwd
}

//RolesInfo 获取角色权限
func RolesInfo(econf *common.EopsConf, role string) *MgoRole {
	rinfo, err := econf.MapCache.Get("role_" + role)
	if err != nil {
		MDBRoles(econf)
		rinfo, _ = econf.MapCache.Get("role_" + role)
	}
	return rinfo.(*MgoRole)
}

//MDBRoles 缓存roles内容
func MDBRoles(econf *common.EopsConf) error {
	rRoles := []common.V{}
	var rolelist []string
	n, err := econf.Mgo.Query(mgoRoles, common.V{}, &rRoles, 2000, 1, nil)
	if n == 0 {
		return err
	}
	for _, role := range rRoles {
		currole := MgoRole{}
		role.Unmarshal(&currole)
		econf.MapCache.Set("role_"+currole.Name, &currole, time.Duration(econf.GOPT.MemExpire)*time.Second)
		rolelist = append(rolelist, currole.Name)
	}
	econf.MapCache.Set("roles", rolelist, time.Duration(econf.GOPT.MemExpire)*time.Second)
	return nil
}
