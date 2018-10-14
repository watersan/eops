/**
 * Created by Administrator on 2015/11/20.
 */
//改造了一下 将标签改为scriptlazy 否则会导致sript标签加载两次的bug
//
(function (ng) {
    'use strict';
    var app = ng.module('ngLoadScript', []);
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
}(angular));