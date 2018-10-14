/**
 * Created by Administrator on 2015/11/19.
 */

//处理session

app.service('SessionService',['$window','$location','permissions',function($window,$location,permissions){

    this.isAnonymus=true;
    this.sessioninit = function($rootScope,username,token,env,menus,userid,roles,fullname){
        var menustext= JSON.stringify(menus);
        $rootScope.userid=userid;
        $rootScope.username=username;
        $rootScope.fullname=fullname;
        $rootScope.roles=roles;
        $rootScope.token=token;
        $rootScope.menus=menus;
        //envnum = $rootScope.menus.Environment.length;
        $rootScope.curenv = env;
        $rootScope.curenvironment = $rootScope.Environment[$rootScope.curenv];
        this.isAnonymus=false;
        $window.sessionStorage.setItem("userid",userid);
        $window.sessionStorage.setItem("username",username);
        $window.sessionStorage.setItem("fullname",fullname);
        $window.sessionStorage.setItem("roles",roles);
        $window.sessionStorage.setItem("token",token);
        $window.sessionStorage.setItem("curenv",env);
        $window.sessionStorage.setItem("menustext",menustext);
        //显示隐藏的div
        //$("#parentDiv").show();
        //$(".spinner").hide();
    }
    this.sessionrecovery=function($rootScope){
        if($window.sessionStorage.getItem("username")!=null){
            this.isAnonymus=false;
            $rootScope.userid=$window.sessionStorage.getItem("userid");
            $rootScope.username=$window.sessionStorage.getItem("username");
            $rootScope.fullname=$window.sessionStorage.getItem("fullname");
            $rootScope.roles=$window.sessionStorage.getItem("roles");
            $rootScope.token=$window.sessionStorage.getItem("token");
            $rootScope.curenv=$window.sessionStorage.getItem("curenv");
            var menustext=$window.sessionStorage.getItem("menustext");
            $rootScope.menus=JSON.parse(menustext);
            if (typeof $rootScope.menus != 'object') {
              $location.path("/login");
            }
            permissions.setPermissions($rootScope.menus);//重新装载权限信息
            $rootScope.curenvironment = $rootScope.Environment[$rootScope.curenv];
            //显示隐藏的div
            //$("#parentDiv").show();
            //$(".spinner").hide();
        }else{
            $location.path("/login");
        }

    }
    this.sessiondestory=function(){

        this.isAnonymus=true;
        $window.sessionStorage.clear();

    }

    this.saveobj=function(k, obj){
        var objtext= JSON.stringify(obj);
        $window.sessionStorage.setItem(k,objtext);
    }
    this.getobj=function(k){
        var objtext= $window.sessionStorage.getItem(k)
        if (objtext == null) {
          return {};
        }
        return JSON.parse(objtext);
    }


}]);
app.service('UtilService',['$window','$location',function($window,$location){
// 获取间隔天数
    this.getDays=function(day1, day2) {
        // 获取入参字符串形式日期的Date型日期
        var d1 = day1.getDate();
        var d2 = day2.getDate();

        // 定义一天的毫秒数
        var dayMilliSeconds  = 1000*60*60*24;

        // 获取输入日期的毫秒数
        var d1Ms = d1.getTime();
        var d2Ms = d2.getTime();

        // 定义返回值
        var ret; //返回逗号分隔的时间 如 2001-01-01,2001-01-02
        var retlist=[]; //返回时间的list 如[2001-01-01,2001-01-02]

        // 对日期毫秒数进行循环比较，直到d1Ms 大于等于 d2Ms 时退出循环
        // 每次循环结束，给d1Ms 增加一天
        for (;d1Ms <= d2Ms; d1Ms += dayMilliSeconds) {

            // 如果ret为空，则无需添加","作为分隔符
            if (!ret) {
                // 将给的毫秒数转换为date日期
                var day = new Date(d1Ms);

                // 获取其年月日形式的字符串
                ret = day.getYMD();
            } else {

                // 否则，给ret的每个字符日期间添加","作为分隔符
                var day = new Date(d1Ms);
                ret = ret + ',' + day.getYMD();
            }
            var day_list = new Date(d1Ms);
            retlist.push({
                "date":day_list.getYMD(),
                "isuniformspeed":"1",
                "throwintime":"7,9;12,14",
                "total":0
            });
        }
        return(retlist); // 或可换为return ret;
    }
// 给Date对象添加getYMD方法，获取字符串形式的年月日
    Date.prototype.getYMD = function(){
        var retDate = this.getFullYear() + "-";  // 获取年份。
        if(this.getMonth()<9){
            retDate +=0;
        }
        retDate += this.getMonth() + 1 + "-";    // 获取月份。
        if(this.getDate()<=9){
            retDate +=0;
        }
        retDate += this.getDate();               // 获取日。
        return retDate;                          // 返回日期。
    }
    // 给Date对象添加changeUTCDatetoLocal方法，获取字符串形式的年月日
    Date.prototype.changeUTCDatetoLocal = function(){
    //  console.log(moment(this).utc().zone(-8).format('YYYY-MM-DD'));
     //  return '1111';
      return  moment(this).utc().zone(-8).format('YYYY-MM-DD');
      //return  moment(this).utc().zone(-8).format('YYYY-MM-DD HH:mm:ss');

    }


    // 给Date对象添加getPreDay方法，获取上一天
    Date.prototype.getPreDay = function(){
        var   yesterday_milliseconds=this.getTime()-1000*60*60*24;
        var   yesterday=new   Date();
        yesterday.setTime(yesterday_milliseconds);
        return new Date(yesterday);
    }
    // 给Date对象添加getNextDay方法，获取下一天
    Date.prototype.getNextDay = function(){
        var   yesterday_milliseconds=this.getTime()+1000*60*60*24;
        var   yesterday=new   Date();
        yesterday.setTime(yesterday_milliseconds);
        return new Date(yesterday);
    }
    // 给Date对象添加getPreWeekBegin方法，获取上一周开始时间
    Date.prototype.getPreWeekBegin = function(){
        var nowDayOfWeek = this.getDay();         //今天本周的第几天
        var nowDay = this.getDate();              //当前日
        var nowMonth = this.getMonth();           //当前月
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        if(nowDayOfWeek != 0){
            var preWeekBegin = new Date(nowYear, nowMonth, nowDay - nowDayOfWeek -6);
        }else{
            var preWeekBegin = new Date(nowYear, nowMonth, nowDay - 13);
        }
        return preWeekBegin;
    }
    // 给Date对象添加getPreWeekEnd方法，获取上一周结束时间
    Date.prototype.getPreWeekEnd = function(){
        var nowDayOfWeek = this.getDay();         //今天本周的第几天
        var nowDay = this.getDate();              //当前日
        var nowMonth = this.getMonth();           //当前月
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        if(nowDayOfWeek != 0){
            var preWeekEnd = new Date(nowYear, nowMonth,  nowDay + (6 - nowDayOfWeek - 6));
        }else{
            var preWeekEnd = new Date(nowYear, nowMonth,  nowDay -7);
        }
        return preWeekEnd;
    }
    // 给Date对象添加getWeekBegin方法，获取本周开始时间
    Date.prototype.getWeekBegin = function(){
        var nowDayOfWeek = this.getDay();         //今天本周的第几天
        var nowDay = this.getDate();              //当前日
        var nowMonth = this.getMonth();           //当前月
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        if(nowDayOfWeek != 0){
            var WeekBegin = new Date(nowYear, nowMonth, nowDay - nowDayOfWeek+1);
        }else{
            var WeekBegin = new Date(nowYear, nowMonth, nowDay - 6);
        }
        return WeekBegin;
    }
    // 给Date对象添加getWeekEnd方法，获取本周结束时间
    Date.prototype.getWeekEnd = function(){
        var nowDayOfWeek = this.getDay();         //今天本周的第几天
        var nowDay = this.getDate();              //当前日
        var nowMonth = this.getMonth();           //当前月
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        if(nowDayOfWeek != 0){
            var WeekEnd = new Date(nowYear, nowMonth, nowDay + (6 - nowDayOfWeek+1));
        }else{
            var WeekEnd = new Date(nowYear, nowMonth, nowDay);
        }
        return WeekEnd;
    }
    // 给Date对象添加getNextWeekBegin方法，获取下一周开始时间
    Date.prototype.getNextWeekBegin = function(){
        var nowDayOfWeek = this.getDay();         //今天本周的第几天
        var nowDay = this.getDate();              //当前日
        var nowMonth = this.getMonth();           //当前月
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        if(nowDayOfWeek != 0){
            var nextWeekBegin = new Date(nowYear, nowMonth, nowDay +(7-nowDayOfWeek+1));
        }else{
            var nextWeekBegin = new Date(nowYear, nowMonth, nowDay +1);
        }
        return nextWeekBegin;
    }
    // 给Date对象添加getNextWeekEnd方法，获取下一周结束时间
    Date.prototype.getNextWeekEnd = function(){
        var nowDayOfWeek = this.getDay();         //今天本周的第几天
        var nowDay = this.getDate();              //当前日
        var nowMonth = this.getMonth();           //当前月
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        if(nowDayOfWeek != 0){
            var nextWeekEnd = new Date(nowYear, nowMonth, nowDay + (7 - nowDayOfWeek)+7);
        }else{
            var nextWeekEnd = new Date(nowYear, nowMonth, nowDay + 7);
        }

        return nextWeekEnd;
    }
    // 给Date对象添加getPreMonthBegin方法，获取上一月开始时间
    Date.prototype.getPreMonthBegin = function(){
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        var lastMonthDate = new Date(this.getTime());  //上月日期
        lastMonthDate.setDate(1);
        lastMonthDate.setMonth(lastMonthDate.getMonth()-1);
        var lastMonth = lastMonthDate.getMonth();
        var preMonthBegin = new Date(nowYear, lastMonth, 1);
        return preMonthBegin;
    }
    // 给Date对象添加getPreMonthEnd方法，获取上一月结束时间
    Date.prototype.getPreMonthEnd = function(){
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        var lastMonthDate = new Date(this.getTime());  //上月日期
        lastMonthDate.setDate(1);
        lastMonthDate.setMonth(lastMonthDate.getMonth()-1);
        var lastMonth = lastMonthDate.getMonth();

        function getMonthDays(myMonth){
            var monthStartDate = new Date(nowYear, myMonth, 1);
            var monthEndDate = new Date(nowYear, myMonth + 1, 1);
            var   days   =   (monthEndDate   -   monthStartDate)/(1000   *   60   *   60   *   24);
            return   days;
        }
        var preMonthEnd = new Date(nowYear, lastMonth, getMonthDays(lastMonth));
        return preMonthEnd;
    }
    // 给Date对象添加getMonthBegin方法，获取本月开始时间
    Date.prototype.getMonthBegin = function(){
        var nowYear = this.getYear();             //当前年
        var nowMonth = this.getMonth();           //当前月
        nowYear += (nowYear < 2000) ? 1900 : 0;
        var monthBegin = new Date(nowYear, nowMonth, 1);
        return monthBegin;
    }
    // 给Date对象添加getMonthEnd方法，获取本月结束时间
    Date.prototype.getMonthEnd = function(){
        var nowYear = this.getYear();             //当前年
        var nowMonth = this.getMonth();           //当前月
        nowYear += (nowYear < 2000) ? 1900 : 0;
        function getMonthDays(myMonth){
            var monthStartDate = new Date(nowYear, myMonth, 1);
            var monthEndDate = new Date(nowYear, myMonth + 1, 1);
            var   days   =   (monthEndDate   -   monthStartDate)/(1000   *   60   *   60   *   24);
            return   days;
        }
        var monthEnd = new Date(nowYear, nowMonth, getMonthDays(nowMonth));
        return monthEnd;
    }
    // 给Date对象添加getNextMonthBegin方法，获取下一月开始时间
    Date.prototype.getNextMonthBegin = function(){
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        var nextMonthDate = new Date(this.getTime());  //上月日期
        nextMonthDate.setDate(1);
        nextMonthDate.setMonth(nextMonthDate.getMonth()+1);
        var nextMonth = nextMonthDate.getMonth();
        var nextMonthBegin = new Date(nowYear, nextMonth, 1);
        return nextMonthBegin;
    }
    // 给Date对象添加getNextMonthEnd方法，获取下一月结束时间
    Date.prototype.getNextMonthEnd = function(){
        var nowYear = this.getYear();             //当前年
        nowYear += (nowYear < 2000) ? 1900 : 0;
        var nextMonthDate = new Date(this.getTime());  //上月日期
        nextMonthDate.setDate(1);
        nextMonthDate.setMonth(nextMonthDate.getMonth()+1);
        var nextMonth = nextMonthDate.getMonth();
        function getMonthDays(myMonth){
            var monthStartDate = new Date(nowYear, myMonth, 1);
            var monthEndDate = new Date(nowYear, myMonth + 1, 1);
            var   days   =   (monthEndDate   -   monthStartDate)/(1000   *   60   *   60   *   24);
            return   days;
        }
        var nextMonthEnd = new Date(nowYear, nextMonth, getMonthDays(nextMonth));
        return nextMonthEnd;
    }



// 给String对象添加getDate方法，使字符串形式的日期返回为Date型的日期
    String.prototype.getDate = function(){
        var strArr = this.split('-');
        var date = new Date(strArr[0], strArr[1] - 1, strArr[2]);
        return date;
    }
    // 给String对象添加getDate方法，使字符串形式的日期返回为Date型的日期
    String.prototype.getDatefromPlanStr = function(){
       // var strArr = this.split('-');
        var year=Number(this.substr(0,4));
        var month=Number(this.substr(4,2));
        var day=Number(this.substr(6,2));
        var date = new Date(year, month - 1, day);
        return date;
    }
}]);

