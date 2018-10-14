app.controller('TaskManageCtrl', ['$uibModalInstance','items','taskst',function ($uibModalInstance, items,taskst) {
  var $ctrl = this;
  $ctrl.step = items;
  $ctrl.taskst = taskst;
  //console.log($ctrl.step);
  $ctrl.ok = function (isValid) {
    if (!isValid) {
      layer.msg('表单填写不正确，请仔细检查！', {"icon": 5});
      return;
    }
    $uibModalInstance.close($ctrl.step);
  };

  $ctrl.delete = function() {
    var step = {
      "index": $ctrl.step.index,
      "delete": true,
    };
    $uibModalInstance.close(step);
  }
  $ctrl.addstep = function(isRescue) {
    $uibModalInstance.close({"isRescue": isRescue});
  }
  $ctrl.cancel = function () {
    $uibModalInstance.dismiss('cancel');
  };
}]);

app.controller('JobsController',['$rootScope','$scope','BaseUrl','$http','$location',
  '$compile','$routeParams','permissions','$interval','$route','$uibModal',
  function ($rootScope,$scope,BaseUrl,$http,$location,$compile,$routeParams,permissions,$interval,$route,$uibModal) {
    $scope.item = {};
    $scope.page = {};
    $scope.page.totalItems = 0;
    $scope.page.currentPage = 1;
    $scope.page.maxSize = 5;
    $scope.page.itemFilter = '';
    $scope.page.numPerPage = "10";
    $scope.page.sortType     = 'name';
    $scope.page.sortReverse  = false;
    $scope.datalist = [];
    $scope.readonly = false;
    $scope.path1readonly = false;
    $scope.path2readonly = false;
    $scope.ready = false;
    $scope.showme = {curr:"",history:""};
    $scope.title = {curr:"",history:""};
    $scope.showme.curr = "jobs";
    $scope.title.curr = "作业";
    $scope.codeChange = false;

    var initcode = "#作者："+$rootScope.username+"\n#版本：V1.0\n#描述：脚本用途说明\n\n";
    var stype ={
      "sh": {"mode": "shell", "initcode": "#!/bin/bash\n"},
      "py": {"mode": "python", "initcode": "#!/usr/bin/python\n# -*- coding: utf-8 -*-\n"},
      "pl": {"mode": "perl", "initcode": "#!/usr/bin/perl\n"},
      "yml": {"mode": "yaml", "initcode": ""},
    };
    var url = BaseUrl + "/jobs/jobs";
    var req = {
      "method": 'GET',
      "url": BaseUrl+'/cmdb/apps',
    };
    var stop;
    var mycode;
    $scope.apps = {};
    $scope.clusters = {};
    $scope.projects = {};
    var cluster = {};
    var project = {};
    $http(req).then(function (response) {
      var data = response.data;
      if(data.code==0){
        for (var id =0 ;id < data.totalRecord; id++) {
          item = data.dataList[id]
          $scope.apps[item._id] = item;
          if (cluster[item.cluster] == undefined) {
            cluster[item.cluster] = []
          };
          cluster[item.cluster].push({
            "id": item._id,
            "text": item.name,
          });
        }
      } else if (data.code == 1199) {
        $rootScope.loginform();
      }
    }).then(function (){
      req.url = BaseUrl+'/cmdb/cluster';
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          for (var id =0 ;id < data.totalRecord; id++) {
            item = data.dataList[id]
            $scope.clusters[item._id] = item;
            if (project[item.project] == undefined) {
              project[item.project] = []
            };
            var nodes = [];
            if (item._id in cluster) {
              nodes = cluster[item._id];
            }
            project[item.project].push({
              "id": item._id,
              "text": item.name,
              "nodes": nodes,
            });
          }
        } else if (data.code == 1199) {
          $rootScope.loginform();
        }
      }).then(function () {
        req.url = BaseUrl+'/cmdb/project';
        $scope.projecttree = [];
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            for (var id =0 ;id < data.totalRecord; id++) {
              item = data.dataList[id]
              $scope.projects[item._id] = item;
              var nodes = [];
              if (item._id in project) {
                nodes = project[item._id];
              }
              $scope.projecttree.push({
                "id": item._id,
                "text": item.name,
                "nodes": nodes,
              });
            }
          } else if (data.code == 1199) {
            $rootScope.loginform();
          }
        }).then(function () {
          $scope.ready = true;
        });
      });
    });

    $scope.stopWait = function() {
      if (angular.isDefined(stop)) {
        $interval.cancel(stop);
      }
    }

    $scope.projectop = function(data, mode, title) {
      $scope.showme.history = $scope.showme.curr;
      $scope.showme.curr = mode;
      if (mode != "" && title != "") {
        $scope.title.history = $scope.title.curr;
        $scope.title.curr = title;
      }
      if ($scope.showme.curr == "runjob") {
        $scope.jobinfo = $scope.item;
        var taskattr = {
          "jobid": $scope.jobinfo._id,
          "name": $scope.jobinfo.name,
          "path": $scope.jobinfo.path,
          "commitid": $scope.jobinfo.commitid,
          "timeout": 60,
          "pri": 19,
          "user": "root",
        };
        //$.extend(taskattr,$scope.item);

        $scope.item = {
        //  taskid: id,
          "name": $scope.jobinfo.name,
          "taskattr": taskattr,
          "target": "",
          "project": [],
          "cluster": [],
          "app": [],
          "list": [],
        };

        //getprojects();
        stop = $interval(function() {
          if (document.getElementById("form_target_project") != null && $scope.ready) {
            $scope.stopWait();
            $("#form_target_project").select2({
              "data": $scope.projecttree,
              //theme: "classic",
              "placeholder": '选择项目',
              "allowClear": true,
            });
            $("#form_target_cluster").select2({
              "data": [],
              //theme: "classic",
              "placeholder": '选择集群',
            });
            $("#form_target_apps").select2({
              "data": [],
              //theme: "classic",
              "placeholder": '选择应用',
            });
            $("#form_target_project").on("change", function(e) {
              var data = $('#form_target_project').select2('data');
              $('#form_target_cluster').val('');
              if (data.length == 1) {
                if(data[0].nodes){
                  $('#form_target_cluster').select2({
                    "data" : data[0].nodes,
                  });
                }
              } else {
                $('#form_target_cluster').select2({
                  "data": [],
                  "placeholder": '不能选择',
                });
              }
              $scope.item.project = [];
              $scope.item.target = "projects";
              for (var i=0;i<data.length;i++) {
                $scope.item.project.push(data[i].id);
              }
            });
            $("#form_target_cluster").on("change", function(e) {
              var data = $('#form_target_cluster').select2('data');
              $('#form_target_apps').val('');
              if (data.length == 1) {
                if(data[0].nodes){
                  $('#form_target_apps').select2({
                    "data" : data[0].nodes,
                  });
                }
              } else {
                $('#form_target_apps').select2({
                  "data": [],
                  "placeholder": '不能选择',
                });
              }
              $scope.item.cluster = [];
              $scope.item.target = "clusters";
              for (var i=0;i<data.length;i++) {
                $scope.item.cluster.push(data[i].id);
              }
            });
            $("#form_target_apps").on("change", function(e) {
              var data = $('#form_target_apps').select2('data');
              $scope.item.app = [];
              $scope.item.target = "apps";
              for (var i=0;i<data.length;i++) {
                $scope.item.app.push(data[i].id);
              }
            });
          }
        },200);
      } else if ($scope.showme.curr == "jobs") {
        stop = $interval(function() {
          if (document.getElementById("scriptcode") != null) {
            $scope.stopWait();
            var mycodest = typeof(mycode);
            if (mycodest == "object") {
              mycode.toTextArea();
            }
            var jobreadonly = false;
            if (permissions.hasPermission("/jobs/jobs:PUT") == false) {
              jobreadonly = true;
            }
            scriptcode = document.getElementById("scriptcode");
            scriptcode.innerHTML = "";
            mycode = CodeMirror.fromTextArea(scriptcode, {
              "lineNumbers": true,
              "theme": "the-matrix",
              "scrollbarStyle": "overlay",
              "readOnly": jobreadonly,
              "tabSize": 2,
              "mode": "python"
            });
            if (data.type == "addfile") {
              $scope.item = {
                "timeout": 60,
                "pri": 19,
                "user": "root",
              };
              $scope.codeChange = true;
              mycode.setValue("");
              $scope.readonly = false;
              if (data.path1 != undefined) {
                $scope.item.path1 = data.path1;
                $scope.path1readonly = true;
              } else {
                $scope.path1readonly = false;
              }
              if (data.path2 != undefined) {
                $scope.item.path2 = data.path2;
                $scope.path2readonly = true;
              } else {
                $scope.path2readonly = false;
              }
              mycode.on("focus", this.onFocus = function() {
                if ("string" == typeof $scope.item.name) {
                  ext = $scope.item.name.match("\.(py|sh|pl|yml)$");
                  if (ext != null) {
                    mycode.setOption('mode', stype[ext[1]].mode);
                    //mycode.setValue(stype[ext[1]].initcode+initcode);
                    mycode.setValue(stype[ext[1]].initcode+initcode);
                  }
                }
              });
            } else if (data.type == "path1" || data.type == "path2") {
              mycode.setValue("");
              $scope.readonly = true;
              $scope.path1readonly = true;
              $scope.path2readonly = true;
              $scope.item = {
                "name": "",
                "path1": data.text,
              };
              if (data.type == "path2") {
                $scope.item["path1"] = data.path1;
                $scope.item["path2"] = data.text;
              }
            } else if (data.type == "file") {
              $scope.data = data;
              $scope.codeChange = false;
              req.url = url;
              if (data.id) {
                req.params = {"id": data.id};
              } else {
                req.params = {"name": data.text,"path":data.path1+"/"+data.path2};
              }
              mycode.off("focus");
              mycode.on("change", this.onChange = function(obj) {
                $scope.codeChange = true;
                // console.log(obj);
              });

              $scope.readonly = true;
              $scope.path1readonly = true;
              $scope.path2readonly = true;
              $http(req).then(function (response) {
                var mydata = response.data;
                if(mydata.code==0){
                  $scope.item = mydata.dataList[0];
                  $scope.item.path1 = data.path1;
                  $scope.item.path2 = data.path2;
                  mycode.setValue(utf8Decode(base64Decode(mydata.dataList[0]["content"])));
                } else if (mydata.code== 1199) {
                  $rootScope.loginform();
                } else {
                  layer.msg(mydata.message, {"icon": 5});
                  //console.log('errorcode:'+data.code+'errorMessage:'+data.message);
                }
              },function(mydata){
                layer.msg('程序内部错误!', {"icon": 5});
                //console.log('request failed!');
              });
            }
          }
        },100);
      } else if ($scope.showme.curr == "shortcut" || $scope.showme.curr == "runlogs") {
        req.url = BaseUrl + "/jobs/" + $scope.showme.curr;
        delete req.params;
        //req.params = {"taskattr.name": $scope.item.name,"taskattr.path": $scope.item.path};
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.datalist = data.dataList;
          } else if (data.code== 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {"icon": 5});
            //console.log('errorcode:'+data.code+'errorMessage:'+data.message);
          }
        },function(data){
          layer.msg('程序内部错误!', {"icon": 5});
          //console.log('request failed!');
        });
      } else if ($scope.showme.curr == "hostslogs") {
        req.url = BaseUrl + "/jobs/runlogs";
        req.params = {"id": data._id,"logdetail": 1};
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.loginfo = data.dataList[0];
          } else if (data.code== 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {"icon": 5});
            //console.log('errorcode:'+data.code+'errorMessage:'+data.message);
          }
        },function(data){
          layer.msg('程序内部错误!', {"icon": 5});
          //console.log('request failed!');
        });
      }
    }
    req.url = BaseUrl + "/jobs/path";
    $http(req).then(function (response) {
      var data = response.data;
      if(data.code==0){
        $('#jobstree').treeview({
          "data": data.dataList,
          //enableLinks: true,
          "expandIcon": 'glyphicon glyphicon-chevron-right',
          "collapseIcon": 'glyphicon glyphicon-chevron-down',
          "color": '#428bca',
          "nodeIcon": '',
          "showBorder": false,
          "levels": 3,
          "showTags": true,
          "onNodeSelected": function(event, data){
            //console.log(event);
            //console.log(data);
            $scope.projectop(data,"jobs","作业");
          }
        });
      } else if (data.code== 1199) {
        $rootScope.loginform();
      } else {
        layer.msg(data.message, {"icon": 5});
        //console.log('errorcode:'+data.code+'errorMessage:'+data.message);
      }
    },function(data){
      layer.msg('程序内部错误!', {"icon": 5});
      //console.log('request failed!');
    });

    if ($location.path() == "/jobs/runlogs") {
      $scope.projectop('','runlogs','执行记录');
    }

    $scope.jumpto = function() {
      $scope.projectop('',$scope.showme.history, $scope.title.history);
    }
    $scope.pidtoname = function(pstr,item,index,arr) {
      if (index == 0) {
        return $scope.projects[item].name;
      } else {
        return pstr + ',' + $scope.projects[item].name;
      }
    }
    $scope.cidtoname = function(pstr,item,index,arr) {
      if (index == 0) {
        return $scope.clusters[item].name;
      } else {
        return pstr + ',' + $scope.clusters[item].name;
      }
    }
    $scope.aidtoname = function(pstr,item,index,arr) {
      if (index == 0) {
        return $scope.apps[item].name;
      } else {
        return pstr + ',' + $scope.apps[item].name;
      }
    }
    //var ddd = ["598580495081faaeac8e5ecb"];
    //ddd.reduce(pidtoname,"");
    $scope.logsdetail = {};
    $scope.joblogsdetail = function (log) {
      $scope.logsdetail.stdout = log.stdout.crlf2html();
      $scope.logsdetail.errout = log.errout.crlf2html();
      $scope.open();
    }

    $scope.open = function (size, parentSelector) {
      var parentElem = parentSelector ?
        angular.element($document[0].querySelector(parentSelector)) : undefined;
      var modalInstance = $uibModal.open({
        "animation": true,
        "ariaLabelledBy": 'modal-title',
        "ariaDescribedBy": 'modal-body',
        "templateUrl": 'taskattr.html',
        "controller": 'TaskManageCtrl',
        "controllerAs": '$ctrl',
        "size": size,
        "appendTo": parentElem,
        "resolve": {
          items: function () {
            return {};
          },
          taskst: function () {
            var taskst = {
              "stdout": $scope.logsdetail.stdout,
              "stderr": $scope.logsdetail.errout,
            };
            return taskst;
          },
        }
      });

    };

    $scope.jobrunshortcut = function(item) {
      var req = {
        "method": 'GET',
        "url": BaseUrl+'/jobs/runshortcut',
        "params": {"id": item._id},
      };
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('操作成功！', {"icon": 1,time:1000,end:function(){
            $route.reload();}});
          $scope.projectop($scope.showme.history,$scope.title.history);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {"icon": 5});
      });
    }
    $scope.ok = function(isValid,dest) {
      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {"icon": 5});
        return;
      }
      $scope.item.release = false;
      if (dest == "jobs-release" && $scope.item.commit_id != undefined
        && ($scope.item.commit_id != $scope.item.commitid)) {
          commitid = $scope.item.commit_id;
          id = $scope.item._id;
          $scope.item = {
            "commitid": commitid,
            "_id": id,
          };
          $scope.item.release = true;
          dest = "jobs";
      }
      var req = {
        "method": 'POST',
        "url": BaseUrl+'/jobs/' + dest,
        "headers": {
          'Content-Type': "application/json; charset=UTF-8"
        },
        "data": $scope.item,
      };
      if (dest == "jobrun" || dest == "shortcut" ) {
        switch ($scope.item.target) {
          case "projects":
            $scope.item.list = $scope.item.project;
            break;
          case "clusters":
            $scope.item.list = $scope.item.cluster;
            break;
          case "apps":
            $scope.item.list = $scope.item.app;
            break;
          default:

        }
        if ($scope.item.list.length == 0) {
          layer.msg("请选择执行目标！", {"icon": 2});
          return;
        }
        if (dest == "jobrun") {
          req.method = "PUT";
          req.url = BaseUrl+'/jobs/run';
        }
      }
      // var data = {type:"file",path1:$scope.item.paht1,path2:$scope.item.path2,text:$scope.item.name};
      if ($scope.item._id != undefined) {
        req.method = "PUT";
        // data.id = $scope.item._id;
      }

      // if (permissions.hasPermission("/jobs/"+dest+":"+req.method) == false) {
      //   layer.msg('没有操作权限！', {"icon": 4});
      // }

      if (dest == "jobs" && $scope.item["release"] == false) {
        if ($scope.codeChange) {
          $scope.item["scriptcode"] = base64Encode(mycode.getValue());
        }
        if (req.method == "POST") {
          $scope.item["path"] = $scope.item["path1"];
          if ($scope.item["path2"] != undefined) {
            $scope.item["path"] += "/" + $scope.item["path2"];
          }
        }
        //$scope.item.operator = $rootScope.fullname;
      }

      //console.log($scope.items);
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('操作成功！', {"icon": 1,"time":1000,"end":function(){
            if ($scope.item["release"] == false) {
              $route.reload();
            }
            }});
          //$scope.projectop($scope.showme.history,$scope.title.history);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {"icon": 5});
      });
    }

    $scope.ExecDelete = function(item) {
      var req = {
        "method": 'DELETE',
        "url": url,
        "headers": {
          'Content-Type': "application/json; charset=UTF-8"
        },
        "data": {"id":item._id}
      }

      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('删除成功！', {"icon": 1});
          $route.reload();
          //$location.path(prefixurl);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      }, function(data){
        layer.msg('程序内部错误！', {"icon": 5});
        //console.log('request failed!');
      });
    }
    $scope.deleteitem = function (item) {
      layer.confirm('您是要删除　<span class="text-red">'+item.path+'/'+item.name+'</span>　？', {
        "title":"删除内容",
        "btn": ['确定','取消'] //按钮
      }, function(){
        $scope.ExecDelete(item);
        $route.reload();
      }, function(){

      });
    };
    $scope.paginate = function (value,index,carr) {
      var begin, end;
      $scope.page.totalItems = carr.length;
      numpp = parseInt($scope.page.numPerPage);
      begin = ($scope.page.currentPage - 1) * numpp;
      end = begin + numpp;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };
    $scope.failedStyle = function (item) {
      if (item.progress < item.hostnum) {
        //return {'background-color':'#FFFFFF'} //#FEF9CA
        return {'color':'#000000'}
      } else if (item.status == 0 && item.progress == item.hostnum) {
        return {'color':'#319D3F'}; //85D64E
      } else if (item.status == 900) {
        return {'color':'#F78D0E','font-weight':'bold'}; //FED02F
      }
      return {'color':'#E13F3F',"font-style":"italic"};
    }
    $scope.failedStyleHlog = function (item) {
      if (item != undefined) {
        if (item.code == 0) {
          return {'background-color':'#85D64E'}
        }
      } else {
        return {'background-color':'#FFFFFF'}
      }
      return {'background-color':'#FD7D89'}
    }

    $scope.runtime = function(item) {
      if (typeof item == "object") {
        if (!(item.endtime)) {
          return -1;
        }
        var endtime = new Date(item.endtime);
        var begintime = new Date(item.begintime);
        if (endtime < begintime)
          return -1;
        return (endtime - begintime) / 1000;
      }
      return ;
    }
}]);

