
app.controller('ItemsController',
  ['$rootScope','$scope','$http','BaseUrl','$route','GlobalData',
  function ($rootScope,$scope, $http,BaseUrl,$route,GlobalData) {
    GlobalData.ttt = {"name":"test"};
    console.log('go into ItemsController');
    var postCfg = {
        headers: {'Content-Type': 'application/json; charset=UTF-8'}
    };
    $scope.enter=function($event){
      if($event.keyCode==13){
        $scope.pageChanged(1);
      }
    }
    var url = BaseUrl+'/cmdb/itemslist'
    $scope.input={
      "pageNo":1,
      "pageSize":200,
    }
    //分页
    $scope.totalItems = 0;
    $scope.currentPage = $scope.input["pageNo"];
    $scope.maxSize       = 5;
    $scope.itemFilter    = '';
    $scope.sortType      = 'name';
    $scope.sortReverse   = false;
    $scope.tablestatus   = "2";//列表是否为空 1  不空 2空
    $scope.numPerPage    = $scope.input["pageSize"];
    $scope.datalist      = [];
    $scope.numPerPage    = "10";
    $scope.currentNum    = 0;
    $scope.selectAll = true;
    $scope.paginate = function (value,index,carr) {
      var begin, end, index;
      if ($scope.itemFilter != '' && carr.length < $scope.numPerPage) {
        //console.log(value)
        return true;
      }
      begin = ($scope.currentPage - 1) * $scope.numPerPage;
      end = begin + $scope.numPerPage;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };

    $scope.doSelectAll = function() {
      if ($scope.selectAll == true) {
        $scope.selectAll = false;
        for (k in $scope.datalist) {
          $scope.datalist[k].selected = true;
        }
      } else {
        $scope.selectAll = true;
        for (k in $scope.datalist) {
          $scope.datalist[k].selected = false;
        }
      }
    };

    $scope.doSelect = function(item) {
      //console.log(item.selected)
      if (item.selected == true) {
        //console.log(index)
        item.selected = false;
      } else {
        item.selected = true;
      }
      //console.log("aa"+$scope.datalist[index].selected)
    };

    $scope.paginate = function (value,index,carr) {
      var begin, end, index;
      if ($scope.opsFilter != '' && carr.length < $scope.numPerPage) {
        //console.log(value)
        return true;
      }
      begin = ($scope.currentPage - 1) * $scope.numPerPage;
      end = begin + $scope.numPerPage;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };
    $scope.pageChanged = function(search) {
      if(search){
        $scope.input.pageNo=1;
        $scope.currentPage=1;
      }else {
        $scope.input.pageNo=$scope.currentPage;
      }
      $scope.requestpageutil($http,url,$scope.input);
    };
    var reqCfg = {
      params: $scope.input,
    };
    //分页end
    $scope.requestpageutil=function($http,url,input){
      $http.get(url, reqCfg).then(function (response) {
        //console.log(response)
        var data = response.data;
        if(data["code"]==0){
          if(data["dataList"].length>0){
            $scope.datalist=data["dataList"];
            $scope.totalItems=data["totalRecord"];
            $scope.tablestatus="1";
          }else{
            $scope.tablestatus="2";
            $scope.totalItems=0;
          }
        }else{
          console.log('errorcode:'+data["code"]+'errorMessage:'+data["message"]);
        }
      },function(data){
        console.log('request failed!');
      });
    }
    $scope.requestpageutil($http,url,$scope.input);

    var delurl = BaseUrl+"/cmdb/itemdel";
    //
    $scope.deleteitem = function (item) {
      console.log("_id: "+item._id);
      $scope.postargs={
        "id":item._id
      }
      $http.put(delurl, $scope.postargs, postCfg)
        .then(function (response) {
          var data = response.data;
          if(data["code"]==0){
            layer.msg('删除成功！', {icon: 1});
          }else{
            console.log('errorcode:'+data["code"]+'errorMessage:'+data["message"]);
            layer.msg('删除失败！', {icon: 2});
          }
        }, function(data){
          layer.msg('程序内部错误！', {icon: 5});
          console.log('request failed!');
        });
    };
    //删除配置项
    $scope.delete = function (item) {
      layer.confirm('您是要删除配置项　<span class="text-red">'+item.nameid+'</span>？', {
        title:"删除配置项",
        btn: ['确定','取消'] //按钮
      }, function(){
        $scope.deleteitem(item);
        $route.reload();
      }, function(){

      });
    };

}]);

