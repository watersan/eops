
// (function(){
//   'use strict';
function yunTpl(rfield,isList) {
  fields = [
    {
      name: "name",
      value: "名称",
      html: `<select ng-model="item.name" ng-disabled="updatereadonly">
            <option ng-repeat="yun in yunproviders" value="{{yun}}">{{yun}}</option>
            </select>`,
    },{
      name: "key",
      value: "公钥",
      islist: false,
      readonly: "updatereadonly",
    },{
      name: "secret",
      value: "私钥",
      islist: false,
      only: "isnew == true",
      readonly: "updatereadonly",
    },{
      name: "proxy",
      value: "代理",
      islist: true,
      pattern:"/^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}:?\d*)?$/",
      readonly: "updatereadonly",
    },{
      name: "proxytype",
      value: "代理类型",
      islist: true,
      readonly: "updatereadonly",
    },{
      name: "regions",
      value: "可用区",
      pattern: "^[a-z0-9A-Z\.\-\\/_,]+$",
      islist: false,
      readonly: "updatereadonly",
    },{
      name: "projectid",
      value: "项目ID",
      islist: false,
      only: "item.name == 'ucloud'",
      readonly: "updatereadonly",
    }
  ];
  if (rfield) {
    return fields;
  } else if (isList) {
    return MyTpl("yunlist",fields,"list");
  } else {
    return MyTpl("yunPost",fields,"two");
  }
}
function appPostTpl(field) {
  // var coltpl = `<div class="col-md-12"><div class="box box-info">` +
  //   boxHeaderTpl("{{action}}应用模版") +
  //   appFormTpl() + boxFootTpl()+`</div></div>`;
  // var tpl = `<div class="row">` + coltpl + `</div>`;
  // return tpl;
  fields = [
    {
      name: "name",
      value: "名称",
      pattern: '/^.{2,}$/',
      required: '请输入模版名称',
      readonly: "updatereadonly",
      placeholder: "请输入模版名称",
    },
    {
      name: "config",
      value: "配置文件",
      pattern: '/^http[s]?://[a-z0-9\.-:]+/.+$/',
      placeholder: "gitlab仓库名称",
    },
    {
      name: "configverify",
      value: "配置校验",
      placeholder: "校验配置文件的作业",
      job: true,
    },
    {
      name: "build",
      value: "构建脚本",
      job: true,
    },
    {
      name: "source",
      value: "安装源",
      html: `<div class="input-group">
          <div class="input-group-btn">
            <button type="button" class="btn btn-info form-control dropdown-toggle" data-toggle="dropdown">{{item.sourcetype}}
              <span class="fa fa-caret-down"></span></button>
            <ul class="dropdown-menu">
              <li><a href="#" ng-click="item.sourcetype='HTTP'">HTTP</a></li>
              <li><a href="#" ng-click="item.sourcetype='GIT'">GIT</a></li>
              <li><a href="#" ng-click="item.sourcetype='SVN'">SVN</a></li>
            </ul>
          </div>
          <input type="text" class="form-control" name="source" placeholder="" ng-model="item.source" ng-readonly="readonly">
        </div>
      `,
    },
    {
      name: "install",
      value: "安装脚本",
      job: true,
    },
    {
      name: "preinstall",
      value: "安装之前",
      job: true,
    },
    {
      name: "postinstall",
      value: "安装之后",
      job: true,
    },
    {
      name: "rollback",
      value: "回滚",
      job: true,
    },
    {
      name: "process",
      value: "进程名称",
    },
    {
      name: "port",
      value: "端口",
      placeholder: "例如: tcp/80,udp/53,mysql/tcp/3307",
      pattern: "/^(,?\w*/?(tcp|udp)/\d+)+$/",
    },
    {
      name: "additionconf",
      value: "附加配置",
      placeholder: "允许增加配置文件的相对路径",
    },
    {
      name: "confpath",
      value: "配置文件路径",
      placeholder: "配置文件的部署路径，如果不是以/开始，则基于安装目录的相对路径部署",
    },
    {
      name: "monitor",
      value: "监控",
    },
    {
      name: "depend",
      value: "依赖",
      html: `<div id="formdepend">
        <input type="text" class="form-control keywork-input" name="depend" placeholder="" ng-model="item.depend" ng-readonly="readonly">
        </div>
      `,
    },
  ];
  if (field) {
    return fields;
  } else {
    return MyTpl("app",fields,"two")
  }
}

