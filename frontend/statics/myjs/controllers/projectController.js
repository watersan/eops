
app.controller('ProjectController',['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','$interval','SessionService','$filter','MyHTTP','permissions',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,$interval,SessionService,$filter,MyHTTP,permissions) {
    var connections = [];
    var r;
    $scope.showsubbox = "hidden";
    $scope.item = {};
    $scope.appstemplates = {};
    $scope.tmp = {};
    $scope.readonly = false;
    $scope.scriptcode = "";
    $scope.dissave = false;
    $scope.page = {};
    $scope.page.totalItems = 0;
    $scope.page.currentPage = 1;
    $scope.page.maxSize = 5;
    $scope.page.itemFilter = '';
    $scope.page.numPerPage = "10";
    $scope.page.sortType     = 'hostname';
    $scope.page.sortReverse  = false;
    $scope.selectitem = {};
    $scope.selectAll = false;
    $scope.allocresource = {};
    $scope.fields = [];
    $scope.users = [];
    $scope.showme = {curr:"business",history:"business"};
    $scope.currentitem = {};
    $scope.title = {curr:"",history:""};
    $scope.treedata = [];
    $scope.projects = {};
    $scope.clusters = {};
    $scope.projectitems = {};
    $scope.clusteritems = {};
    $scope.appitems = {};
    $scope.resourceshownoalloc = false;
    $scope.allocedresource = {};
    $scope.topology = {width: 800,height:1200};
    $scope.sourcetype = "HTTP";
    var mycode;
    //console.log();
    /*
    function getAbsoluteLeft(objectId) {
      o = document.getElementById(objectId);
      oLeft = o.offsetLeft;
      console.log(oLeft);
      while(o.offsetParent!=null) {
        oParent = o.offsetParent;
        console.log(oParent.offsetLeft);
        oLeft += oParent.offsetLeft;
        o = oParent;
      }
      return oLeft;
      }
    //获取控件上绝对位置
    function getAbsoluteTop(objectId) {
      o = document.getElementById(objectId);
      oTop = o.offsetTop;
      while(o.offsetParent!=null) {
        oParent = o.offsetParent;
        oTop += oParent.offsetTop;  // Add parent top position
        o = oParent;
      }
      return oTop;
    }
    */
    $scope.paginate = function (value,index,carr) {
      var begin, end, index;
      $scope.page.totalItems = carr.length;
      numpp = parseInt($scope.page.numPerPage);
      begin = ($scope.page.currentPage - 1) * numpp;
      end = begin + numpp;
      //index = $scope.datalist.indexOf(value);
      return (begin <= index && index < end);
    };
    $scope.doSelectAll = function() {
      var filtered = $filter('filter')($scope.datalist, $scope.page.itemFilter);
      sorted = $filter('orderBy')(filtered,$scope.page.sortType,$scope.page.sortReverse);
      numpp = parseInt($scope.page.numPerPage);
      begin = ($scope.page.currentPage - 1) * numpp;
      end = begin + numpp;
      if ($scope.selectAll == true) {
        $scope.selectAll = false;
      } else {
        $scope.selectAll = true;
      }
      var i = 0;
      angular.forEach(sorted, function(item) {
         if (begin <= i && i < end) {
           $scope.allocresource[$scope.appsid]["environment"][$rootScope.curenv][item._id] = $scope.selectAll;
           //item.selected = false;
         }
         i++;
      });
    };

    $scope.doSelect = function(item) {
      //console.log(item);
      id = item._id;
      if ($scope.allocresource[$scope.appsid][item._id] == undefined || $scope.allocresource[$scope.appsid][item._id] == false) {
        //console.log(index)
        $scope.allocresource[$scope.appsid][item._id] = true;
      } else {
        $scope.allocresource[$scope.appsid][item._id] = false;
      }
      //console.log("aa"+$scope.datalist[index].selected)
    };

    var textdialog = function(event, data, aaa) {
      //console.log(event);
      //o = document.getElementById('holder');
      x1 = event.target.x.baseVal.value + 10;
      y1 = event.target.y.baseVal.value + 100;
      //console.log(y1);
      if ($scope.showsubbox == "visible") {
        $scope.showsubbox = "hidden";
      } else {
        $scope.showsubbox = "visible";
      }
      var ss = {
        'backgroundColor': '#ff0',
        'left': x1+'px',
        'top': y1+'px',
        "visibility": $scope.showsubbox,
        //'left': x1+"px",
        //'top': y1+"px",
        "z-index": "10",
      };
	    $('#subbox')
        .stop()
        .css(ss)
        .animate({backgroundColor: '#ddd'}, 1000);
    };

    Raphael.fn.connection = function (obj1, obj2, line, bg) {
      if (obj1.line && obj1.from && obj1.to) {
        line = obj1;
        obj1 = line.from;
        obj2 = line.to;
      }
      //obj2为拖动的对象
      var bb1 = obj1.getBBox(),
        bb2 = obj2.getBBox(),
        p = [{x: bb1.x + bb1.width / 2, y: bb1.y - 1},          //上0
        {x: bb1.x + bb1.width / 2, y: bb1.y + bb1.height + 1},  //下1
        {x: bb1.x - 1, y: bb1.y + bb1.height / 2},              //左2
        {x: bb1.x + bb1.width + 1, y: bb1.y + bb1.height / 2},  //右3
        {x: bb2.x + bb2.width / 2, y: bb2.y - 1},               //上4
        {x: bb2.x + bb2.width / 2, y: bb2.y + bb2.height + 1},  //下5
        {x: bb2.x - 1, y: bb2.y + bb2.height / 2},              //左6
        {x: bb2.x + bb2.width + 1, y: bb2.y + bb2.height / 2}], //右7
        d = {}, dis = [];
      var b1c={x: bb1.x + bb1.width / 2, y: bb1.y + bb1.height / 2},
        b2c={x: bb2.x + bb2.width / 2, y: bb2.y + bb2.height / 2};
      /*
      //最短路径算法
      for (var i = 0; i < 4; i++) {
        for (var j = 4; j < 8; j++) {
          var dx = Math.abs(p[i].x - p[j].x),
          dy = Math.abs(p[i].y - p[j].y);
          if ((i == j - 4) || (((i != 3 && j != 6) || p[i].x < p[j].x) && ((i != 2 && j != 7) ||
            p[i].x > p[j].x) && ((i != 0 && j != 5) || p[i].y > p[j].y) && ((i != 1 && j != 4) ||
            p[i].y < p[j].y))) {
            dis.push(dx + dy);
            d[dis[dis.length - 1]] = [i, j];
          }
        }
      }
      if (dis.length == 0) {
        res = [0, 4];
      } else {
        res = d[Math.min.apply(Math, dis)];
      }
      */
      //自己写的路径算法。
      var res;
      var tan25 = Math.tan(25*Math.PI/180);
      var tan65 = Math.tan(65*Math.PI/180);
      if (p[2].x > p[7].x) {   //左侧
        if (tan25 > Math.abs(b1c.y - b2c.y)/(b1c.x - b2c.x) ||
          Math.abs(p[2].y - p[7].y) - Math.abs(bb1.height - bb2.height) / 2 < 30) {
          res = [2,7];
        } else if (p[2].y > p[7].y) {
          res = [0,5];
        } else {
          res = [1,4];
        }
      } else if (p[4].y > p[1].y) {
        if (tan65 >= Math.abs(b1c.x - b2c.x)/(b2c.y - b1c.y) ||
          Math.abs(p[1].x - p[4].x) - Math.abs(bb1.width - bb2.width) / 2 < 30) {
          res = [1,4];
        } else if (p[1].x > p[4].x) {
          res = [2,7];
        } else {
          res = [3,6];
        }
      } else if (p[6].x > p[3].x) {
        if (tan25 > Math.abs(b1c.y - b2c.y)/(b2c.x - b1c.x) ||
          Math.abs(p[3].y - p[6].y) - Math.abs(bb1.height - bb2.height) / 2 < 30) {
          res = [3,6];
        } else if (p[3].y > p[6].y){
          res = [0,5];
        } else {
          res = [1,4];
        }
      } else if (p[0].y > p[5].y) {
        if (tan65 >= Math.abs(b1c.x - b2c.x)/(b1c.y - b2c.y) ||
          Math.abs(p[0].x - p[5].x) - Math.abs(bb1.width - bb2.width) / 2 < 30) {
          res = [0,5];
        } else if (p[0].x > p[5].x) {
          res = [2,7];
        } else {
          res = [3,6];
        }
      } else {
        res = [0, 4];
      }

      var x1 = p[res[0]].x,
        y1 = p[res[0]].y,
        x4 = p[res[1]].x,
        y4 = p[res[1]].y;
      dx = Math.max(Math.abs(x1 - x4) / 2, 10);
      dy = Math.max(Math.abs(y1 - y4) / 2, 10);
      var x2 = [x1, x1, x1 - dx, x1 + dx][res[0]].toFixed(3),
        y2 = [y1 - dy, y1 + dy, y1, y1][res[0]].toFixed(3),
        x3 = [0, 0, 0, 0, x4, x4, x4 - dx, x4 + dx][res[1]].toFixed(3),
        y3 = [0, 0, 0, 0, y1 + dy, y1 - dy, y4, y4][res[1]].toFixed(3);
      var path = ["M", x1.toFixed(3), y1.toFixed(3), "C", x2, y2, x3, y3, x4.toFixed(3), y4.toFixed(3)].join(",");
      if (line && line.line) {
        line.bg && line.bg.attr({path: path});
        line.line.attr({path: path});
      } else {
        var color = typeof line == "string" ? line : "#000";
        return {
          bg: bg && bg.split && this.path(path).attr({
            stroke: bg.split("|")[0],
            fill: "none",
            "stroke-width": bg.split("|")[1] || 3
          }),
          line: this.path(path).attr({stroke: color, fill: "none"}),
          from: obj1,
          to: obj2
        };
      }
    };
    //给矩形增加居中的文字
    function insertRectText(root,rectangle,str){
      var x = Math.round($(rectangle.node).attr("x"));
      var y = Math.round($(rectangle.node).attr("y"));
      var w = $(rectangle.node).attr("width");
      var h = $(rectangle.node).attr("height");
      var textStr = root.text(x + w / 2,y + h / 2,str).attr({fill:"#444444","font-size":"15px"});
      // textStr.attr({
      //    "fill":"#444444",
      //    "font-size":"15px",
      // });
      rectangle.data("cooperative1", textStr);
    };
    var dragger = function () {
      this.ox = this.type == "ellipse" ? this.attr("cx") : this.attr("x");
      this.oy = this.type == "ellipse" ? this.attr("cy") : this.attr("y");
      this.animate({fill: "#4A9BFF"}, 500);
    };
    var move = function (dx, dy) {
      var att;
      if (this.type == "ellipse") {
        att = {cx: this.ox + dx, cy: this.oy + dy}
      } else if (this.type == "rect") {
        att = {x: this.ox + dx, y: this.oy + dy};
      } else if (this.type == "text") {
        att = {x: this.ox + dx, y: this.oy + dy};
      }
      if (att.x <=5 || att.x >=$scope.topology.width) {
        att.x = this.ox;
      }
      if (att.y <=5 || att.y >= $scope.topology.height) {
        att.y = this.oy;
      }
      this.attr(att);
      var ttt = this.data("cooperative1");
      if (ttt) {
        var w = $(this.node).attr("width");
        var h = $(this.node).attr("height");
        att.x = att.x + w / 2;
        att.y = att.y + h / 2;
        ttt.attr(att);
      }
      for (var i = connections.length; i--;) {
        r.connection(connections[i]);
      }
      //r.safari();
    };
    var up = function () {
      this.animate({fill: '#4A9BFF'}, 500);
    };
    var restorecb = function(el, data) {
      el.drag(move, dragger, up);
      return el;
    }
    //$scope.title = "";
    // var lpatharr = $location.path().lpath.split('/');
    // if (lpatharr.length == 3) {
    //   $scope.changeview(lpatharr[3],"");
    // }

    $scope.changeview = function(mode, title) {
      $scope.showme.history = $scope.showme.curr;
      $scope.showme.curr = mode;
      if (title != "") {
        $scope.title.history = $scope.title.curr;
        $scope.title.curr = title;
      }
      if (mode == "appadd" || mode == "deploy") {
        //$scope.item = {};
        req = {
          "method": 'GET',
          "url": BaseUrl+'/cmdb/apptpl',
          "params": {"pageNo":1,"pageSize":1000}, //,"smalllist":"yes"
        }
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.appstemplates = data.dataList;
            if ($scope.isappadd == false) {
              $scope.itemtpl = $scope.tplname($scope.item.from,false);
              if (!($scope.item.sourcetype)) {
                $scope.item.sourcetype = $scope.itemtpl.sourcetype;
              }
            }
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {icon: 2});
          }
        });
      }

      //Bug[001]: 需要判断是否有post和put权限。
      if (mode == "project" || mode == "cluster" ) {
        req = {
          method: 'GET',
          url: BaseUrl+'/auth/user',
          params: {"pageNo":1,"pageSize":1000,"smalllist":"yes"},
        }
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.users = data.dataList
            // for (var id =0 ;id < data.totalRecord; id++) {
            //   item = data.dataList[id]
            //   $scope.users.push(item.fullname);
            // }
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {icon: 2});
          }
        });
        $( "#form_opser" ).autocomplete({
          minLength: 1, //输入1个符触发搜索,
          delay: 0,
          source: function( request, response ) {
              var matcher = new RegExp(request.term, "i");
              var results =$.grep($scope.users, function(val){
                return matcher.test(val.fullname);
              });
              response(results.slice(0, 20));
          },
          "select": function(event, ui){
            this.value = ui.item.fullname;
            $scope.item.opser = ui.item.fullname;
            return false;
          },
          //在bootstrap的modal下，此项是必须的。
          appendTo: ".eventInsForm",
        });
        $( "#form_secondopser" ).autocomplete({
          minLength: 1, //输入1个符触发搜索,
          delay: 0,
          source: function( request, response ) {
              var matcher = new RegExp(request.term, "i");
              var results =$.grep($scope.users, function(val){
                return matcher.test(val.fullname);
              });
              response(results.slice(0, 20));
          },
          "select": function(event, ui){
            this.value = ui.item.fullname;
            $scope.item.secondopser = ui.item.fullname;
            return false;
          },
          //在bootstrap的modal下，此项是必须的。
          appendTo: ".eventInsForm",
        });

      }
      if (mode == "business") {
        //画拓扑图。首先获取层级关系
        var shapes = {};
        var depend = {};
        if ($scope.currentitem.hasOwnProperty("nodes") == false)
          return;
        for (var i =0; i < $scope.currentitem.nodes.length; i++) {
          cluster = $scope.currentitem.nodes[i];
          if (cluster.oid) {
            //console.log(cluster);
            shapes[cluster.oid] = {
              name: cluster.text,
              depend: cluster.depend,
            };
            for (var j =0; j < cluster.depend.length; j++) {
              dp = cluster.depend[j];
              depend[dp] = 1;

            }
          }
        }
        for (dp in depend) {
          if (!shapes.hasOwnProperty(dp)) {
            shapes[dp] = {
              name: $scope.projectitems[$scope.clusteritems[dp].project].name + ": " + $scope.clusteritems[dp].name,
              depend: [],
            }
          }
        }
        topoid = "";
        for (oid in shapes) {
          //console.log(oid);
          if (!depend.hasOwnProperty(oid)) {
            topoid = oid;
            break;
          }
        }
        if (topoid == "")
          return;
        var x=250,y=50;
        // console.log(topoid);
        // console.log(shapes[topoid]);
        var l = strlen(shapes[topoid].name);
        var w = l*10+20;
        //r.clear();
        if (typeof(r) == "object") {
          r.remove();
        }
        stop = $interval(function() {
          if (document.getElementById("topology") != null) {
            $scope.stopWait();
            //console.log($scope.projectitems[$scope.projectid].topology);
            r = Raphael("topology", $scope.topology.width, $scope.topology.height);
            if ($scope.projectitems[$scope.projectid].topology != undefined && $scope.projectitems[$scope.projectid].topology != "") {
              r.fromJSON($scope.projectitems[$scope.projectid].topology);
              return;
            }
            gdata = r.rect(x, y, w, 40, 10);
            gdata.attr({
              "font-size": 25,
              fill: '#4A9BFF',
              stroke: '#4A9BFF',
              "stroke-width": 1,
              cursor: "move", r: 10
            });
            gdata.drag(move, dragger, up);
            insertRectText(r,gdata,shapes[topoid].name);
            //console.log(gdata.data("cooperative1"));
            shapes[topoid].gdata = gdata;
            //shapes[topoid].gdata.dblclick(textdialog);
            drawtopology(topoid,x+w/2,y+100);
            function drawtopology(oid, x, y) {
              var depcount = 0;
              var alen = 0;
              for (var i=0; i<shapes[oid].depend.length;i++) {
                dp = shapes[oid].depend[i];
                if (!shapes[dp].hasOwnProperty("gdata")) {
                  alen = alen + strlen(shapes[dp].name)*10 + 20;
                  depcount++;
                }
                //console.log(dp);
              }
              alen = alen + ((depcount - 1) * 60);
              //console.log(shapes[oid].name + ":alen:" + alen);
              var bx = 10;
              if (alen / 2 < x - 10) {
                bx = x - (alen / 2);
              }
              nbx = bx;
              for (var i=0; i<shapes[oid].depend.length;i++) {
                dp = shapes[oid].depend[i];
                if (!shapes[dp].hasOwnProperty("gdata")) {
                  l = strlen(shapes[dp].name);
                  //console.log(shapes[dp].name+":bx:"+bx+":"+x);
                  gdata = r.rect(bx, y, l*10+20, 40, 10);
                  gdata.attr({
                    "font-size": 25,
                    fill: '#4A9BFF',
                    stroke: '#4A9BFF',
                    "stroke-width": 1,
                    cursor: "move", r: 10
                  });
                  gdata.drag(move, dragger, up);
                  insertRectText(r,gdata,shapes[dp].name);
                  shapes[dp].gdata = gdata;
                  bx = bx + l*10+20 + 60;
                }
                connections.push(r.connection(shapes[oid].gdata, shapes[dp].gdata, "#46C37B", "#46C37B|3"));
                //drawtopology(dp,bx,y+100);
              }
              for (var i=0; i<shapes[oid].depend.length;i++) {
                dp = shapes[oid].depend[i];
                l = strlen(shapes[oid].name);
                drawtopology(dp,nbx+(l*10+20)/2,y+100);
              }
            }
          }
        },100);
      } else if (mode == "project") {
        if ($scope.projectid != "") {
          $scope.item = $scope.projectitems[$scope.projectid];
        }
      } else if (mode == "cluster") {
        if ($scope.clusterid != "") {
          $scope.item = $scope.clusteritems[$scope.clusterid];
        }
        results = [];
        for (i = 0,l=$scope.treedata.length; i < l; i++) {
          p = $scope.treedata[i];
          if (p.text != "add" && p.hasOwnProperty("nodes") && p.nodes.length > 0) {
            children = [];
            for (k = 0,kl=p.nodes.length; k < kl; k++) {
              c = p.nodes[k];
              if (c.nodes && c.nodes.length > 0 && c.shared == true)
                children.push({id: c.id,text: c.text,selected: false});
            }
            if (children.length > 0) {
              results.push({text: p.text, children: children});
            }
          }
        }
        stop = $interval(function() {
          if (document.getElementById("form_depend") != null) {
            $scope.stopWait();
            $.fn.select2.amd.require([
              'select2/data/array',
              'select2/utils'
            ], function (ArrayData, Utils) {
              function CustomData ($element, options) {
                CustomData.__super__.constructor.call(this, $element, options);
              }
              function formatState(state) {
                //'<span><i class="fa fa-minus-square-o"></i>'+p.text+"</span>"
                //console.log(state);
                if (!state.id) {
                  return $('<span><i class="fa fa-minus-square-o"></i>  '+state.text+'</span>');
                }
                return $('<span style="text-indent:2em;"><i class="fa fa-asterisk"></i>  '+state.text+'</span>');
              }

              Utils.Extend(CustomData, ArrayData);

              CustomData.prototype.current = function (callback) {
                var data = [];
                if (!$scope.item.depend)
                  $scope.item.depend = []
                for (i=0,l=$scope.item.depend.length;i<l;i++) {
                  oid = $scope.item.depend[i];
                  data.push({
                    id: oid,
                    text: $scope.clusteritems[oid].name,
                    selected: true,
                  });
                }
                callback(data);
              };

              $("#form_depend").select2({
                data: results,
                templateResult: formatState,
                theme: "classic",
                tags: "true",
                //debug: true,
                dataAdapter: CustomData
              });
            });
          }
        },100);
      } else if (mode == "appconfadd") {
        //$scope.item = {};
        $scope.additionconfs = [];
        var initstr = "";
        //$scope.namereadonly = false;
        if (title.indexOf("变量") > -1) {
          $scope.item.name = "Veriables";
          initstr = "# Key = value; key和value不能有空格\n";
          $scope.readonly = true;
        } else {
          req = {
            "method": 'GET',
            "url": BaseUrl+'/cmdb/apptpl',
            'params': {"_id":$scope.appitems[$scope.appsid].from,"additionconf":"yes"},
          }
          $http(req).then(function (response) {
            var data = response.data;
            if(data.code==0){
              if (data.totalRecord > 0) {
                $scope.additionconfs = data.dataList[0].additionconf;
                //$scope.readonly = false;
              } else {
                layer.msg('此应用不允许增加配置文件！', {icon: 1,time:1000,end:function(){
                  $scope.changeview("appconf",title.history);}});
                //$scope.readonly = true;
              }
              //console.log($scope.additionconfs);
            } else if (data.code == 1199) {
              $rootScope.loginform();
            } else {
              layer.msg(data.message, {icon: 2});
            }
          });
        }
        stop = $interval(function() {
          if (document.getElementById("scriptcode") != null) {
            $scope.stopWait();
            mycode = CodeMirror.fromTextArea(document.getElementById("scriptcode"), {
              lineNumbers: true,
              theme: "the-matrix",
              scrollbarStyle: "overlay",
              readOnly: false,
              mode: "textile"
            });
            if ($scope.scriptcode != "") {
              initstr = $scope.scriptcode;
            }
            mycode.setValue(initstr);
          };
        },100,5);
        //console.log($scope.appitems[$scope.appsid]);
        //mycode.setValue(utf8Decode(base64Decode(data.dataList["content"])));
        //scriptcode: base64Encode(mycode.getValue())
      } else if (mode == "appconf") {
        req = {
          method: 'GET',
          url: BaseUrl+'/cmdb/appconf',
          params: {"pageNo":1,"pageSize":1000,"appid":$scope.appsid},
        }
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.datalist = data.dataList
            //console.log($scope.additionconfs);
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            $scope.datalist = [];
            layer.msg(data.message, {icon: 2});
          }
        });
      } else if (mode == "hostlist") {
        url = BaseUrl+'/cmdb/resource';
        req = {
          method: 'GET',
          url: url,
          params: {"pageNo":1,"pageSize":1000,environment: $rootScope.curenv},
        }
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            $scope.datalist=data.dataList;
            $scope.totalItems=data.totalRecord;
            $scope.tablestatus="1";
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            layer.msg(data.message, {icon: 2});
          }
        },function(data){
          layer.msg('无法获取信息!', {icon: 5});
        });
      } else if (mode == "deploy") {
        $scope.allowupload = true;
        req = {
          method: 'GET',
          url: BaseUrl+'/deploy/history',
          params: {"pageNo":1,"pageSize":1000,"appid": $scope.appsid,"history":"yes"},
        }
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0 ){
            $scope.totalItems=data.totalRecord;
            if (data.totalRecord > 0) {
              $scope.deployinfo=data.dataList;
            }
          } else if (data.code == 1199) {
            $rootScope.loginform();
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
        var from = $scope.appitems[$scope.appsid].from;
        var itemtpl = $scope.tplname(from,false);
        if (typeof itemtpl == "object") {
          if (itemtpl.build.length > 0) {
            $scope.allowupload = false;
          }
        }
        stop = $interval(function() {
          if (document.getElementById("fileupload") != null) {
            $scope.stopWait();
            cid = $scope.appitems[$scope.appsid].cluster;
            pid = $scope.clusteritems[cid].project;
            // $('#progress .progress-bar').css(
            //     'display',
            //     'none'
            // );
            //$('#fileupload').addClass('fileupload-processing');
            $('#fileupload').bind('fileuploadprogress', function (e, data) {
              var progress = parseInt(data.loaded / data.total * 100, 10);
              $('#progress .progress-bar').attr("aria-valuenow",progress).css(
                  'width',
                  progress + '%'
              ).text(progress + '%');
            });
            $('#fileupload').fileupload({
              url: BaseUrl + '/deploy/uploadapp?p='+$scope.projectitems[pid].name+
                '&c=' + $scope.clusteritems[cid].name + '&a=' + $scope.appitems[$scope.appsid].name,
              maxChunkSize: 10485760,
              uploadTemplateId: 'template-upload',
              add: function (e, data) {
                data.submit();
              },
              done: function(e,data) {
                // console.log(e);
                // console.log(data);
                $('#progress .progress-bar').css(
                    'display',
                    'none'
                );
                layer.msg("上传成功！", {icon: 1});
              },
            });
          }
        },100);
        //var overallProgress = $('#fileupload').fileupload('progress');

      }
    }

    function strlen(str) {
      var i,sum;
      sum=0;
      //console.log(str);
      for(i=0;i<str.length;i++) {
        if (str.charCodeAt(i)<=255)
          sum=sum+1;
        else
          sum=sum+2;
      }
      return sum;
    };
    $scope.changed = function() {
      //dd = $scope.item.dd;
      //$scope.item = {};
      if ($scope.page.apptpl == null) {
        return;
      }
      $scope.itemtpl = $scope.page.apptpl;
      $scope.item.sourcetype = $scope.page.apptpl.sourcetype;
      $scope.item.from = $scope.page.apptpl["_id"];
      $scope.item.cluster = $scope.clusterid;
      delete $scope.item._id;
      //console.log($scope.item);
    };
    $scope.appconfop = function(item,op) {
      req = {
        method: 'GET',
        url: BaseUrl+'/cmdb/appconf',
      }

      if (op == "var") {
        //$scope.scriptcode = "# Key = value; key和value不能有空格\n";
        //$scope.readonly = true;
        req.url = BaseUrl+'/cmdb/appver';
        req.params = {"id": $scope.appsid};
        $scope.fields = [];
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            if (typeof data.dataList === 'object') {
              var variables = data.dataList;
              for (k in variables) {
                $scope.fields.push({name: k,value:variables[k],status:0});
              }
            }
            //console.log($scope.additionconfs);
          } else if (data.code == 1199) {
            $rootScope.loginform();
          } else {
            $scope.scriptcode = "# Key = value; key和value不能有空格\n";
          }
        }).then(function () {
          $scope.changeview('appvar','添加变量: '+$scope.title.source);
        });

        return;
      } else if (op == "add") {
        $scope.item = {};
        $scope.changeview('appconfadd','添加配置: '+$scope.title.source);
        return;
      } else if (op == "view") {
        $scope.readonly = true;
        $scope.dissave = true;
      } else if (op == "update") {
        $scope.dissave = false;
      }
      $scope.item = item;
      req.params = {"id": item._id};
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          tmp = data.dataList[0];
          $scope.scriptcode = base64Decode(tmp.content);
          //console.log(tmp);
          delete tmp.content;
          $scope.item = tmp;
          //console.log($scope.additionconfs);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      }).then(function () {
        $scope.changeview('appconfadd',item.name);
      });
    };
    $scope.projectop = function(data) {
      $scope.type = data.type;
      if ($scope.type == "project") {
        if (data.nameid == "add") {
          $scope.item = {};
          $scope.projectid = "";
          //$scope.doshowwhat("projectattr");
        } else {
          $scope.projectid = data.oid;
          $scope.currentitem = data;
          $scope.type = "business";
        }
      } else if ($scope.type == "cluster") {
        $scope.clusterid = data.oid;
        if (data.nameid == "add") {
          $scope.clusterid = "";
          $scope.item = {};
          $scope.item.project = data.project;
        }
      } else if ($scope.type == "apps") {
        $scope.appsid = data.oid;
        $scope.clusterid = data.cluster;
        if (data.nameid == "add") {
          $scope.type = "appadd";
          $scope.item = {};
          $scope.item.sourcetype = "GIT";
          $scope.page.apptpl = {name:"选择模板",_id:"",};
          $scope.itemtpl = {};
          $scope.isappadd = true;
        } else {
          $scope.type = "appadd";
          $scope.isappadd = false;
          $scope.item = {};
          //$scope.item = $scope.appitems[data.oid];
          $scope.item._id = data.oid;
          $scope.item.name = $scope.appitems[data.oid].name;
          $scope.item.from = $scope.appitems[data.oid].from;
          $scope.item.config = $scope.appitems[data.oid].config;
          if ($scope.appitems[data.oid].sourcetype && $scope.appitems[data.oid].sourcetype != "") {
            $scope.item.sourcetype = $scope.appitems[data.oid].sourcetype;
            $scope.item.source = $scope.appitems[data.oid].source;
          }
        }
      }
      $scope.title.source = data.text;
      $scope.changeview($scope.type, data.text);
      $scope.$apply();
    }
    $scope.tplname = function(id,getname) {
      for (var i=0;i<$scope.appstemplates.length;i++) {
        if ($scope.appstemplates[i]._id == id) {
          if (getname) {
            return $scope.appstemplates[i].name;
          }
          return $scope.appstemplates[i]
        }
      }
      return "";
    }
    $scope.deleteItem = function (id, kind) {
      var chkid;
      chkid = $scope.displayDelBtn(kind)
      if (id == "") {
        chkid = $scope.displayDelBtn(kind)
        if (chkid == "") {
          layer.confirm('请先删除此项的依赖！', {
            title:"无法删除",
            btn: ['确定'] //按钮
          });
        } else {
          id = chkid
        }
      }

      layer.confirm('您是要删除此项吗？', {
        title:"删除操作",
        btn: ['确定','取消'] //按钮
      }, function(){
        $scope.Delete(id, kind);
        $route.reload();
      }, function(){

      });
    };

    //删除操作通过service异步执行
    $scope.Delete = function(id, kind) {
      var req = {
        method: 'DELETE',
        url: BaseUrl+'/cmdb/' + kind,
        headers: {
          'Content-Type': "application/json; charset=UTF-8"
        },
        data: {"id":id}
      }
      MyHTTP.http(req,"yes");
    }
    $scope.displayDelBtn = function(kind) {
      switch (kind) {
        case 'project':
          if ($scope.projectid != "" && $scope.projects.hasOwnProperty($scope.projectid) &&
          $scope.projects[$scope.projectid].length > 0) {
            return $scope.projectid
          }
          break;
        case 'cluster':
          if ($scope.clusterid != "" && $scope.clusters.hasOwnProperty($scope.clusterid) &&
          $scope.clusters[$scope.clusterid].length > 0) {
            return $scope.clusterid
          }
          break;
        case 'apps':
          if ($scope.appsid != "" && $scope.appitems.hasOwnProperty($scope.appsid)) {
            if ($scope.appitems[$scope.appsid].depend.length == 0) {
              return ""
            } else {
              for (var dpid in $scope.appitems[$scope.appsid].depend) {
                if ($scope.appitems.hasOwnProperty(dpid)) {
                  return $scope.appsid
                }
              }
              return ""
            }
          }
          break;
        default:
          return ""
      }
      return ""
    }

    $scope.resourceFilter = function(value,index,carr) {
      if ($scope.resourceshow == "alloced") {
        var display = $scope.allocedresource[$scope.appsid][$rootScope.curenv][value.hostname];
        return display;
      } else if ($scope.resourceshow == "noalloced" || $scope.resourceshow == "addresource") {
        if ($scope.resourceshownoalloc) {
          return true;
        } else {
          return value.usage == 0 && $scope.allocedresource[$scope.appsid][$rootScope.curenv][value.hostname] != true;
        }
      }
    }
    $scope.allocResources = function(op) {
      t = "hostlist";
      if ($scope.appsid == undefined) {
        $scope.changeview("nohostlist","没有选择应用");
        return;
        //t = "nohostlist";
      }
      if ($scope.appitems.hasOwnProperty($scope.appsid) && $scope.appitems[$scope.appsid].depend == 1) {
        t = "denyalloc"
      }
      $scope.resourceshownoalloc = !$scope.resourceshownoalloc;
      if ($scope.allocedresource.hasOwnProperty($scope.appsid) == false) {
        $scope.allocedresource[$scope.appsid] = {};
        $scope.allocedresource[$scope.appsid][$rootScope.curenv] = {};
      }
      if ($scope.allocresource.hasOwnProperty($scope.appsid) == false ) {
        $scope.allocresource[$scope.appsid] = {};
        $scope.allocresource[$scope.appsid][$rootScope.curenv] = {};
        if ($rootScope.curenv in $scope.appitems[$scope.appsid].environment &&
        $scope.appitems[$scope.appsid].environment[$rootScope.curenv].hosts) {
          angular.forEach($scope.appitems[$scope.appsid].environment[$rootScope.curenv].hosts,
          function(item){
            $scope.allocresource[$scope.appsid][$rootScope.curenv][item] = true;
            $scope.allocedresource[$scope.appsid][$rootScope.curenv][item] = true;
          });
        }
      }
      var curenv = $scope.appitems[$scope.appsid].environment;
      if (!($rootScope.curenv in curenv)) {
        $scope.resourceshow = "noalloced";
        $scope.resourceop = "分配";
      } else if (curenv[$rootScope.curenv].hosts instanceof Array && curenv[$rootScope.curenv].hosts.length == 0) {
        $scope.resourceshow = "noalloced";
        $scope.resourceop = "分配";
      } else if (op == "addresource") {
        $scope.resourceshow = "addresource";
        $scope.resourceop = "分配";
      } else {
        $scope.resourceshow = "alloced";
        $scope.resourceop = "释放";
      }
      $scope.changeview(t, "资源分配-"+$scope.appitems[$scope.appsid].name);
    };
    //提交资源分配
    $scope.DoAlloc = function() {
      var pushlist = [];
      var pulllist = [];
      for (var k in $scope.allocresource[$scope.appsid][$rootScope.curenv]) {
        if ($scope.allocresource[$scope.appsid][$rootScope.curenv][k]) {
          if ($scope.allocedresource[$scope.appsid][$rootScope.curenv][k] != true) {
            pushlist.push(k);
          }
          $scope.allocedresource[$scope.appsid][$rootScope.curenv][k] = true;
          //console.log($scope.datalist[i].hostname);
        } else {
          if ($scope.allocedresource[$scope.appsid][$rootScope.curenv][k] == true) {
            pulllist.push(k);
          }
          $scope.allocedresource[$scope.appsid][$rootScope.curenv][k] = false;
        }
      }
      req = {
        "method": 'PUT',
        "url": BaseUrl+'/cmdb/allocresource',
        "headers": {'Content-Type': "application/json; charset=UTF-8"},
        "data": {"appid": $scope.appsid, "resourcelist": pushlist},
      }
      if ($scope.resourceshow == "addresource" && $scope.resourceop == "分配") {
        req.data.alloctype = "push";
      } else if ($scope.resourceop == "释放") {
        req.data.alloctype = "pull";
        req.data.resourcelist = pulllist;
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('操作成功！', {icon: 1, end:function(){$route.reload();}});
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      });
    };

    $scope.inittopology = function() {
      delete $scope.projectitems[$scope.projectid].topology;
      $scope.changeview("business", $scope.title.curr);
    };
    $scope.savetopology = function() {
      // json = r.toJSON(function(el,data){
      //   eldata = el.data("cooperative1");
      //   if (eldata == undefined) {
      //     return data;
      //   }
      //   return {"id": eldata.id};
      // });
      json = r.toJSON();
      //console.log(json);
      req = {
        "method": 'PUT',
        "url": BaseUrl+'/cmdb/project',
        "headers": {'Content-Type': "application/json; charset=UTF-8"},
        "data": {"_id": $scope.projectid, "topology": json},
      }
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('操作成功！', {icon: 1});
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      });
    }
    $scope.addField = function() {
      var len = $scope.fields.length;
      $scope.fields.push({
        "id": len,
        "name": "",
        "value": "",
        "status": 1,
      });
    }
    $scope.checkbox = {deployHistroy: false};
    $scope.newdate = function(time) {
      return new Date(time);
    }
    $scope.allowDeploy = function() {
      if ($scope.appitems[$scope.appsid].environment &&
      typeof $scope.appitems[$scope.appsid].environment == "object") {
        if ($scope.totalItems > 0) {
          return "allow";
        }
        return "nodata";
      } else {
        return "noenv";
      }
    }
    $scope.DoDeploy = function(id) {
      var req = {
        "method": 'GET',
        "url": BaseUrl+'/deploy/do',
        "params": {"deployid": id},
      };
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg("部署任务开始！", {icon: 1,time:1000,end:function(){
            $route.reload();}});
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      });
    }

    $scope.setStatus = function(item,tr) {
      // console.log(item);
      // console.log(tr);
      var req = {
        "method": 'PUT',
        "url": BaseUrl+'/deploy/testresult',
        "headers": {'Content-Type': "application/json; charset=UTF-8"},
        "data": {"deployid": item._id, "tr": tr},
      };
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg("设置成功！", {icon: 1,time:1000,end:function(){
            $route.reload();}});
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      });
    }
    $scope.chkenvperm = function(env) {
      if (env == $rootScope.curenv || env in $rootScope.menus.environment) {
        return true;
      }
      return false;
    }
    $scope.getCdenv = function(status,w) {
      var envindex = 0;
      if (status >= 10) {
        envindex = parseInt(status / 10) - 1;
      }
      var appenv = $scope.appitems[$scope.appsid].environment;
      var nextenvindex = 0;
      for (i = envindex+1; i < $scope.cdenv.length;i++) {
        if ($scope.cdenv[i] in appenv && appenv[$scope.cdenv[i]].hostscount &&
        appenv[$scope.cdenv[i]].hostscount > 0) {
          nextenvindex = i;
          break;
        }
      }
      if (w) {
        envindex = nextenvindex;
      }
      //return $rootScope.Environment[$scope.cdenv[envindex]]
      return $scope.cdenv[envindex];
    }

    $scope.deployFilter = function(value,index,carr) {
      if (($scope.checkbox.deployHistroy && (value.status == 0 || value.status == 2)) ||
      (value.status != 0 && value.status != 2)) {
        return true;
      }
      return false;
    }

    $scope.ok = function(isValid,dest) {
      switch (dest) {
        case "appconf":
          $scope.item.appid = $scope.appsid;
          $scope.item.from = $scope.appitems[$scope.appsid].from;
          $scope.item.content = base64Encode(mycode.getValue());
          $scope.item.operator = $rootScope.name;
          break;
        case "appvar":
          $scope.item = {};
          $scope.item._id = $scope.appsid;
          $scope.item.variables = {
            del: {},
            add: {},
            change: {},
          };
          for (k in $scope.fields) {
            name = $scope.fields[k].name;
            value = $scope.fields[k].value;
            if ($scope.fields[k].status == 0) {
              $scope.item.variables.change[name] = value;
            } else if ($scope.fields[k].status == 1) {
              $scope.item.variables.add[name] = value;
            } else {
              $scope.item.variables.del[name] = value;
            }
          }
          //delfields=fields.splice($index,1)
          //console.log($scope.item);
          dest = "appver";
          break;
        case "appadd":
          if (!($scope.item.from)) {
            layer.msg('请选择应用模板！', {icon: 5});
            return;
          }
          // if ($scope.item.source != "") {
          //   $scope.item.sourcetype = $scope.sourcetype;
          // }
          dest = "apps";
          break;
        default:

      }
      // if (dest == "appconf") {
      //   $scope.item.appid = $scope.appsid;
      //   $scope.item.from = $scope.appitems[$scope.appsid].from;
      //   $scope.item.content = base64Encode(mycode.getValue());
      //   console.log($scope.item);
      //   //return;
      // }
      var req = {
        "method": 'POST',
        "url": BaseUrl+'/cmdb/' + dest,
        "headers": {
          'Content-Type': "application/json; charset=UTF-8"
        },
        "data": $scope.item
      }
      if ($scope.item._id != undefined) {
        req.method = "PUT";
        delete $scope.item.createtime;
      }
      if (!isValid) {
        layer.msg('表单填写不正确，请仔细检查！', {icon: 5});
        return;
      }
      if (permissions.hasPermission("/cmdb/"+dest+":"+req.method) == false) {
        layer.msg('没有操作权限！', {icon: 4});
      }
      //console.log($scope.items);
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          layer.msg('操作成功！', {icon: 1,time:1000,end:function(){
            $route.reload();}});
          $scope.changeview($scope.showme.history,$scope.title.history);
        } else if (data.code == 1199) {
          $rootScope.loginform();
        } else {
          layer.msg(data.message, {icon: 2});
        }
      },function(data){
        layer.msg('无法获取信息!', {icon: 5});
      });
    }
    var stop;
    $scope.stopWait = function() {
      if (angular.isDefined(stop)) {
        $interval.cancel(stop);
      }
    }
    var req = {
      "method": 'GET',
      "url": BaseUrl+'/cmdb/apps',
    }
    $http(req).then(function (response) {
      var data = response.data;
      if(data.code==0){
        var depend = {};
        var clusterOid;
        var dlist;
        if (data.totalRecord > 0) {
          dlist = data.dataList.sort(function(a,b) {
            return a.name.localeCompare(b.name);
          });
        }
        for (var id =0 ;id < data.totalRecord; id++) {
          item = dlist[id]
          oid = item._id;
          $scope.appitems[oid] = item;
          var hcount = 0;
          if (item.environment && $rootScope.curenv in item.environment) {
            hcount = item.environment[$rootScope.curenv].hostscount;
          }
          //node.tags[0] = hcount;
          var node = {
            "text": item["name"],
            "oid": oid,
            "tags": [hcount],
            "type": 'apps',
          };
          if (item.status != 0 && $rootScope.username != item.operator && $rootScope.roles != 'admin') {
            node.tags.push('<i class="fa fa-lock"></i>')
          }
          clusterOid = item.cluster;
          if (clusterOid == "") {
            continue;
          }
          if ($scope.clusters[clusterOid] == undefined) {
            $scope.clusters[clusterOid] = [];
          }
          $scope.clusters[clusterOid].push(node);
          //console.log(item);
          //console.log(clusters);
        }
        for (var id =0 ;id < data.totalRecord; id++) {
          item = data.dataList[id]
          for (var j =0 ;j < item.depend.length; j++) {
            dp = item.depend[j];
            depend[dp] = 1;
            $scope.appitems[dp].depend = 1;
          }
        }
        //angular.forEach(
        for (coid in $scope.clusters) {
          for (var id=0 ;id < $scope.clusters[coid].length; id++) {
            oid = $scope.clusters[coid][id].oid;
            if (depend.hasOwnProperty(oid)) {
              if (parseInt($scope.clusters[coid][id].tags[0]) >= 0) {
                $scope.clusters[coid][id].tags.shift();
              }
              $scope.clusters[coid][id].color = "#ff6600";
            }
          }
        }
      } else if (data.code== 1199) {
        $rootScope.loginform();
      }
    }).then(function () {
      req.url = BaseUrl + '/cmdb/cluster';
      $http(req).then(function (response) {
        var data = response.data;
        if(data.code==0){
          var dlist;
          if (data.totalRecord > 0) {
            dlist = data.dataList.sort(function(a,b) {
              return a.name.localeCompare(b.name);
            });
          }
          for (var id =0 ;id < data.totalRecord; id++) {
            item = dlist[id]
            oid = item._id;
            $scope.clusteritems[oid] = item;
            var node = {
              text: item.name,
              oid: oid,
              tags: [],
              shared: item.shared,
              depend: item.depend,
              type: "cluster",
              state: {expanded: false},
              nodes: [],
            }
            projectOid = item.project;
            if (projectOid == "") {
              continue;
            }
            if ($scope.clusters[oid] != undefined) {
              hcount = 0;
              for (var i =0; i < $scope.clusters[oid].length; i++) {
                h = parseInt($scope.clusters[oid][i].tags[0])
                if (h > 0)
                  hcount += h;
              }
              node.nodes = $scope.clusters[oid];
              node.tags[0] = hcount;
            }
            if (permissions.hasPermission("/cmdb/apps:POST")) {
              node.nodes.push({
                "text": "添加应用",
                "type": "apps",
                "cluster": oid,
                "nameid": "add",
              });
            }
            if ($scope.projects[projectOid] == undefined) {
              $scope.projects[projectOid] = [];
            }
            $scope.projects[projectOid].push(node);
          }
        } else if (data.code== 1199) {
          $rootScope.loginform();
        }
      }).then(function () {
        req.url = BaseUrl + '/cmdb/project';
        $http(req).then(function (response) {
          var data = response.data;
          if(data.code==0){
            var dlist;
            if (data.totalRecord > 0) {
              dlist = data.dataList.sort(function(a,b) {
                return a.name.localeCompare(b.name);
              });
            }
            for (var id =0 ;id < data.totalRecord; id++) {
              //console.log(data.dataList[id])
              item = dlist[id]
              oid = item._id;
              $scope.projectitems[oid] = item;
              var node = {
                text: item.name,
                oid: oid,
                tags: [],
                state: {expanded: false},
                type: "project",
                nodes: [],
              }
              // if (id == 0) {
              //   node.state.selected = true;
              // }
              if ($scope.projects.hasOwnProperty(oid)) {
                node.nodes = $scope.projects[oid];
                hcount = 0;
                for (i =0; i < $scope.projects[oid].length; i++) {
                  if (parseInt($scope.projects[oid][i].tags[0]) > 0)
                    hcount += parseInt($scope.projects[oid][i]["tags"][0]);
                }
                node.tags[0] = hcount;
              }
              if (permissions.hasPermission("/cmdb/cluster:POST")) {
                node.nodes.push({
                  text: "添加集群",
                  type: "cluster",
                  project: oid,
                  nameid: "add",
                });
              }
              $scope.treedata[id] = node;
            }
            if (permissions.hasPermission("/cmdb/project:POST")) {
              $scope.treedata.push({
                text: '添加业务',
                type: 'project',
                nameid: 'add',
              });
            }
            //console.log($scope.treedata);
            $('#projecttree').treeview({
              "data": $scope.treedata,
              //enableLinks: true,
              "expandIcon": 'glyphicon glyphicon-chevron-right',
              "collapseIcon": 'glyphicon glyphicon-chevron-down',
              "color": '#428bca',
              "nodeIcon": '',
              "showBorder": false,
              "levels": 3,
              "showTags": true,
              onNodeSelected: function(event, data){
                $scope.projectop(data);
              }
            });
          } else if (data.code== 1199) {
            $rootScope.loginform();
          }
        });
      });
    });
}]);

