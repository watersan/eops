app.controller('CMDBController',['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','SessionService','orderByFilter',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,SessionService,orderBy) {
    $scope.input={
      "pageNo": 1,
      "pageSize": 2000,
    };

    //$rootScope.cmdbitems = SessionService.getobj("cmdbtype");
    $scope.items = [];
    var items = [];
    $scope.allowCustomFields = true;
    $scope.totalItems = 0;
    $scope.currentPage = 1;
    $scope.maxSize       = 5;
    $scope.itemFilter    = '';
    $scope.sortType      = 'hostid';
    $scope.sortReverse   = false;
    $scope.tablestatus   = "2";//列表是否为空 1  不空 2空
    $scope.datalist      = [];
    $scope.numPerPage    = "10";
    $scope.currentNum    = 0;
    $scope.selectAll = true;
    $scope.cmdbnameid = $routeParams.type;
    var prefixurl = "/cmdbitem/";
    if ($scope.cmdbnameid == "cmdbtype") {
      prefixurl = "/cmdb/";
    }
    $scope.thedata = {
      "addperm": prefixurl + $scope.cmdbnameid + "/add",
      "addurl": prefixurl + $scope.cmdbnameid + "/add",
      "operations": [
        {
          "href": prefixurl + $scope.cmdbnameid + "/info/",
          "perm": prefixurl + $scope.cmdbnameid + "/info/",
          "name": "查看",
        },
        {
          "href": prefixurl + $scope.cmdbnameid + "/update/",
          "perm": prefixurl + $scope.cmdbnameid + "/update/",
          "name": "编辑",
        },
        {
          "href": "javascript:;",
          "perm": prefixurl + $scope.cmdbnameid + "/delete/",
          "name": "删除",
        },
      ],
    };
    $scope.sortBy = function(propertyName) {
      $scope.sortReverse = (propertyName !== null && $scope.sortType === propertyName)
        ? !$scope.sortReverse : false;
      $scope.sortType = propertyName;
      $scope.items = orderBy(items, $scope.sortType, $scope.sortReverse);
    };
    $scope.paginate = function (value,index,carr) {
      var begin, end;
      numpp = parseInt($scope.numPerPage);
      if ($scope.itemFilter != '' && carr.length < numpp) {
        //console.log(value)
        return true;
      }
      begin = ($scope.currentPage - 1) * numpp;
      end = begin + numpp;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };

    $scope.doSelectAll = function() {
      numpp = parseInt($scope.numPerPage);
      var begin = ($scope.currentPage - 1) * numpp;
      var end = begin + numpp;
      if ($scope.selectAll == true) {
        $scope.selectAll = false;
        for (k in $scope.items) {
          if (begin <= k && k < end) {
            $scope.items[k].selected = true;
          }
        }
      } else {
        $scope.selectAll = true;
        for (k in $scope.items) {
          if (begin <= k && k < end) {
            $scope.items[k].selected = false;
          }
        }
      }
    };
    $scope.upDisplayFields = function(field) {
      //console.log($rootScope.cmdbitems[$scope.itemnameid][field.nameid].display);
      SessionService.saveobj("cmdbitems", $rootScope.cmdbitems);
    };
    $scope.checkdisplay = function(value,index,carr) {
      if ($rootScope.cmdbitems[$scope.cmdbnameid].fields[index] == undefined) {
        return false;
      }
      if ($rootScope.cmdbitems[$scope.cmdbnameid].fields[index].display == true) {
        return true;
      }
      return false;
    }
    $scope.deleteitem = function (item) {
      var req = {
        "method": 'DELETE',
        "url": BaseUrl+'/cmdb/' + $scope.cmdbnameid,
        "headers": {
          'Content-Type': "application/json; charset=UTF-8"
        },
        "data": {"id":item._id}
      }

      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('删除成功！', {icon: 1});
          //window.history.back();
          $location.path('/cmdb/' + $scope.cmdbnameid + '/');
        } else if (data.code == 1199) {
          $rootScope.loginform();
          //$rootScope.referer = $location.path();
          //$location.path('/login');
        } else {
          layer.msg(data.message, {icon: 2});
        }
      }, function(data){
        layer.msg('程序内部错误！', {icon: 5});
        // console.log('request failed!');
      });
    };
    /*
    if ($rootScope.cmdbitems[$scope.cmdbnameid] == undefined) {
      $rootScope.cmdbitems[$scope.cmdbnameid] = {};
      var reqCfg = {
        params: {"nameid": $scope.cmdbnameid},
      };
      var url = BaseUrl + '/cmdb/' + $scope.cmdbnameid;
      $http.get(url, {})
        .then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.fields = data.dataList["fields"];
            $rootScope.cmdbitems[$scope.cmdbnameid] = data.dataList;
            SessionService.saveobj("cmdbtype", $rootScope.cmdbitems);
            $scope.cmdbname = data.dataList["name"];
          }
        },function(data) {
          console.log('request failed!');
        });
    } else {
    */
    $scope.fields = $rootScope.cmdbitems[$scope.cmdbnameid].fields;
    $scope.cmdbname = $rootScope.cmdbitems[$scope.cmdbnameid].name;

    var req = {
      "method": 'GET',
      "url": BaseUrl + '/cmdb/' + $scope.cmdbnameid + '/all',
      "params": $scope.input,
    }
    $http(req).then(function (response) {
      var data = response.data;
      if(data.code==0){
        items = data.dataList;
        $scope.items = orderBy(items, $scope.sortType, $scope.sortReverse);
        $scope.tablestatus = "1";
        //console.log($scope.items);
      } else if (data.code == 1199) {
        $rootScope.loginform();
        //$rootScope.referer = $location.path();
        //$location.path('/login');
      } else {
        layer.msg(data.message, {icon: 2});
      }
    },function(data){
      layer.msg('程序内部错误!', {icon: 5});
      // console.log('request failed!');
    });

}]);