function apptplListTpl() {
  fields = [
    {
      name: "name",
      value: "名称",
      islist:true,
    },
    {
      name: "source",
      value: "安装源",
      islist:true,
    },
    {
      name: "depend",
      value: "依赖",
      islist:true,
    },
    {
      name: "usage",
      value: "使用数",
      islist:true,
    },
  ];
  return MyTpl("apptpllist",fields,"list")
}

function RunLogTpl() {
  return MyTpl("runlog","","");
}

function ImportResource(src) {
  return MyTpl("import",src,"one")
}

function MyTpl(subtpl,obj,footdata) {
  var coltpl = `<div class="col-md-12"><div class="box box-info">` +
    boxHeaderTpl();
  switch (subtpl) {
    case "app":
    //name: "tt",value: "",pattern: "",required: "",readonly:"",placeholder:"",html:""
      coltpl += FormTpl(obj);
      break;
    case "apptpllist":
      coltpl += ListTpl("/cmdb/apptpl",obj);
      break;
    case "yunlist":
      coltpl += ListTpl("/cmdb/yunprovider",obj);
      break;
    case "yunPost":
      coltpl += FormTpl(obj);
      break;
    case "runlog":
      coltpl += jobRunLog();
      break;
    case "import":
      coltpl += ImportTpl(obj);
      break;
    default:

  }
  if (typeof(footdata) == "string" && footdata != "") {
    coltpl += boxFootTpl({kind: footdata});
  } else if (typeof(footdata) == "object") {
    coltpl += boxFootTpl(footdata);
  }
  coltpl += `</div></div>`;
  var tpl = `<div class="row">` + coltpl + `</div>`;
  return tpl;
}

/*
fields = [{name: "tt",value: "测试"},{}]
*/
function ListTpl(api, fields) {
  var tpl = `<div class="clearfix">
          <div style="float: right; padding:0 15px;">
            <div has-permission="zdopsapi:POST" style="float: left;margin-right:10px;padding-bottom:10px;">
              <a href="zdopsapi/add"  class="btn btn-sm">
                <i class="fa fa-plus fa-2x"></i>
              </a>
            </div>
          </div>
  			  <div class="input-group input-group-sm" style="width:180px;padding-bottom:10px;padding-left:10px;">
            <input type="text" class="form-control pull-right"
              placeholder="查找..." ng-keypress="enter($event)"
              ng-model="page.itemFilter">
            <div class="input-group-btn">
              <button type="submit" class="btn btn-default">
                <i class="fa fa-search"></i>
              </button>
            </div>
          </div>
        </div>
        <div ng-switch="tablestatus">
          <div class="box-body table-responsive no-padding" ng-switch-when="1">
            <table id="usertable" class="table table-bordered table-striped">
              <thead>
                <tr>`.replace(/zdopsapi/g,api);
  var th = `<th>
    <a href="#" ng-click="page.sortType = 'varname'; page.sortReverse = !page.sortReverse">namestr
      <span ng-show="page.sortType == 'varname' && !page.sortReverse" class="fa fa-caret-down">
      </span>
      <span ng-show="page.sortType == 'varname' && page.sortReverse" class="fa fa-caret-up">
      </span>
    </a>
  </th>
  `;

  var tr = `<td class="name ng-scope"> <p class="name-wrapper ng-binding" ng-bind="item.name"></p></td>`;

  for (var i=0; i<fields.length;i++) {
    var field = fields[i];
    if (field.islist) {
      if (i > 0) {
        tr += `<td ng-bind="item.`+field.name+`"></td>`;
      }
      tpl += th.replace(/varname/,field.name).replace(/namestr/,field.value);
    }
  }
  tpl += `<th class="numeric">操作</th></tr></thead><tbody>
    <tr ng-repeat="item in datalist | orderBy:sortType:sortReverse | filter:page.itemFilter | filter:paginate" select-on-click class="ng-scope">
     `+ tr;
  tpl += `
    <td class="numeric" valign="middle">
      <a href="zdopsapi/info/{{item._id}}" has-permission="zdopsapi:GET" class="btn btn-primary btn-xs">查看</a>&nbsp;&nbsp;
      <a href="zdopsapi/update/{{item._id}}" has-permission="zdopsapi:PUT" class="btn btn-primary btn-xs">修改</a>&nbsp;&nbsp;
      <a href="javascript:;" ng-click="deleteitem(item)" has-permission="zdopsapi:DELETE" class="btn btn-danger btn-xs">删除</a>
`.replace(/zdopsapi/g,api);
  if (api == "/cmdb/yunprovider") {
    tpl += `&nbsp;&nbsp;<a href="/cmdb/resource/yunimport/{{item._id}}" has-permission="/cmdb/resource/import:GET;/cmdb/resource/import:POST" class="btn btn-warning btn-xs">导入</a>`;
  } else if (api == "/cmdb/idcprovider") {
    tpl += `&nbsp;&nbsp;<a href="/cmdb/resource/idcimport/{{item._id}}" has-permission="/cmdb/resource/import:GET;/cmdb/resource/import:POST" class="btn btn-warning btn-xs">导入</a>`;
  }
  tpl += '</td>';
  tpl += `</tr></tbody></table></div>
    <div class="box-body table-responsive no-padding" ng-switch-when="2">
      <center style="padding:20px 0;">无数据 .....</center>
    </div></div>
  `;
  return tpl;
}