app.controller('SSHController',['$rootScope','$scope','BaseUrl','$http','$route','$location',
  '$compile','$routeParams','SessionService','MyHTTP',
  function ($rootScope,$scope,BaseUrl,$http,$route,$location,$compile,$routeParams,SessionService,MyHTTP) {
    var newTerminal = function() {
        // Introducing the superSandbox()!  Use it to wrap any code that you don't want to load until dependencies are met.
        // In this example we won't call newTerminal() until GateOne.Terminal and GateOne.Terminal.Input are loaded.
        console.log("asdf");
        GateOne.Base.superSandbox("NewExternalTerm", ["GateOne.Terminal", "GateOne.Terminal.Input"], function(window, undefined) {
            "use strict";
            var existingContainer = GateOne.Utils.getNode('#'+GateOne.prefs.prefix+'container');
        // var container = GateOne.Utils.createElement('div', {
        //         'id': 'container', 'class': 'terminal', 'style': {'height': '100%', 'width': '100%'}
        // });
        //var gateone = GateOne.Utils.getNode('#gateone');
        // Don't actually submit the form
        // if (!existingContainer) {
        //         gateone.appendChild(container);
        // } else {
        //         container = existingContainer;
        // }
        // Create the new terminal
        GateOne.Terminal.newTerminal($rootScope.termNum, null, existingContainer);
        });
    };
    var req = {
      method: 'GET',
      url: BaseUrl+'/auth/sshauth',
    }
    $http(req).then(function (response) {
      var authobj = response.data;
      //GateOne.Logging.init();
      //GateOne.Logging.setLevel(10);
      GateOne.init({url: 'http://console.zdops.com:10443/',auth: authobj, embedded: true}, newTerminal);
    });
}]);