//添加配置项
app.controller('ItemsAddController', ['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','GlobalData','MyHTTP',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,GlobalData,MyHTTP) {
    console.log('go into ItemAddController');

    $scope.fields=[];
    $scope.item={};
    $scope.skey={};
    $scope.ftypes = [
      {name : "字符串", value : 'str'},
      {name : "整数", value : "int"},
      {name : "时间", value : "time"},
      {name : "数组", value : "array"},
      {name : "外键", value : "foreignKey"},
      {name : "子表", value : "subsetArray"},
    ];
    var i = 0;
    var url = BaseUrl+'/cmdb/itemadd';
    //var urlInfo = BaseUrl + '/cmdb/iteminfo';
    var postCfg = {
        headers: {'Content-Type': 'application/json; charset=UTF-8'}
    };
    var lpath = $location.path();
    /*
    if (lpath == "/itemadd1") {
      var searchObject = $location.search();
      if (searchObject.name == undefined) {
        $location.path('/itemadd');
      }
      $scope.item = searchObject;
    } else if (lpath == "/itemadd") {
      $scope.item.type="table";
    }
    */
    $scope.item.type="table";
    //$scope.fields =  new Array();
    $scope.fieldtype = function(id){
      if ($scope.fields[id].type == "foreignKey" || $scope.fields[id].type == "subsetArray") {
        $scope.fields[id].isvalue = true;
      } else {
        $scope.fields[id].isvalue = false;
      }
    }
    $scope.addField = function() {
      //var sinfo = document.getElementById("addfield").innerHTML;
      var len = $scope.fields.length;
      $scope.fields.push({
        id: len,
        nameid : "",
        name: "",
        type: 'str',
        value: "",
        mode: false,
      });
      //$scope.fields[len].fieldtype = "int";
    }
    $scope.delField = function(i) {
      $scope.fields.splice(i,1);
    }

    $scope.ok = function (isValid) {

      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {icon: 5});
        return;
      }
      $scope.item = GlobalData.item;
      GlobalData.item = undefined;
      $scope.item.fields = $scope.fields;
      $scope.item.indexs = [];
      for (id in $scope.fields) {
        console.log("Field: "+$scope.fields[id])
        if ($scope.fields[id].skey == true) {
          $scope.item.indexs.push({
            name : $scope.fields[id].nameid,
            field : $scope.fields[id].nameid,
            unique : false,
            orderby : true,
          });
          break;
        }
      }

      //console.log($scope.items);
      $http.post(url, $scope.item, postCfg)
        .then(function (response) {
          var data = response.data;
          if(data["code"]==0){
            //MyHTTP.cmdblist("table");
            layer.msg('添加成功!', {icon: 1,end:function(){
              window.location='/itemslist';
            }});
          }else{
            layer.msg(data["message"], {icon: 5});
            console.log('errorcode:'+data["code"]+'errorMessage:'+data["message"]);
          }
        },function(data){
          layer.msg('程序内部错误!', {icon: 5});
          console.log('request failed!');
        });
    };

    $scope.cancel = function () {
      //$location.search({});
      window.history.back();
      //$location.path('/itemslist');
    };

    $scope.next = function (isValid) {
      console.log($scope.item);
      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {icon: 5});
        return;
      }
      //$location.search($scope.item);
      GlobalData.item = $scope.item;
      $location.path('/itemadd1');
    };

}]);


//配置项信息
app.controller('ItemsInfoController', ['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','GlobalData',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,GlobalData) {
    console.log('go into ItemInfoController');
    var qargs = {
      "params": {"id":$routeParams.id},
    };
    var url = BaseUrl + '/cmdb/iteminfo';
    $http.get(url, qargs)
      .then(function (response) {
        var data = response.data;
        if(data["code"]==0){
          $scope.item = data["dataList"];
          //console.log($scope.item);
          //layer.msg('成功!', {icon: 1,end:function(){
            //window.location='/itemslist';
          //}});
        }else{
          layer.msg(data["message"], {icon: 5});
          console.log('errorcode:'+data["code"]+'errorMessage:'+data["message"]);
        }
      },function(data){
        layer.msg('程序内部错误!', {icon: 5});
        console.log('request failed!');
      });
    $scope.cancel = function () {
      //$location.search({});
      window.history.back();
      //$location.path('/itemslist');
    };

}]);