function boxHeaderTpl() {
  var tpl = `<div class="box-header with-border">
    <h3 class="box-title">{{title.curr}}</h3>
  </div>`;
  return tpl
}

/*
data = {
  kind: "",
  back: 'jumpto()',
  other: 'html code',
}
*/
function boxFootTpl(data) {
  var tpl = `<!-- /.box-body --><div class="box-footer">`;
  if (data.kind == 'one' || data.kind == 'two') {
    var back = 'jumpto()';
    if (typeof(data) == "object" && data.back) {
      back = data.back;
    }
    tpl += `<button type="button" class="btn btn-default" ng-click="`+back+`">返回</button>`;
  }
  if (data.kind == 'two') {
    tpl += `<button type="button" class="btn btn-info pull-right" ng-click="ok(!itemForm.$invalid)" ng-disabled="readonly">保存</button>`;
  } else if (data.kind == 'list') {
    tpl += `          <div class="col-sm-6">
                <ul uib-pagination total-items="page.totalItems" ng-model="page.currentPage"
                  max-size="page.maxSize" boundary-link-numbers="true" rotate="false"
                  items-per-page="page.numPerPage" class="pagination-sm pull-left">
                </ul>
              </div>
              <div class="col-sm-6">
                <label for="usertable_length" class="pull-right" style="margin:20px 0;">每页显示的行数：
                  <select name="usertable_length" ng-model="page.numPerPage" aria-controls="usertable">
                    <option value="5" ng-selected="true" selected="yes">5</option>
                    <option value="10">10</option>
                    <option value="25">25</option>
                    <option value="50">50</option>
                    <option value="100">100</option>
                  </select>
                </label>
              </div>
              `;
  } else {
    if (typeof(data) == "object" && data.hasOwnProperty("other")) {
      tpl += data.other;
    }
  }
  tpl += '</div>';
  return tpl;
}