app.service('ArrayService',['$window','$location',function($window,$location){
// 获取间隔天数
   this.ishasvalue=function(value,list){
       var flag=list.indexOf(value);
       if(flag==-1){
           return false;
       }
       return true;

   }

}]);



app.service('ComService',['$rootScope',function($rootScope){
    var cl=function(a){
        console.log(a);
    }
    var record=function(result,itemname,oldvalue,newvalue){
        result.push(itemname+'|旧值:'+oldvalue+"|新值:"+newvalue);
    }
    function isZArray(obj) {
        return Object.prototype.toString.call(obj) === '[object Array]';
    }
    function isZObject(obj){
        return (typeof obj)=='object'
    }
    function isZEqualsForArray(obj1,obj2){
        if(isZArray(obj1)&&isZArray(obj2)){
            if(obj1.length==obj2.length){
                var samearr=[];
                for(var s in obj1){
                    for(var x in obj2){
                        if(obj1[s]===obj2[x]){
                            samearr.push(obj1[s]);
                        }
                    }
                }
                if(samearr.length==obj1.length){
                    return true;
                }
            }
        }
        return false;
    }
    var intermap={
        "id":"编号",
        "obj.kitty.name":"obj的kittyname属性",
        "obj.kitty.cp":"obj的kittycp属性",
        "obj.long":"obj的long属性",
        "arr":"数组arr"
    }
    //原版
    // a oldvalue b newvalue
    this.diff=function(a,b,father){
        var result=[];
        for(var item in a){
            if(isZObject(a[item])){
                if(isZArray(a[item])){
                    if(!isZEqualsForArray(a[item],b[item])){
                        if(father){
                            record(result,father+'.'+item,a[item],b[item]);
                        }else{
                            record(result,item,a[item],b[item]);
                        }
                    }
                }else{
                    if(father){
                        diff(a[item],b[item],father+'.'+item);
                    }else{
                        diff(a[item],b[item],item);
                    }
                }
            }else{
                if(a[item]!=b[item]){//此处用双等  宽泛的认为2 和'2' 一样
                    if(father){
                        record(result,father+'.'+item,a[item],b[item]);
                    }else{
                        record(result,item,a[item],b[item]);
                    }
                }
            }
        }
        //此处ajax 入库
        cl(result);

    }
    //国际化版
    this.diff2=function(a,b,intermap,father){
        var result=[];
        for(var item in a){
            if(isZObject(a[item])){
                if(isZArray(a[item])){
                    if(!isZEqualsForArray(a[item],b[item])){
                        if(father){
                            record(result,intermap[father+'.'+item],a[item],b[item]);
                        }else{
                            record(result,intermap[item],a[item],b[item]);
                        }
                    }
                }else{
                    if(father){
                        diff2(a[item],b[item],father+'.'+item,intermap);
                    }else{
                        diff2(a[item],b[item],item,intermap);
                    }
                }
            }else{
                if(a[item]!=b[item]){//此处用双等  宽泛的认为2 和'2' 一样
                    if(father){
                        record(result,intermap[father+'.'+item],a[item],b[item]);
                    }else{
                        record(result,intermap[item],a[item],b[item]);
                    }
                }
            }
        }
  /*      //此处ajax 入库
        var url = BaseUrl + '/syslog/addLog'
        var  postCfg = {
            headers: {'Content-Type': 'application/json; charset=UTF-8'}
         }
        var input = {
            "tableId":intermap.tableId,
            "bizId":intermap.bizId,
            "bizName":intermap.bizName,
            "bizType":intermap.bizType,
            "content":result
        }
        $http.post(url, input, postCfg)
            .success(function (data) {
                if (data["code"] == 1) {
                    console.log('success log');
                } else {
                    console.log('errorcode:' + data["errorCode"] + 'errorMessage:' + data["errorMessage"]);
                }
            }).error(function (data) {
                console.log('request failed!');
            });*/
        $rootScope.logList.push({"result":result,"intermap":intermap});
        cl(result);
    }

}]);

app.factory('MyHTTP',['$location','$rootScope','$interval','$http','$route',
  function($location,$rootScope,$interval,$http,$route){
  return {
    http: function (req,path) {
      $http(req).then(function (response) {
        var data = response.data;
        if(data["code"]==0){
          if (path == "yes") {
            layer.msg(data["message"], {icon: 1,end:function(){
              $route.reload();
            }});
            return;
          } else if (path != "") {
            layer.msg(data["message"], {icon: 1});
            $location.path(path);
            return;
          } else {
            layer.msg(data["message"], {icon: 1});
          }
          if(data["totalRecord"]>0){
            console.log(data);
            return data;
          }
        } else if (data["code"] == 1199) {
          //console.log("asdfasdf");
          $rootScope.referer = $location.path();
          $location.path('/login');
        } else {
          layer.msg(data["message"], {icon: 2});
        }
      },function(data){
        finish = 1;
        layer.msg('程序内部错误!', {icon: 5});
        console.log('request failed: '+ req.url);
      }).then(function () {return});
    },
  };
}]);
