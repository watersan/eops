app.controller('ApptplController',
  ['$rootScope','$scope','$http','BaseUrl','$location','$route','$routeParams','$interval','MyHTTP',
  function ($rootScope,$scope, $http,BaseUrl,$location,$route,$routeParams,$interval,MyHTTP) {
    //console.log('go into UserListController');

    //分页
    $scope.page = {};
    $scope.page.totalItems = 0;
    $scope.page.currentPage = 1;
    $scope.page.maxSize = 5;
    $scope.page.itemFilter = '';
    $scope.page.numPerPage = "10";
    $scope.page.sortType     = 'name';
    $scope.page.sortReverse  = false;

    // $scope.page.totalItems = 0;
    // $scope.page.currentPage = 1;
    // $scope.maxSize = 5;
    // $scope.itemFilter = '';
    // $scope.page.numPerPage = "10";
    // $scope.sortType     = 'name';
    // $scope.sortReverse  = false;
    $scope.tablestatus="2";//列表是否为空 1  不空 2空
    $scope.datalist = [];
    //$scope.currentNum = 0;
    $scope.selectAll = true;
    $scope.isnew = true;
    $scope.item = {};
    $scope.readonly = true;
    $scope.updatereadonly = true;
    $scope.isplist = {};
    $scope.tplTags = [];
    $scope.import = {};
    $scope.importMethod = {
      method: "",
      scan: false,
      csv: false,
    };
    $scope.title = {
      curr: "应用模版管理",
    };
    $scope.yunproviders = ["aliyun","aws","Ucloud"];
    $scope.kind = "";
    var dict = {
      add: "添加",
      update: "更新",
      info: "查看",
    };
    var url;
    $scope.paginate = function (value,index,carr) {
      var begin, end;
      $scope.page.totalItems = carr.length;
      begin = ($scope.page.currentPage - 1) * $scope.page.numPerPage;
      end = begin + $scope.page.numPerPage;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };
    var lpath = $location.path();
    lpatharr = lpath.split('/');
    //console.log(lpatharr);
    var stop;
    $scope.stopWait = function() {
      if (angular.isDefined(stop)) {
        $interval.cancel(stop);
      }
    }
    if (lpatharr[2] == "apptpl") {
      url = BaseUrl+'/cmdb/apptpl';
      $scope.kind = "应用模板";
      if (lpatharr[3] == "add" || lpatharr[3] == "update") {
        req = {
          "method": 'GET',
          "url": BaseUrl+'/jobs/jobs',
          "params": {"pageNo":1,"pageSize":1000,},
        }
        $scope.apptplAll = {};
        $scope.joblist = [];
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            for (i=0;i<data.totalRecord;i++) {
              $scope.joblist.push(data.dataList[i].path+'/'+data.dataList[i].name);
            }
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {icon: 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {icon: 5});
        }).then(function () {
          fields = appPostTpl(true);
          $.each( fields, function( index, field ) {
            if (field.job) {
              $( "#job_"+field.name+" .keywork-input" ).autocomplete({
                  "minLength": 1, //输入1个符触发搜索,
                  "delay": 0,
                  "source": $scope.joblist,
              });
            }
          });
        });
        req.url = url;
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            if (lpatharr[2] == "apptpl") {
              for (i=0;i<data.totalRecord;i++) {
                $scope.tplTags.push(data.dataList[i].name);
              }
            }
            $scope.page.totalItems=data.totalRecord;
            $scope.tablestatus="1";
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {icon: 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {icon: 5});
        }).then(function () {
          $( "#formdepend .keywork-input" ).autocomplete({
            "minLength": 1, //输入1个符触发搜索,
            "source": function( request, response ) {
              response( $.ui.autocomplete.filter(
                $scope.tplTags, extractLast( request.term ) ) );
            },
            "focus": function() {
                // 阻止选择项后输入框获取焦点
                return false;
            },
            "select": function( event, ui ) {
              var terms = split( this.value );
              terms.pop();
              terms.push( ui.item.value );
              // terms.push( "" );
              this.value = terms.join( "," );
              // this.value = terms;
              return false;
            }
          });
        });
      }
    } else if (lpatharr[2] == 'resource') {
      $scope.kind = "资源";
      url = BaseUrl+'/cmdb/resource';
      if (lpatharr[3] == "add" || lpatharr[3] == "update" || lpatharr[3] == "info") {
        req = {
          "method": 'GET',
          "url": BaseUrl+'/cmdb/svcprovider',
          "params": {"pageNo":1,"pageSize":1000},
        };
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.svcproviders=data.dataList;
          } else {
            layer.msg(data.message, {icon: 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {icon: 5});
        });
        req.url = BaseUrl + "/auth/env";
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.cdenv=data.dataList;
          } else {
            layer.msg(data.message, {icon: 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {icon: 5});
        });
      }
    } else if (lpatharr[2] == 'yunprovider') {
      $scope.kind = "云服务商";
      url = BaseUrl+'/cmdb/yunprovider';
    } else if (lpatharr[2] == 'idcprovider') {
      $scope.kind = "IDC服务商";
      url = BaseUrl+'/cmdb/idcprovider';
      if (lpatharr[3] == "add" || lpatharr[3] == "update") {
        req = {
          "method": 'GET',
          "url": BaseUrl+'/cmdb/isp',
        };
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.isplist=data.dataList;
          } else {
            layer.msg(data.message, {icon: 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {icon: 5});
        });
        $('#begindate').datepicker({
          "autoclose": true,
          "language": 'zh-CN',
          'format': 'yyyy年mm月dd日',
          "todayHighlight": true,
          "minViewMode": 0,
        });
        $('#enddate').datepicker({
          "autoclose": true,
          "language": 'zh-CN',
          "format": 'yyyy年mm月dd日',
          "todayHighlight": true,
          "minViewMode": 0,
        });
      }
    }
    $scope.title.curr = dict[lpatharr[3]] + $scope.kind;
    if (lpatharr[3] == 'list') {
      $scope.title.curr = $scope.kind + "管理";
      req = {
        "method": 'GET',
        "url": url,
        "params": {"pageNo":1,"pageSize":1000,"environment": $rootScope.curenv},
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          if (lpatharr[2] == "apptpl") {
            for (i=0;i<data.totalRecord;i++) {
              if (data.dataList[i].depend) {
                data.dataList[i].depend = data.dataList[i].depend.join(',');
              }
            }
          }
          $scope.datalist=data.dataList;
          $scope.page.totalItems=data.totalRecord;
          $scope.tablestatus="1";
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {icon: 5});
      });
    } else if (lpatharr[3] == 'add') {
      $scope.action = "添加";
      $scope.isnew = true;
      $scope.readonly = false;
      $scope.updatereadonly = false;
      $scope.item.sourcetype = "GIT";
      //$scope.item.additionconf = "";
    } else if (lpatharr[3] == 'update' || lpatharr[3] == 'info') {
      $scope.action = "查看";
      //$scope.item.additionconf = "";
      $scope.isnew = false;
      if (lpatharr[3] == 'update') {
        $scope.action = "更新";
        $scope.readonly = false;
        $scope.updatereadonly = false;
      }
      req = {
        "method": 'GET',
        "url": url,
        "params": {"id":$routeParams.id,environment: $rootScope.curenv},
      };
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          $scope.item = data.dataList[0];
          if ($scope.item.additionconf) {
            $scope.item.additionconf = $scope.item.additionconf.join(',');
          }
          if ($scope.item.depend) {
            $scope.item.depend = $scope.item.depend.join(',');
          }
          if ($scope.item.regions) {
            $scope.item.regions = $scope.item.regions.join(',');
          }
          delete $scope.item.createtime;
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {icon: 5});
      });
    } else if (lpatharr[3] == 'yunimport') {
      $scope.title.curr = "从公有云导入";
      $scope.importsrc = "yun";
    } else if (lpatharr[3] == "idcimport") {
      $scope.title.curr = "从IDC导入";
      $scope.importsrc = "idc";
      $scope.importResult = "";
      stop = $interval(function() {
        if (document.getElementById("fileupload") != null) {
          $scope.stopWait();
          $('#fileupload').bind('fileuploadprogress', function (e, data) {
            var progress = parseInt(data.loaded / data.total * 100, 10);
            $('#progress .progress-bar').attr("aria-valuenow",progress).css(
                'width',
                progress + '%'
            ).text(progress + '%');
          });
          $('#fileupload').fileupload({
            "url": BaseUrl + '/cmdb/resource/import?id='+$routeParams.id+'&env='+$scope.import.cdenv,
            "maxChunkSize": 10485760,
            "uploadTemplateId": 'template-upload',
            "add": function (e, data) {
              data.submit();
            },
            "done": function(e,data) {
              // console.log(e);
              // console.log(data._response.result);
              $('#progress .progress-bar').css(
                  'display',
                  'none'
              );
              if (data._response.result.code == 0) {
                $scope.importResult = "提供资源总数："+data._response.result.totalRecord+"</br>成功导入资源数量："+data._response.result.dataList.insert+"</br>";
                $('#import-result').html($scope.importResult);
              } else {
                $scope.importResult = "上传失败！";
              }
            },
          });
        }
      },100);
    }

    $scope.ok = function (isValid) {
      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {icon: 5});
        return;
      }
      //console.log($scope.item);
      //return;
      if (lpatharr[2] == "apptpl") {
        if ($scope.item.additionconf) {
          if (typeof $scope.item.additionconf == "string") {
            dd = $scope.item.additionconf.split(",");
            $scope.item.additionconf = dd;
          }
        } else {
          delete $scope.item.additionconf;
        }
        if ($scope.item.depend) {
          if (typeof $scope.item.depend == "string") {
            dd = $scope.item.depend.split(",");
            $scope.item.depend = dd;
          }
        } else {
          delete $scope.item.depend;
        }
        $scope.item.usage = 0;
      } else if (lpatharr[2] == "resource") {
        $scope.item.status = 0;
        $scope.item.usage = 0;
      } else if (lpatharr[2] == "yunprovider") {
        if ($scope.item.regions) {
          if (typeof $scope.item.regions == "string") {
            dd = $scope.item.regions.split(",");
            $scope.item.regions = dd;
          }
        } else {
          $scope.item.regions = [];
        }
      }
      req = {
        "method": 'POST',
        "url": url,
        "params": {"environment": $rootScope.curenv},
      }
      req.headers = {'Content-Type': "application/json; charset=UTF-8"};
      if ($routeParams.id) {
        req.method = 'PUT';
        delete $scope.item.createtime;
        delete $scope.item.usage;
      }
      req.data = $scope.item;
      MyHTTP.http(req,"/cmdb/"+lpatharr[2]+"/list");
    }

    $scope.ExecDelete = function(item) {
      req = {
        "method": 'DELETE',
        "url": url,
        "params": {"id": item._id},
      }
      req.headers = {'Content-Type': "application/json; charset=UTF-8"};
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code ==0){
          layer.msg('删除成功！', {icon: 1,end:function(){
            $route.reload();
          }});
        } else if (data.code == 1199) {
          $rootScope.loginform();
          // $rootScope.referer = $location.path();
          // $location.path('/login');
        } else {
          layer.msg(data.message, {icon: 2});
        }
      });
    }
    $scope.deleteitem = function (item) {
      if (item.name) {
        title = item.name;
      } else {
        title = item.hostid;
      }
      layer.confirm('您是要删除　<span class="text-red">'+title+'</span>　？', {
        title:"删除内容",
        btn: ['确定','取消'] //按钮
      }, function(){
        $scope.ExecDelete(item);
        $route.reload();
      }, function(){

      });
    };
    // sdddd
    $scope.importhost = function() {
      req = {
        "method": 'GET',
        "url": BaseUrl+'/cmdb/resource/import',
        "params": {"id": $routeParams.id,"env":$scope.import.cdenv},
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code ==0){
          $scope.importResult = "成功导入资源数量："+data.dataList.insert+
            "； 更新资源的数量："+data.dataList.update;
          $('#import-result').html($scope.importResult);
          // layer.msg('导入完成！', {icon: 1,end:function(){
          //   $route.reload();
          // }});
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      });
    }
    $scope.jumpto = function (path) {
      if (path == undefined) {
        window.history.back();
      } else {
        $location.path(path);
      }
       // $uibModalInstance.dismiss('cancel');
    };
    function split( val ) {
      return val.split( /,\s*/ );
    }
    function extractLast( term ) {
      return split( term ).pop();
    }

}]);
