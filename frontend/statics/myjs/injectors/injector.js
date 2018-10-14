/**
 * Created by Administrator on 2015/11/19.
 */


//          拦截器对请求
app.factory('sessionInjector', ['ComService','SessionService','$rootScope','$window','$q', function(ComService,SessionService,$rootScope,$window,$q) {
  var sessionInjector = {
    request: function(config) {
      if(config.url=='login.html'){return config;}
      if(config.url.indexOf('https://')!=-1){
        config.withCredentials=true;
        return config;
      }
      if (!SessionService.isAnonymus) {
        //已登录
        config.headers['x-session-token'] = $rootScope.token;
      }else{
        $rootScope.cmdbitems = SessionService.getobj("cmdbitems");
        //从sessionstorage中取信息
        SessionService.sessionrecovery($rootScope);
      }
      return config;
    },
    response: function(response) {
      /*
      response属性：
      {
      "code":
      "dataList": http响应的正文
      "totalRecord":
      "message":
      }
      */
      return response;
    },
    responseError: function(errorResponse) {
      switch (errorResponse.status) {
        case 403:
          $window.location = '/403';
          break;
        case 500:
          $window.location = '/500';
          break;
        case 900:
          $window.location = '/logout';
          break;
        }
        return $q.reject(errorResponse);
      }
    };
    return sessionInjector;
}]);

//拦截请求。没有使用
app.provider('securityInterceptor', function() {
    this.$get = function($location, $q) {
        return function(promise) {
            return promise.then(null, function(response) {
                if(response.status === 403 || response.status === 401) {
                    $location.path('/error');
                }
                return $q.reject(response);
            });
        };
    };
});