function ImportTpl(src) {
  var tpl = `      <div class="box-body">
          <div class="col-sm-10">
            <div class="row">
              <div class="col-sm-6">
                <select ng-model="import.cdenv">
                  <option ng-repeat="(k,v) in Environment" value="{{k}}">{{v}}</option>
                </select>
              </div>
              <div class="col-sm-4">
                <button type="button" class="btn btn-info" ng-click="importhost()">导入</button>
              </div>
            </div>
          </div>`;
  if (src == "idc") {
    tpl += `          <div class="col-sm-10" ng-if="importsrc == 'idc'">
                <div class="row">
                  <div class="col-sm-8">
                    <input type="text" class="form-control input-sm" name="ipstart" ng-model="import.ipstart">-
                    <input type="text" class="form-control input-sm" name="ipend" ng-model="import.ipend">
                  </div>
                  <div class="col-sm-2">
                    <button type="button" class="btn btn-info" ng-click="scanhost()">扫描</button>
                  </div>
                </div>
                <div class="row">
                  <div class="col-sm-8">
                    <span class="btn btn-primary fileinput-button">
                      <i class="glyphicon glyphicon-plus"></i>
                      <!--  -->
                      <span>导入CSV</span>
                      <!-- The file input field used as target for the file upload widget -->
                      <input id="fileupload" type="file" name="files" class="btn btn-primary" ng-model="upfilename">
                    </span>
                    <!-- <input id="fileupload" type="file" name="files" class="btn btn-primary" ng-model="upfilename"> -->
                    <div id="progress" class="progress">
                      <div class="progress-bar progress-bar-success" aria-valuemin="0" aria-valuemax="100"></div>
                    </div>
                  </div>
                </div>
              </div>
  `;
  }
  tpl += '</div>';
  return tpl;
}
/*
fields = [{name: "tt",value: "",pattern: "",required: true,readonly:"",placeholder:"",html:""},]
*/
function FormTpl(fields) {
  var tpl = `<form class="form-horizontal" name="itemForm">
    <div class="box-body">
  `;
  var formtpl = "";
  for (var i = 0; i<fields.length;i++) {
    field = fields[i];
    tpl += `<div class="form-group"`;
    if (field.only) {
      tpl += ' ng-if="'+field.only+'"';
    }

    if ("required" in field && field.required != "") {
      tpl += ` ng-class="{'has-success':!itemForm.varname.$invalid, 'has-warning':itemForm.varname.$error.required,'has-error':(!itemForm.varname.$error.required) && itemForm.varname.$invalid}">`.replace(/varname/g,field.name);
    } else {
      tpl += '>';
    }
    tpl += `<label class="col-sm-2 control-label">`+field.value+'</label>';
    tpl += `<div class="col-sm-10"><div class="row">`;
    if (field.hasOwnProperty("job") && field.job == true) {
      tpl += '<div class="col-sm-8" id="job_'+field.name+'">';
    } else {
      tpl += '<div class="col-sm-8">';
    }
    if ("html" in field && field.html != "") {
      tpl += field.html;
      tpl += '</div></div></div></div>';
      continue;
    }
    tpl +=`<input type="text" class="form-control input-sm keywork-input" name="varname" ng-model="item.varname"`;
    if ("placeholder" in field && field.placeholder != "") {
      tpl += ` placeholder="` + field.placeholder + '"';
    }
    if ("required" in field && field.required != "") {
      tpl += " required";
    }
    if ("pattern" in field && field.pattern != "") {
      tpl += ' ng-pattern="' + field.pattern + '"';
    }
    var readonly = ' ng-readonly="readonly"';
    if ("readonly" in field && field.readonly != "") {
      readonly = ' ng-readonly="'+field.readonly+'"';
    }
    tpl += " " + readonly + "></div>";
    if ("pattern" in field && field.pattern != "") {
      var tmp = `<div class="col-sm-2">
          <span class="help-block" ng-show="itemForm.varname.$error.required"><i class="fa fa-bell-o"></i>`+field.required+`</span>
          <span class="help-block" ng-show="!itemForm.varname.$invalid"><i class="fa fa-check"></i></span>
        </div>
      `;

      // tpl += tmp.replace(/varname/g,field.name)
    }
    tpl += '</div></div></div>';
    tpl = tpl.replace(/varname/g,field.name);
  }
  tpl += '</form>';
  return tpl;
}

function basetpl(obj) {
  //console.log(obj);
  var tpl = `<div class="row">
      <h1 style="font-size: 50px"><center>欢迎进入运维控制台!</center> </h1>
  </div>`;
  return tpl;
}

