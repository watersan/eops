
预定义5个部署环境：
1. CDEnvBuild   = "build"
2. CDEnvDevelop = "develop"
3. CDEnvTest    = "test"
4. CDEnvPre     = "pre-release"
5. CDEnvRelease = "release"
自定义的部署环境也是从上面中选择，不管选择几个环境，部署任务的执行顺序也是按此进行。
为满足自动化部署需求，作业系统增加自定义回调接口，用于自动触发下一个环境的部署。
部署任务的执行结果是从作业系统的日志中获取，因此，部署系统要保存环境和作业id的关系。

tpl的处理流程：
build，上传，更新版本，创建所依赖app的部署任务
tplconf的处理流程：
更新版本，创建所依赖app的部署任务
tpl和tplconf在部署时，要同步部署每个环境的所有依赖app。避免有的app更新成功，而部分app因使用的模块或配置问题导致更新失败

EnvDevelopAgent
EnvTestAgent
EnvReleaseAgent
不同环境分别制定一个agent。job的执行由控制台转发给对应环境的agent，再由agent下发到执行的主机。
job的回调api通过agent转发给控制台。

webhook
if 模板 {
  获取模板信息
} else if "cmdb-" {
  if 配置 {
    设置appconfver
  } else {
    设置appver
  }
  创建部署任务
} else {
  查找cmdb-apps中source或config的值为GitHTTP的记录
  if 有记录 {
    if config == GitHTTP {
      设置appconfver
    } else {
      设置appver
    }
    创建部署任务
    return
  } else {
    查找cmdb-apptpl中source或config的值为GitHTTP的记录
  }
}
if 模板 {
  if config == GitHTTP {
    同步依赖应用的config的版本，并为每个应用创建部署任务
  } else {
    执行构建
  }
} else {
  执行部署任务
}
