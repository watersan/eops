package cmdb

import (
	"opsv2.cn/opsv2/api/auth"
	"opsv2.cn/opsv2/api/common"
)

//RouteList api路由表
func RouteList(route *common.APIRoute) {
	route.API["/cmdb/project|POST"] = common.APIInfo{
		Handler: AddItem,
		Name:    "添加项目",
		Perm:    auth.APIPermAllow,
	}
	route.API["/cmdb/project|PUT"] = common.APIInfo{
		Handler: UpdateItem,
		Name:    "修改项目",
		Perm:    auth.APIPermAllow,
	}
	route.API["/cmdb/project|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看项目",
		Perm:    auth.APIPermAllow,
	}
	route.API["/cmdb/project|DELETE"] = common.APIInfo{
		Handler: DeleteItem,
		Name:    "删除项目",
		Perm:    auth.APIPermAllow,
	}
	route.API["/cmdb/cluster|POST"] = common.APIInfo{Handler: AddItem, Name: "添加集群", Perm: auth.APIPermAllow}
	route.API["/cmdb/cluster|PUT"] = common.APIInfo{Handler: UpdateItem, Name: "修改集群", Perm: auth.APIPermAllow}
	route.API["/cmdb/cluster|GET"] = common.APIInfo{Handler: Items, Name: "查看集群", Perm: auth.APIPermAllow}
	route.API["/cmdb/cluster|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除集群", Perm: auth.APIPermAllow}
	route.API["/cmdb/apptpl|POST"] = common.APIInfo{Handler: AddItem, Name: "添加应用模版", Perm: auth.APIPermAllow}
	route.API["/cmdb/apptpl|PUT"] = common.APIInfo{Handler: UpdateItem, Name: "修改应用模版", Perm: auth.APIPermAllow}
	route.API["/cmdb/apptpl|GET"] = common.APIInfo{Handler: Items, Name: "查看应用模版", Perm: auth.APIPermAllow}
	route.API["/cmdb/apptpl|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除应用模版", Perm: auth.APIPermAllow}
	route.API["/cmdb/apps|POST"] = common.APIInfo{Handler: AddItem, Name: "添加应用", Perm: auth.APIPermAllow}
	route.API["/cmdb/apps|PUT"] = common.APIInfo{Handler: UpdateItem, Name: "修改应用", Perm: auth.APIPermAllow}
	route.API["/cmdb/apps|GET"] = common.APIInfo{Handler: Items, Name: "查看应用", Perm: auth.APIPermAllow}
	route.API["/cmdb/apps|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除应用", Perm: auth.APIPermAllow}
	route.API["/cmdb/appver|PUT"] = common.APIInfo{
		Handler: UpdateItem,
		Name:    "更新应用变量",
		Perm:    auth.APIPermAllow | auth.APIPermPCAConfig | auth.APIPermPCARead | auth.APIPermCDEnv,
	}
	route.API["/cmdb/appver|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看应用变量",
		Perm:    auth.APIPermAllow | auth.APIPermPCARead | auth.APIPermCDEnv,
	}
	route.API["/cmdb/appconf|POST"] = common.APIInfo{
		Handler: AddItem,
		Name:    "添加应用配置",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/appconf|PUT"] = common.APIInfo{
		Handler: UpdateItem,
		Name:    "修改应用配置",
		Perm:    auth.APIPermAllow | auth.APIPermPCAConfig | auth.APIPermPCARead | auth.APIPermCDEnv,
	}
	route.API["/cmdb/appconf|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看应用配置",
		Perm:    auth.APIPermAllow | auth.APIPermPCARead | auth.APIPermCDEnv,
	}
	route.API["/cmdb/appconf|DELETE"] = common.APIInfo{
		Handler: DeleteItem,
		Name:    "删除应用配置",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/resource|POST"] = common.APIInfo{
		Handler: AddResource,
		Name:    "添加资源",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/resource|PUT"] = common.APIInfo{
		Handler: UpdateResource,
		Name:    "修改资源",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/resource|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看资源",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/resource|DELETE"] = common.APIInfo{
		Handler: DeleteItem,
		Name:    "删除资源",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/resourceall|GET"] = common.APIInfo{
		Handler: Items,
		Name:    "查看资源",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/yunprovider|POST"] = common.APIInfo{Handler: AddItem, Name: "添加云服务商", Perm: auth.APIPermAllow}
	route.API["/cmdb/yunprovider|PUT"] = common.APIInfo{Handler: UpdateItem, Name: "修改云服务商", Perm: auth.APIPermAllow}
	route.API["/cmdb/yunprovider|GET"] = common.APIInfo{Handler: Items, Name: "查看云服务商", Perm: auth.APIPermAllow}
	route.API["/cmdb/yunprovider|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除云服务商", Perm: auth.APIPermAllow}
	route.API["/cmdb/idcprovider|POST"] = common.APIInfo{Handler: AddItem, Name: "添加IDC", Perm: auth.APIPermAllow}
	route.API["/cmdb/resource/import|GET"] = common.APIInfo{Handler: ImportFromYun, Name: "导入资源", Perm: auth.APIPermAllow}
	route.API["/cmdb/resource/import|POST"] = common.APIInfo{Handler: ImportFromCSV, Name: "从CSV导入资源", Perm: auth.APIPermAllow}
	route.API["/cmdb/idcprovider|PUT"] = common.APIInfo{Handler: UpdateItem, Name: "修改IDC", Perm: auth.APIPermAllow}
	route.API["/cmdb/idcprovider|GET"] = common.APIInfo{Handler: Items, Name: "查看IDC", Perm: auth.APIPermAllow}
	route.API["/cmdb/idcprovider|DELETE"] = common.APIInfo{Handler: DeleteItem, Name: "删除IDC", Perm: auth.APIPermAllow}
	route.API["/cmdb/isp|GET"] = common.APIInfo{Handler: ISPList, Name: "ISP列表", Perm: auth.APIPermVerified}
	route.API["/cmdb/svcprovider|GET"] = common.APIInfo{Handler: SVCProvider, Name: "服务商列表", Perm: auth.APIPermVerified}
	route.API["/cmdb/allocresource|PUT"] = common.APIInfo{
		Handler: AllocResource,
		Name:    "为应用分配资源",
		Perm:    auth.APIPermAllow | auth.APIPermCDEnv,
	}
	route.API["/cmdb/projecttree|GET"] = common.APIInfo{Handler: ProjectTree, Name: "项目树", Perm: auth.APIPermVerified}

}
