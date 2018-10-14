/**
 * Created by Administrator on 2015/11/19.
 */

 (function(){
 'use strict';
   var app = angular.module('ngLoadScript', []);
   app.directive('scriptlazy', function() {
     return {
       restrict: 'E',
       scope: false,
       link: function(scope, elem, attr){
         var s = document.createElement("script");
         s.type = "text/javascript";
         var src = elem.attr('src');
         if(src!==undefined){
             s.src = src;
         }else{
             var code = elem.text();
             s.text = code;
         }
         document.head.appendChild(s);
         elem.remove();
       }
     };
   });

   angular.module('routeStyles', ['ngRoute'])
     .directive('head', ['$rootScope','$compile','$interpolate',
     function($rootScope, $compile, $interpolate){
       // this allows for support of custom interpolation symbols
       var startSym = $interpolate.startSymbol(),
         endSym = $interpolate.endSymbol(),
         html = ['<link rel="stylesheet" ng-repeat="(routeCtrl, cssUrl) in routeStyles" ng-href="',startSym,'cssUrl',endSym,'">'].join('');
         return {
         restrict: 'E',
         link: function(scope, elem){
           elem.append($compile(html)(scope));
           scope.routeStyles = {};
           $rootScope.$on('$routeChangeStart', function (e, next) {
             if(next && next.$$route && next.$$route.css){
               if(!angular.isArray(next.$$route.css)){
                 next.$$route.css = [next.$$route.css];
               }
               angular.forEach(next.$$route.css, function(sheet){
                 scope.routeStyles[sheet] = sheet;
               });
             }
           });
           $rootScope.$on('$routeChangeSuccess', function(e, current, previous) {
             if (previous && previous.$$route && previous.$$route.css) {
               if (!angular.isArray(previous.$$route.css)) {
                 previous.$$route.css = [previous.$$route.css];
               }
               if (current.$$route && current.$$route.css && !angular.isArray(current.$$route.css)) {
                 current.$$route.css = [current.$$route.css];
               }
               angular.forEach(previous.$$route.css, function (sheet) {
                 if (!current.$$route || !current.$$route.css || current.$$route.css.indexOf(sheet) === -1) {
                   // Only remove if not also required in the current page.
                   scope.routeStyles[sheet] = undefined;
                 }
               });
             }
           });
         }
       };
     }
   ]);
 })();