app.controller('JobsFlowController',['$rootScope','$scope','BaseUrl','$http','$location',
  '$compile','$routeParams','permissions','$interval','$route','$uibModal','$document',
  function ($rootScope,$scope,BaseUrl,$http,$location,$compile,$routeParams,permissions,$interval,$route,$uibModal,$document) {
    var $ctrl = this;
    $scope.flow = [];
    $scope.page = {};
    $scope.page.totalItems = 0;
    $scope.page.currentPage = 1;
    $scope.page.maxSize = 5;
    $scope.page.itemFilter = '';
    $scope.page.numPerPage = "10";
    $scope.page.sortType     = 'name';
    $scope.page.sortReverse  = false;
    $scope.datalist = [];
    $scope.showme = {};
    $scope.title = {};
    $scope.showme.curr = "list";
    $scope.title.curr = "作业流管理";
    $scope.step = {};
    $scope.flowstep = 0;
    $scope.joblist = [];
    $scope.data = {};
    $scope.taskattr = false;
    $scope.nodesnum = 0;
    $scope.item = {};
    $scope.taskst = {noNext: false,noRescue: false,readonly:false};
    $scope.readonly = false;
    $scope.datalist = [];
    var flowurl = BaseUrl+'/jobs/flow';
    var req = {
      "method": 'GET',
      "url": flowurl,
      "params": {"pageNo":1,"pageSize":2000},
    };

    var stop;
    $scope.stopWait = function() {
      if (angular.isDefined(stop)) {
        $interval.cancel(stop);
      }
    }

    $scope.paginate = function (value,index,carr) {
      var begin, end;
      $scope.page.totalItems = carr.length;
      numpp = parseInt($scope.page.numPerPage);
      begin = ($scope.page.currentPage - 1) * numpp;
      end = begin + numpp;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };

    $scope.newtask = function() {
      var task = {
        "timeout": 60,
        "pri": 19,
        "user": "root",
        "argv": "",
      };
      return task;
    }
    //var lpath = $location.path();
    //var lpath = $location.path().split('/');
    //获取所有作业流
    $http(req).then(function (response) {
      var data = response.data;
      if(data.code==0){
        $scope.datalist=data.dataList;
        $scope.showme.curr = "list";
      } else if (data.code == 1199) {
        $rootScope.loginform();
      } else {
        layer.msg(data.message, {"icon": 2});
      }
    },function(data){
      layer.msg('无法获取信息!', {"icon": 5});
    });

    /*
      参照http://angular-ui.github.io/bootstrap/#!#%2Fmodal的实例实现。
      resolve：是将本控制器内的对象传递给TaskManageCtrl。
    */
    $scope.open = function (size, parentSelector) {
      var parentElem = parentSelector ?
        angular.element($document[0].querySelector(parentSelector)) : undefined;
      var modalInstance = $uibModal.open({
        "animation": true,
        "ariaLabelledBy": 'modal-title',
        "ariaDescribedBy": 'modal-body',
        "templateUrl": 'taskattr.html',
        "controller": 'TaskManageCtrl',
        "controllerAs": '$ctrl',
        "size": size,
        "appendTo": parentElem,
        "resolve": {
          "items": function () {
            return $scope.step;
          },
          "taskst": function () {return $scope.taskst},
        }
      });
      stop = $interval(function() {
        if (document.getElementById("jobselector") != null) {
          $scope.stopWait();
          $( "#jobselector" ).zdops_autocomplete({
              "minLength": 1, //输入1个符触发搜索,
              "delay": 0,
              // source: $scope.joblist,
              "source": function( request, response ) {
                // response( $.ui.autocomplete.filter(
                //   $scope.joblist,  request.term  ) );
                var matcher = new RegExp(request.term, "i");
                var results =$.grep($scope.joblist, function(val){
                  return matcher.test(val.text);
                });
                response(results.slice(0, 100));
              },
              "select": function(event, ui){
                this.value = ui.item.text;
                $scope.step["jobid"]= ui.item.id;
                return false;
              },
              //在bootstrap的modal下，此项是必须的。
              appendTo: ".eventInsForm",
          });
        }
      },200);

      modalInstance.result.then(function (selectedItem) {
        //console.log(selectedItem);
        $scope.step = selectedItem;
        if ($scope.step.hasOwnProperty("isRescue")) {
          $scope.addstep($scope.step["isRescue"]);
        } else if ($scope.step.hasOwnProperty("delete")) {
          $scope.delstep($scope.step["index"]);
        } else {
          $scope.save();
        }
      }, function () {
        //console.log('Modal dismissed at: ' + new Date());
      });
    };
    $scope.changeview = function(mode, item) {
      $scope.showme.history = $scope.showme.curr;
      $scope.showme.curr = mode;
      $scope.title.history = $scope.title.curr;
      if (mode == "list") {
        $scope.title.curr = "作业流管理";
        return;
      }
      if ($scope.joblist.length == 0) {
        req.url = BaseUrl+'/jobs/jobs';
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            for (i=0;i<data.totalRecord;i++) {
              var item = data.dataList[i];
              $scope.joblist.push({id: item._id,text: item.path+'/'+item.name});
            }
          } else {
            layer.msg(data.message, {"icon": 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {"icon": 5});
        }).then(function(){
          // console.log("OK");
        });
      }
      $scope.item = {"_id": item._id,"name": item.name,"describe":item.describe};
      var x = 100;
      var y = 20;
      var h = 50;
      var w= 150;
      //$scope.nodesnum = 0;
      var prev = 0;
      if(typeof item == "object") {
        $scope.data = {nodes: [],edges:[]};
        flowtmp = item.task;
        var i = 0;
        $scope.nodesnum = 0;
        while(true) {
          var node = {
            "id": "job"+i.toString(),
            "x": x,
            'y': y+(h+60)*($scope.nodesnum),
            "height": h,
            "width": w,
            "text":flowtmp.alias,
            "className": "node success",
            "link": "stepform("+i.toString()+")",
            "task": $.extend({},flowtmp),
            "index": $scope.nodesnum,
          };
          delete node.task["next"];
          delete node.task["rescue"];
          //node.task.index = $scope.nodesnum;
          $scope.data.nodes.push(node);
          if(i > 0) {
            $scope.data.edges.push({
              "source": "job"+prev.toString(),
              "sDirection":'bottom',
              "target": node.id,
              "tDirection":'top',
              "edgesType":"success",
            });
          }
          prev = i;
          if(flowtmp.hasOwnProperty("rescue") && flowtmp.rescue != null) {
            i++;
            node.rescue = i;
            var rnode = {
              "id": "job"+i.toString()+"r",
              'x': x + w + 100,
              "y": y+(h+60)*($scope.nodesnum),
              'height': h,
              "width": w,
              "text":flowtmp.rescue.alias,
              "className": "node warning",
              "link": "stepform("+i.toString()+")",
              "task": $.extend({},flowtmp.rescue),
              "index": $scope.nodesnum,
            };
            $scope.data.nodes.push(rnode);
            $scope.data.edges.push({
              "source": node.id,
              "sDirection":'right',
              "target": rnode.id,
              "tDirection":'left',
              "edgesType":"warning",
            });
          }
          i++;
          if(flowtmp.hasOwnProperty("next") && flowtmp.next != null) {
            flowtmp = flowtmp.next;
          } else {
            break;
          }
          $scope.nodesnum++;
        }
      }
      switch (mode) {
        case "info":
          $scope.title["curr"] = "查看作业流："+item.name;
          $scope.readonly = true;
          $scope.taskst.readonly = true;
          break;
        case "update":
          $scope.title.curr = "编辑作业流："+item.name;
          $scope.readonly = false;
          $scope.taskst.readonly = false;
          break;
        case "add":
          $scope.title.curr = "添加作业流";
          $scope.readonly = false;
          $scope.taskst.readonly = false;
          $scope.data = {
            "nodes": [
              {
                "id": "job0",
                "x": x,
                'y': 20,
                "height": 50,
                "width": 150,
                "text":"添加作业",
                "className": "node success",
                "link": "stepform(0)",
                "index": 0,
                //task: $scope.newtask(),
              },
            ],
            "edges": [],
          };

          break;
        default:

      }
      stop = $interval(function() {
        if (document.getElementById("jobsflow") != null) {
          $scope.stopWait();
          //console.log("OK");
          $scope.topology = $('#jobsflow .flow_box').bkTopology({
              "lineWidth": 2,
              "data": $scope.data,      //配置数据源
              "drag": true,      //是否支持拖拽移动
              "lineType": [      //配置线条的类型
                  {"type":'success',"lineColor":'#46C37B'},
                  {"type":'info',"lineColor":'#4A9BFF'},
                  {"type":'warning',"lineColor":'#f0a63a'},
                  {"type":'danger',"lineColor":'#c94d3c'},
                  {"type":'default',"lineColor":'#aaa'}
              ]
          });
          $("#node-templates").popover();
        }
      },100);
    }
    //$scope.changeview("add","添加作业流");

    $scope.delstep = function(i) {
      var node = $scope.data.nodes[i];
      layer.confirm('您是要删除　<span class="text-red">'+node.text+'</span>　？', {
        title:"删除内容",
        btn: ['确定','取消'] //按钮
      }, function(index){
        // if ((node.task.hasOwnProperty("rescue") && node.task.rescue != null) ) {
        //   layer.msg("请先删除子节点。", {"icon": 2});
        //   return;
        // } else if (node.index != $scope.nodesnum ) {
        //   layer.msg("请先删除子节点。", {"icon": 2});
        //   return;
        // }
        if ((!node.task.hasOwnProperty("rescue") || (node.task.hasOwnProperty("rescue") && node.task.rescue == null)) ||
        (node.index == $scope.nodesnum && !node.task.hasOwnProperty("rescue")) ) {
          k = $scope.data.edges.indexOfBy(function(e){
            return e.target == node.id ? true : false;
          });
          r = $scope.data.nodes.indexOfBy(function(e){
            return "job"+e.rescue+"r" == node.id ? true : false;
          });
          if (r >= 0) {
            delete $scope.data.nodes[r].rescue;
          }
          $scope.data.nodes.splice(i,1);
          $scope.data.edges.splice(k,1);
          $scope.topology.reLoad("node",$scope.data.nodes);
          console.log($scope.data);
        } else {
          layer.msg("请先删除子节点。", {"icon": 2});
        }
        layer.close(index);
      }, function(){

      });
    }
    $scope.addstep = function(isRescue) {
      var datalen = $scope.data.nodes.length;
      var prevnode = $scope.data.nodes[$scope.flowstep];
      //$scope.flowstep++;
      var id = datalen.toString();
      $scope.nodesnum++;
      var node = {
        "id": "job"+id,
        "x": prevnode.x,
        "y":  prevnode.y + prevnode.height + 60,
        "height": 50,
        "width": prevnode.width,
        "text":"添加作业",
        "className": "node success",
        "link": "stepform("+id+")",
        "prev": $scope.flowstep,
        "index": $scope.nodesnum,
        //task: $scope.newtask(),
      };
      if (isRescue) {
        node.id = "job" + id + "r";
        node.x = prevnode.x + prevnode.width + 100;
        node.y = prevnode.y;
        node.className = "node warning";
        prevnode.rescue = datalen;
        $scope.data.edges.push({
          "source": prevnode.id,
          "sDirection":'right',
          "target": node.id,
          "tDirection":'left',
          "edgesType":"warning",
        });
        //$scope.isRescue = true;
      } else {
        //$scope.isRescue = false;
        $scope.data.edges.push({
          "source": prevnode.id,
          "sDirection":'bottom',
          "target": node.id,
          "tDirection":'top',
          "edgesType":"success",
        });
      }
      //$scope.flowstep++;
      $scope.data.nodes.push(node);
      $scope.topology.reLoad("edge",$scope.data.edges);
    }

    $scope.defaultargv = {
      "timeout": 60,
      "pri": 19,
      "user": "root",
    };
    $scope.save = function() {
      var node = $scope.data["nodes"][$scope.flowstep];

      node.task = $.extend({},$scope.defaultargv,$scope.step);
      // if($scope.step.hasOwnProperty("timeout")) {
      //   node.task.timeout = $scope.step.timeout;
      // }
      // if($scope.step.hasOwnProperty("pri")) {
      //   node.task.pri = $scope.step.pri;
      // }
      // if($scope.step.hasOwnProperty("user")) {
      //   node.task.user = $scope.step.user;
      // }
      n = $scope.step.job.lastIndexOf("/");
      node["task"].name = $scope.step.job.substr(n+1);
      node["task"].path = $scope.step.job.substr(0,n);
      node["text"] = $scope.step.alias;
      //node.task.index = node.index;
      node["title"] = '作业：'+ node["task"].job;
      //delete node["task"].job;
      $scope.topology.reLoad("node",$scope.data.nodes);
      //$('#taskattr').modal('hide');
    }

    stepform = function(i) {
      // console.log(i);
      // $scope.data["nodes"][i].text = "改名";
      $scope.flowstep = i;
      var node = $scope.data["nodes"][i];
      $scope.taskst.noRescue = false;
      $scope.taskst.noNext = false;
      $scope.taskst.nodel = true;
      if (node.id.endsWith("r")) {
        $scope.taskst.noRescue = true;
        $scope.taskst.noNext = true;
        $scope.taskst.nodel = false;
      } else {
        if (node.hasOwnProperty("rescue")) {
          $scope.taskst.noRescue = true;
        }
        if (node.index < $scope.nodesnum ) {
          $scope.taskst.noNext = true;
        } else if (!node.hasOwnProperty("rescue")) {
          $scope.taskst.nodel = false;
        }
      }
      $scope.step = {};
      if ("task" in node) {
        $scope.step = node["task"];
        $scope.step.job = node["task"].path + '/' + node["task"].name;
        //$( "#jobselector" ).val(node["task"].path + '/' + node["task"].name);
      }
      $scope.step.index = i;
      $scope.open();
      //$('#taskattr').css({top: node.y+node.height,left:node.x+node.width});
      //,visibility:"visible"
      // $scope.taskattr = true;
      // $('#taskattr').modal('handleUpdate');
      // $('#taskattr').modal('show');
      // $( "#jobselector" ).autocomplete( "enable" );
      // dd = $( "#jobselector" ).autocomplete( "instance" );
      // console.log(dd);
      // $scope.topology.reLoad("node",$scope.data);
    };

    $scope.deleteitem = function (item) {
      layer.confirm('您是要删除　<span class="text-red">'+item.name+'</span>　？', {
        title:"删除内容",
        btn: ['确定','取消'] //按钮
      }, function(){
        $scope.ExecDelete(item);
        $route.reload();
      }, function(){

      });
    }
    $scope.ExecDelete = function(item) {
      var req = {
        "method": 'DELETE',
        "url": flowurl,
        "headers": {
          'Content-Type': "application/json; charset=UTF-8"
        },
        "data": {"id":item._id}
      }

      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('删除成功！', {"icon": 1});
          $route.reload();
          //$location.path(prefixurl);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      }, function(data){
        layer.msg('程序内部错误！', {"icon": 5});
        //console.log('request failed!');
      });
    };

    $scope.ok = function(isValid) {
      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {"icon": 5});
        return;
      }
      var flowinfo = {};
      $scope.item.task = {};
      var nodes = $scope.data.nodes;
      $scope.item.task = nodes[0].task;
      if(nodes[0].hasOwnProperty("rescue")) {
        $scope.item.task.rescue = nodes[nodes[0].rescue].task;
      }
      flowtmp = $scope.item.task;
      for(i=1;i<nodes.length;i++) {
        if(!nodes[i].id.endsWith("r")) {
          flowtmp.next = nodes[i].task;
          if(nodes[i].hasOwnProperty("rescue")) {
            r = $scope.data.nodes.indexOfBy(function(e){
              return "job"+nodes[i].rescue+"r" == e.id ? true : false;
            });
            flowtmp.next.rescue = nodes[r].task;
          }
          flowtmp = flowtmp.next;
        }
      }
      var req = {
        "method": 'POST',
        "url": flowurl,
        "data": $scope.item,
      };
      if ($scope.item.hasOwnProperty("_id")) {
        req.method = 'PUT';
      }
      //console.log($scope.item);
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('操作成功！', {"icon": 1,"time":1000,"end":function(){
            $route.reload();}});
          //$scope.changeview($scope.showme.history,$scope.title.history);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {"icon": 5});
      });
    }


}]);
