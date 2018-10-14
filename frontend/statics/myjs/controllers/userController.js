/**
 * Created by mp on 2016/1/14.
 */
//用户列表
app.controller('UserController',
  ['$rootScope','$scope','$http','BaseUrl','$location','$route','$routeParams','MyHTTP',
  function ($rootScope,$scope, $http,BaseUrl,$location,$route,$routeParams,MyHTTP) {
    // console.log('go into UserListController');
    $scope.enter=function($event){
      if($event.keyCode==13){
        $scope.pageChanged(1);
      }
    }
    //分页
    $scope.totalItems = 0;
    $scope.currentPage = 1;
    $scope.maxSize = 5;
    $scope.userFilter = '';
    $scope.sortType     = 'name';
    $scope.sortReverse  = false;
    $scope.tablestatus="2";//列表是否为空 1  不空 2空
    $scope.numPerPage = "10";
    $scope.datalist = [];
    $scope.currentNum = 0;
    $scope.selectAll = false;
    $scope.isnew = true;
    $scope.user={
      "name":"",
      "fullname":"",
      "mobile":"",
      "roles": [],
    };

    $scope.paginate = function (value,index,carr) {
      var begin, end;
      $scope.totalItems = carr.length;
      numpp = parseInt($scope.numPerPage);
      begin = ($scope.currentPage - 1) * numpp;
      end = begin + $scope.numPerPage;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };
    $scope.doSelectAll = function() {
      var filtered = $filter('filter')($scope.datalist, $scope.itemFilter);
      sorted = $filter('orderBy')(filtered,$scope.sortType,$scope.sortReverse);
      numpp = parseInt($scope.numPerPage);
      begin = ($scope.currentPage - 1) * numpp;
      end = begin + numpp;
      if ($scope.selectAll == true) {
        $scope.selectAll = false;
      } else {
        $scope.selectAll = true;
      }
      var i = 0;
      angular.forEach(sorted, function(item) {
         if (begin <= i && i < end) {
           item.selected = $scope.selectAll;
         }
         i++;
      });
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
    $scope.pageChanged = function(search) {
        if(search){
            $scope.input.pageNo=1;
            $scope.currentPage=1;
        }else {
            $scope.input.pageNo=$scope.currentPage;
        }
        $scope.requestpageutil($http,url,$scope.input);

    };

    var reqrole = {
      "method": 'GET',
      "url": BaseUrl+'/auth/roles',
    };
    $http(reqrole).then(function (response) {
      var data = response.data;
      if(data.code==0){
        $scope.rolelist=data.dataList;
      }
    });
    var url = BaseUrl+'/auth/user';
    var lpath = $location.path();
    if (lpath == "/userlist") {
      req = {
        "method": 'GET',
        "url": url,
        "params": {"pageNo":1,"pageSize":1000,},
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          $scope.datalist=data.dataList;
          $scope.totalItems=data.totalRecord;
          $scope.tablestatus="1";
        } else if (data.code == 1199) {
          $rootScope.loginform();
          // $rootScope.referer = $location.path();
          // $location.path('/login');
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {"icon": 5});
      });
    } else if (lpath == "/useradd") {
      $scope.isnew = true;
      $scope.userop = "添加用户";
    } else if (lpath.indexOf('/userupdate/') > -1) {
      $scope.isnew = false;
      $scope.userop = "更新用户";
      req = {
        "method": 'GET',
        "url": BaseUrl+'/auth/user',
        params: {"id":$routeParams.id},
      };
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          $scope.user.name = data.dataList[0].name;
          $scope.user.fullname = data.dataList[0].fullname;
          $scope.user.mobile = data.dataList[0].mobile;
          $scope.user.roles = data.dataList[0].roles;
          $scope.user.passwd = "12345678";
          $scope.passwdagain = "12345678";
        } else if (data.code == 1199) {
          $rootScope.loginform();
          //$rootScope.referer = $location.path();
          //$location.path('/login');
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {"icon": 5});
      });
    }
    $scope.ok = function (isValid) {
      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {"icon": 5});
        return;
      }
      req = {
        "method": 'POST',
        "url": BaseUrl+'/auth/user',
      }
      req.headers = {'Content-Type': "application/json; charset=UTF-8"};
      req.data = $scope.user;
      if ($routeParams.hasOwnProperty("id")) {
        req.method = 'PUT';
        delete $scope.user.passwd;
      }
      MyHTTP.http(req,"/userlist");
    }
    // 用户删除
    $scope.deleteOk = function () {
      //console.log( $scope.user);
      req.method = 'DELETE';
      req.url = BaseUrl+'/auth/user';
      req.headers = {'Content-Type': "application/json; charset=UTF-8"};
      req.params = {"name": $scope.user.name};
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('删除成功！', {"icon": 1,"end":function(){
            $route.reload();
          }});
        } else if (data.code == 1199) {
          $rootScope.loginform();
          // $rootScope.referer = $location.path();
          // $location.path('/login');
        } else {
          layer.msg(data.message, {"icon": 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {"icon": 5});
      });
    };
    var updateSelected = function(action,name){
      if(action == 'add' && $scope.user.roles.indexOf(name) == -1){
        $scope.user.roles.push(name);
      }
      if(action == 'remove' && $scope.user.roles.indexOf(name)!=-1){
        var idx = $scope.user.roles.indexOf(name);
        $scope.user.roles.splice(idx,1);
      }
    }

    $scope.updateSelection = function($event, id){
      var checkbox = $event.target;
      var action = (checkbox.checked?'add':'remove');
      updateSelected(action,checkbox.name);
    }
    $scope.isSelected = function(id){
      return $scope.user.roles.indexOf(id)>=0;
    }

    $scope.cancel = function () {
      window.history.back();
       // $uibModalInstance.dismiss('cancel');
    };
    $scope.newrow = function(index) {
      if (index % 3 === 0 && index != 0) {
        return "row"
      }
      return ""
    }
    //用户删除
    $scope.deleteuser = function (user) {
      $scope.user=user;
      layer.confirm('您是要删除　<span class="text-red">'+user.name+'</span>　用户？', {
        "title":"删除用户",
        "btn": ['确定','取消'] //按钮
      }, function(){
        $scope.deleteOk();
        $route.reload();
      }, function(){

      });
    };

    //重置用户密码
    $scope.resetpwd = function (user) {
      var prompt1,prompt2,prompt3;
      var content = '&nbsp;&nbsp;&nbsp;&nbsp;新密码： <input type="password" class="pp2" style="margin:5px" size="30">'+
      '<br> &nbsp;&nbsp;&nbsp;&nbsp;新密码： <input type="password" class="pp3" style="margin:5px" size="30">';
      if ($rootScope.roles != "admin") {
        content = '&nbsp;&nbsp;&nbsp;&nbsp;旧密码： <input type="password" class="pp1" style="margin:5px" size="30"> <br> ' + content;
      }
      layer.open({
        "type": 1 //Page层类型
        ,"btn": ['&#x786E;&#x5B9A;','&#x53D6;&#x6D88;']
        ,"area": ['300px', '250px']
        ,"title": '重置密码：'
        ,"content": content
        ,"resize": false
        ,"btnAlign": 'c'
        ,"shade": 0.6 //遮罩透明度
        ,"anim": 1 //0-6的动画形式，-1不开启
        ,"success": function(layero){
          prompt1 = layero.find('.pp1');
          prompt2 = layero.find('.pp2');
          prompt3 = layero.find('.pp3');
          prompt1.focus();
        }
        ,"yes": function(index){
          var oldpwd = prompt1.val();
          var newpwd = prompt2.val();
          var newpwd1 = prompt3.val();
          layer.close(index);
          if(newpwd != newpwd1) {
            layer.msg("密码输入不一致！", {"icon": 5});
            return;
          }
          req = {
            "method": 'PUT',
            "url": BaseUrl+'/auth/pwd',
            "headers": {
              'Content-Type': "application/json; charset=UTF-8"
            },
            "data": {"name": user,"newpasswd": newpwd,"oldpasswd": oldpwd},
          }
          $http(req).then(function (response) {
            var data = response.data;
            if(data.code==0){
              layer.msg('密码修改成功！', {"icon": 1});
            } else if (data.code == 1199) {
              $rootScope.loginform();
              // $rootScope.referer = $location.path();
              // $location.path('/login');
            } else {
              layer.msg(data.message, {"icon": 2});
            }
          },function(data){
            layer.msg('无法获取信息!', {"icon": 5});
          });
        },
      });
    };
}]);