(function() {
	Raphael.fn.toJSON = function(callback) {
		var
			data,
			elements = new Array,
			paper    = this
			;

		for ( var el = paper.bottom; el != null; el = el.next ) {
			data = callback ? callback(el, new Object) : new Object;

			if ( data ) elements.push({
				data:      data,
				type:      el.type,
				attrs:     el.attrs,
				transform: el.matrix.toTransformString(),
				id:        el.id
				});
		}
    //console.log(elements);
		var cache = [];
		var o = JSON.stringify(elements, function (key, value) {
      //console.log(value);
		    //http://stackoverflow.com/a/11616993/400048
		    if (typeof value === 'object' && value !== null) {
		        if (cache.indexOf(value) !== -1) {
		            // Circular reference found, discard key
		            return;
		        }
		        // Store value in our collection
		        cache.push(value);
		    }
		    return value;
		});
		cache = null;
		return o;
	}

	Raphael.fn.fromJSON = function(json, callback) {
		var
			el,
			paper = this
			;

		if ( typeof json === 'string' ) json = JSON.parse(json);

		for ( var i in json ) {
			if ( json.hasOwnProperty(i) ) {
				el = paper[json[i].type]()
					.attr(json[i].attrs)
					.transform(json[i].transform);

				el.id = json[i].id;

				if ( callback ) el = callback(el, json[i].data);

				if ( el ) paper.set().push(el);
			}
		}
	}
})();