function jobRunLog() {
  var tpl =`<div class="box-body empty" ng-switch="showme.curr">
          <!-- 作业日志 -->
          <div id="itemlist" class="animate-switch" ng-switch-when="runlogs">
            <div class="input-group input-group-sm" style="width:180px;padding-bottom:10px;padding-left:10px;">
              <input type="text" class="form-control pull-right"
                placeholder="查找..."
                ng-model="page.itemFilter">
              <div class="input-group-btn">
                <button type="submit" class="btn btn-default">
                  <i class="fa fa-search"></i>
                </button>
              </div>
            </div>

            <table id="usertable" class="table table-bordered table-striped with-border">
              <thead>
                <tr>
                  <th>
                    <a href="#" ng-click="page.sortType = 'name'; page.sortReverse = !page.sortReverse">
                      作业名称
                      <span ng-show="page.sortType == 'name' && !page.sortReverse"
                        class="fa fa-caret-down">
                      </span>
                      <span ng-show="page.sortType == 'name' && page.sortReverse"
                        class="fa fa-caret-up">
                      </span>
                    </a>
                  </th>
                  <th>
                    执行人
                  </th>
                  <th>
                    执行时间
                  </th>
                  <th>
                    作业参数
                  </th>
                  <th>
                    执行结果
                  </th>
                  <th>
                    详细信息
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr ng-repeat="item in datalist | orderBy:page.sortType:page.sortReverse | filter:page.itemFilter | filter:paginate" select-on-click class="ng-scope" ng-style="failedStyle(item)">
                  <td class="name ng-scope" ng-bind="item.taskname"></td>
                  <td ng-bind="item.operator"></td>
                  <td>{{newdate(item.begintime).format("%Y年%m月%d日 %H:%M:%S")}}</td>
                  <td ng-bind="item.taskattr.argv"></td>
                  <td>
                    <span ng-if="item.failed == 0">成功</span>
                    <span ng-if="item.failed > 0 && item.progress == item.hostnum">失败：{{ item.failed }}</span>
                    <span ng-if="item.failed > 0 && item.progress != item.hostnum">以完成：{{item.progress}}/{{item.hostnum}}</span>
                  </td>
                  <td>
                    <button type="button" class="btn btn-info btn-xs" ng-click="projectop(item,'hostslogs','作业详细日志:')" has-permission="/jobs/runlogs|GET">详细</button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div class="col-sm-4">
              <ul uib-pagination total-items="page.totalItems" ng-model="page.currentPage"
                max-size="page.maxSize" boundary-link-numbers="true" rotate="false"
                items-per-page="page.numPerPage" class="pagination-sm pull-left">
              </ul>
            </div>
            <div class="col-sm-8">
              <label for="usertable_length" class="pull-right" style="margin:20px 0;">每页显示的行数：
                <select name="usertable_length" ng-model="page.numPerPage" aria-controls="usertable">
                  <option value="5" ng-selected="true" selected="yes">5</option>
                  <option value="10">10</option>
                  <option value="25">25</option>
                  <option value="50">50</option>
                  <option value="100">100</option>
                </select>
              </label>
            </div>
          </div>
          <!-- 执行记录 -->
          <div id="itemlist" class="animate-switch" ng-switch-when="hostslogs">
            <script type="text/ng-template" id="taskattr.html">
              <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                  <span ng-click="$ctrl.cancel()">&times;</span></button>
                <h4 class="modal-title">任务输出</h4>
              </div>
              <div class="modal-body">
                <label class="control-label">标准输出：</lable><br/>
                <p ng-bind-html="$ctrl.taskst.stdout"></p>
                <label class="control-label">错误输出：</lable><br/>
                <p ng-bind-html="$ctrl.taskst.stderr"></p>
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-primary pull-right" ng-click="$ctrl.cancel()">关闭</button>
              </div>
            </script>

            <div class="row">
              <label class="col-sm-8 control-label">作业： {{loginfo.taskattr.path}}/{{loginfo.taskattr.name}} {{loginfo.taskattr.argv}}</label>
              <label class="col-sm-4 control-label">失败数： {{loginfo.failed}}</label>
            </div>
            <div class="row">
              <label class="col-sm-4 control-label">执行人： {{loginfo.operator}}</label>
              <label class="col-sm-4 control-label">总执行时间： {{runtime(loginfo)}}s</label>
            </div>
            <div class="row">
              <label class="col-sm-4 control-label">超时时间： {{loginfo.taskattr.timeout}}</label>
              <label class="col-sm-4 control-label">优先级： {{loginfo.taskattr.priority}}</label>
              <label class="col-sm-4 control-label">用户： {{loginfo.taskattr.user}}</label>
            </div>
            <div class="row">
              <label class="col-sm-12 control-label">作业时间： {{newdate(loginfo.begintime).format("%Y年%m月%d日 %H:%M:%S")}}</label>
            </div>
            <table id="usertable" class="table table-bordered table-striped">
              <thead>
                <tr>
                  <th>
                    <a href="#" ng-click="page.sortType = 'hostname'; page.sortReverse = !page.sortReverse">
                      主机名称
                      <span ng-show="page.sortType == 'hostname' && !page.sortReverse"
                        class="fa fa-caret-down">
                      </span>
                      <span ng-show="page.sortType == 'hostname' && page.sortReverse"
                        class="fa fa-caret-up">
                      </span>
                    </a>
                  </th>
                  <th>
                    别名
                  </th>
                  <th>
                    作业
                  </th>
                  <th>
                    状态码
                    <span ng-show="page.sortType == 'code' && !page.sortReverse"
                      class="fa fa-caret-down">
                    </span>
                    <span ng-show="page.sortType == 'code' && page.sortReverse"
                      class="fa fa-caret-up">
                    </span>
                  </a>
                  </th>
                  <th>
                    执行时间
                  </th>
                  <th>
                    资源使用
                  </th>
                  <th>
                    输出信息
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr data-ng-repeat-start="item in loginfo.hostlogs | orderBy:page.sortType:page.sortReverse | filter:paginate" select-on-click class="ng-scope" ng-style="failedStyleHlog(item.jobs[0])">
                  <td class="name ng-scope" ng-bind="item.hostname" rowspan="{{item.jobs.length}}"></td>
                  <td ng-bind="item.jobs[0].alias"></td>
                  <td ng-bind="item.jobs[0].name"></td>
                  <td ng-bind="item.jobs[0].code"></td>
                  <td>{{item.jobs[0].runtime/1000}}</td>
                  <td>{{(item.jobs[0].usedcpu/100).toFixed(2)}}%|{{item.jobs[0].usedmem}}MB</td>
                  <td>
                    <button type="button" class="btn btn-info btn-xs"
                      ng-click="joblogsdetail(item.jobs[0])" ng-if="item.jobs[0].stdout != '' || item.jobs[0].stderr !=''">
                      详细
                    </button>
                  </td>
                </tr>
                <tr data-ng-repeat-end ng-repeat="job in item.jobs" select-on-click class="ng-scope" ng-style="failedStyleHlog(job)" ng-hide="$first">
                  <td ng-bind="job.alias"></td>
                  <td ng-bind="job.name"></td>
                  <td ng-bind="job.code"></td>
                  <td>{{job.runtime/1000}}</td>
                  <td>{{(job.usedcpu/100).toFixed(2)}}%|{{job.usedmem}}MB</td>
                  <td>
                    <button type="button" class="btn btn-info btn-xs"
                      ng-click="joblogsdetail(job)" ng-if="job.stdout != '' || job.stderr !=''">
                      详细
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
            <div class="col-sm-4">
              <ul uib-pagination total-items="page.totalItems" ng-model="page.currentPage"
                max-size="page.maxSize" boundary-link-numbers="true" rotate="false"
                items-per-page="page.numPerPage" class="pagination-sm pull-left">
              </ul>
            </div>
            <div class="col-sm-8">
              <label for="usertable_length" class="pull-right" style="margin:20px 0;">每页显示的行数：
                <select name="usertable_length" ng-model="page.numPerPage" aria-controls="usertable">
                  <option value="5" ng-selected="true" selected="yes">5</option>
                  <option value="10">10</option>
                  <option value="25">25</option>
                  <option value="50">50</option>
                  <option value="100">100</option>
                </select>
              </label>
            </div>
          </div>
        </div>
        <div class="box-footer" ng-switch="showme.curr">
          <button type="button" class="btn btn-default" ng-click="projectop('',showme.history, title.history)" ng-switch-when="hostslogs">返回</button>
        </div>
`;
  return tpl;
}
// })();
