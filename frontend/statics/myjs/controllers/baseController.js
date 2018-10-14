
//登陆控制器
app.controller('LoginController', ['$scope', '$http', '$location', '$rootScope',
  'SessionService','BaseUrl','$uibModal','$timeout',function ($scope, $http,
  $location, $rootScope,SessionService,BaseUrl,$uibModal,$timeout) {
    // console.log('go into loginController');
    $scope.open = function (size) {
        var modalInstance = $uibModal.open({
            animation: false,
            templateUrl: 'myModalContent.html',
            controller: 'ModalInstanceCtrl',
            size: 'lgggg',
            backdrop:'static',//是否有背景 true   false  static 点击模块外部不关闭
            keyboard:false,//禁用esc
            resolve: {
                items: function () {
                    return $scope.items;
                }
            }
        });
        modalInstance.result.then(function (selectedItem) {

        }, function () {
            //console.info('Modal dismissed at: ' + new Date());
        });
    };
    $scope.open();
    // $timeout(function() {
    //
    //     angular.element(document.querySelector('#modalclick')).trigger('click');
    // },0);



}]);


//登陆modal控制器
app.controller('ModalInstanceCtrl', ['$scope', '$uibModalInstance','BaseUrl','$rootScope',
  '$location','$http','SessionService','permissions',function ($scope,
    $uibModalInstance,BaseUrl,$rootScope,$location,$http,SessionService,permissions) {
    // console.log('go into ModalInstanceCtrl');
    $scope.login = function (userdata) {
        //console.log('调用service 验证登录信息及相关操作');
        var url = BaseUrl+'/auth/login'
        reqCfg = {
          "params": {"name": userdata.username,"passwd": userdata.password},
        };

        $http.get(url, reqCfg).then(function (response) {
          /*  data={"code":0,"status":"SUCCESS","message":"登录成功","errors":null,"user":{"user_id":"9999","user_name":"tom","phone":"","email":""},"roles":["admin"],"permissionlist":[
                "system:manage:page", "adv:manage:page", "media:manage:page", "developer:manage:page", "tag:manage:page","channel:manage:page","put:manage:page"
            ]}]};
          */
          //console.log(data);
          var data = response.data;
          if(data.code==0){
            var menus = data.dataList.PermissionList;
            SessionService.sessioninit($rootScope,
              userdata.username,
              'tokenid11111',
              data.dataList.workenv,
              menus,
              data.dataList.name,
              data.dataList.roles.join(","),
              data.dataList.fullname,
            );
            //存储权限信息
            permissions.setPermissions(menus);
            //console.log(permissionList);
            //console.log($rootScope.referer);
            path = "/base";
            $uibModalInstance.dismiss('cancel');
            if ($rootScope.hasOwnProperty("referer")) {
              path = $rootScope.referer;
            }
            //console.log(path);
            $location.path(path);
          } else if(data.code ==1) {
            $location.path('/login');
          } else {
              $scope.loginmessage = data.message;
          }
        }, function(){
          $scope.loginmessage = "登录异常";
        });
    }


}]);

//登出控制器
app.controller('LogoutController', ['$scope', '$http', '$location', '$rootScope','SessionService',function ($scope, $http, $location, $rootScope,SessionService) {
    console.log('go into LogoutController');

    SessionService.sessiondestory();
    $location.path("/login");


}]);


app.controller('BaseController', ['$scope', '$location', 'permissions','MyHTTP',
  function($scope, $location, permissions,MyHTTP) {
    $scope.$on('$routeChangeStart', function(scope, next, current) {
      var permission = next.$$route.permission;
      if(angular.isString(permission) && !permissions.hasPermission(permission))
        $location.path('/');
    });
    //初始化本地缓存
    //MyHTTP.cmdblist();

}]);