var app = angular.module('myApp', ['ng','ngRoute','ngTouch','ngAnimate','ngResource','ngLoadScript','routeStyles','ui.bootstrap','ngSanitize']);
var Environment = {
  "build": "构建环境",
  "develop": "开发环境",
  "pre-release": "预发布环境",
  "release": "正式环境",
  "test": "测试环境",
};
app.run(['$rootScope', '$window', '$location', '$log','$http','BaseUrl','SessionService','permissions','$route',
  function ($rootScope, $window, $location, $log,$http,BaseUrl,SessionService,permissions,$route) {
    var routeChangeSuccessOff = $rootScope.$on('$routeChangeSuccess', routeChangeSuccess);
    $rootScope.Environment = Environment;
    SessionService.sessionrecovery($rootScope);
    //$rootScope.menus = Environment;
    //$rootScope.myenvs = ["test","official"];
    //$rootScope.curenvironment = "develop";
    $rootScope.newdate = function(time) {
      return new Date(time);
    }
    $rootScope.changeenv = function(env) {
      req = {
        method: 'GET',
        url: BaseUrl+'/auth/env',
        params: {"env":env,},
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data["code"]==0){
          $rootScope.curenvironment = $rootScope.Environment[env];
          $rootScope.curenv = env;
          $window.sessionStorage.setItem("curenv",env);
          $route.reload();
        } else if (data["code"] == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data["message"], {icon: 2});
        }
      });
      //console.log($rootScope.curenvironment);
    }
    $rootScope.isOpenLogin = false;
    $rootScope.login = function(index,user,pwd) {
      if (index != -1) {
        layer.close(index);
      }
      var req = {
        method: 'GET',
        url: BaseUrl+'/auth/login',
        params: {name: user,passwd: pwd},
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          var menus = data.dataList.PermissionList;
          SessionService.sessioninit(
            $rootScope,
            user,
            'tokenid11111',
            data.dataList.workenv,
            menus,
            data.dataList.name,
            data.dataList.roles.join(","),
            data.dataList.fullname,
          );
          //存储权限信息
          permissions.setPermissions(menus);
          $rootScope.isOpenLogin = false;
          $window.location.reload();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {icon: 5});
      });
    }
    $rootScope.loginform = function() {
        if ($rootScope.isOpenLogin) {
          return;
        }
        $rootScope.isOpenLogin = true;
        var username,passwd;
      var content = `<div style="margin : 10px 10px 20px 30px;">
      用户： <input type="text" class="l-username" name="username" style="margin:5px" size="30">
      </br>
      密码： <input type="password" class="l-password" name="password" style="margin:5px" size="30">
      </div>`;
      layer.open({
        type: 1 //Page层类型
        ,btn: ['登录']
        ,area: ['400px', '200px']
        ,title: '登录：'
        ,content: content
        ,resize: false
        ,btnAlign: 'c'
        ,shade: 0.6 //遮罩透明度
        ,anim: 1 //0-6的动画形式，-1不开启
        ,success: function(layero){
          username = layero.find('.l-username');
          passwd = layero.find('.l-password');
          username.focus();
        }
        ,yes: function(index){
          var user = username.val();
          var pwd = passwd.val();
          $rootScope.login(index,user,pwd);
        },
      });
    }
    $rootScope.resetpwd = function () {
      var prompt1,prompt2,prompt3;
      var content = '&nbsp;&nbsp;&nbsp;&nbsp;新密码： <input type="password" class="pp2" style="margin:5px" size="30">'+
      '<br> &nbsp;&nbsp;&nbsp;&nbsp;新密码： <input type="password" class="pp3" style="margin:5px" size="30">';
      content = '&nbsp;&nbsp;&nbsp;&nbsp;旧密码： <input type="password" class="pp1" style="margin:5px" size="30"> <br> ' + content;
      layer.open({
        "type": 1 //Page层类型
        ,"btn": ['&#x786E;&#x5B9A;','&#x53D6;&#x6D88;']
        ,"area": ['300px', '300px']
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
            "data": {"name": $rootScope.username,"newpasswd": newpwd,"oldpasswd": oldpwd},
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
    function routeChangeSuccess(event,current,previous) {
        jQuery(document).scrollTop(0);//页面跳转完成时 scrollTop
        //$log.log('routeChangeSuccess');
        //$log.log(current);
        //$log.log(previous);
    }

    $rootScope.logList=[];
    $rootScope.$watch("logList",function(newVal,oldVal){
        //console.log(oldVal,newVal);
        if(oldVal==newVal){
            return;
        }
        if(newVal.length==0){
            return;
        }
        newVal.pop()
    },true);


}]);


//定义permissions服务
app.factory('permissions', ['$rootScope',function ($rootScope) {
    var permissionList;
    return {
        "setPermissions": function(permissions) {
            permissionList = JSON.parse(JSON.stringify(permissions));
            $rootScope.$broadcast('permissionsChanged')
        },
        "hasPermission": function (permissions) {
            if (typeof(permissionList) != "object") {
              return false;
            }
            if ($rootScope.roles == "admin")
             return true;
            perms = permissions.split(";");
            for (k = 0; k < perms.length; k++) {
              if("object" == typeof permissionList && permissionList.api.hasOwnProperty(perms[k]) && permissionList.api[perms[k]] == true) {
                return true;
              }
            }
            if (permissionList.environment.hasOwnProperty($rootScope.curenv)) {
              for (k = 0; k < perms.length; k++) {
                if (permissionList.environment[$rootScope.curenv].api.hasOwnProperty(perms[k]) && permissionList.environment[$rootScope.curenv].api[perms[k]]) {
                  return true;
                }
              }
            }
            // if(permissionList.api.hasOwnProperty(permapi)) {
            //   if(permissionList.api[permapi][permmethod] == true) {
            //     //if(permissionList.indexOf(permission.trim()) > -1 || permissionList[0] == "admin"){
            //     return true;
            //   }
            // }
            return false;
        }
    };
}])

