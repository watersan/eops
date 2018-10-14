# EOPS 初始化
<pre><code>use eops
db.createUser({user: "eopsadmin",pwd: "h39igXFCuLZHfuPq7CLx",customData:{},roles:["dbOwner"]})
db.createUser({user: "eops",pwd: "lejkkgt0uPXw6ArNGJKZ",customData:{},roles:["dbOwner"]})
db.auth("eops","lejkkgt0uPXw6ArNGJKZ")
db.users.getIndexes()
密码是: 12345678
db.zd_users.createIndex({name: 1},{unique: 1,})
db.zd_jobs.ensureIndex({"name":1,"path":1},{"unique":true})
db.zd_users.insert({
"name" : "dale_di@126.com", "passwd" : "1709fbcb842e2ad8c29718b74ac2875b8c958364bac49ae7a9fa055c96b7b", "mobile" : "18610116661", "fullname" : "daledi", "createtime" : new Date(), "modifytime" : new Date(), "logintime" : "", "roles" : [ "admin","all" ]
})
db.zd_users.insert({
"name" : "test@126.com", "passwd" : "1709fbcb842e2ad8c29718b74ac2875b8c958364bac49ae7a9fa055c96b7b", "mobile" : "18610116662", "fullname" : "test", "createtime" : new Date(), "modifytime" : new Date(), "logintime" : "", "roles" : [ ]
})
db.zd_roles.createIndex({name: 1},{unique: 1,})
db.zd_apptpl.createIndex({name: 1},{unique: 1,})
db.zd_project.createIndex({name: 1},{unique: 1,})
db.zd_cluster.createIndex({name: 1,project: 1},{unique: 1,})
db.zd_app.createIndex({name: 1,cluster: 1},{unique: 1,})
db.zd_roles.remove({});
db.zd_roles.insert({
  name:"admin",
  permissions:{
    api: {
      "/auth/user:GET":true,
      "/auth/user:PUT":true,
      "/auth/user:POST":true,
      "/auth/user:DELETE":true,
      "/auth/user2off:PUT":true,
      "/cmdb/project:POST":true,
      "/cmdb/project:GET":true,
      "/cmdb/project:PUT":true,
      "/cmdb/project:DELETE":true,
    },
    environment: ["develop","test","official"],
  },
});
db.zd_roles.insert({
  name:"all",
  permissions:{
    api: {
      "/auth/login:GET":true,
      "/auth/pwd:PUT":true,
    },
    environment: [],
  },
})
db.zd_app.insert({"name" : "bbs", "cluster" : "598596f25081faaeac8e5eed", "depend" : [ "598599185081faaeac8e5f07" ], "from" : "", "type" : "apptpl", "appconf" : [ ], "hosts" : [ ], "deploy" : 1, "hostscount" : 3})
db.zd_app.insert({"name" : "php", "cluster" : "598596f25081faaeac8e5eed", "depend" : [ ], "from" : "", "type" : "apptpl", "appconf" : [ ], "hosts" : [ ], "deploy" : 1, "hostscount" : 3})
db.zd_app.insert({"name" : "nginx", "cluster" : "598596f25081faaeac8e5eed", "depend" : [ ], "from" : "", "type" : "apptpl", "appconf" : [ ], "hosts" : [ ], "deploy" : 1, "hostscount" : 3})
db.zd_cluster.insert({"name" : "MySQL主", "describe" : "", "project" : "598580495081faaeac8e5ecb", "depend" : [ ], "shared" : true, "opser" : "daledi", "secondopser" : "daledi"})
db.zd_cluster.insert({"name" : "MySQL从", "describe" : "", "project" : "598580495081faaeac8e5ecb", "depend" : [ "59868e5aac91868bcb6029e1" ], "shared" : true, "opser" : "daledi", "secondopser" : "daledi"})
db.zd_cluster.insert({"name" : "负载均衡", "describe" : "", "project" : "598580495081faaeac8e5ecb", "depend" : [ ], "shared" : true, "opser" : "daledi", "secondopser" : "daledi"})
db.zd_cluster.insert({"name" : "Redis", "describe" : "", "project" : "598580495081faaeac8e5ecb", "depend" : [ ], "shared" : true, "opser" : "daledi", "secondopser" : "daledi"})
db.zd_basedata.insert({key: "isplist",type: "isp",value: {"telecom": "电信","unicom": "联通","cernet": "教育","cmnet": "移动", "crtc": "铁通"}})
</code></pre>

# 所有API返回的数据结构：
<pre><code>输出格式：
{
  "code":int           //状态码，0代表成功，其他请参考common.InitErrCode
  "dataList":[]        //数据列表，
  "totalRecord":int    //总记录数
  "message":str        //错误信息
}
</pre></code>

# 所有mongodb表结构：
<pre><code>
所有表都包含以下字段：
Status：以下两个值的定义全局统一
  -1000: 被删除
  -2000: 隐藏，系统自动创建。不允许前端管理
Modifytime:
Createtime:
</pre></code>

# 待完成内容
* 项目，集群，应用，配置，模版等名称不允许应中文。方便脚本引用处理。
* 项目，集群，应用的名称不允许修改。会破坏依赖关系。（待考虑）
* cmdb在删除时，需要对依赖进行判断，有依赖的不能删除。必须从依赖最下层开始删除。

# bug列表
* 在有过滤条件的情况下，全选功能还需要调试。

# 待做的功能:
### 变量处理
* 配置文件要能支持变量：{{variablename}}。这个变量可以在应用中配置，也要能够自动支持拓扑图
中的依赖的集群名称。
* 集群名称可以作为变量名，值是集群下应用的ip。变量名：集群名称+"-port"=应用IP+":"+应用的port。
变量名：[]集群名称，值为集群下所有应用的ip列表，同样支持端口模式。
* 应用相关的变量要随任务下发。在任务执行时进行变量替换。

### 部署
* 全局配置环境列表，并配置环境部署顺序。比如先部署测试环境，再部署正式环境。
* 资源，作业和应用（包括应用变量）要有环境属性。
* 要严格按部署顺序进行部署。
* 部署是针对应用的。允许选择部署的设备。
* 实现方法：
  1. 应用模板的配置中，将


### 作业
* 可以为作业分配允许执行的环境。默认情况下作业要继承创建者的环境配置。
* 作业在执行时，要指定环境。作业修改后，也要遵循部署顺序执行一次，然后才可以随时在正式环境执行。

### 内存缓存
* 增加转存本地文件和加载功能
