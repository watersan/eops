
APIPermNoCheck   = -1
APIPermVerified    = 0x0
APIPermAllow     = 0x1 //api
APIPermJobRun    = 0x2 //jobrun,jobread
//APIPermJobAdmin   = 0x4 //jobread
//APIPermJobModify = 0x8 //jobmodify
//project 权限，包括集群和应用
APIPermPCADeploy = 0x16 //deploy
APIPermPCAConfig = 0x32 //modify config
APIPermPCARead   = 0x64 //read config
//APIPermPCAAdmin    = 0x128
所有权限key都通过RawQuery获取
# 作业系统：
* 作业，作业流可配权限：api，jobrun+cdenv，PCARead。权限key：path1,path1/path2,path1/path2/name,flowname
* 快捷方式：jobrun+cdenv，PCARead。权限key：path1,path1/path2,path1/path2/name
# 配置管理系统：
* 项目：可配权限：api，PCARead+anycdenv。权限key：name,id。只允许admin可增改删。
* 集群：可配权限：api，PCARead+anycdenv。权限key：project+name,id
* 应用：可配权限：api，PCADeploy+cdenv，PCAConfig+cdenv，PCARead+anycdenv。权限key：project+cluster+name,id
# 资源管理
* 可配权限：api+cdenv