//配置项信息
app.controller('ItemsdbController', ['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','SessionService',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,SessionService) {
    console.log('go into ItemdbController');
    //$("#customlist").removeBox();
    $scope.input={
      "pageNo":1,
      "pageSize":2000,
    }
    $scope.cmdbtype = SessionService.getobj("cmdbtype");
    $scope.totalItems    = 0;
    $scope.currentPage   = 1;
    $scope.itemFilter    = '';
    $scope.sortType      = 'name';
    $scope.sortReverse   = false;
    $scope.tablestatus   = "2";//列表是否为空 1  不空 2空
    $scope.datalist      = [];
    $scope.numPerPage    = "10";
    $scope.currentNum    = 0;
    $scope.selectAll     = true;
    $scope.itemnameid = $routeParams.id;

    $scope.upDisplayFields = function(field) {
      //$scope.cmdbtype[$scope.itemnameid][field.nameid].display = field.display;
      console.log($scope.cmdbtype[$scope.itemnameid][field.nameid].display);
      /*
      if ($scope.cmdbtype[$scope.itemnameid][field.nameid].display != true) {
        $scope.cmdbtype[$scope.itemnameid][field.nameid].display = true;
      } else {
        $scope.cmdbtype[$scope.itemnameid][field.nameid].display = false;
      }
      */
      SessionService.saveobj("cmdbtype", $scope.cmdbtype);
    };

    $scope.doSelect = function(item) {
      //console.log(item.selected)
      if (item.selected == true) {
        //console.log(index)
        item.selected = false;
      } else {
        item.selected = true;
      }
      //console.log("aa"+$scope.datalist[index].selected)
    };
    $scope.checkdisplay = function(value,index,carr) {
      //console.log(value);
      //console.log("index: "+ index);
      if ($scope.cmdbtype[$scope.itemnameid][value.nameid] == undefined) {
        //console.log(value)
        return false
      }
      //console.log($rootScope.dispfield.cmdb[$scope.itemnameid][value.nameid]);
      if ($scope.cmdbtype[$scope.itemnameid][value.nameid]["display"] == true) {
        return true
      }
      return false
    }
    $scope.paginate = function (value,index,carr) {
      var begin, end, index;
      if ($scope.opsFilter != '' && carr.length < $scope.numPerPage) {
        //console.log(value)
        return true;
      }
      begin = ($scope.currentPage - 1) * $scope.numPerPage;
      end = begin + $scope.numPerPage;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };
    var reqCfg = {
      params: {"nameid": $scope.itemnameid},
    };
    var url = BaseUrl + '/cmdb/iteminfo';

    $http.get(url, reqCfg)
      .then(function (response) {
        var data = response.data;
        if(data["code"]==0){
          $scope.fields = data["dataList"]["fields"];
          if ($scope.cmdbtype[$scope.itemnameid] == undefined) {
            $scope.cmdbtype[$scope.itemnameid] = {};
          } else {
            //console.log($scope.cmdbtype);
            return;
          }
          var dispnum = $scope.fields.length;
          for (var i=0;i<dispnum;i++) {
            if (i >= 6) {
              break;
            }
            var fname = $scope.fields[i].nameid;
            $scope.fields[i].display = true;
            $scope.cmdbtype[$scope.itemnameid][fname] = $scope.fields[i];
          }
          SessionService.saveobj("cmdbtype", $scope.cmdbtype);
          $scope.itemname = data["dataList"]["name"];
        }
      });
    reqCfg = {
      params: $scope.input,
    };
    url = BaseUrl + '/cmdb/itemsdb/' + $scope.itemnameid;
    if ($scope.itemnameid == "cmdbtype") {
      url = BaseUrl + '/cmdb/itemslist'
    }
    $http.get(url, reqCfg)
      .then(function (response) {
        var data = response.data;
        if(data["code"]==0){
          $scope.items = data["dataList"];
          $scope.tablestatus = "1";
          //console.log($scope.items);
        }else{
          layer.msg(data["message"], {icon: 5});
          console.log('errorcode:'+data["code"]+'errorMessage:'+data["message"]);
        }
      },function(data){
        layer.msg('程序内部错误!', {icon: 5});
        console.log('request failed!');
      });
    $scope.cancel = function () {
      //$location.search({});
      window.history.back();
      //$location.path('/itemslist');
    };
  $scope.isshow = 0;
  $scope.customlistshow = function () {
    if($scope.isshow == 0) {
      $("#customlist").activateBox();
      $scope.isshow = 1;
    } else {
      $("#customlist").removeBox();
      $scope.isshow = 0;
    }
  }
}]);