//定义has-Permission指令
app.directive('hasPermission', ['permissions',function(permissions) {
    return {
        restrict: 'AECM',
        link: function (scope, element, attrs) {
            if (!angular.isString(attrs["hasPermission"]))
                throw "hasPermission value must be a string";

            var value = attrs["hasPermission"].trim();
            var notPermissionFlag = value[0] === '!';
            if (notPermissionFlag) {
                value = value.slice(1).trim();
            }

            function toggleVisibilityBasedOnPermission() {
                var hasPermission = permissions["hasPermission"](value);

                if (hasPermission && !notPermissionFlag || !hasPermission && notPermissionFlag)
                    element.show();
                else
                    element.hide();
            }

            toggleVisibilityBasedOnPermission();
            scope.$on('permissionsChanged', toggleVisibilityBasedOnPermission);
            }
        };
    }])
    .directive('jqSortable', ['$parse', function($parse) {
    return function(scope, element, attrs) {
        /*解析并取得表达式*/
        var expr = $parse(attrs['jqSortable']);
        var $oldChildren;

        element.sortable({
            "opacity": 0.7,
            "scroll": false,
            "tolerance": "pointer",
            "start": function() {
                /*纪录移动前的 children 顺序*/
                $oldChildren = element.children('[ng-repeat]');
            },
            "update": function(){
                var newList = [];
                var oldList = expr(scope);
                var $children = element.children('[ng-repeat]');

                /*产生新顺序的数组*/
                $oldChildren.each(function(i){
                    var index = $children.index(this);
                    if(index == -1){ return; }
                    oldList[i]['sort']=(index+1);
                    newList[index] = oldList[i];
                });
                /*将新顺序的数组写回 scope 变量*/
                expr.assign(scope, newList);
                /*通知 scope 有异动发生*/
                scope.$digest();
            }
        });

        /*在 destroy 时解除 Sortable*/
        scope.$on('$destroy', function(){
            element.sortable('destroy');
        });
    };
}]);



