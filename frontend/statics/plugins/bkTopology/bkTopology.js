/*!
 * bkTopology v1.0
 * author : 蓝鲸智云 
 * Copyright (c) 2012-2017 Tencent BlueKing. All Rights Reserved. 
 */
(function($){
	function Topology(config){
		var defaultConfig = {
			lineWidth: 2
		};
		this.config = $.extend(defaultConfig, config);
		this.curDom = config.curDom;
		this.data = config.data;
		this.lineBtnGroup ={};
		this.lineColor =[];
		this.initEvent(this.curDom);
		//this.dragToConnection();
		this.topologyContainerInfo =[];
		this.topologyContainerInfo['w'] = 0;
		this.topologyContainerInfo['h'] = 0;
	}
	Topology.prototype = {
		constructor : Topology,
		/*组件初始化方法*/
		init:function(){
			var tempData = [];
			this.curDom.addClass('topology-container');
			this.initLine(this.config.lineType);
			if(this.config && this.config.autoPosition){
				this.autoPosition(this.data);
			}
			this.initData(this.data,this.curDom);
			if(this.config.drag == true){
				var curDomWidth = Number(this.curDom.css('width').replace('px',''));
				var curDomHeight = Number(this.curDom.css('height').replace('px',''));
				if(curDomWidth <this.topologyContainerInfo['w'] || curDomHeight <this.topologyContainerInfo['h']){
					this.reSize();
				}
			}
		},
		/*初始化节点和线条信息*/
		initData:function(data){
			var topThis = this;
			data.nodes.forEach(function(v,i){
				topThis.drawNode(v);
			});
			data.edges.forEach(function(v,i){
				topThis.drawPathyun_shi(v);
			});
		},
		/*渲染节点*/
		drawNode:function(v){
			var uid="",v;
			if(v && v.id && v.id.length>1){
				uid = v.id;
			}else{
				uid = this.getUUid();
			}
			if(this.topologyContainerInfo['w'] < Number(v.x)){
				this.topologyContainerInfo['w'] = Number(v.x);
			}
			if(this.topologyContainerInfo['h'] < Number(v.y)){
				this.topologyContainerInfo['h'] = Number(v.y);
			}
			var templates = $('#node-templates').clone().removeAttr('id')
				.addClass(v.className)
				.removeClass("none").attr('data-id',uid)
				.css({
					top:(v.y)+'px',
					left:(v.x)+'px',
					height:v.height,
					width:v.width
			});
			if(v.title){
				templates.attr('data-content',v.title);
			}
			templates.find('.node-text').html(v.text);
			templates.find('i').attr('data-node-id',uid);
			templates.data('dataType',"node");
			this.curDom.append(templates);
		},
		/*重新计算屏幕*/
		reSize:function(){
			this.curDom.css({
				width:(this.topologyContainerInfo['w']+200)+'px',
				height:(this.topologyContainerInfo['h']+200)+'px'
			})
		},
		/*使用 dagre 组件计算节点和线条位置*/
		autoPosition:function(data){
			var topThis = this;
			var g = new dagre.graphlib.Graph({

			});
			g.setGraph({
				rankdir: 'LR'
				// nodesep: 100,
				// align: 'DR'
			});
			g.setDefaultEdgeLabel(function() { return {}; });

      		data.nodes.forEach(function(v,i){
        		g.setNode(v.id, v);
      		});

      		data.edges.forEach(function(v,i){
        		g.setEdge(v.source, v.target,{});
      		});

      		dagre.layout(g);

			var temp = [];
      		g.edges().forEach(function(e){
      			topThis.data.edges.forEach(function(v){
      				if(e.v == v.source && e.w == v.target){
      					temp.push({
		      				"source":e.v,
		      				"target":e.w,
		      				"points":g.edge(e).points,
		      				"edgesType":v.edgesType
		      			})
      				}
      			})
			});
			topThis.data['edges'] = temp;
		},
		/*	根据源节点的出线点和目标节点的入线点计算线条的点位置信息*/
		computePathPostion:function(data){
			var target =$('[data-id="'+data.target+'"]');
		   	var source =$('[data-id="'+data.source+'"]');
		   	var s=[],t=[],pathPostion=[];

			s['l'] = source.position().left;
			s['t'] = source.position().top;
			s['w'] = source.innerWidth();
			s['h'] = source.innerHeight();

			t['l'] = target.position().left;
			t['t'] = target.position().top;
			t['w'] = target.innerWidth();
			t['h'] = target.innerHeight();

			// 中心点
			pathPostion.push({
				x:s['w']/2+s['l'],
				y:s['h']/2+s['t']
			});

			/*当源节点的线在*/
			var variant = t['h']/2;
			if(data.sDirection == 'top' && data.tDirection == 'top'){
				if(s['l'] > t['l']){
					pathPostion.push({
						x:s['l'] < t['l'] ? t['l']+t['w']/2 : s['l']+s['w']/2,
						y:s['t'] < t['t'] ? s['t']-variant : t['t']-variant
					});
					pathPostion.push({
						x:s['l'] < t['l'] ? s['l']+s['w']/2 :t['l']+t['w']/2,
						y:s['t'] < t['t'] ? s['t']-variant : t['t']-variant
					});
				}else{
					pathPostion.push({
						x:s['l'] < t['l'] ? s['l']+s['w']/2 :t['l']+t['w']/2,
						y:s['t'] < t['t'] ? s['t']-variant : t['t']-variant
					});
					pathPostion.push({
						x:s['l'] < t['l'] ? t['l']+t['w']/2 : s['l']+s['w']/2,
						y:s['t'] < t['t'] ? s['t']-variant : t['t']-variant
					});
				}
			}else if(data.sDirection == 'bottom' && data.tDirection == 'top'){
				// 开始结点比目标结点位置高，且不在一条垂直线上
				if(s['t']< t['t'] && (s['l']+s['w']/2) != (t['l']+t['w']/2)){
					pathPostion.push({
						x:s['l']+s['w']/2,
						y:(t['t'] - s['t']+s['h'])/2+s['t']
					});
					pathPostion.push({
						x:t['l']+t['w']/2,
						y:(t['t'] - s['t']+s['h'])/2+s['t']
					});
				}
			}else if(data.sDirection == 'bottom' && data.tDirection == 'bottom'){
				if(s['l'] < t['l']){
					pathPostion.push({
						x:(s['w']/2)+s['l'],
						y:s['h']+s['t']+variant
					});

					pathPostion.push({
						x:(t['w']/2)+t['l'],
						y:t['h']+t['t']+variant
					});
				}else{
					pathPostion.push({
						x:(s['w']/2)+s['l'],
						y:s['h']+s['t']+variant
					});
					pathPostion.push({
						x:(t['w']/2)+t['l'],
						y:t['h']+t['t']+variant
					});
				}
			}else if(data.sDirection == 'bottom' && data.tDirection == 'right'){
				if(s['t']< t['t']){
					var maxTop = 99999;
					var right =0;
					this.data.nodes.forEach(function(v){
						if(Math.floor(s['t']) < v.y && v.y < Math.floor(t['t'])){
							if(maxTop > v.y){
								maxTop = v.y;
							}
							if(t['l'] < v.x){
								if(right < v.x){
									right = v.width/2;
								}
							}
						}
					});

					if(maxTop == 99999){
						maxTop =t['t']
					}
					pathPostion.push({
						x:s['l']+s['w']/2,
						y:(maxTop - s['t']+s['h'])/2+s['t']
					});
					pathPostion.push({
						x:t['l']+t['w']+variant+right,
						y:(maxTop - s['t']+s['h'])/2+s['t']
					});
					pathPostion.push({
						x:t['l']+t['w']+variant+right,
						y:t['t']+t['h']/2
					});
				}
			}else if(data.sDirection == 'bottom' && data.tDirection == 'left'){
				if(s['t']< t['t']){
					var maxTop = 99999;
					var left = 0;
					this.data.nodes.forEach(function(v){
						if(Math.floor(s['t']) < v.y && v.y < Math.floor(t['t'])){
							if(maxTop > v.y){
								maxTop = v.y;
							}
							if(t['l']>v.x){
								if(left < v.x){
									left = v.width/2;
								}
							}
						}
					})
					if(maxTop == 99999){
						maxTop =t['t']
					}
					pathPostion.push({
						x:s['l']+s['w']/2,
						y:(maxTop - s['t']+s['h'])/2+s['t']
					});

					pathPostion.push({
						x:t['l']-variant-left,
						y:(maxTop - s['t']+s['h'])/2+s['t']
					});

					pathPostion.push({
						x:t['l']-variant-left,
						y:t['t']+t['h']/2
					});
				}
			}else if(data.sDirection == 'right' && data.tDirection == 'top'){
				pathPostion.push({
					x:s['l']+s['w']+variant,
					y:s['t']+s['h']/2
				});

				pathPostion.push({
					x:s['l']+s['w']+variant,
					y:t['t']-variant
				});

				pathPostion.push({
					x:t['l']+t['w']/2,
					y:t['t']-variant
				});
			}else if(data.sDirection == 'right' && data.tDirection == 'left'){
				// 开始结点比目标结点位置靠左，且不在一条水平线上
				if(s['l']< t['l'] && (s['t']+s['h']/2) != (t['t']+t['h']/2)){
					pathPostion.push({
						x:s['l']+s['w']+variant,
						y:s['t']+s['h']/2
					});

					pathPostion.push({
						x:s['l']+s['w']+variant,
						y:t['t']+variant
					});
				}
			}else if (data.sDirection == "right" && data.tDirection == "right"){
				if(s['l'] <= t['l']){
					pathPostion.push({
						x:s['l']+s['w']+variant,
						y:s['t']+s['h']/2
					})

					pathPostion.push({
						x:t['l']+t['w']+variant,
						y:t['t']+t['h']/2
					})
				}else{
					var left = 0;
					this.data.nodes.forEach(function(v){
						if(Math.floor(s['t']) < v.y && v.y < Math.floor(t['t'])){
							if(t['l'] < v.x){
								if(left < v.x){
									left = v.width/2;
								}
							}
						}
					})
					pathPostion.push({
						x:t['l']+t['w']+left,
						y:s['t']+s['h']/2
					})
					pathPostion.push({
						x:t['l']+t['w']+left,
						y:t['t']+t['h']/2
					})
				}
			}else if (data.sDirection == "left" && data.tDirection == "top"){
				pathPostion.push({
					x:s['l']-variant,
					y:s['t']+s['h']/2
				})

				pathPostion.push({
					x:s['l']-variant,
					y:t['t']-variant
				})

				pathPostion.push({
					x:t['l']+t['w']/2,
					y:t['t']-variant
				})

			}else if (data.sDirection == "left" && data.tDirection == "left"){
				pathPostion.push({
					x:s['l']-variant,
					y:s['t']+s['h']/2
				})

				pathPostion.push({
					x:s['l']-variant,
					y:t['t']+variant
				})

			}
			//目标中心点
			pathPostion.push({
				x:t['w']/2+t['l'],
				y:t['h']/2+t['t'],
				t:t
			});
			return pathPostion;
		},
		/* 根据线条点信息绘制路径*/
		drawPathyun_shi:function(data){
		   	var v = this.computePathPostion(data);
		   	var isLine = true;
		   	var d="";
		   	d += "M "+v[0].x+" "+v[0].y;
		   	var MTemp = " M x y ";
		   	var LTemp = " L x y ";
		   	var ATemp = " A 5 5 0 0,f x y ";
		   	var ARROW_OFFSET = 5;


		   	if(v.length > 2){
		   		for(var i = 1 ;i<v.length-1;i++){
		   			if(v[i].x == v[i-1].x){
		   				d += LTemp.replace('x',v[i].x).replace('y',v[i].y + (v[i].y > v[i-1].y ? -ARROW_OFFSET:+ARROW_OFFSET));
		   				d += MTemp.replace('x',v[i].x).replace('y',v[i].y + (v[i].y > v[i-1].y ? -ARROW_OFFSET:+ARROW_OFFSET));
		   				if(v[i].x > v[i+1].x){
		   					d += ATemp.replace('f',v[i].y > v[i-1].y ? 1:0).replace('x',v[i].x-ARROW_OFFSET).replace('y',v[i].y);
		   					d += MTemp.replace('x',v[i].x-ARROW_OFFSET).replace('y',v[i].y);
		   				}else{
		   					d += ATemp.replace('f',v[i].y < v[i-1].y ? 1:0).replace('x',v[i].x+ARROW_OFFSET).replace('y',v[i].y);
		   					d += MTemp.replace('x',v[i].x+ARROW_OFFSET).replace('y',v[i].y);
		   				}
		   			}else{
		   				var lx = 0;
		   				if(Math.floor(v[i].x) > Math.floor(v[i-1].x)){
		   					lx = v[i].x - ARROW_OFFSET;
		   				}else if(Math.floor(v[i].x) < Math.floor(v[i-1].x)){
		   					lx = v[i].x + ARROW_OFFSET;
		   				}else{
		   					lx = v[i].x;
		   				}
		   				d += LTemp.replace('x',lx).replace('y',v[i].y);
		   				d += MTemp.replace('x',v[i].x + (v[i].x > v[i-1].x ? -ARROW_OFFSET:+ARROW_OFFSET)).replace('y',v[i].y);
		   				if(v[i].y > v[i+1].y){
		   					d += ATemp.replace('f',v[i].x < v[i-1].x ? 1:0).replace('x',v[i].x).replace('y',v[i].y-ARROW_OFFSET);
		   					d += MTemp.replace('x',v[i].x).replace('y',v[i].y-ARROW_OFFSET);
		   				}else{
		   					d += ATemp.replace('f',v[i].x > v[i-1].x ? 1:0).replace('x',v[i].x).replace('y',v[i].y+ARROW_OFFSET);
		   					d += MTemp.replace('x',v[i].x).replace('y',v[i].y+ARROW_OFFSET);
		   				}
		   			}
		   			if(i == v.length-2){
		   				d += LTemp.replace('x',v[i+1].x).replace('y',v[i+1].y);
		   			}
				}
		   	}else{
				d += ' L '+v[1].x +" "+v[1].y;
		   	}

			var arrowInfo = v[v.length-1];
			var tDirection = data.tDirection;

			var arrow_d =""
			var MTemp =" M x1 y1";
			var LTemp =" L x1 y1";
			var arrowWidth = this.config.lineWidth*1.5;
			var variant = this.config.lineWidth;

			if(tDirection == "top"){
				var ax = arrowInfo.x,ay= arrowInfo.y-arrowInfo.t['h']/2 - variant;
				arrow_d = MTemp.replace("x1",ax).replace("y1",ay);
				arrow_d += LTemp.replace("x1",ax+arrowWidth/2).replace("y1",ay-arrowWidth);
				arrow_d += LTemp.replace("x1",ax).replace("y1",ay-arrowWidth);
				arrow_d += LTemp.replace("x1",ax-arrowWidth/2).replace("y1",ay-arrowWidth);
				arrow_d += LTemp.replace("x1",ax).replace("y1",ay);
			}else if (tDirection == "left"){
				var ax = arrowInfo.x-arrowInfo.t['w']/2 -variant,ay= arrowInfo.y;
				arrow_d = MTemp.replace("x1",ax).replace("y1",ay);
				arrow_d += LTemp.replace("x1",ax-arrowWidth).replace("y1",ay-arrowWidth/2);
				arrow_d += LTemp.replace("x1",ax-arrowWidth).replace("y1",ay);
				arrow_d += LTemp.replace("x1",ax-arrowWidth).replace("y1",ay+arrowWidth/2);
				arrow_d += LTemp.replace("x1",ax).replace("y1",ay);
			}else if (tDirection == "right"){
				var ax = arrowInfo.x+arrowInfo.t['w']/2 +variant*2,ay= arrowInfo.y;
				arrow_d = MTemp.replace("x1",ax).replace("y1",ay);
				arrow_d += LTemp.replace("x1",ax+arrowWidth).replace("y1",ay+arrowWidth/2);
				arrow_d += LTemp.replace("x1",ax+arrowWidth).replace("y1",ay);
				arrow_d += LTemp.replace("x1",ax+arrowWidth).replace("y1",ay-arrowWidth/2);
				arrow_d += LTemp.replace("x1",ax).replace("y1",ay);
			}else if (tDirection == "bottom"){

				var ax = arrowInfo.x,ay= arrowInfo.y+arrowInfo.t['h']/2 +variant*2;
				arrow_d = MTemp.replace("x1",ax).replace("y1",ay);
				arrow_d += LTemp.replace("x1",ax-arrowWidth/2).replace("y1",ay+arrowWidth);
				arrow_d += LTemp.replace("x1",ax).replace("y1",ay+arrowWidth);
				arrow_d += LTemp.replace("x1",ax+arrowWidth/2).replace("y1",ay+arrowWidth);
				arrow_d += LTemp.replace("x1",ax).replace("y1",ay);
			}

			var data_uid = this.getUUid();

			var tempPath =$('<svg style="width: 100%;height: 100%;" pointer-events="none">'+
		            		'<path data-id="'+data_uid+'" d="'+d+'" pointer-events="visibleStroke"  fill="none" stroke="'+this.lineColor[data.edgesType]+'" stroke-width="'+this.config.lineWidth+'px";"></path>'+
		            		'<path d="'+arrow_d+'" stroke="'+this.lineColor[data.edgesType]+'" fill="'+this.lineColor[data.edgesType]+'" class="node-path" style="stroke-width: 5px;"></path>'+
		         	  	'</svg>');

		    var result = this.curDom.append(tempPath);
		    data.dataType="path";
		    tempPath.find('[data-id]').data(data)

		    return;
		},
		getUUid:function(){
			var s = [];
		    var hexDigits = "0123456789abcdef";
		    for (var i = 0; i < 36; i++) {
		        s[i] = hexDigits.substr(Math.floor(Math.random() * 0x10), 1);
		    }
		    s[14] = "4";
		    s[19] = hexDigits.substr((s[19] & 0x3) | 0x8, 1);
		    s[8] = s[13] = s[18] = s[23] = "-";

		    var uuid = s.join("");
		    return uuid;
		},
		initLine:function(lineType){
			var tempArray = [];
			lineType.forEach(function(v,i){
				tempArray[v.type] = v.lineColor;
			})
			this.lineColor = tempArray;
		},
		/*获取节点信息*/
		getNodes:function(){
			var tempArray = [];
			this.curDom.find('.node').each(function(i,v){
				var dom = $(v);
				tempArray.push({
					id:dom.attr('data-id'),
					x:Number(dom.css('left').replace("px","")),
					y:Number(dom.css('top').replace("px","")),
					height:dom.css('height').replace('px',''),
					width:dom.css('width').replace('px',''),
					text:dom.find('.node-text').html(),
					title:dom.find('[data-content]').val(),
					className:dom.prop('class')
				});
			})
			return tempArray;
		},
		/*获取连线信息*/
		getEdges:function(){
			var tempArray = [];
			this.curDom.find('path[data-id]').each(function(i,v){
				var dom = $(v);
				var jqData = dom.data();
				tempArray.push(jqData);
			})
			return tempArray;
		},
		/* 删除节点或者线条*/
		remove:function(arguments){
			var jqDom,topThis = this;
			if(arguments[0] && typeof(arguments[0]) == 'string'){
				jqDom = $(arguments[0]);
			}else if(arguments[0] instanceof jQuery){
				jqDom = arguments[0];
			}else{
				return ;
			}
			var dataId = jqDom.attr('data-id');
			if(dataId){
				var jqData = jqDom.data();
				if(jqData.dataType == 'node'){
					topThis.getEdges().forEach(function(v,i){
						if(v.source == dataId || v.target == dataId){
							topThis.curDom.find('[data-id="'+v.id+'"]').parent().remove();
						}
					})
					jqDom.remove();
				}else if(jqData.dataType == 'path'){
					jqDom.parent().remove();
				}
			}
		},
		/*重新初始化*/
		reLoad:function(type, data){
			if(type == "node"){
				this.data['nodes'].push(data);
			}
			if(type == "edge"){
				this.data['edges'].push(data);
			}
			this.curDom.empty();
			this.init();
		},
		/*拖拽连线*/
		dragToConnection:function(){
			var isMouseDown =false;
			var topThis = this;
			var position = [];
			var edgesType ="";
			var dataId=null;
			var nodeChildrenClassArray =[];
			this.curDom.on('mousedown','.node',function(e){
				var className = $(e.toElement).prop('class');
				dataId = $(this).attr('data-id');
				if(className && className.indexOf('option-btn') != -1){
					getChildrenClassName();
					isMouseDown = true;
					position['start_x'] = e.pageX;
					position['start_Y'] = e.pageY;
			        if(className.indexOf('success') != -1){
			            edgesType= 'success';
			        }else if(className.indexOf('failure') != -1){
			            edgesType= 'failure'
			        }else if(className.indexOf('other') != -1){
			            edgesType= 'check'
			        }
			        var tempPath =$('<svg style="width: 100%;height: 100%;" pointer-events="none">'+
		            		'<path class="dragToConnectionSvge" d="" pointer-events="visibleStroke" stroke-width="2px";"></path>'+
		            		/*'<path d="'+arrow['d']+'" stroke="'+this.lineColor[data.edgesType]+'" fill="'+this.lineColor[data.edgesType]+'" class="node-path" style="stroke-width: 1px;"></path>'+*/
		         	  	'</svg>');
					topThis.curDom.append(tempPath);

			        topThis.curDom.find('.dragToConnectionSvge').attr({
			        	"stroke":topThis.lineColor[edgesType],
			        	"fill":topThis.lineColor[edgesType]
			        })
				}
			})
			.on('mouseup',function(e){
				var className = $(e.toElement).prop('class');
				try
				{
				   	className = className.split(' ');
				   	if(className.length>0 && nodeChildrenClassArray.indexOf(className[className.length-1]) != -1){
				   		var nodeId = $(e.toElement).parentsUntil('.topology-container').find('i').data('nodeId');
				   		if(nodeId && nodeId != dataId){
				   			var result = topThis.config.onConnection.apply(this,[e,topThis.getEdges()]);
				   			if(result == false){
								return;
							}
				   			var temp = [];
					        temp['edges'] ={"source": dataId, "target": nodeId, edgesType:edgesType}
				   			topThis.reLoad(temp)
				   		}
					}
				}catch(err){

				}finally{
					topThis.curDom.find('.dragToConnectionSvge').parent().remove();
					isMouseDown = false;
				}
			})
			.on('mousemove',function(e){
				if(isMouseDown){
					showLine(e);
				}
			})
			function showLine(e){
				var y = Number(topThis.curDom.css('top').replace('px',"").replace('auto','0'));
				var x = Number(topThis.curDom.css('left').replace('px',"").replace('auto','0'));

				var d = " M "+(position['start_x'] - x) +" "+(position['start_Y']-y )+ " L "+(e.pageX -x)+" "+(e.pageY-y);
				topThis.curDom.find('.dragToConnectionSvge').attr('d',d);
			}
			function getChildrenClassName(){
				if(nodeChildrenClassArray.length == 0){
					var temp ="";
					topThis.curDom.find('.node:eq(1) *').each(function(i,v){
						temp += $(v).prop('class')+" ";
					});
					nodeChildrenClassArray = temp.split(' ');
					nodeChildrenClassArray[nodeChildrenClassArray.length-1] ='node';
				}
			}
		},
		/*初始化通用事件*/
		initEvent:function(curDom){
			var topThis = this;
			var isMouseDown = false;
			var mouseXY = [];

			curDom
			.on('mouseenter','svg path',function(e){
				e.stopPropagation();
				e.preventDefault();
				$(this).css({
					'stroke-width': (topThis.config.lineWidth * 1.3) + 'px',
					'z-index':'1'
				});
			})
			.on('mouseleave','svg path',function(e){
				e.stopPropagation();
				e.preventDefault();
				$(this).css({'stroke-width': topThis.config.lineWidth+ 'px', 'z-index':'inherit'});
			})
			.on('mousemove',function(e){
				if(isMouseDown && topThis.config.drag == true){
					var move = [],moveTo = [],div = [];
					var $this = $(this);

					move['x'] = e.pageX;
					move['y'] = e.pageY;

					moveTo['t'] = move['y'] - mouseXY['y'];
					moveTo['l'] = move['x'] - mouseXY['x'];

					curDom.css({
						'top':moveTo['t']+mouseXY['top'],
						'left':moveTo['l']+mouseXY['left']
					})
				}
			}).on('mousedown',function(e){
				if(topThis.config.drag == true){
					mouseXY['x'] = e.pageX;
					mouseXY['y'] = e.pageY;
					mouseXY['top'] = Number(curDom.css('top').replace('px',"").replace('auto',"0"));
					mouseXY['left'] = Number(curDom.css('left').replace('px',"").replace('auto',"0"));
					if($(this).prop('class') == $(e.toElement).prop('class')){
						isMouseDown = true;
					}
				}
			});
			if(topThis.config.drag == true){
				$('body').on('mouseup',function(e){
					isMouseDown = false;
					mouseXY=[];
				});
			}
		}
	}
	$.fn.bkTopology = function(config){
		config.curDom = $(this)
		var topo={};
		if(typeof(arguments[0]) == 'object'){
			if($(this).prop('class').indexOf('topology-container') ==-1){
				topo = new Topology(config);
				topo.init();
			}
		}
		var methods={
			remove:function(){
				topo.remove(arguments);
			},
			getNodes:function(){
				return topo.getNodes();
			},
			getEdges:function(){
				return topo.getEdges();
			},
			drawNode:function(){
				topo.drawNode(arguments);
			},
			reLoad:function(type,data){
				topo.reLoad(type,data);
			},
			getUUid:function(){
				return topo.getUUid();
			}
		}
		return methods;
	}
})(jQuery);