app.controller('CMDBPostController', ['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','MyHTTP','SessionService','$timeout',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,MyHTTP,SessionService,$timeout) {
    $scope.item={};
    $scope.fields = [];
    $scope.infomsgsuo = false;
    $scope.showsubmit = true;
    $scope.readonly = false;
    // $scope.ftypeindex = {
    //   "字符串": "str",
    //   "文本": "text",
    //   "整数": "int",
    //   "浮点": "float",
    //   "布尔": "bool",
    //   "时间": "time",
    //   "外键": "foreignKey",
    //   "配置": "conf",
    //   "应用包": "source",
    // };
    $scope.sourcelist = {
      "HTTP": "http",
      "GIT": "git",
    };
    $("#zdopslock").popover({"placement":'top'});
    $("#zdopsnotnull").popover({"placement":'top'});
    $("#zdopsdisp").popover({"placement":'top'});
    $scope.cmdbnameid = $routeParams.type;
    //$rootScope.cmdbitems = SessionService.getobj("cmdbtype");
    //console.log($scope.cmdbnameid);
    $scope.cmdbname = $rootScope.cmdbitems[$scope.cmdbnameid].name;

    $scope.addField = function() {
      //var sinfo = document.getElementById("addfield").innerHTML;
      var len = $scope.fields.length;
      $scope.fields.push({
        "id": len,
        "nameid" : "",
        "name": "",
        "type": 'str',
        "value": "",
        "mode": false,
      });
      //$scope.fields[len].fieldtype = "int";
    }
    $scope.isOpen = false;
    $scope.open = function($event) {
      $event.preventDefault();
      $event.stopPropagation();
      $scope.isOpen = true;
    };
    $scope.itemfieldshow = function(field, type) {
      if (field.type == "source" && $scope.item[field.nameid] == undefined) {
        $scope.item[field.nameid] = "http";
      }
      if (field.type == type) {
        //console.log(field.nameid +":"+type);
        return true;
      }
      return false;
    }
    $scope.delField = function(i) {
      $scope.fields.splice(i,1);
    }
    if ($routeParams.id != undefined) {
      var req = {
        "url": BaseUrl + '/cmdb/' + $scope.cmdbnameid + '/' + $routeParams.id,
        "method": "GET",
      };
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          $scope.item = data.dataList[0];
          if ($scope.item.fields != undefined) {
            $scope.fields = $scope.item.fields;
          }
          $scope.tablestatus = "1";
          //console.log($scope.items);
        } else if (data.code == 1199) {
          $rootScope.loginform();
          // $rootScope.referer = $location.path();
          // $location.path('/login');
        } else {
          layer.msg(data.message, {icon: 2});
        }
      },function(data){
        layer.msg('程序内部错误!', {icon: 5});
        // console.log('request failed!');
      });
    }
    var lpath = $location.path();
    var rlevel = lpath.split('/');
    if (rlevel[3] == "info") {
      $scope.showsubmit = false;
      $scope.readonly = true;
    }
    if (lpath != "/cmdb/cmdbtype/add") {
      $scope.fields = $rootScope.cmdbitems[$scope.cmdbnameid].fields;
    }
    $scope.ok = function (isValid) {

      // if (!isValid) {
      //   layer.msg('表单填写不正确，请仔细检查！', {icon: 5});
      //   return;
      // }
      if (rlevel[1] == "cmdb") {
        $scope.item.fields = $scope.fields;
        $scope.item.indexs = [];
      }
      var req = {
        "method": 'POST',
        "url": BaseUrl+'/cmdb/' + $scope.cmdbnameid,
        "headers": {
          'Content-Type': "application/json; charset=UTF-8"
        },
        "data": $scope.item
      }
      if (rlevel[3] == "update") {
        req.method = "PUT";
      }
      //console.log($scope.items);
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          if ($scope.cmdbnameid == "cmdbtype") {
            nameid = $scope.item.nameid;
            $rootScope.cmdbitems[nameid] = $scope.item;
          }
          SessionService.saveobj("cmdbitems", $rootScope.cmdbitems);
          layer.msg('添加成功!', {icon: 1,end:function(){
            //window.location='/cmdb/' + $scope.cmdbnameid + '/';
            $route.reload();
          }});
        } else if (data.code == 1199) {
          $rootScope.loginform();
          // $rootScope.referer = $location.path();
          // $location.path('/login');
        } else {
          layer.msg(data.message, {icon: 2});
        }
      },function(data){
        layer.msg('程序内部错误!', {icon: 5});
        // console.log('request failed!');
      });
    };

    $scope.cancel = function () {
      //$location.search({});
      window.history.back();
      //$location.path('/itemslist');
    };

}]);