//将权限放入内存中，初始化permissionList值，没有任何权限
// app.run(function(permissions) {
//     permissions.setPermissions(permissionList);
// });
/*
app.config(['$httpProvider', function($httpProvider) {
  //  $httpProvider.interceptors.push('securityInterceptor');
    $httpProvider.interceptors.push('sessionInjector');
}
]);*/
app.config(['$routeProvider','$locationProvider','$httpProvider', function ($routeProvider,$locationProvider,$httpProvider) {
    $locationProvider.html5Mode(true);
    $httpProvider.interceptors.push('sessionInjector');//注入自定义拦截器
    //$httpProvider.defaults.withCredentials = true;
    $routeProvider.
        //登录
        when('/login', {
            templateUrl: 'html/login.html',
            controller: 'LoginController',
            css: []}).
        //主页
        when('/base', {
            //templateUrl: 'html/base.html',
            template: basetpl(),
            controller: 'BaseController',
            css: ['']}).
        //配置项管理  ->  配置项列表
        when('/cmdb/app', {
            templateUrl: 'html/lists.html',
            controller: 'CMDBController',
            css: []}).
        when('/cmdb/project/add', {
            templateUrl: 'html/cmdb_post.html',
            controller: 'CMDBPostController',
            css: []}).
        when('/cmdb/project/info/:id', {
            templateUrl: 'html/cmdb_post.html',
            controller: 'CMDBPostController',
            css: []}).
        when('/cmdb/project/update/:id', {
            templateUrl: 'html/cmdb_post.html',
            controller: 'CMDBPostController',
            css: []}).
        when('/cmdb/projectview', {
            templateUrl: 'html/project_view.html',
            controller: 'ProjectController',
            css: []}).
        when('/jobs/jobs', {
            templateUrl: 'html/jobs_view.html',
            controller: 'JobsController',
            css: []}).
        when('/jobs/flow', {
            templateUrl: 'html/jobsflow.html',
            controller: 'JobsFlowController',
            css: []}).
        when('/jobs/runlogs', {
            template: RunLogTpl(),
            controller: 'JobsController',
            css: []}).
        when('/jobs/ssh', {
            templateUrl: 'html/webssh.html',
            controller: 'SSHController',
            css: []}).
        when('/cmdb/apptpl/add', {
            //templateUrl: 'html/apptpl_post.html',
            template: appPostTpl(),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/apptpl/info/:id', {
            //templateUrl: 'html/apptpl_post.html',
            template: appPostTpl(),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/apptpl/update/:id', {
            //templateUrl: 'html/apptpl_post.html',
            template: appPostTpl(),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/apptpl/list', {
            //templateUrl: 'html/apptpl_list.html',
            template: apptplListTpl(),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/resource/add', {
            templateUrl: 'html/resource_post.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/resource/info/:id', {
            templateUrl: 'html/resource_post.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/resource/update/:id', {
            templateUrl: 'html/resource_post.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/resource/list', {
            templateUrl: 'html/resource_list.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/yunprovider/add', {
            // templateUrl: 'html/yunprovider_post.html',
            template: yunTpl(false,false),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/yunprovider/info/:id', {
            templateUrl: 'html/yunprovider_post.html',
            // template: yunTpl(false,false),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/yunprovider/update/:id', {
            templateUrl: 'html/yunprovider_post.html',
            // template: yunTpl(false,false),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/yunprovider/list', {
            templateUrl: 'html/yunprovider_list.html',
            // template: yunTpl(false,true),
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/idcprovider/add', {
            templateUrl: 'html/idcprovider_post.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/idcprovider/info/:id', {
            templateUrl: 'html/idcprovider_post.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/idcprovider/update/:id', {
            templateUrl: 'html/idcprovider_post.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/idcprovider/list', {
            templateUrl: 'html/idcprovider_list.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/resource/yunimport/:id', {
            templateUrl: 'html/import.html',
            controller: 'ApptplController',
            css: []}).
        when('/cmdb/resource/idcimport/:id', {
            templateUrl: 'html/import.html',
            controller: 'ApptplController',
            css: []}).
        //登出
        when('/logout', {
            templateUrl: 'html/login.html',
            controller: 'LogoutController',
            css: []}).
        //用户管理
        when('/userlist', {
            templateUrl: 'html/user_list.html',
            controller: 'UserController',
            css: []}).
        //用户添加
        when('/useradd', {
            templateUrl: 'html/user_add.html',
            controller: 'UserController',
            css: []}).
        //用户修改
        when('/userupdate/:id', {
            templateUrl: 'html/user_add.html',
            controller: 'UserController',
            css: []}).
        //用户密码重置
        when('/updatepwd/:id', {
            templateUrl: 'html/user_updatepwd.html',
            controller: 'UserController',
            css: []}).
        when('/rolelist', {
            templateUrl: 'html/roles.html',
            controller: 'RolesController',
            css: []}).


        otherwise({redirectTo: '/base'});
}]);



app.filter('propsFilter', function() {
    return function(items, props) {
        var out = [];

        if (angular.isArray(items)) {
            var keys = Object.keys(props);

            items.forEach(function(item) {
                var itemMatches = false;

                for (var i = 0; i < keys.length; i++) {
                    var prop = keys[i];
                    var text = props[prop].toLowerCase();
                    if (item[prop].toString().toLowerCase().indexOf(text) !== -1) {
                        itemMatches = true;
                        break;
                    }
                }

                if (itemMatches) {
                    out.push(item);
                }
            });
        } else {
            // Let the output be the input untouched
            out = items;
        }

        return out;
    };
});



app.constant('BaseUrl', 'http://console.zdops.com/eops');

//common labrary

(function(glable) {

    glable._assert = function(condition, message){
        if (!condition) {
            message = message || "Assertion failed";
            if (typeof Error !== "undefined") {
                throw new Error(message);
            }
            throw message; // Fallback
        }
    };
    if (!String.prototype.format) {
        /**
         * 格式化字符串
         * @Usage
         *     '{0} {1}'.format(null, 'b') => '{0} b'
         *     '{0} {1}'.format(true, null, 'b') => ' b'
         */
        String.prototype.format = function () {
            var args = arguments;
            var is_non_match_to_empty = false;
            if (typeof args[0] === 'boolean') {
                is_non_match_to_empty = args[0];
            }
            return this.replace(/{(\d+)}/g, function (match, number) {
                number = parseInt(number, 10);

                if (is_non_match_to_empty) {
                    return args[number + 1] != null ? args[number + 1] : '';
                } else {
                    return args[number] != null ? args[number] : match;
                }
            });
        };
    }
    String.prototype.zhlen = function() {
      var zhlen = 0;
      for (i=0;i<this.length;i++) {
        zhlen++;
        var char = this.charCodeAt(i)
        if(char >= 19968 && char <= 40869) {
          zhlen++;
        }
      }
      return zhlen;
    }
    String.prototype.bytelen = function() {
      var bytelen = 0;
      for (i=0;i<this.length;i++) {
        var char = this.charCodeAt(i)
        while(char>0) {
          char = char >> 8;
          bytelen++;
        }
      }
      return bytelen;
    }
    String.prototype.beginsWith = function(prefix){
        return this.lastIndexOf(prefix, 0) === 0;
    };

    String.prototype.endsWith = function(subfix){
        return this.lastIndexOf(subfix) === this.length-1;
    };

    String.prototype.crlf2html = function(){
        return this.replace(/\n/g,"<br/>");
    };
    // convert js string to html entity.
    // http://stackoverflow.com/questions/18749591/encode-html-entities-in-javascript
    String.prototype.escapeHtml = function(){
      //[\u00A0-\u9999<>\&]
        return this.replace(/[^\u0030-\u0039\u0041-\u005A\u0061-\u007A]/gim, function(i) {
            return '&#'+i.charCodeAt(0)+';';
        });
    };

    //@deprecated, use _.flatten instead
    if(!Array.prototype.flatten){
        Array.prototype.flatten = function flatten(){
            var flat = [];
            for (var i = 0, l = this.length; i < l; i++){
                var type = Object.prototype.toString.call(this[i]).split(' ').pop().split(']').shift().toLowerCase();
                if (type) { flat = flat.concat(/^(array|collection|arguments|object)$/.test(type) ? flatten.call(this[i]) : this[i]); }
            }
            return flat;
        };
    }

    //todo: replace _indexOf with this function
    if(!Array.prototype.indexOfBy){
        Array.prototype.indexOfBy = function(predict){
            _assert(typeof predict === 'function', 'invalid param');

            for (var _i = 0; _i < this.length; _i++) {
                var e = this[_i];
                if (predict(e)){
                    return _i;
                }
            }

            return -1;
        };
    }
    if(!Date.prototype.format) {
      Date.prototype.format = function(format) {
        var o = {
          "y": this.getFullYear() - 2000,
          "Y": this.getFullYear(),
          "m" : this.getMonth()+1, //month
          "d" : this.getDate(), //day
          "H" : this.getHours(), //hour
          "M" : this.getMinutes(), //minute
          "S" : this.getSeconds(), //second
          "s" : this.getTime() //millisecond
        };
        var datestr = "";
        for (var i=0,l=format.length;i<l;i++) {
          var s = format[i];
          if (s == '%') {
            s = format[++i];
            if (s in o) {
              datestr += o[s]>9?o[s]:'0'+o[s];
            }
          } else {
            datestr += s;
          }

        }
        return datestr;
      }
    }
})(window);

app.directive('icheck', function($timeout, $parse) {
  return {
    link: function($scope, element, $attrs) {
      return $timeout(function() {
        var ngModelGetter, value;
        ngModelGetter = $parse($attrs['ngModel']);
        value = $parse($attrs['ngValue'])($scope);
        if ($attrs['ngShow'] != undefined) {
          if ($scope.$eval($attrs['ngShow']) ==  false) {
            return;
          }
        }
        return $(element).iCheck({
          checkboxClass: 'icheckbox_flat-blue',
          radioClass: 'iradio_flat-blue',
          increaseArea: '20%'
        }).on('ifChanged', function(event) {
          if ($(element).attr('type') === 'checkbox' && $attrs['ngModel']) {

            $scope.$apply(function() {
              return ngModelGetter.assign($scope, event.target.checked);
            });
            if ($attrs["name"] == "customFields") {
              $scope.$parent.upDisplayFields($scope.field);
            }
            //console.log($attrs['ngModel'].substr(0,8));
            //console.log($attrs);
          }
          if ($(element).attr('type') === 'radio' && $attrs['ngModel']) {
            return $scope.$apply(function() {
              return ngModelGetter.assign($scope, value);
            });
          }
        });
      });
    }
  };
});

app.directive('datetimepicker', function($timeout, $parse) {
  return {
    link: function($scope, element, $attrs) {
      return $timeout(function() {
        return $(element).datetimepicker({
          language:  'zh-TW',
        weekStart: 1,
        todayBtn:  true,
        autoclose: true,
        todayHighlight: true,
        startView:2,
        forceParse: false,
        });
      });
    }
  };
});


function base64Encode(input){
  var rv;
  rv = encodeURIComponent(input);
  rv = unescape(rv);
  rv = window.btoa(rv);
  return rv;
}


function base64Decode(input){
  var rv;
  rv = window.atob(input);
  //rv = encodeURIComponent(rv);
  //rv = escape(rv);
  return rv;
}

function utf8Decode(string) {
  if (typeof string !== 'string') return string;
  var output = "", i = 0, charCode = 0;

  while (i < string.length) {
    charCode = string.charCodeAt(i);

    if (charCode < 128)
      output += String.fromCharCode(charCode),
      i++;
    else if ((charCode > 191) && (charCode < 224))
      output += String.fromCharCode(((charCode & 31) << 6) | (string.charCodeAt(i + 1) & 63)),
      i += 2;
    else
      output += String.fromCharCode(((charCode & 15) << 12) | ((string.charCodeAt(i + 1) & 63) << 6) | (string.charCodeAt(i + 2) & 63)),
      i += 3;
  }

  return output;
}

$.widget( "custom.zdops_autocomplete", $.ui.autocomplete, {
  _renderItemData: function(ul,item){
    var li = $('<li aria-label="'+item.text+'" data-id="'+item.id+'"><p class="zdops-menu-item-text">'+item.text+'</p></li>');
    li.data('ui-autocomplete-item', item)
    ul.append(li);
    return li;
  },
});
