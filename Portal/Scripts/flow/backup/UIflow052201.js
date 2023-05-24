// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*'use strict';
function _0x30f6(_0x180a7f,_0x5d5a98){var _0x42049c=_0x4204();return _0x30f6=function(_0x30f6b3,_0x100ba6){_0x30f6b3=_0x30f6b3-0x122;var _0x115fb4=_0x42049c[_0x30f6b3];return _0x115fb4;},_0x30f6(_0x180a7f,_0x5d5a98);}(function(_0x48010b,_0x19cbdd){var _0x36ee8e=_0x30f6,_0x5a0168=_0x48010b();while(!![]){try{var _0x3f1e92=parseInt(_0x36ee8e(0x132))/0x1*(-parseInt(_0x36ee8e(0x129))/0x2)+-parseInt(_0x36ee8e(0x137))/0x3*(-parseInt(_0x36ee8e(0x123))/0x4)+parseInt(_0x36ee8e(0x124))/0x5*(parseInt(_0x36ee8e(0x133))/0x6)+-parseInt(_0x36ee8e(0x12b))/0x7+parseInt(_0x36ee8e(0x12d))/0x8+-parseInt(_0x36ee8e(0x131))/0x9+parseInt(_0x36ee8e(0x12c))/0xa;if(_0x3f1e92===_0x19cbdd)break;else _0x5a0168['push'](_0x5a0168['shift']());}catch(_0x122998){_0x5a0168['push'](_0x5a0168['shift']());}}}(_0x4204,0x65ee6),(!function(_0x52457c){var _0x3ce8ad=_0x30f6,_0x406208=[];_0x52457c[_0x3ce8ad(0x12a)](!0x0,{'import_js':function(_0x2f9a80){var _0x461c98=_0x3ce8ad;for(var _0x51c96c=!0x1,_0x401e12=0x0;_0x401e12<_0x406208['length'];_0x401e12++)if(_0x406208[_0x401e12]==_0x2f9a80){_0x51c96c=!0x0;break;}0x0==_0x51c96c&&(_0x52457c(_0x461c98(0x125))[_0x461c98(0x136)](_0x461c98(0x130)+_0x2f9a80+_0x461c98(0x128)),_0x406208[_0x461c98(0x126)](_0x2f9a80));}});}(jQuery),function(){var _0x4c4a55=_0x30f6,_0x5a2660=_0x4c4a55(0x122);$[_0x4c4a55(0x12e)](_0x5a2660+'D3.V5.0/d3.min.js'),$['import_js'](_0x5a2660+'/Dagre/dagre.min.js'),$[_0x4c4a55(0x12e)](_0x5a2660+_0x4c4a55(0x134)),$[_0x4c4a55(0x12e)](_0x5a2660+_0x4c4a55(0x135)),$[_0x4c4a55(0x12e)](_0x5a2660+_0x4c4a55(0x127)),$[_0x4c4a55(0x12e)](_0x5a2660+'svc/flow/joint.js'),$[_0x4c4a55(0x12e)](_0x5a2660+'svc/flow/svg-pan-zoom.js'),svclpmsolution&&null!=svclpmsolution||$[_0x4c4a55(0x12e)](_0x5a2660+_0x4c4a55(0x12f));}()));function _0x4204(){var _0x13bc01=['push','svc/flow/backbone.js','\x22></script>','6808rZxEEQ','extend','3812760FSSlEC','8344400jHosdv','1938152kVlFMa','import_js','svc/uiflow_lpm_core.min.js','<script\x20type=\x22text/javascript\x22\x20src=\x22','3178584BZjKoc','109zPxABb','6CsPHER','svc/flow/lodash.js','svc/flow/graphlib.js','append','3aBfUgA','/Apriso/Portal/scripts/','2145056BkbtWG','367145nohBuh','head'];_0x4204=function(){return _0x13bc01;};return _0x4204();}

*/
(function($)
{

    var import_js_imported = [];

    $.extend(true,
    {
        import_js : function(script)
        {
            var found = false;
            for (var i = 0; i < import_js_imported.length; i++)
                if (import_js_imported[i] == script) {
                    found = true;
                    break;
                }

            if (found == false) {
                $("head").append('<script type="text/javascript" src="' + script + '"></script>');
                import_js_imported.push(script);
            }
        }
    });

})(jQuery);

(function(){		

	var path = '/portal/scripts/'
	$.import_js(path + "D3.V5.0/d3.min.js")
	$.import_js(path + "Dagre/dagre.min.js")
	$.import_js(path + "flow/lodash.js")
	$.import_js(path + "flow/graphlib.js")
	$.import_js(path + "flow/backbone.js") 
	$.import_js(path + "flow/joint.js")
	$.import_js(path + "flow/svg-pan-zoom.js")  
	$.import_js(path + "flow/filesave.js")  
	$.import_js(path + "UIForm.js")  
	$.import_js(path + "contextmenu/jquery.contextMenu.js")  

})() 

var UIFlow;
(function (UIFlow) {   
	function generateUUID(){
		var d = new Date().getTime();
		var uuid = 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
			var r = (d + Math.random()*16)%16 | 0;
			d = Math.floor(d/16);
			return (c=='x' ? r : (r&0x3|0x8)).toString(16);
		});
		return uuid;    
	}
	UIFlow.generateUUID = generateUUID;
	function safeName(name){
		return name.replace(/[^a-zA-Z0-9]/g, "_");
	}
	UIFlow.safeName = safeName;
	function safeId(id){
		return id.replace(/[^a-zA-Z0-9]/g, "_");
	}   
	UIFlow.safeId = safeId; 
	function safeClass(className){
		return className.replace(/[^a-zA-Z0-9]/g, "_");
	}   
	UIFlow.safeClass = safeClass;
	function replaceAll(target, search, replacement) {
		return target.replace(new RegExp(search, "g"), replacement);
	}
	UIFlow.replaceAll = replaceAll;
	
})(UIFlow || (UIFlow = {}));

UIFlow = UIFlow || {};

joint.shapes.basic.MergePoint = joint.shapes.basic.Generic.extend({

    markup: '<g class="rotatable"><g class="scalable"><rect/></g><text/></g>',

    defaults: joint.util.deepSupplement({

        type: 'basic.Rect',
        attrs: {
            'rect': { 
				fill: 'darkblue', 
				stroke: 'darkblue', 
				"stroke-width":0.5,
				width: 0.5, 
				height: 0.5,
				transform: 'rotate(45)' 
				},
            'text': { 
				'font-size': 14, text: '', 'ref-x': .5, 'ref-y': .5, ref: 'rect', 'y-alignment': 'middle', 'x-alignment': 'middle', fill: 'black', 'font-family': 'Arial, helvetica, sans-serif' }
        }

    }, joint.shapes.basic.Generic.prototype.defaults)
});

joint.shapes.standard.Rectangle.define('ProcessFlow.OperationBlock', {
        attrs: {
	        body: {
				rx:10,
				ry:10,
	            refWidth: '100%',
	            refHeight: '100%',
	            strokeWidth: 0,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
	        header: {
				rx:5,
				ry:5,
	            refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
	        headerText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',
	            refX: '50%',
	            refY: 12,
	            fontSize: 16,
	            fill: '#333333'
	        },
	        workcenter: {
				rx:5,
				ry:5,
	            refY: '25',
				refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
			descriptionbox: {
				rx:5,
				ry:5,
	            refY: '50',
				refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
			featuresbox: {
				rx:5,
				ry:5,
	            refY: '75',
				refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
	        workcenterText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',
	            refX: '50%',
	            refY: 37,
	            fontSize: 16,
	            fill: '#333333'
	        },
	        descriptionText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',				
	            refX: '50%',
	            refY2: 60,
	            fontSize: 12,
	            fill: '#333333'
	        },
			featuresText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',				
	            refX: '50%',
	            refY2: 85,
	            fontSize: 12,
	            fill: '#333333'
	        },
	    }
    }, {
        // inherit joint.shapes.standard.Rectangle.markup
		
	    markup: [{
	        tagName: 'rect',
	        selector: 'body'
	    }, {
	        tagName: 'rect',
	        selector: 'header'
	    }, {
	        tagName: 'text',
	        selector: 'headerText'
	    },{
	        tagName: 'rect',
	        selector: 'workcenter'
	    }, {
	        tagName: 'rect',
	        selector: 'descriptionbox'
	    },{
	        tagName: 'rect',
	        selector: 'featuresbox'
	    },{
	        tagName: 'text',
	        selector: 'workcenterText'
	    },{
	        tagName: 'text',
	        selector: 'descriptionText'
	    },{
	        tagName: 'text',
	        selector: 'featuresText'
	    }]
    }, {
        
    });

joint.shapes.standard.Rectangle.define('ProcessFlow.StepBlock', {
        attrs: {
	        body: {
				rx:10,
				ry:10,
	            refWidth: '100%',
	            refHeight: '100%',
	            strokeWidth: 0,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
	        header: {
				rx:5,
				ry:5,
	            refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
	        headerText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',
	            refX: '50%',
	            refY: 12,
	            fontSize: 16,
	            fill: '#333333'
	        },
	        stepname: {
				rx:5,
				ry:5,
	            refY: '25',
				refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
			features: {
				rx:5,
				ry:5,
	            refY: '50',
				refWidth: '100%',
	            height: 50,
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#FFFFFF'
	        },
			routerflag: {
				refX:"0",
				refY:"50",           
				width: "35",
	            height: "35",
				transform: "translate(-50%, -50%) rotate(45deg)",					
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: 'none',
				x: '160',
				y: '-90'
	        },
	        stepnameText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',
	            refX: '50%',
	            refY: 37,
	            fontSize: 16,
	            fill: '#333333'
	        },
	        featuresText: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',
	            refX: '50%',
	            refY: 60,
	            fontSize: 12,
	            fill: '#333333'
	        }
			
	    }
    }, {
        // inherit joint.shapes.standard.Rectangle.markup
		
	    markup: [
		{
	        tagName: 'rect',
	        selector: 'body'
	    }, {
	        tagName: 'rect',
	        selector: 'header'
	    },{
	        tagName: 'rect',
	        selector: 'features'
	    },
		
		{
	        tagName: 'text',
	        selector: 'headerText'
	    },{
	        tagName: 'rect',
	        selector: 'stepname'
	    }, {
	        tagName: 'text',
	        selector: 'stepnameText'
	    },{
	        tagName: 'text',
	        selector: 'featuresText'
	    },{
	        tagName: 'rect',
	        selector: 'routerflag'
	    }]
    }, {
        
    });

joint.shapes.standard.Rectangle.define('ProcessFlow.StepBlock.Function', {
        attrs: {
			root:{
			//	magnet: false,
			},			
	        body: {
				rx:0,
				ry:0,
	            refWidth: '100%',
	            refHeight: '50',
	            strokeWidth: 2,
	            stroke: '#000000',
	            fill: '#8ECAE6',
			//	magnet: false,
	        },
	        functionheader: {
				rx:0,
				ry:0,
	            refWidth: '100%',
	            height: 25,
	            strokeWidth: 2,
	            stroke: '#000000',
			//	magnet: false,
	            fill: '#8ECAE6'
	        },
	        functionname: {
	            textVerticalAnchor: 'middle',
	            textAnchor: 'middle',
	            refX: '50%',
	            refY: 12,
	            fontSize: 16,
	            fill: 'black'
	        },
	    /*    functionblock: {
				rx:0,
				ry:0,
	            refY: '25',
				refWidth: '100%',
	            height: '100%'+25,
	            strokeWidth: 2,
	            stroke: '#000000',				
	            fill: '#8ECAE6',
			//	magnet: false
	        }		*/	
	    },
		ports: {
			items: [{}],
			groups: {
				input: {
					position: { name: 'left' },
					attrs: {
						circle: {
							fill: 'transparent',
							stroke: '#31d0c6',
							strokeWidth: 1,
							r: 6,
							cx: -6,
							cy: 25,
							magnet: "passive",
							function:'',
							port: ''
						},
						rect: {
							fill: '#31d0c6'
						},
						FunctionInputSource:{
							x:-10,
							y:0,
							textAnchor: 'end',
						},
						FunctionInputName: {
							textAnchor: 'begin',
							x: 15,
							y: 0,
							fill: 'black'
						}  
					},					
					label: {
						position: {
							name: 'left',
							args: {
								y: 30,	
								x: -5,
								attrs: {
									portLabel: { textAnchor: 'begin' }
								}
							}
						},
						markup: [
							
							{
							tagName: 'rect',
							selector: 'FunctioninputRect',
							groupSelector: 'FunctioninputRects'
						}, {
							tagName: 'text',
							selector: 'FunctionInputSource',
							groupSelector: 'FunctionInputSources'
						},{
							tagName: 'text',
							selector: 'FunctionInputName',
							groupSelector: 'FunctionInputNames'
						}]
					}  
				},
				output: {
					position: { name: 'right'},
					attrs: {
						circle: {
							fill: 'transparent',
							stroke: '#31d0c6',
							strokeWidth: 1,
							r: 6,
							cx:0,
							cy: 25,
							magnet: true,
							cursor: 'pointer', // show pointer cursor when over the port
							perimeter: {
								name: 'ellipse',
								args: {
									rx: 10,
									ry: 10
								}
							}
						},
						rect: {
							fill: '#31d0c6'
						},
						FunctionOutputDest:{
							textAnchor: 'begin',
						},
						FunctionOutputName: {
							x: -5,
							y: 0,
							textAnchor: 'end',
							fill: 'black'
						}
					},
					label: {
						position: {
							name: 'right',
							args: {
								y: 30,	
								x: -5,
								attrs: {
									portLabel: { textAnchor: 'end' }
								}
							}
						},
						markup: [
							
							{
							tagName: 'rect',
							selector: 'FunctionOutputRect',
							groupSelector: 'FunctionOutputRects'
						}, {
							tagName: 'text',
							selector: 'FunctionOutputDest',
							groupSelector: 'FunctionOutputDests'
						},{
							tagName: 'text',
							selector: 'FunctionOutputName',
							groupSelector: 'FunctionOutputNames'
						}]
					}					
					
				}
			}
		}
    }, {
        // inherit joint.shapes.standard.Rectangle.markup
		
	    markup: [
		{
	        tagName: 'rect',
	        selector: 'body'
	    }, {
	        tagName: 'rect',
	        selector: 'functionheader'
	    },{
	        tagName: 'text',
	        selector: 'functionname'
	    }],
		portMarkup: [{ tagName: 'circle', selector: 'portBody' }],
		MINIMUM_NUMBER_OF_PORTS: 0,
		getport(port){
			return this.getPorts().find(function(p) {
				return p.id === port;
			});
		},
		getGroupPorts: function(group) {
			return this.getPorts().filter(function(port) {
				return port.group === group;
			});
		},
				
		getUsedInPorts: function() {
			var graph = this.graph;
			if (!graph) return [];
			var connectedLinks = graph.getConnectedLinks(this, { inbound: true });
			return connectedLinks.map(function(link) {
				return this.getPort(link.target().port);
			}, this);
		}
		
    }, {
        
    } 
	);

	joint.shapes.standard.Link.define('Function.Link', null, {

		hide: function() {
			this.set('hidden', true);
		},
	
		show: function() {
			this.set('hidden', false);
		},
	
		isVisible: function() {
			return !this.get('hidden');
		},
	
		resolveOrientation: function(element) {
	
			var source = this.get('source');
			var target = this.get('target');
			var result;
	
			if (source && source.id !== element.id) {
				result = {
					oppositeEnd: source,
					currentEnd: target
				};
			}
	
			if (target && target.id !== element.id) {
				result = {
					oppositeEnd: target,
					currentEnd: source
				};
			}
	
			return result;
		}
	});

const Function_DataType_List =["String", "Integer", "Float", "Bool", "DateTime", "Object"]
const Function_DataType_Color_List	=['#A0D8B3','#A2A378','#DBDFEA','#FFEAD2', '#FEFF86','#FFEBEB']
const Function_Source_Color_List = ['#82CD47', '#6DA9E4', '#F6BA6F', '#BFCCB5', '#FFEBEB']
const Function_Dest_Color_List = ['#F6BA6F', '#BFCCB5', '#82CD47']
const Function_Source_List =["Constant", "Previous function", "system Session", "User Session", "External"]
const Function_Dest_List=["", "Session", "External"]
const Function_Type_List =["ParameterMap", "Csharp Script", "Javascript", "Database Query", "StoreProcedure", "SubTranCode", "TableInsert", "TableUpdate", "TableDelete"]
const Function_Type_Color_List = ['#82CD47', '#6DA9E4', '#F6BA6F', '#BFCCB5', '#FFEBEB', '#FFEBEB', '#FFEBEB', '#FFEBEB', '#FFEBEB']

/*
const (
	InputMap FunctionType = iota
	Csharp
	Javascript
	Query
	StoreProcedure
	SubTranCode
	TableInsert
	TableUpdate
	TableDelete
)
*/
var ProcessFlow = (function(){
	'use strict';

	$.on = (element, event, selector, callback) => {
		if(!element || element == null)
			return;
		if (!callback) {
			callback = selector;
			$.bind(element, event, callback);
		} else {
			$.delegate(element, event, selector, callback);
		}
	};

	$.off = (element, event, handler) => {
		element.removeEventListener(event, handler);
	};

	$.bind = (element, event, callback) => {
		if(!element || element == null)
			return;
		event.split(/\s+/).forEach(function(event) {
			element.addEventListener(event, callback);
		});
	};

	$.delegate = (element, event, selector, callback) => {
		if(!element || element == null)
			return;
		element.addEventListener(event, function(e) {
			const delegatedTarget = e.target.closest(selector);
			if (delegatedTarget) {
				e.delegatedTarget = delegatedTarget;
				callback.call(this, e, delegatedTarget);
			}
		});
	};

	$.closest = (selector, element) => {
		if (!element) return null;

		if (element.matches(selector)) {
			return element;
		}

		return $.closest(selector, element.parentNode);
	};
	$.clearcontent = (element) => {
		element.empty();
	}
	$.attr = (element, attr, value) => {
		
		if(!element || element == null)
			return;
		
		if (!value && typeof attr === 'string') {
			return element.getAttribute(attr);
		}

		if (typeof attr === 'object') {
			for (let key in attr) {
				$.attr(element, key, attr[key]);
			}
			return;
		}

		element.setAttribute(attr, value);
	};

	
	
	class Block{
		constructor(flow, data,type){
			this.flow = flow;
			this.data = data;
			this.type = this.data.type;
			this.id = this.type=='START'? 'START': this.data.id;
			this.build_block();
		//	this.set_events();
			this.set_events();
		}

		build_block(){

			var rect, headeredRectangle
			switch (this.type.toUpperCase()) {

				case 'START':

					rect = new joint.shapes.standard.Rectangle(); 
					rect.position(this.data.x, this.data.y );
					rect.resize(this.data.width, this.data.height);
					rect.attr({
						'nodeid': 'START',
								body: {
									rx: 10,
									ry: 10,
									fill: 'blue'
								},
								label: {
									text: 'Start',
									fill: 'white'
								}  
							});
					rect.addTo(this.flow.Graph);
							
					this.node =  {
						id: 'START',
						shape: rect
					}
					break;

				case 'STARTNODE':

					rect = new joint.shapes.standard.Rectangle(); 
					rect.position(this.data.x, this.data.y );
					rect.resize(this.data.width, this.data.height);
					rect.attr({
						'nodeid': this.data.id,
								body: {
									rx: 10,
									ry: 10,
									fill: this.data.fillcolor
								},
								label: {
									text: 'Start',
									fill: 'white'
								}  
							});
					rect.addTo(this.flow.Graph);
							
					this.node =  {
						id: this.data.id,
						shape: rect
					}
					break;

				case 'OPERATION':

					headeredRectangle = new joint.shapes.ProcessFlow.OperationBlock();//  .standard.HeaderedRectangle();
					headeredRectangle.position(this.data.x, this.data.y);
					headeredRectangle.resize(this.data.width, this.data.height < 100? 100:this.data.height );
					headeredRectangle.attr('root/title', this.data.OprSequenceNo +' - ' +this.data.WorkCenter+ ' - '+this.data.Description);
					headeredRectangle.attr('nodeid', this.data.OprSequenceNo);
					if(this.data.fillcolor == null)
						headeredRectangle.attr('header/fill', 'Yellow');
					else
						headeredRectangle.attr('header/fill', this.data.fillcolor);
					headeredRectangle.attr('headerText/text', '\ue122 ' + this.data.OprSequenceNo);
					
					//headeredRectangle.attr('bodyText/text', this.data.WorkCenter); 
					headeredRectangle.attr('workcenterText/text',  '\ue357 '+ this.data.WorkCenter); 
					
					if(this.data.fillcolor == null)
						headeredRectangle.attr('workcenter/fill', 'Yellow');
					else
						headeredRectangle.attr('workcenter/fill', this.data.fillcolor);
					
					headeredRectangle.attr('descriptionText/text', this.data.Description); 
					headeredRectangle.attr('featuresText/text', (this.data.Elements==null? '': this.data.Elements));
						
					if(this.data.desfillcolor == null)
						headeredRectangle.attr('descriptionbox/fill', "none"); 
					else
						headeredRectangle.attr('descriptionbox/fill', this.data.desfillcolor); 

					if(this.data.desfontsize == null)
						headeredRectangle.attr('descriptionText/fontSize', '12')
					else
						headeredRectangle.attr('descriptionText/fontSize', this.data.desfontsize)
					


					if(this.data.featurefillcolor == null)
						headeredRectangle.attr('featuresbox/fill', "none"); 
					else
						headeredRectangle.attr('featuresbox/fill', this.data.featurefillcolor);
					
					if(this.data.featurefontsize == null)
						headeredRectangle.attr('featuresText/fontSize', '12')
					else
						headeredRectangle.attr('featuresText/fontSize', this.data.featurefontsize)
					
					headeredRectangle.addTo(this.flow.Graph); 
					
					this.node = {
						id: this.data.OprSequenceNo,
						shape: headeredRectangle
					}
					break;
				case 'STEP':
				case 'FUNCGROUP':

					headeredRectangle = new joint.shapes.ProcessFlow.StepBlock() //standard.HeaderedRectangle();
					headeredRectangle.position(this.data.x, this.data.y);
					headeredRectangle.resize(this.data.width, this.data.height);
					headeredRectangle.attr('root/title', this.data.name);
					headeredRectangle.attr('nodeid', this.data.name);
					headeredRectangle.attr('header/fill', '#8dcfec');
					headeredRectangle.attr('headerText/text', this.data.name);
					
					if(!this.data.routing){
						headeredRectangle.attr('routerflag/strokeWidth', 0);
						
					}
					else{
						headeredRectangle.attr('routerflag/fill', '#f7d7dc');
					}					
					
					headeredRectangle.attr('stepnameText/text', this.data.description);
					headeredRectangle.attr('stepname/fill', '#8dcfec');
					headeredRectangle.attr('features/fill', '#8dcfec');
					headeredRectangle.attr('featuresText/text', (this.data.Elements==null? '': this.data.Elements)); 
					headeredRectangle.addTo(this.flow.Graph);
					
					this.node = {
						id: this.data.id,
						shape: headeredRectangle
					}
					break;	

				case 'FUNCTION':	
				//	console.log(this.data.Inputs, this.data.Outputs )
					let ports =[];

					for(var i=0;i<this.data.inputs.length;i++){
						let y= 25 + i*20;
						ports.push({
							group: 'input',
							id: this.data.inputs[i].id,
							args: {x: 0, y: y},
							attrs: { 
								circle: { 									
										fill: Function_DataType_Color_List[this.data.inputs[i].datatype],
										functionid:this.data.id,
										port: this.data.inputs[i].id,
										portname: this.data.inputs[i].name,   
									},
									FunctionInputName:{
										port:this.data.inputs[i].id,
										text: this.data.inputs[i].name,
									//	y: 0,	
									//	x: -10,
									//	fill: Function_DataType_Color_List[this.data.Inputs[i].datatype],
									//	source: Function_Source_List[this.data.inputs[i].source] + ' / '+ this.data.inputs[i].aliasname
									},
									rect:{
										width: this.data.inputs[i].source == undefined || this.data.inputs[i].source == "1"? 0: 100,
										height: (this.data.inputs[i].source == undefined || this.data.inputs[i].source == "1")? 0: 20,
										x:this.data.inputs[i].source == undefined || this.data.inputs[i].source == "1"? 0:-110,
										y:-15,									
										fill: this.data.inputs[i].source == undefined || this.data.inputs[i].source == "1"? 'none':Function_Source_Color_List[this.data.inputs[i].source],
										
									},
									FunctionInputSource:{
										text: this.data.inputs[i].source == undefined || this.data.inputs[i].source == "0"? (this.data.inputs[i].value==undefined? '': this.data.inputs[i].value) : this.data.inputs[i].source == "1"? '': this.data.inputs[i].aliasname

									}
								}
						});												
					} 
					
					for(var i=0;i<this.data.outputs.length;i++){
						
						let y= 25 + i*20;
						let x = this.data.width +6;
						ports.push({
							group: 'output',
							position:{name: "right"},
							id: this.data.outputs[i].id,
							args: {x: x, y: y},
							attrs: { 
								circle: { 									
									fill: Function_DataType_Color_List[this.data.outputs[i].datatype],
									functionid:this.data.id,
									port: this.data.outputs[i].id,
									portname: this.data.outputs[i].name, 
								},
								FunctionOutputName:{
									port:this.data.outputs[i].id,
									text: this.data.outputs[i].name
								//	y: 0,	
								//	x: 20,						
								//	fill: Function_DataType_Color_List[this.data.Outputs[i].datatype]
								},
								rect:{
									width: this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0"? 0: 150,
									height: (this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0")? 0: 20,
									x:this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0"? 0:10,
									y:-15,									
									fill: this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0"? 'none':Function_Dest_Color_List[this.data.outputs[i].outputdest],
									//text: this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0"? '': Function_Dest_Color_List[this.data.outputs[i].outputdest] + ' / '+ this.data.outputs[i].aliasname

								},
								FunctionOutputDest:{
									x:this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0"? 0:15,
									y:0,
									text: this.data.outputs[i].outputdest == undefined || this.data.outputs[i].outputdest == "0"? '': this.data.outputs[i].aliasname
								}
							}
						});
					} 					
					headeredRectangle = new joint.shapes.ProcessFlow.StepBlock.Function({
						ports: {
							items: ports
						}
					}) 
				//	headeredRectangle = new joint.shapes.ProcessFlow.StepBlock.Function()
					headeredRectangle.position(this.data.x, this.data.y);
					headeredRectangle.resize(this.data.width, this.data.height);
				//	headeredRectangle.attr('root/title', this.data.FunctionName);
					headeredRectangle.attr('nodeid', this.data.id);
					headeredRectangle.attr('functionheader/fill', Function_Type_Color_List[this.data.functype]);
					headeredRectangle.attr('functionheader/functionname', this.data.functionName);
					headeredRectangle.attr('functionname/text', this.data.functionName);					
					headeredRectangle.addTo(this.flow.Graph);
				//	headeredRectangle.addPorts(ports);
				//	console.log(headeredRectangle.getGroupPorts("input"))

					this.node = {
						id: this.data.id,
						shape: headeredRectangle
					}  
					break;
			}
		}
		
		update(data, subtype=''){
			let that = this;
			
			switch(that.type.toUpperCase()){

				case "FUNCTION":
					// update block self
					if(subtype == ''){
						that.data = Object.assign(that.data,data);
						data = that.data;
						that.node.shape.attr('functionheader/functionname', that.data.functionName);
						that.node.shape.attr('functionname/text', that.data.functionName);
						that.node.shape.attr('functionheader/fill', Function_Type_Color_List[that.data.functype]);
					//	that.node.shape.attr('nodeid', data.id);
						that.node.shape.resize(that.data.width, that.data.height);
						that.node.shape.position(that.data.x, that.data.y);						
					}else if(subtype.toUpperCase() == 'INPUT'){
						// update input
						
						for(var i=0;i<that.data.inputs.length;i++){
							if(that.data.inputs[i].id == data.id){
								that.data.inputs[i] = Object.assign(that.data.inputs[i],data);
								that.remove_events();
								that.node.shape.remove();
								that.build_block();
								that.set_events();
								break;
							}
						}					
					}
					else if(subtype.toUpperCase() == 'OUTPUT'){
						// update output
						for(var i=0;i<that.data.outputs.length;i++){
							if(that.data.outputs[i].id == data.id){
								that.data.outputs[i] = Object.assign(that.data.outputs[i],data);
								that.remove_events();
								that.node.shape.remove();
								that.build_block();
								that.set_events();
								break;
							}
						}					
					}

					for(var i=0;i<that.flow.nodes.length;i++){
						if(that.flow.nodes[i].id == that.data.id){
							that.flow.nodes[i] = that.data;
							break;
						}
					}

					for(var i=0;i<that.flow.flowobj.functiongroups.length;i++){
						if(that.flow.flowobj.functiongroups[i].name == that.flow.funcgroupname){
							for(var j=0;j<that.flow.flowobj.functiongroups[i].functions.length;j++){
								if(that.flow.flowobj.functiongroups[i].functions[j].id == that.data.id){
									that.flow.flowobj.functiongroups[i].functions[j] = that.data;
									break;
								}
							}
							break;
						}
					}

					break;
				case "FUNCGROUP":
					that.data = Object.assign(that.data,data);
					that.node.shape.attr('funcgroupheader/funcgroupname', that.data.name);
					that.node.shape.attr('funcgroupname/text', that.data.description);
				//	that.node.shape.attr('funcgroupheader/fill', Function_Group_Color_List[that.data.type]);
				//	that.node.shape.attr('nodeid', data.id);
				//	that.node.shape.resize(data.width, data.height);
					that.node.shape.position(that.data.x, that.data.y);
					for(var i=0;i<that.flow.nodes.length;i++){
						if(that.flow.nodes[i].id == that.data.id){
							that.flow.nodes[i] = that.data;
							break;
						}
					}

					for(var i=0;i<that.flow.flowobj.functiongroups.length;i++){
						if(that.flow.flowobj.functiongroups[i].id == that.data.id){						
							that.flow.flowobj.functiongroups[i] = that.data;
							break;
						}
					}  
					break;
				case "START":
					that.data = Object.assign(that.data,data);
					that.flow.flowobj.startnode = that.data;
					break;
			}
		
		}

		
		delete(){
			let that = this;
			if(this.node){
				this.node.shape.remove();
				this.node = null;
			}
			let index = -1;
			for(var i=0;i<that.flow.nodes.length;i++){
				if(that.flow.nodes[i].id == that.data.id){
					index = i;
					break;
				}
			}
			if(index >=0)
				that.flow.nodes.splice(index,1);
			
			if(that.type.toUpperCase() == "FUNCTION"){
				for(var i=0;i<that.flow.flowobj.functiongroups.length;i++){
					if(that.flow.flowobj.functiongroups[i].name == that.flow.funcgroupname){
						let index = -1;
						for(var j=0;j<that.flow.flowobj.functiongroups[i].functions.length;j++){
							if(that.flow.flowobj.functiongroups[i].functions[j].id == that.data.id){
								index = j;
								break;
							}
						}
						if(index >=0)
							that.flow.flowobj.functiongroups[i].functions.splice(index,1);
						break;
					}
				}
			}else if(that.type.toUpperCase() == "FUNCGROUP"){
				let index = -1;
				for(var i=0;i<that.flow.flowobj.functiongroups.length;i++){
					if(that.flow.flowobj.functiongroups[i].id == that.data.id){
						index = i;
						break;
					}
				}
				if(index >=0)
					that.flow.flowobj.functiongroups.splice(index,1);
			}
		}

		set_events(){
			if(!this.node)
				return;
			let that = this;
			
			this.node.shape.on('change:position', function(element, newPosition) {
				console.log('change the position',element, newPosition)
				let data = {
					x: newPosition.x,
					y: newPosition.y
				}
				that.update(data,'')
			});
			
		}
		remove_events(){
			if(!this.node)
				return;
			let that = this;
			
			this.node.shape.off('change:position', function(element, newPosition) {
			//	console.log('change the position',element, newPosition)
				let data = {
					x: newPosition.x,
					y: newPosition.y
				}
				that.update(data,'')
			});

		}
		
	}

	class MergePoint{
		constructor(flow,data){
			this.flow = flow;
			this.id = data.id;
			this.data = data;
			this.build();
		}
		
		
		build(){
			let mp = new joint.shapes.basic.MergePoint({
				size: { width: 10, height: 10 },
				attrs: { MergePoint: { width: 10, height: 10 } }
			});
			
			mp.addTo(this.flow.Graph);
						
			this.node =  {
				id: 'mp_'+this.id,
				shape: mp
			}
			
		}
	}
	class FunctionLink{
		constructor(flow, sourcenode, sourceport,destnode,destport, data = null){	
			this.flow = flow;
			this.data = data;
			this.sourcenodeid = sourcenode;
			this.sourceportid = sourceport;
			this.destnodeid = destnode;
			this.destportid = destport;

			this.build_link(sourcenode, sourceport,destnode,destport);
		}

		build_link(sourcenode, sourceport,destnode,destport){
		
			var _link = new joint.shapes.Function.Link({
				source: {id: sourcenode.shape.id,  port: sourcenode.shape.getport(sourceport).id},
				target: {id: destnode.shape.id, port: destnode.shape.getport(destport).id}
			  });
			
			this.flow.Graph.addCell(_link); 
			this._link = _link;
		}

		update(sourcenode, sourceport,destnode,destport){
			this._link.source({id: sourcenode.shape.id,  port: sourcenode.shape.getport(sourceport).id});
			this._link.target({id: destnode.shape.id, port: destnode.shape.getport(destport).id});		
		}

		delete(){
			this._link.remove();
			this.flow.functionlinklines.splice(this.flow.functionlinklines.indexOf(this),1);
			let index =-1
			for(var i=0;i<this.flow.functionlinks.length;i++){
				if(this.functionlinks[i].sourcefunctionid == this.sourcenodeid &&
					this.functionlinks[i].sourceoutputid == this.sourceportid &&
					this.functionlinks[i].targetfunctionid == this.destnodeid &&
					this.functionlinks[i].targetinputid == this.destportid){
						index = i;
						break;
					}
				}
			if(index>=0){
				this.flow.functionlinks.splice(index,1);
				
				let targetfunctionid = this.destnodeid;
				let targetinputid = this.destportid;

				for(var i=0;i<this.nodes.length;i++){
					if(this.nodes[i].id == targetfunctionid){
						for(var j=0;j<this.nodes[i].outputs.length;j++){
							if(this.nodes[i].inputs[j].id ==  targetinputid){
								this.nodes[i].inputs[j].source = 0;
								this.nodes[i].inputs[j].aliasname = ''
								break;
							}
						}
						break;
					}
				}
			//	console.log('update flowobj:', this.funcgroup,this.flowobj )
				for(var n=0;n<this.flowobj.functiongroups.length;n++){
					if(this.flowobj.functiongroups[n].name = this.flow.funcgroupname){
						//console.log('found function group:',n)
						for(var i=0;i<this.flowobj.functiongroups[n].functions.length;i++){
							if(this.flowobj.functiongroups[n].functions[i].id == targetfunctionid){
							//	console.log('found functions', i, this.flowobj.functiongroups[n].functions[i].name)
	
								for(var j=0;j<this.flowobj.functiongroups[n].functions[i].outputs.length;j++){
									if(this.flowobj.functiongroups[n].functions[i].inputs[j].id ==  targetinputid){
									//	console.log('found input', j, this.flowobj.functiongroups[n].functions[i].inputs[j].name)
	
										this.flowobj.functiongroups[n].functions[i].inputs[j].source =0;
										this.flowobj.functiongroups[n].functions[i].inputs[j].aliasname = ''
										break;
									}
								}
								break;	
							}
							
						}
						break;
					}
					
				}
			
			
			}
		}
	}
	class Link{
		constructor(flow, fromnode,tonode, data = null, mergepoint = null){	
			this.flow = flow;
			this.data = data;
			
			//console.log(this.data)
			if(!mergepoint)
				this.build_link(fromnode, tonode);
			else
			{
				this.build_link(fromnode, mergepoint);
				this.build_link(mergepoint, tonode);
			}
		//	this.make_link_tools();

		}
		build_link(fromnode,tonode){
				let _link = new joint.shapes.standard.Link();
				
							
				_link.source(fromnode.shape);
				_link.target(tonode.shape);
				
				//if(this.data){
				//	if(this.data.Lable)
						_link.appendLabel({
							attrs: {
								text: {
									text: this.data.Label ? this.data.Label : ''
								}
							}
						})
				//}
				_link.addTo(this.flow.Graph);	
				this._link = _link;
		}
		
		make_link_tools(){
			let that =this;
			
		/*	var verticesTool = new joint.linkTools.Vertices();
			var segmentsTool = new joint.linkTools.Segments();
			var sourceArrowheadTool = new joint.linkTools.SourceArrowhead();
			var targetArrowheadTool = new joint.linkTools.TargetArrowhead();
			var sourceAnchorTool = new joint.linkTools.SourceAnchor();
			var targetAnchorTool = new joint.linkTools.TargetAnchor();
			var boundaryTool = new joint.linkTools.Boundary();  */
			var removeButton = new joint.linkTools.Remove();
			
			/*verticesTool, segmentsTool,
					sourceArrowheadTool, targetArrowheadTool,
					sourceAnchorTool, targetAnchorTool,
					boundaryTool, */
			
			var toolsView = new joint.dia.ToolsView({
				tools: [
					removeButton
				]
			});

			var linkView = this._link.findView(that.flow.Paper);
			linkView.addTools(toolsView);
		}
		
	}
	
	class Toolbar{
		constructor(flow, data){
			this.flow = flow;
			this.data = data;
			this.build_toolbar();
			this.set_event();
		}
		
		build_toolbar(){
			let toolbar =  document.createElement('div');
			toolbar.classList.add('uiflow_process_flow_toolbar_container_toolbar');
			this.flow.$toolbar_container.appendChild(toolbar);
			
			let icon = document.createElement('span');
		/*	icon.classList.add('wux-ui-3ds'); */
			icon.classList.add(this.data.type);
			$(icon).attr('draggable', 'true')
			toolbar.appendChild(icon); 
			
			let desc = document.createElement('span');
			desc.classList.add('uiflow_process_flow_toolbar_container_toolbar_desc')
			toolbar.appendChild(desc);
			$(desc).html(this.data.description);
			
			$(toolbar).attr('data-key', this.data.datakey);
			$(toolbar).attr('title', this.data.description);
			$(toolbar).attr('draggable', 'true')
			this.toolbar = toolbar;

		}
		
		set_event(){
			if(!this.toolbar)
				return;
			let that = this;
			
			const dragStart = event => {
				
				event.currentTarget.classList.add('dragging');
				event.dataTransfer.setData('tooldata', that.data);
				
				$("body").css("cursor","move");
				
				event.dataTransfer.effectAllowed = "move";
				
				that.flow.trigger_event('tool_dragstart', [event]); 
				
				that.flow.svgZoom.disablePan();
				
			};

			const dragEnd = event => {
				$("body").css("cursor","");
				
							
				let block  = that.flow.get_element_byPos(event.x, event.y);
				
				if(!block){
					let flowarea = document.getElementById(that.flow.wrapper)
					let rect = flowarea.getBoundingClientRect();

					let x = event.x;
					let y = event.y;
					if(rect.x < x &&  x < rect.right && rect.y < y && y < rect.bottom)
					{						
						block ={
							type: that.flow.options.flowtype
						}
					}
					
				}
			
				
				event.preventDefault();

				event.currentTarget.classList.remove('dragging');				
				
				if(block && that.data.category.toUpperCase().includes(block.type.toUpperCase())  )
					that.flow.trigger_event('tool_dragend', [event,that.data,block]); 
				else 
					Apr.UserMessages.showMessage({
						message:  'The target element does not support drop element!',
						severity: Apr.SeverityLevel.error,
						type: Apr.MessageType.nonModal,
						cssClass:null,
						icon:null
					})	
				that.flow.svgZoom.enablePan();
			};
					
			const drag = event => {
				event.preventDefault();
				event.currentTarget.style.cursor = 'copy';
				return false;
			};		

			
			this.toolbar.addEventListener('dragstart', dragStart);
			this.toolbar.addEventListener('drag', drag); 
			this.toolbar.addEventListener('dragend', dragEnd); 
			

			$.on(this.toolbar, 'click', e => {
			//	console.log(this,e)
				/*
				if(that.flow.options.flowtype.toUpperCase() == 'PROCESS' && that.data.datakey.toUpperCase() == 'OPERATION'){
						console.log(that.flow.options.flowtype, that.data.datakey )
						let block = new Block(that.flow,{
							OprSequenceNo: '',
							WorkCenter: '',
							Description: '',
							type: 'OPERATION',
							id: 0,
							x: 100,
							y:100,
							width: 200,
							height: 100
							
						},'OPERATION');
						
						console.log(block);
						
						
						return;			
				}
				else if(that.flow.options.flowtype.toUpperCase() == 'PROCESS' && that.data.datakey.toUpperCase() == 'STEP'){
					return;
					
				}
				else if(that.flow.options.flowtype.toUpperCase() == 'OPERATION' && that.data.datakey.toUpperCase() == 'STEP'){
					let block = new Block(that.flow,{
							SequenceNo: 0,
							Description: '',
							type: 'STEP',
							id: 0
							
						},'STEP');
					
						return;
				}
				else if(that.flow.options.flowtype.toUpperCase() == 'OPERATION' && that.data.datakey.toUpperCase() == 'OPERATION'){
					return;
					
				}
				*/
				if(that.data.datakey == 'Refresh'){
					that.flow.refresh();
				}
				else
					that.flow.trigger_event('tool_click', [that.data]); 
			})  
			
		}
		
	} 
	
	class MenuBar{
		constructor(flow, data){
			this.flow = flow;
			this.data = data;
			this.build_menubar();
			this.set_events();
		}
		
		build_menubar(){
			let menubar =  document.createElement('div');
			menubar.classList.add('uiflow_process_flow_menubar_container_menubar');
			
			
			let icon = document.createElement('span');
			icon.classList.add('uiflow_menubar_'+this.data.type);
		//	$(icon).attr('draggable', 'true')
			menubar.appendChild(icon); 
			
			let desc = document.createElement('span');
			desc.classList.add('uiflow_process_flow_menubar_container_menubar_desc')
			menubar.appendChild(desc);
			$(desc).html(this.data.description);
			
			$(menubar).attr('data-key', this.data.datakey);
			$(menubar).attr('title', this.data.description);
		//	$(menubar).attr('draggable', 'true')
			this.flow.menu_panel.appendChild(menubar);
			
			this.menubar = menubar;
		}
		set_events(){
			$.on(this.menubar, 'click', e => {
				this.flow.menu_click(this.data); 
			})
		}

	}

	class JSONWrapper {
		constructor(json, onChange) {
			this.data = json;
			this.onChange = onChange;
			return new Proxy(this, {
				get(target, property) {
				//console.log(target)
				  return target.data[property];
				},
				set(target, property, value) {
					console.log('json change')
				  target.data[property] = value;
				  target.onChange(property, value);
				  return true;
				},
				deleteProperty(target, property) {
				  delete target.data[property];
				  target.onChange(property, undefined);
				  return true;
				}
			  });
		  }
	  }

	class ProcessFlow{
		constructor(wrapper,flowobj, options, funcgroup){
			this.flowobjchange = false;
			let that = this;
			
			this.flowobj  = flowobj; /*new JSONWrapper(flowobj, (property, value) => {
				console.log(`Property ${property} has been updated with value ${value}`);
				that.flowobjchange = true;
			  }); */
			
			this.setup_wrapper(wrapper);

			this.setup_objects(options, funcgroup);	
				
			
		}

		setup_objects(options, funcgroup){
			
			this.funcgroupname = funcgroup;

			this.setup_options(options);			
			this.flowtype = this.options.flowtype

			if($('#'+this.sectionwrapper).width() > 800)
				this.options.width = $('#'+this.sectionwrapper).width();
			if($('#'+this.sectionwrapper).height() > 600)
				this.options.height = $('#'+this.sectionwrapper).height();


			if(this.options.flowtype == 'FUNCGROUP')
				this.setup_paper_fg();
			else 
				this.setup_Paper();

			this.setup_Toolbar();			

			this.setup_Menubar();

			let obj = {};

			if(this.options.flowtype == 'FUNCGROUP')
			{
			//	console.log(this.flowobj, funcgroup)
				let fgobj ={};
				if(funcgroup == "" || !funcgroup){
					fgobj = this.flowobj.functiongroups[0]
					this.funcgroupname = fgobj.name;
				}
				else{
					fgobj = this.flowobj.functiongroups.find(fg=>fg.name == funcgroup)

				}
			//	console.log(fgobj)
				obj = this.get_process_Object(fgobj)				
			}
			else
				obj = this.get_process_Object(this.flowobj)

			this.setup_nodes(obj.nodes);
			
			this.setup_mergegroup(obj.mergegroups);
			//
			this.setup_functionlinks(obj.functionlinks)
			
			this.setup_links(obj.links);	
	
			this.initial_const();
			
			this.render();	
		}
		
		setup_options(options){
			this.toolbars = [];
			const default_options = {
				gridsize: 10,
				drawgrid: true,
				width: 1400,
				height: 1000,
				backgroundcolor: 'white',
				interactive: true,
				nodewidth: 200,
				nodeheight: 100,
				colspace: 80,
				rowspace:50,
				colmargin:20,
				rowmargin: 20,
				rankdir: 'TB',
				align: "",
				marginx: 30,
				marginy: 30,
				nodesep: 50,
				ranksep: 50,
				edgesep: 30,
				ranker: "longest-path",
				showtoolbar: true,
				flowtype: 'FUNCGROUP',
				skipstartnode: false,
				showlinkmergepoint: true
			};
			
			this.options = Object.assign({}, default_options, options);

			
			if(this.options.flowtype == 'FUNCGROUP')
				this.options.skipstartnode= true
			else
				this.options.skipstartnode= false  
		
		}	
		
		setup_wrapper(wrapper){
			let section = document.getElementById(wrapper)			
			section.style.display = "flex";
			section.style.flexDirection = "row"
			section.style.flexWrap = "nowrap"
			section.style.width = "100%"
			section.style.height = "100%"
			this.sectionwrapper = wrapper

			let attrs={
				'class':'processflow_container uiflow_process_flow_menubar_container',
				'id':this.wrapper+'_flow_menu_panel'
			}
			this.menu_panel = (new UI.FormControl(section, 'div', attrs)).control;

			this.wrapper = wrapper +'_flow_container'
			attrs={
				'class':'processflow_container',
				'id':this.wrapper,
				'style':'width:100%;height:100%;display:flex'
			}
			this.wrappercontainer = (new UI.FormControl(section, 'div', attrs)).control;
			
			attrs={
				'class':'processflow_container',
				'id':wrapper+'_flow_property_panel',
				'style':'width:0px;height:100%;float:right;position:absolute;top:0px;right:0px;background-color:lightgrey;overflow:auto;' +
								'border-left:2px solid #ccc;resize:horizontal;z-index:9'
			}
			this.property_panel = (new UI.FormControl(section, 'div', attrs)).control;

			attrs={
				'class':'processflow_items_panel',
				'id':wrapper+'_flow_items_panel',
				'style':'width:0px;height:100%;float:left;position:absolute;top:0px;left:0px;background-color:lightgrey;overflow:auto;' +
								'border-left:2px solid #ccc;resize:horizontal;z-index:9'
			}
			this.item_panel = (new UI.FormControl(section, 'div', attrs)).control;
			
		}

		get_process_Object(flowobj){
			console.log('get_process_Object',flowobj)
			if(!flowobj.hasOwnProperty('uuid') || flowobj.uuid == "")
				flowobj.uuid = UIFlow.generateUUID();

			let nodes =[];
			let links =[];
			let mergegroups =[];
			let functionlinks =[];
			/*	
					node:
						{
						id: outputs.ProcessOperationStepIDList[i],
						OprSequenceNo: PPR_ProcessFlow.Context.inputs.OprSequenceNo,
						SequenceNo: outputs.SequenceNoList[i],
						StepName: outputs.StepNameList[i],
						Description: outputs.DescriptionList[i],
						Elements: elements,
						type: "STEP"
					}
					links:
					{
						fromnode: outputs.SourceStepIDList[i],
						tonode: outputs.DestinationStepIDList[i],
						wipcontentid: 0,
						reasoncode:'',
						Label: ''
					}

					*/


			switch (this.options.flowtype.toUpperCase()) {
				case 'PROCESS':
					nodes = flowobj.Operations;
					links = flowobj.OperationLinks;
					mergegroups = flowobj.MergeGroups;
					break;
				case 'OPERATION':
				case 'TRANCODE':
					// build the nodes
					let firstnodeid = "";
					let index = 0;
					if(flowobj.functiongroups)
						flowobj.functiongroups.forEach(functiongroup => {
							let routerdef = functiongroup.RouterDef;
							let routing = false;
							if(routerdef){
								if(routerdef.value.length > 0){
									routing = true;
								}
							}
							else
								routerdef = {
									variable:'',
									value:[],
									destinations:[],

								};
							
							let nodeid = ""
							if(!functiongroup.id){
								nodeid = UIFlow.generateUUID();
								flowobj.functiongroups[index].id = nodeid;
							}
							else
								nodeid = functiongroup.id;


							if(functiongroup.name == flowobj.firstfuncgroup)
								firstnodeid = nodeid;


						//	console.log("position:",functiongroup.x,functiongroup.y)
							let node = {
								id: nodeid,
								name: functiongroup.name,
								functiongroupname:functiongroup.functiongroupname,
								description: functiongroup.description,
								routerdef: routerdef,
								elements: [],
								x:functiongroup.x,
								y:functiongroup.y,
								routing:routing,
								type: "FUNCGROUP"
							};
						//	console.log(node)
							nodes = nodes.concat(node);
							index = index + 1;
						});
					else 
						this.flowobj.functiongroups = [];
					// build the links
					
					let link = {
						fromnode:"START",
						tonode: firstnodeid,
						wipcontentid: 0,
						reasoncode:'',
						Label: ''
					};

					links = links.concat(link);

					if(flowobj.functiongroups){
						flowobj.functiongroups.forEach(functiongroup => {
							let routerdef = functiongroup.routerdef;
							console.log(routerdef)
							if(routerdef){
								let variable = routerdef.variable
								let values = routerdef.value;
								let nextfuncgroups = routerdef.nextfuncgroups;
								let defaultfuncgroup = routerdef.defaultfuncgroup;
								console.log(nextfuncgroups,defaultfuncgroup)
								if(Array.isArray(nextfuncgroups)){
									nextfuncgroups.forEach(nextfuncgroup => {
										let link = {
											fromnode:this.get_itemidbyname(nodes,functiongroup.name),
											tonode: this.get_itemidbyname(nodes,nextfuncgroup),
											wipcontentid: 0,
											reasoncode: "",
											Label: variable + '=' + values[nextfuncgroups.indexOf(nextfuncgroup)]
										};
										links = links.concat(link);
									
										
									});
								}
								if (defaultfuncgroup != "") {
									let link = {
										fromnode:this.get_itemidbyname(nodes,functiongroup.name),
										tonode: this.get_itemidbyname(nodes,defaultfuncgroup),
										wipcontentid: 0,
										reasoncode: "",
										Label: variable==''? 'default' : (variable + '=default')

									};
									links = links.concat(link);
								}
							}

						});
					}
					//console.log(links)
					break;

				case 'FUNCGROUP':
					if(!flowobj.functions){
						for(var i=0;i<this.flowobj.functiongroups.length;i++){
							if(this.flowobj.functiongroups[i].name == this.funcgroupname){
								this.flowobj.functiongroups[i].functions = [];
								flowobj.functions = [];
							}
						}						
					}
					let findex = 0;

					flowobj.functions.forEach(functionobj => {
						let inputs =[];
						let outputs =[];
						let inindex = 0;

						functionobj.inputs.forEach(input => {

							let nodeid = "";

							if(!input.id){
								nodeid = UIFlow.generateUUID();
								flowobj.functions[findex].inputs[inindex].id = nodeid;
							}
							else
								nodeid = input.id;

							let inputobj = {
								id: nodeid,
								name: input.name,
								datatype: input.datatype,
								description: input.description,
								source:	input.source,
								aliasname: input.aliasname,
								defaultvalue: input.defaultvalue
							}
							inputs = inputs.concat(inputobj);
							if(input.source == "1"){
								let arr = input.aliasname.split(".");
								if(arr.length ==2)
								{
									functionlinks.push(
										{
											sourcefunction: arr[0],
											sourceoutput: arr[1],
											targetfunction: functionobj.name,
											targetinput: input.name
										}
									)	
								}
							}

							inindex = inindex + 1;
						});
						let outindex = 0;
						functionobj.outputs.forEach(output => {
							let nodeid = "";
							if(!output.id){
								nodeid = UIFlow.generateUUID();
								flowobj.functions[findex].outputs[outindex].id = nodeid;
							}
							else
								nodeid = output.id;

							let outputobj = {
								id: nodeid,
								name: output.name,
								datatype: output.datatype,
								description: output.description,
								outputdest:	output.outputdest,
								aliasname: output.aliasname,
								defaultvalue: output.defaultvalue
							}
							outputs = outputs.concat(outputobj);
							outindex = outindex + 1;
						});

						let nodeid = "";
						if(!functionobj.id){
							nodeid = UIFlow.generateUUID();
							flowobj.functions[findex].id = nodeid;
						}
						else
							nodeid = functionobj.id;

						let node = {
							id: nodeid,
							name: functionobj.name,	
							functionName: functionobj.name,
							description: functionobj.description,
							content: functionobj.content,
							functype: functionobj.functype,
							inputs: inputs,
							outputs: outputs,
							type: "FUNCTION",
							x: functionobj.x,
							y: functionobj.y
						};
						nodes = nodes.concat(node);
						findex = findex + 1;
					});						

					break;

				default:
					nodes = flowobj.Operations;
					links = flowobj.OperationLinks;
					functionlinks = [];
					mergegroups = flowobj.MergeGroups;
					break;
			}
		//	console.log('get object:', nodes)
			return {
				nodes: nodes,
				links: links,
				functionlinks: functionlinks,
				mergegroups: mergegroups,
			}
		}
		setup_paper_fg(){
			let that = this;
			this.Graph = new joint.dia.Graph;

			var magnetAvailabilityHighlighter = {
				name: 'stroke',
				options: {
					padding: 6,
					attrs: {
						'stroke-width': 3,
						'stroke': 'red'
					}
				}
			};

		    this.Paper = new joint.dia.Paper({
				el: this.wrappercontainer, // document.getElementById(wrapper),
				model: this.Graph,
				marginx: this.marginx,
				marginy:this.marginy,
				width: (this.options.width-45),
				height: this.options.height,
				gridSize: this.options.gridsize,
				drawGrid: this.options.drawgrid,
				interactive: this.options.interactive,
				addLinkFromMagnet: true,
			//	elementView: ClickableView,
				magnetThreshold: 'onleave',
				background: {
					color: this.options.backgroundcolor
				},
				defaultConnectionPoint: { name: 'boundary' },
				defaultLink: new joint.shapes.Function.Link({ z: - 1 }),
			//	defaultConnector: { name: 'line' },
			//	defaultConnectionPoint: { name: 'port' },
				markAvailable: true,
				validateConnection: function(cellViewS, magnetS, cellViewT, magnetT, end, linkView) {
					// Prevent loop linking
				//	console.log('validate connection:',cellViewS, magnetS, cellViewT, magnetT, end, linkView)
					
					let bvalidate = that.validate_functionlink(cellViewS, magnetS, cellViewT, magnetT);
					/*if (!bvalidate){
						console.log($(linkView.el))
						$(linkView.el).remove();
					} */
						
					return bvalidate; //(magnetS !== magnetT); 
					
				},
				// Enable link snapping within 20px lookup radius
				snapLinks: { radius: 20 },
			//	markAvailable: true,
			//	snapLinks: { radius: 40 },
			/*	defaultRouter: {
					name: 'mapping',
					args: { padding: 30 }
				},  
				defaultConnectionPoint: { name: 'anchor' },  */
			//	defaultAnchor: { name: 'mapping' },
				/*defaultConnector: {
					name: 'jumpover',
					args: { jump: 'cubic' }
				}, */
				highlighting: {
					'magnetAvailability': {
						name: 'stroke',
						options: {
							padding: 0,
							attrs: {
								'stroke-width': 1,
								'stroke': 'red'
							}
						}
					},
					'elementAvailability': {
						name: 'stroke',
						options: {
							padding: 0,
							attrs: {
								'stroke-width': 1,
								'stroke': '#ED6A5A'
							}
						}
					}
				}
			});		
			
		//	this.Paper.options.highlighting.magnetAvailability = magnetAvailabilityHighlighter;
		}
		setup_Paper(){
		/*	var ClickableView = joint.dia.ElementView.extend({
				pointerdown: function () {
					this._click = true;
					joint.dia.ElementView.prototype.pointerdown.apply(this, arguments);
				},
				pointermove: function () {
					this._click = false;
					joint.dia.ElementView.prototype.pointermove.apply(this, arguments);
				},
				pointerup: function (evt, x, y) {
					if (this._click) {
						this.notify('cell:click', evt, x, y);
					} else {
						joint.dia.ElementView.prototype.pointerup.apply(this, arguments);
					}
				}
			});  */
			
		    this.Graph = new joint.dia.Graph;

		    this.Paper = new joint.dia.Paper({
				el: this.wrappercontainer, //document.getElementById(wrapper),
				model: this.Graph,
				marginx: this.marginx,
				marginy:this.marginy,
				width: this.options.width,
				height: this.options.height,
				gridSize: this.options.gridsize,
				drawGrid: this.options.drawgrid,
				interactive: this.options.interactive,
				addLinkFromMagnet: true,
			//	elementView: ClickableView,
				magnetThreshold: 'onleave',
				background: {
					color: this.options.backgroundcolor
				},
				markAvailable: true,
				snapLinks: { radius: 40 },
			/*	defaultRouter: {
					name: 'mapping',
					args: { padding: 30 }
				},  
				defaultConnectionPoint: { name: 'anchor' },  */
			//	defaultAnchor: { name: 'mapping' },
				defaultConnector: {
					name: 'jumpover',
					args: { jump: 'cubic' }
				},
				highlighting: {
					magnetAvailability: {
						name: 'addClass',
						options: {
							className: 'record-item-available'
						}
					},
					connecting: {
						name: 'stroke',
						options: {
							padding: 8,
							attrs: {
								'stroke': 'none',
								'fill': '#7c68fc',
								'fill-opacity': 0.2
							}
						}
					}
				}
			});

		}
		
		setup_Menubar(){
			let menubars=[];
			menubars.push({
				type: 'Repository',
				datakey: 'Repository',
				description: 'Repository',
				category: 'trancode'
			})
			menubars.push({
				type: 'New',
				datakey: 'New',
				description: 'New',
				category: 'trancode'
			})
			menubars.push({
				type: 'Save',
				datakey: 'Save',
				description: 'Save',
				category: 'trancode'
			})
			menubars.push({
				type: 'Saveas',
				datakey: 'Saveas',
				description: 'Save as',
				category: 'trancode'
			})
			menubars.push({
				type: 'Load',
				datakey: 'Load',
				description: 'Load',
				category: 'trancode'
			})
			menubars.push({
				type: 'Export',
				datakey: 'Export',
				description: 'Export',
				category: 'trancode'
			})
			menubars.push({
				type: 'Import',
				datakey: 'Import',
				description: 'Import',
				category: 'trancode'
			})
			menubars.push({
				type: 'Sessions',
				datakey: 'Sessions',
				description: 'Sessions',
				category: 'trancode'
			})
			menubars.push({
				type: 'Parameters',
				datakey: 'Parameters',
				description: 'Parameters',
				category: 'trancode'
			})

			this.Menubars = menubars;
		}
		setup_Toolbar(){
			if(!this.options.showtoolbar)
				return;
			
			let that = this;
			
			let parentcontainer = $('#'+this.wrapper).parent()[0];

			
			this.$toolbar_container = document.createElement('div');
			this.$toolbar_container.classList.add('uiflow_process_flow_toolbar_container');	
			this.$toolbar_container.classList.add('dragscroll');

			parentcontainer.appendChild(this.$toolbar_container);
			
			let toolbars = [];

			toolbars.push({
				type: 'Refresh',
				datakey: 'Refresh',
				description: 'Refresh the flow',
				category: '',
				shows: 'Process,Operation'
			})
			
			toolbars.push({
				type: 'Operation',
				datakey: 'Operation',
				description: 'Process Operation',
				category: 'Process',
				shows: 'Process'
			})

			toolbars.push({
				type: 'Step',
				datakey: 'Step',
				description: 'Operation Step',
				category: 'Operation',
				shows: 'Operation'
			})

			toolbars.push({
				type: 'WorkCenter',
				datakey: 'WorkCenter',
				description: 'Work Center',
				category: 'Operation',
				shows: 'Process,Operation'
			}) 	

			toolbars.push({
				type: 'Product',
				datakey: 'Product',
				description: 'Process Product',
				category: 'Process',
				shows: 'Process,Operation'
				
			})

			toolbars.push({
				type: 'BOM',
				datakey: 'BOM',
				description: 'Product BOM',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			})

			toolbars.push({
				type: 'Alert',
				datakey: 'Alert',
				description: 'Alert / Notice',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'
			}) 
			
			toolbars.push({
				type: 'Component',
				datakey: 'Component',
				description: 'Component',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'CheckList',
				datakey: 'CheckList',
				description: 'Check List',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'
			}) 
			
			toolbars.push({
				type: 'Document',
				datakey: 'Document',
				description: 'Document',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'WorkInstruction',
				datakey: 'WorkInstruction',
				description: 'Work Instruction',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'Resource',
				datakey: 'Resource',
				description: 'Resource',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'ResourceClass',
				datakey: 'ResourceClass',
				description: 'Resource Class',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'Characteristic',
				datakey: 'Characteristic',
				description: 'Characteristic',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'Specification',
				datakey: 'Specification',
				description: 'Specification',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'DataCollectionPlan',
				datakey: 'DataCollectionPlan',
				description: 'Data Collection Plan',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'Skill',
				datakey: 'Skill',
				description: 'Skill',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'
			}) 

			toolbars.push({
				type: 'Role',
				datakey: 'Role',
				description: 'Role',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'				
			}) 
		
			toolbars.push({
				type: 'Employee',
				datakey: 'Employee',
				description: 'Employee',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'
			}) 
/*
			toolbars.push({
				type: 'EmployeeClass',
				datakey: 'EmployeeClass',
				description: 'Employee Class',
				category: 'Operation,Step',
				shows: 'Process,Operation,Step'
			}) */
			
			toolbars.push({
				type: 'Signature',
				datakey: 'Signature',
				description: 'Signature',
				category: 'Process,Operation,Step',
				shows: 'Process,Operation,Step'
			}) 
			
			this.toolbars = toolbars.map((toolbar,i) => {
				return toolbar;
			})
			
		}

		setup_startnode(){
			if(this.flowobj.startnode)
				return this.flowobj.startnode;
			else
				return {
					x: 100,
					y:100,
					width: this.options.nodewidth * 0.6,
					height: this.options.nodewidth * 0.2,
					type: "START"
				}
			
		}
		
		setup_functionlinks(functionlinks){
			let that = this;
			this.functionlinks = functionlinks.map((functionlink,i) => {

				return {
					type: "FUNCTIONLINK",
					sourcefunctionid: that.get_itemidbyname(that.nodes,functionlink.sourcefunction),
					sourceoutputid: that.get_itemidbyname(that.get_itembyname(that.nodes,functionlink.sourcefunction).outputs,functionlink.sourceoutput),
					targetfunctionid: that.get_itemidbyname(that.nodes,functionlink.targetfunction),
					targetinputid: that.get_itemidbyname(that.get_itembyname(that.nodes,functionlink.targetfunction).inputs,functionlink.targetinput)
				}	

			})

		}

		setup_nodes(nodes){		
			console.log("setup_nodes",nodes)
			let tempnodes = nodes.map((node,i) =>{
				console.log(node,i)
				if(!node.x)
					node.x = 100;
				
				if(!node.y)
					node.y = (this.options.nodeheight +20)* (i + 1)
				
				if(!node.width)
					node.width = this.options.nodewidth;
				
				if(!node.height)
					node.height = this.options.nodeheight;
				
				if(!node.type)
					node.type = "OPERATION"
			//	console.log(node)
				return node;		
			})
			
			this.nodes = [];
			
			if(!this.options.skipstartnode && this.options.flowtype !='FUNCGROUP')
				this.nodes.push(this.setup_startnode());
			
			this.nodes = this.nodes.concat(tempnodes);
	
		}

		setup_mergegroup(_mergegroups){
			this.mergegroups =[];
			this.mergegroups = _mergegroups.map((_mp,i) =>{
				return _mp;
			}
			)
		}
		
		setup_links(_links){
			this.links =[];
			this.links = _links.map((_link,i) => {
			//	console.log(_link);
				return _link;			
			});	
		
		}
		
		render(){
			
			this.Graph.clear();
			
			this.make_blocks();
			
			let flowarea = this.wrappercontainer ;// document.getElementById(this.wrapper)
			//let rect = flowarea.getBoundingClientRect();
			//console.log('window resize',flowarea,rect);
			
			
		//	console.log(this.nodes, this.blocks)
			this.make_mergepoint();
			
			this.make_functionlink();

			this.make_links();
			
			this.make_Toolbar();

			this.make_Menubar();
			//console.log(this.links, this.linklines)
		//	this.auto_layout();

			this.zoom();
			
			this.resize();
			
			this.make_link_tools();
			
			this.make_element_tools();
			
			this.create_events();
			
			$('html,body').css('cursor','pointer');
		}
		
		
		refresh(){
			
			this.Graph.clear();
			
			this.make_blocks();
			
			let flowarea = document.getElementById(this.wrapper)
		//	let rect = flowarea.getBoundingClientRect();
			//console.log('window resize',flowarea,rect);
			
			
		//	console.log(this.nodes, this.blocks)
			this.make_mergepoint();
			
			this.make_functionlink();

			this.make_links();
			
			this.make_Toolbar();

			this.make_Menubar();
			//console.log(this.links, this.linklines)
						
		//	this.auto_layout();

			this.zoom();
			
			this.resize();
			
			this.make_link_tools();
			
			this.make_element_tools();
			
		//	this.create_events();
			
			$('html,body').css('cursor','pointer');
			
		}
		
		reload(){
			this.destry();
			this.setup_objects(this.options, this.funcgroupname)
		}

		resize(){
						
			let rect = this.Paper.viewport.getBoundingClientRect();
			
		//	console.log(rect, this.options)
			
			if(rect.width > this.options.width )
				this.Paper.scaleContentToFit({ padding: 20 });
			
			
			rect = this.Paper.viewport.getBoundingClientRect();
			
			if(rect.height > this.options.width)
				$('#'+ this.wrapper).css('overflow-y','auto')  
		}
		
		zoom(){
			console.log(this.wrapper,this.container,$(this.container))
			this.svgZoom = svgPanZoom($(this.wrappercontainer).find('svg')[0], {
			  center: true,
			  zoomEnabled: true,
			  panEnabled: true,
			  controlIconsEnabled: true,
			  fit: false,
			  minZoom: 0.05,
			  maxZoom:50,
			  zoomScaleSensitivity: 0.1
			});
			
		}

		auto_layout(){
			joint.layout.DirectedGraph.layout(this.Graph, { 
				setLinkVertices: false, 
				nodeSep: this.options.nodeSep,
				edgeSep: this.options.edgeSep,
				rankDir: this.options.rankdir,
				align: this.options.align,
				marginX: this.options.marginx,
				marginY: this.options.marginy,
				ranker: this.options.ranker			
			}); 
		}

		
		destry(){
			window.removeEventListener('resize', this.windows_resize,false);
			this.blocks = [];
			this.nodes =[];
			this.links =[];
			this.functionlinklines=[];
			this.functionlinks =[];
			this.mergepoints =[];
			this.mergegroups =[];
			this.linklines =[];
			this.linktools =[];
			this.elementtools =[];
			this.toolbar = null;

			this.windows_resize = null;
			
			this.svgZoom = null;
		    this.Paper = null
			this.Graph = null;

		}

		make_blocks(){
			let that = this;
			that.blocks = [];
			this.nodes.forEach(function(node){
				that.blocks.push(new Block(that, node));			
			});
		}
		
		add_block(data){
			this.blocks.push(new Block(this, data));
		}
		
		make_mergepoint(){
			let that = this;
			that.mergepoints = [];
			this.mergegroups.forEach(function(data){
				that.mergepoints.push(new MergePoint(that, data));			
			});
		}
		
		add_mergepoint(data){
			this.mergepoints.push(new MergePoint(this, data));
		}	

		add_new_mergepoint(_link){
			let maxid = 0;
			this.mergegroups.forEach(function(data){
				if(data.id > maxid)
					maxid = data.id;
			})
			maxid +=1;
			
			let data = {
				id:maxid				
			}
			
			this.mergegroups.push(data)
			
		//	console.log(this.mergepoints)
			
			/*for(var i=0;i<this.linklines.length;i++){
				
				if(this.linklines[i] == _link){
					console.log(this.linklines[i])
					
					this.linklines[i].mergegroup = maxid;
					console.log(this.linklines[i])
					
				}				
			} */
			for(var i=0;i<this.links.length;i++){
				if(this.links[i] = _link.data){
					//console.log(this.links[i])
					this.links[i].mergegroup = maxid;					
					break;
				}				
			}
			
			this.refresh();
			return maxid; 			
		}	
		make_functionlink(){
			/*
			sourcefunction: arr[0],
			sourceoutput: arr[1],
			targetfunction: functionobj.name,
			targetinput: input.name
			
			*/
			let that = this;
			that.functionlinklines = [];

			if(this.functionlinks.length == 0)
				return;

			this.functionlinks.forEach(function(_link){
				//console.log(_link)
				let sourcenode = that.get_block(_link.sourcefunctionid).node;
				let destnode = that.get_block(_link.targetfunctionid).node;

				if(sourcenode && destnode){
					let link = new FunctionLink(that, sourcenode, _link.sourceoutputid,destnode, _link.targetinputid, {})
					that.functionlinklines.push(link);
				}
	
			})

		}
		validate_functionlink(sourceelement,sourceport, destelement, destport){

		//	console.log(sourceelement.el, $(sourceport).attr('port-group'), $(sourceport).attr('port'))
			if(this.options.flowtype !== 'FUNCGROUP')
				return false;

			if(!sourceelement || !sourceport || !destelement || !destport)
				return false;			
			
			if(sourceelement === destelement || sourceport === destport)
				return false;
			
			if($(sourceport).attr('port-group') !== 'output' || $(destport).attr('port-group') !== 'input')
				return false;

			let sourceid = $(sourceelement.el).attr("model-id")
			let targetid = $(destelement.el).attr("model-id")

			if(!sourceid || !targetid)
				return false;

			let sourcenode = this.get_block_byelementid(sourceid).data
			let destnode= this.get_block_byelementid(targetid).data


			if(!sourcenode || !destnode)
				return false;

			//let sourceouputid = $(sourceport).attr('port')
			//let destinputid = $(destport).attr('port')

			 let bloop = !(this.validate_looplink(sourcenode, destnode))
		     console.log(sourcenode, destnode, bloop)
			return bloop;
		}
		validate_looplink(sourcenode, destnode){
			/* to avoid the loop link between 2 functions*/
			for(var i=0;i<sourcenode.inputs.length;i++){
				if(sourcenode.inputs[i].source =="1" )
				{
					let aliasname = sourcenode.inputs[i].aliasname;
					let arr = aliasname.split('.');
					if(arr.length == 2){
						if(arr[0] == destnode.name){
							console.log(sourcenode.name, sourcenode.inputs[i].aliasname, destnode.name)
							return true;
						}
					}
				}
			}	
			return false;	
		}

		add_functionlink(sourcefunctionid, sourceoutputid, targetfunctionid, targetinputid){

			if(this.options.flowtype !== 'FUNCGROUP')
				return;
		//	console.log('add function link',sourcefunction, sourceoutput, targetfunction, targetinput)

			if(!sourcefunctionid || !sourceoutputid || !targetfunctionid || !targetinputid)
				return;
			/* add link to the functions */
			let sourcefunction = this.get_itembyid(this.nodes, sourcefunctionid);
			if(!sourcefunction)
				return;

			let sourceoutput = this.get_itemnamebyid(sourcefunction.outputs, sourceoutputid);

			let targetfunction = this.get_itembyid(this.nodes, targetfunctionid);
			if(!targetfunction)
				return;

			let targetinput = this.get_itemnamebyid(targetfunction.inputs, targetinputid);
			if(sourceoutput =="" || targetinput =="")
				return;
		//	console.log('add function link',sourcefunction, sourceoutput, targetfunction, targetinput)
		//	console.log('add function link')
			this.functionlinks.push({
				type: "FUNCTIONLINK",
				sourcefunctionid: sourcefunctionid,
				sourceoutputid: sourceoutputid,
				targetfunctionid: targetfunctionid,
				targetinputid: targetinputid
			})
			
		//	console.log('update nodes:', )
			for(var i=0;i<this.nodes.length;i++){
				if(this.nodes[i].id == targetfunctionid){
					for(var j=0;j<this.nodes[i].outputs.length;j++){
						if(this.nodes[i].inputs[j].id ==  targetinputid){
							this.nodes[i].inputs[j].source = 1;
							this.nodes[i].inputs[j].aliasname = sourcefunction.name +'.'+ sourceoutput
							break;
						}
					}
					break;
				}
			}
		//	console.log('update flowobj:', this.funcgroup,this.flowobj )
			for(var n=0;n<this.flowobj.functiongroups.length;n++){
				if(this.flowobj.functiongroups[n].name = this.funcgroupname){
					//console.log('found function group:',n)
					for(var i=0;i<this.flowobj.functiongroups[n].functions.length;i++){
						if(this.flowobj.functiongroups[n].functions[i].id == targetfunctionid){
						//	console.log('found functions', i, this.flowobj.functiongroups[n].functions[i].name)

							for(var j=0;j<this.flowobj.functiongroups[n].functions[i].outputs.length;j++){
								if(this.flowobj.functiongroups[n].functions[i].inputs[j].id ==  targetinputid){
								//	console.log('found input', j, this.flowobj.functiongroups[n].functions[i].inputs[j].name)

									this.flowobj.functiongroups[n].functions[i].inputs[j].source = 1;
									this.flowobj.functiongroups[n].functions[i].inputs[j].aliasname = sourcefunction.name +'.'+ sourceoutput
									break;
								}
							}
							break;	
						}
						
					}
					break;
				}
				
			}
		//	console.log('complete update:', )
		}

		remove_functionlink(sourcefunctionid, sourceoutputid, targetfunctionid, targetinputid){
			let index = -1;
		//	console.log('remove link:',sourcefunctionid, sourceoutputid, targetfunctionid, targetinputid)
		//	console.log(this.functionlinks)

			if(!sourcefunctionid || !sourceoutputid || !targetfunctionid || !targetinputid)
				return;

			let sourcefunction = this.get_itembyid(this.nodes, sourcefunctionid);
			if(!sourcefunction)
				return;

			let sourceoutput = this.get_itemnamebyid(sourcefunction.outputs, sourceoutputid);

			let targetfunction = this.get_itembyid(this.nodes, targetfunctionid);
			if(!targetfunction)
				return;

			let targetinput = this.get_itemnamebyid(targetfunction.inputs, targetinputid);
			if(sourceoutput =="" || targetinput =="")
				return;

			for(var i=0;i<this.functionlinks.length;i++){
				if(this.functionlinks[i].sourcefunctionid == sourcefunctionid &&
					this.functionlinks[i].sourceoutputid == sourceoutputid &&
					this.functionlinks[i].targetfunctionid == targetfunctionid &&
					this.functionlinks[i].targetinputid == targetinputid){
						index = i;
						break;
					}
			}

			if(index >=0 ){
				this.functionlinks.splice(index,1);
			//	this.refresh();
			}
			index =-1;
			for(var i=0;i<this.functionlinklines.length;i++){
				if(this.functionlinklines[i].data.sourcefunction == sourcefunctionid &&
					this.functionlinklines[i].data.sourceoutput == sourceoutputid &&
					this.functionlinklines[i].data.targetfunction == targetfunctionid &&
					this.functionlinklines[i].data.targetinput == targetinputid){
						index = i;
						break;
					}
			}
			if(index >=0 ){
				this.functionlinklines.splice(index,1);
			//	this.refresh();
			}

			for(var i=0;i<this.nodes.length;i++){
				if(this.nodes[i].id == targetfunctionid){
					for(var j=0;j<this.nodes[i].outputs.length;j++){
						if(this.nodes[i].inputs[j].id ==  targetinputid){
							this.nodes[i].inputs[j].source = 0;
							this.nodes[i].inputs[j].aliasname = ''
							break;
						}
					}
					break;
				}
			}
		//	console.log('update flowobj:', this.funcgroup,this.flowobj )
			for(var n=0;n<this.flowobj.functiongroups.length;n++){
				if(this.flowobj.functiongroups[n].name = this.funcgroupname){
					//console.log('found function group:',n)
					for(var i=0;i<this.flowobj.functiongroups[n].functions.length;i++){
						if(this.flowobj.functiongroups[n].functions[i].id == targetfunctionid){
						//	console.log('found functions', i, this.flowobj.functiongroups[n].functions[i].name)

							for(var j=0;j<this.flowobj.functiongroups[n].functions[i].outputs.length;j++){
								if(this.flowobj.functiongroups[n].functions[i].inputs[j].id ==  targetinputid){
								//	console.log('found input', j, this.flowobj.functiongroups[n].functions[i].inputs[j].name)

									this.flowobj.functiongroups[n].functions[i].inputs[j].source =0;
									this.flowobj.functiongroups[n].functions[i].inputs[j].aliasname = ''
									break;
								}
							}
							break;	
						}
						
					}
					break;
				}
				
			}
		}	

		make_links(){
			let that = this;
			that.linklines = [];
			this.links.forEach(function(_link){
				
				let fromnode = that.get_block(_link.fromnode);
				let tonode = that.get_block(_link.tonode);	
				
				let mergepoint = null;
				if(!_link.mergegroup)
					mergepoint = null;
				else if(_link.mergegroup >0 )	
					 mergepoint = that.get_mergepoint(_link.mergegroup).node;
				else
					 mergepoint = null;
				
			//	console.log(_link.fromnode,_link.tonode)
				
			//	console.log(fromnode,tonode)
				
				if(fromnode && tonode)
					that.linklines.push(new Link(that, fromnode.node,tonode.node, _link,mergepoint));			
			});	
			
		}
		
		add_link(sourceblock, targetblock,groupid){
			let that = this
			if(that.flowtype== "FUNCGROUP")
				return;

			
			let _link = {
					fromnode: sourceblock.data.type == 'START' ? 'START' : this.options.flowtype == 'PROCESS'? sourceblock.data.OprSequenceNo: sourceblock.data.id, 
					tonode: this.options.flowtype == 'PROCESS'? targetblock.data.OprSequenceNo : targetblock.data.id,
					Label: this.options.flowtype == 'PROCESS'?  'Good' : '',
					wipcontentclassid:1,
					mergegroup:groupid==0? '':groupid,
					reasoncode:''
				}

			this.links.push(_link);
			
		//	console.log('add link',_link)
			
			let mergepoint = null;
				if(!_link.mergegroup)
					mergepoint = null;
				else if(_link.mergegroup >0 )	
					 mergepoint = that.get_mergepoint(_link.mergegroup).node;
				else
					 mergepoint = null;
			
			let fromnode = this.get_block(_link.fromnode);
			let tonode = this.get_block(_link.tonode);	
			
			console.log(fromnode, tonode)

			if(fromnode && tonode){
				let newlinkline = 	new Link(this, fromnode.node,tonode.node, _link, mergepoint);
				
				this.linklines.push(newlinkline);	

				this.add_funcgrouplinktoflowobject(fromnode.node, tonode.node)
			}
		}
		
		update_link(sourceoprsequenceno, targetoprsequenceno,wipcontentclassid,reasoncode, description){
			
			let _linkline = this.get_linkview_byopr(sourceoprsequenceno,targetoprsequenceno)
						
			_linkline.data.Label = description;
			_linkline.data.wipcontentclassid = wipcontentclassid;
			_linkline.data.reasoncode = reasoncode; 
			
		//	console.log(_linkline.id, )
			
			$('g[model-id="'+_linkline._link.id+'"]').find('text').find('tspan').html(description);
			
			let _link = this.get_link_byopr(sourceoprsequenceno,targetoprsequenceno)
			_link.Label = description;
			_linkline.wipcontentclassid = wipcontentclassid;
			_linkline.reasoncode = reasoncode; 
			
		//	console.log(_linkline.id, this.links)
			
		}
		
		make_Toolbar(){
			if(!this.options.showtoolbar)
				return;
			
			$('.uiflow_process_flow_toolbar_container').html('');
			
			let that = this;
		//	console.log(this.toolbars)
			this.toolbars.forEach(function(toolbar){
				if(toolbar.shows.toUpperCase().includes(that.options.flowtype.toUpperCase()))
					return new Toolbar(that,toolbar);	
				else 
					return;
			}) 
			
		}
		
		make_Menubar(){

			this.menu_panel.innerHTML ="";
			let that = this;
			this.Menubars.forEach(function(menu){
				return new MenuBar(that, menu);
			})
		}
		
		make_link_tools(){
			if(!this.options.interactive)
					return;
			
			let that =this;
				
			this.Paper.on('link:mouseenter', function(linkView) {
				that.svgZoom.disablePan();
			//	console.log(that.linktoolsView)
				linkView.addTools(that.linktoolsView);
			//	console.log(linkView)
			});

			this.Paper.on('link:mouseleave', function(linkView) {
				linkView.removeTools();
						
				that.svgZoom.enablePan();
			});
			
			this.Paper.on('link:remove', function(linkView) {
				//linkView.removeTools();
				let _link = that.get_link_bylinkview(linkView);
			//	console.log('removelink', _link)								
				//that.svgZoom.enablePan();
			});
			
		}
		
		initial_const(){
			let that = this;
	
			joint.linkTools.InfoButton = joint.linkTools.Button.extend({
				name: 'edit-button',
				options: {
					markup: [{
						tagName: 'circle',
						selector: 'button',
						attributes: {
							'r': 7,
							'fill': '#001DFF',
							'cursor': 'pointer'
						}
					}, {
						tagName: 'path',
						selector: 'icon',
						attributes: {
							'd': 'M -2 4 2 4 M 0 3 0 0 M -2 -1 1 -1 M -1 -4 1 -4',
							'fill': 'none',
							'stroke': '#FFFFFF',
							'stroke-width': 2,
							'pointer-events': 'none'
						}
					}],
					distance: 10,
					offset: 0,
					action: function(evt) {
						let _link = that.get_link_bylinkview(this);
					//	alert('View id: ' + this.id + '\n' + 'Model id: ' + this.model.id +'\n'+ _link.data);
						
						if(_link)
							that.trigger_event('link_change', [_link,this]); 
					}
				}
			});
			
			
			joint.linkTools.mergeButton = joint.linkTools.Button.extend({
				name: 'edit-button',
				options: {
					markup: [{
						tagName: 'rect',
						selector: 'button',
						attributes: {
							fill: 'darkblue', 
							stroke: 'darkblue', 
							"stroke-width":0.5,
							width: 15, 
							height: 15,
							transform: 'rotate(45)',
							cursor: 'pointer'
						}
					}],
					distance: '80%',
					offset: 0,
					action: function(evt) {
						let _link = that.get_link_bylinkview(this);
					//	alert('View id: ' + this.id + '\n' + 'Model id: ' + this.model.id +'\n'+ _link.data);
						console.log(_link)
						let mgid = that.add_new_mergepoint(_link);
						that.trigger_event('link_merge', [_link,this, mgid]); 
					}
				}
			});
		
			const linkmergeButton = new joint.linkTools.mergeButton();
			
			const linkinfoButton = new joint.linkTools.InfoButton();

			const linkremoveButton = new joint.linkTools.Remove({
					useModelGeometry: true,
					action: function(_evt, view) {
					   let _link = that.get_link_bylinkview(this);
					//   alert('View id: ' + this.id + '\n' + 'Model id: ' + this.model.id +'\n'+ _link.data);
						console.log(_link)
						if(_link)
							that.trigger_event('link_remove', [_link,this]); 
					}
				});

			/*var verticesTool = new joint.linkTools.Vertices();
			var segmentsTool = new joint.linkTools.Segments();
			var sourceArrowheadTool = new joint.linkTools.SourceArrowhead();
			var targetArrowheadTool = new joint.linkTools.TargetArrowhead();
			var sourceAnchorTool = new joint.linkTools.SourceAnchor();
			var targetAnchorTool = new joint.linkTools.TargetAnchor();
			var boundaryTool = new joint.linkTools.Boundary(); */
		//	console.log(this.showlinkmergepoint)
			
		/*	if(!this.showlinkmergepoint)
				this.linktoolsView = new joint.dia.ToolsView({
					tools: [
						linkinfoButton, linkremoveButton
						
					]
				});
			else  */
			if(that.flowtype != 'FUNCGROUP' ){
			}
			else if(that.flowtype != 'PROCESS' ){

				this.linktoolsView = new joint.dia.ToolsView({
					tools: [
						linkinfoButton, linkremoveButton
					]
				});

			}
			else{
				this.linktoolsView = new joint.dia.ToolsView({
					tools: [
						linkinfoButton, linkremoveButton,linkmergeButton
					]
				});				
					
			}
			
			const elementboundaryTool = new joint.elementTools.Boundary({
					padding: 10,
					rotate: true,
					useModelGeometry: true
				});
				
			const elementremoveButton = new joint.elementTools.Remove({
					useModelGeometry: true,
					action: function(_evt, view) {
					   console.log('joint.elementTools.Remove',_evt,view);
					   that.trigger_event('node_remove', [_link,this]); 
					}
				}); 
			
			const linkfromButton = new joint.elementTools.Button({
					markup: [{
						tagName: 'circle',
						selector: 'button',
						attributes: {
							'r': 7,
							'magnet': 'true',
							'fill': '#025718',
							'cursor': 'crosshair',
							'stroke-width': 2,
							'draggable':true
						}
					}],
					x: '100%',
					y: '50%',
					offset: {
						x: -5,
						y: 0
					},
					rotate: true,
					magnet: true,
					action: function(evt,view) {
						if(that.options.flowtype == "FUNCGROUP"){
							
							that.linkelements = null;
							that.linkfromelementview = null;
							that.linkline = null;
							$('#uiflow_temp_link_line').remove()
							return;
						}
							
						evt.preventDefault()
						that.linkelements = true;
						that.linkfromelementview = view;
					//	console.log('View id: ' + this.id + '\n' + 'Model id: ' + this.model.id, view);
					//	this.addClass('dragging');
						$('html,body').css('cursor','crosshair');
						
					//	evt.dataTransfer.setData('sourceview', view);
					//	evt.dataTransfer.effectAllowed = "crosshair";
						
							joint.highlighters.mask.add(view, { selector: 'root' }, 'my-element-highlight', {
								deep: true,
								attrs: {
									'stroke': '#3FFF33',
									'stroke-width': 4
								}
							});	
							
						that.linkline = document.createElementNS("http://www.w3.org/2000/svg", "line");
						that.linkline.setAttribute("marker-end", "url(#arrowhead)");
						that.linkline.setAttribute("stroke", "red");
						that.linkline.setAttribute("stroke-width", "2");
						that.linkline.setAttribute("id", "uiflow_temp_link_line");
						that.linkline.setAttribute("x1", evt.offsetX);
						that.linkline.setAttribute("y1", evt.offsetY);
						that.linkline.setAttribute("x2", evt.offsetX);
						that.linkline.setAttribute("y2", evt.offsetY);
						$("svg").append(that.linkline); 
						//that.linkline = linkline; 
						
					}
				});
				
			this.elementtoolsView = new joint.dia.ToolsView({
					tools: [
						elementboundaryTool,
						elementremoveButton,
						linkfromButton	//, linkfromButton,removeButton
					]
				});	
		
		}
		
		make_element_tools(){
			if(!this.options.interactive)
				return;
				
			this.linkelements = false; 
			

			let that =this;
			
			
			this.Paper.on('element:mouseenter', function(elementView) {
			//	console.log("enter the element",elementView)
				if(!that.linkelements || that.linkfromelementview != elementView){
					
					var color = '#FF4365'
					
					if(that.linkelements)
						 color = '#3FFF33'
					
					
						joint.highlighters.mask.add(elementView, { selector: 'root' }, 'my-element-highlight', {
							deep: true,
							attrs: {
								'stroke': color,
								'stroke-width': 2
							}
						});	
				
					that.svgZoom.disablePan();

					if(that.options.flowtype != "FUNCGROUP")
						elementView.addTools(that.elementtoolsView);
				}
			});
			
			this.Paper.on('element:mouseleave', function(elementView) {
				
				if(!that.linkelements || that.linkfromelementview != elementView){
					elementView.hideTools();
					joint.dia.HighlighterView.remove(elementView);					
					that.svgZoom.enablePan();
				}
			}); 
			
		}
		
		create_events(){
			
			let that =this;
			
			this.Paper.on('blank:pointerclick', function(){
				
				if(that.linkfromelementview){
					that.linkfromelementview.hideTools();
					
					joint.dia.HighlighterView.remove(that.linkfromelementview);
				}
				
				that.linkelements = null;
				that.linkfromelementview = null;	

				if(that.linkline)
					$('#uiflow_temp_link_line').remove()
				that.linkline = null;

				$('html,body').css('cursor','pointer');
			});
			
			this.Paper.on('blank:pointerdbclick', function(){
				
				if(that.linkfromelementview){
					that.linkfromelementview.hideTools();
					
					joint.dia.HighlighterView.remove(that.linkfromelementview);
				}
				
				that.linkelements = null;
				that.linkfromelementview = null;
				//that.linkline.remove();
				if(that.linkline)
					$('#uiflow_temp_link_line').remove()
				that.linkline = null;

				$('html,body').css('cursor','pointer');
			});
			
			this.Paper.on('element:mouseover', function(elementView) {
				
				if(!that.linkelements || that.linkfromelementview != elementView){
					var color = '#FF4365'
					
					if(that.linkelements)
						 color = '#3FFF33'
	
				
					joint.highlighters.mask.add(elementView, { selector: 'root' }, 'my-element-highlight', {
						deep: true,
						attrs: {
							'stroke': color,
							'stroke-width': 3
						}
					});				
				}
				
			});

			this.Paper.on('element:pointerdown', function(elementView) {
	
				joint.highlighters.mask.add(elementView, { selector: 'root' }, 'my-element-highlight', {
					deep: true,
					attrs: {
						'stroke': '#FF4365',
						'stroke-width': 3
					}
				});
				
				elementView.showTools();
				
			});

			this.Paper.on('element:pointerup', function(elementView) {
				
			//	console.log('element:pointerup',that.linkelements, that.linkfromelementview,elementView)
				if(that.linkelements && that.linkfromelementview && that.linkfromelementview != elementView)
				{
					//that.linkfromelementview,elementView, 
					
					let fromelement = that.get_object_byelementid(that.linkfromelementview.model.id)
					let toelement = that.get_object_byelementid(elementView.model.id)
					
					if(fromelement.type =='block' && toelement.type == 'block'){
							that.trigger_event('link_add', [fromelement.obj, toelement.obj,0]); 
							//that.trigger_event('link_add', [that.get_block_byelementid(that.linkfromelementview.model.id), that.get_block_byelementid(elementView.model.id)]); 
							that.add_link(fromelement.obj, toelement.obj,0)
					}
					else{
						console.log(fromelement,toelement);
						
						if(fromelement.type =='mergepoint' && toelement.type== 'block'){
							let nodes = that.get_mergepoint_linkedblock(fromelement.obj.data.id).fromnodes;
							
							for(var i=0;i<nodes.length;i++){
								let block = that.get_block(nodes[i]);
								
								if(block){
									that.add_link(block, toelement.obj,fromelement.obj.data.id)
									that.trigger_event('link_add', [block, toelement.obj,fromelement.obj.data.id]); 
								}
							}
							
						}
						else if(toelement.type=='mergepoint' && fromelement.type == 'block'){
							let nodes = that.get_mergepoint_linkedblock(toelement.obj.data.id).fromnodes;
							
							for(var i=0;i<nodes.length;i++){
								let block = that.get_block(nodes[i]);
								
								if(block){
									that.add_link(fromelement.obj, block, toelement.obj.data.id)
									that.trigger_event('link_add', [fromelement.obj, block, toelement.obj.data.id]); 
								}
							}
							
						}
						
					}
				/*	that.add_link(that.linkfromelementview, elementView); */
					
					that.linkfromelementview.hideTools();
					
					joint.dia.HighlighterView.remove(that.linkfromelementview);
					
					that.linkelements = null;
					that.linkfromelementview = null;	
					if(that.linkline)
						$('#uiflow_temp_link_line').remove();
					that.linkline = null;

					 $('html,body').css('cursor','pointer');
				}
				else if((!that.linkelements || that.linkelements == undefined) && (that.linkfromelementview == undefined || !that.linkfromelementview ) ){
					that.linkfromelementview = null;
					that.linkelements = null;
					
					joint.highlighters.mask.add(elementView, { selector: 'root' }, 'my-element-highlight', {
							deep: true,
							attrs: {
								'stroke': '#FF4365',
								'stroke-width': 3
							}
						});
						
					var nodeid = elementView.model.attr('nodeid')
				//	console.log(nodeid)
					that.trigger_event('block_click', [nodeid]);		
					
				}
				
				that.linkfromelementview = null;
				that.linkelements = null;
				if(that.linkline)
					$('#uiflow_temp_link_line').remove();
			//	that.linkline.remove();
				that.linkline = null;
				
				elementView.hideTools();
				joint.dia.HighlighterView.remove(elementView);
				that.svgZoom.enablePan();
				$('html,body').css('cursor','pointer');				
			});
			
			this.Paper.on('element:mouseout', function(elementView) {
				//resetAll(this);
				if(!that.linkelements || that.linkfromelementview != elementView)
					joint.dia.HighlighterView.remove(elementView);
				
			});	

			this.Paper.on('element:pointerdblclick', function(elementView) {
				that.linkelements = null;
				that.linkfromelementview = null;
				
				joint.highlighters.mask.add(elementView, { selector: 'root' }, 'my-element-highlight', {
					deep: true,
					attrs: {
						'stroke': '#FF4365',
						'stroke-width': 3
					}
				});
				
				var nodeid = elementView.model.attr('nodeid')
		
				that.trigger_event('block_dbclick', [nodeid]);
				
			});

			this.Paper.on('element:pointerlclick', function(elementView) {
				joint.highlighters.mask.add(elementView, { selector: 'root' }, 'my-element-highlight', {
					deep: true,
					attrs: {
						'stroke': '#FF4365',
						'stroke-width': 3
					}
				});
				
				var nodeid = elementView.model.attr('nodeid')
				that.trigger_event('block_click', [nodeid]); 
				
			});		
			
			if(this.options.flowtype == 'FUNCGROUP'){
				this.Paper.on('port:mouseenter', function(event, port) {
					console.log('port:mouseenter', event, port)
				})
				this.Paper.on('port:pointerclick', function(event, port) {
					console.log('port:pointerclick',event, port)
				})
				this.Paper.on('link:mouseenter', function(linkView) {
					var tools = new joint.dia.ToolsView({
						tools: [
							new joint.linkTools.TargetArrowhead(),
							new joint.linkTools.Remove({ distance: -30 })
						]
					});
					linkView.addTools(tools);
				});
				
				this.Paper.on('link:mouseleave', function(linkView) {
					linkView.removeTools();
				});
				
				this.Paper.on('link:connect link:disconnect', function(linkView, evt, elementView) {
					var element = elementView.model;
					console.log('link:connect link:disconnect:', linkView, evt, elementView,element)
				//	console.log(linkView.sourceView,$(linkView.sourceMagnet).attr('port'), linkView.targetView,$(linkView.targetMagnet).attr('port'))
					var sourcenodeid = linkView.sourceView.model.attr('nodeid')
					var destnodeid = linkView.targetView.model.attr('nodeid')
				//	console.log(sourcenodeid, $(linkView.sourceMagnet).attr('port'),destnodeid, $(linkView.targetMagnet).attr('port'))
				
					that.add_functionlink(sourcenodeid, $(linkView.sourceMagnet).attr('port'),destnodeid, $(linkView.targetMagnet).attr('port'))
				});
				
				this.Graph.on('remove', function(cell, collection, opt) {
				//	console.log('remove', cell, collection, opt)
					if (!cell.isLink() || !opt.ui) return;
					if(!cell.target().id || !cell.source().id) return;

					var target = this.getCell(cell.target().id).attr('nodeid');
					var source = this.getCell(cell.source().id).attr('nodeid');
				//	console.log(source, target, cell.target().port, cell.source().port)
					that.remove_functionlink(source, cell.source().port,target, cell.target().port)
				//	if (target instanceof Shape) target.updateInPorts();
				});
			}
			
			this.attach_dropeventstoport();
		//	window.addEventListener('resize', joint.util.debounce(that.rescale), false);
			window.addEventListener('resize',that.windows_resize,false);
			this.attach_contextmenu();
			
			that.trigger_event('process_ready', this); 
		/*	
			this.Paper.el.addEventListener("mousemove", (event) =>  {
				
				if(that.linkline){
					console.log(event)
					that.linkline.setAttribute("x2", event.offsetX);
					that.linkline.setAttribute("y2", event.offsetY);
				}
				else if($('#uiflow_temp_link_line').length > 0)
					$('#uiflow_temp_link_line').remove();
				
				event.stopPropagation();
			});  
			this.Paper.el.addEventListener("mouseup", (event) =>  {
				$('#uiflow_temp_link_line').remove();
			//	that.linkline.remove();
				that.linkline = null;
				event.stopPropagation();
			});   */

			/*	
			const dragEnter = event => {
				console.log('drag enter element', event)
				
				event.preventDefault();
				
				event.currentTarget.classList.add('drop');
			};
		
			const dragOver = event => {
		//		console.log('drag over element', event)
				
				//event.preventDefault();
				
				return false;
			};			

			const dragLeave = event => {
				console.log('drag leave element', event)
			//	event.preventDefault();
				
				event.currentTarget.classList.remove('drop');
			};	

			const dragDrop= event => {
				console.log('drag drop element', event)
				event.preventDefault();
			//	concole.log('drag drop elemen', event.dataTransfer.getData('text/html'));
				event.currentTarget.classList.remove('drop');
			};	
			
			document.querySelectorAll('.joint-element').forEach(join_element => {
				join_element.addEventListener('dragenter', dragEnter);
		//		join_element.addEventListener('dragover', dragOver);
		//		join_element.addEventListener('dragleave', dragLeave);
				join_element.addEventListener('drop', dragDrop);
			});
			*/
			
						
		}
		windows_resize(that){
			
			joint.util.debounce(function(){
				//var that =this;		
				
				
				if($('#'+this.sectionwrapper).width() > 800)
					this.options.width = $('#'+this.sectionwrapper).width();
				if($('#'+this.sectionwrapper).height() > 600)
				this.options.height = $('#'+this.sectionwrapper).height();

				let width = this.options.width;
				let height = this.options.height;

				if(that.Paper){			
					let originalwidth = that.Paper.options.width;
					let originalheight = that.Paper.options.height;
						
					$('#'+that.sectionwrapper).css('width', (width) + 'px');
					$('#'+that.sectionwrapper).css('height', (height) + 'px');
						
					let widthscale = width / originalwidth;
					let heightscale = height / originalheight;
						
					console.log('resize', that.Paper, widthscale, heightscale)
						
					that.Paper.scale(widthscale,heightscale);
					//console.log('resize', that.Paper)
					that.Paper.options.width = width;
					that.Paper.options.height = height;
					//	that.Paper.scaleContentToFit({ padding: 50 });
					that.refresh();
					/*	joint.util.debounce(function(){	
							console.log('resize', that.Paper)				
							that.Paper.scaleContentToFit({ padding: 50 });
							//that.zoom();
					})  */
				}
			})
		}
		
		attach_dropeventstoport(){
			if(this.options.flowtype !="FUNCGROUP")
				return;

			var that = this;
			$('.joint-port').each(function(){
				$(this).find('circle').on('drop', function(event) {
					event.preventDefault();
					event.stopPropagation();
					console.log(event.currentTarget)
					let category = event.originalEvent.dataTransfer.getData("category");
					if(category == "session"){
						let type = event.originalEvent.dataTransfer.getData("type");					
						let variable = event.originalEvent.dataTransfer.getData("variable");
						let fucntionid = $(event.currentTarget).attr('functionid');
						let portgroup = $(event.currentTarget).attr('port-group');
						let port = $(event.currentTarget).attr('port');
						that.function_parameter_assignment(fucntionid, portgroup, port, category, type, variable);
						//console.log($(event.currentTarget).attr('port-group'),$(event.currentTarget).attr('functionid'), category, type, variable)
					}
					

				})

				$(this).find('circle').on('dragover', function(event) {
					event.preventDefault();
					event.stopPropagation();
				//	console.log('dragover', event)
				})

			})

		}

		function_parameter_assignment(functionid, paramtype, parameterid,category,type, variable){
			if(paramtype =='output' && category == 'session' && type=='system')
			{
				alert('System variable cannot be assigned to output parameter');
				return;
			}
			let that = this;
			console.log(functionid, paramtype, parameterid,category,type, variable)
			let block = that.get_block_bydataid(functionid);
			if(block){
				if(paramtype == 'input'){
					let data={
						id:parameterid,
						source: (type == 'system'? 2: 3),
						aliasname: variable
					}
					block.update(data, paramtype);
				}
				else if(paramtype == 'output'){
					let outputdest =[];
					let aliasname =[];
	
					for(var j=0;j<block.data.outputs.length;j++){
						if(block.data.outputs[j].id == parameterid){
							if(Array.isArray(block.data.outputs[j].outputdest )){
								outputdest = block.data.outputs[j].outputdest;
								aliasname = block.data.outputs[j].aliasname;
							}
							else if(block.data.outputs[j].outputdest !='' && block.data.outputs[j].outputdest != null){
								outputdest.concat([block.data.outputs[j].outputdest]);
								aliasname.concat([block.data.outputs[j].aliasname]);
							}

							break;
						}
					}
		
					let data={
						id:parameterid,
						outputdest: outputdest.concat([1]),
						aliasname: aliasname.concat([variable])
					}
					console.log(block,data)
					block.update(data, paramtype);
				}
				
			}

			return;

		}
		update_node_Elements(id,Element){
			let elementsstr = ''
		//	console.log(id,Element)
			for(var i=0;i< this.blocks.length;i++){
				if(this.blocks[i].id == id){
					elementsstr = this.blocks[i].data.Elements;	
					let code = this.get_code_Element(Element);
					
					elementsstr = ((elementsstr ==undefined || !elementsstr) ? '': elementsstr);
					
				//	console.log(elementsstr,code,elementsstr.indexOf(code))
									
					
					if(code !='' && elementsstr.indexOf(code) < 0){
						this.blocks[i].data.Elements = elementsstr +  code;
						
					//	console.log(this.blocks[i].data.Elements);
						this.render();
						
						return;
					}
					
					
					return;
				}				
			}	
			
		}
		
		get_code_Element(Element){
			let code = '';
			switch(Element.toUpperCase()){
				case 'STEP':
					code = '\ue408';
					break;
				case 'CHECKLIST':
					code = '\ue014';
					break;
				case 'DOCUMENT':
					code = '\uf1ea';
					break;
				case 'WORKINSTRUCTION':
					code = '\uf15c';
					break;
				case 'COMPONENT':
					code = '\uf12e';
					break;
				case 'RESOURCE':
					code = '\uf0e3';
					break;
				case 'RESOURCECLASS':
					code = '\ue115';
					break;
				case 'ROLE':
					code = '\ue344';
					break;
				case 'SKILL':
					code = '\ue118';
					break;
				case 'CHARACTERISTIC':
					code = '\uf02c';
					break;
				case 'DATACOLLECT':
					code = '\ue339';
					break;
				case 'ALERT':
					code = '\uf003';
					break;
				case 'SIGNATURE':
					code = '\uf040';
					break;					

			}

			return code;	
			
		}
		get_itembyid(items,id){
			return items.find(item => {
				return item.id == id;
			});			
		}

		get_itembyname(items,name){
		//	console.log('get_itembyname',items,name)
			return items.find(item => {
				return item.name == name;
			});			
		}
		get_itemnamebyid(items,id){
			let item = this.get_itembyid(items,id);
			if(item	){
				return item.name;
			}			
			
			return '';
		}
		get_itemidbyname(items,name){
		//	console.log('get_itembyname',items,name)
			let item = this.get_itembyname(items,name);

			if(item	){
				return item.id;
			}			
			
			return '';
		}

		get_node(id){
			return this.get_itembyid(this.nodes,id);
			/*
			return  this.nodes.find(node => {
			//	console.log(node, id)
				return node.id == id;
			});		*/
		}
		
		get_block(id){
			return  this.blocks.find(block => {
			//	console.log(block)
				return block.id == id;
			});
		}
		
		get_mergepoint(id){
			return  this.mergepoints.find(mp => {
			//	console.log(block)
				return mp.id == id;
			});
		} 
		
		get_mergepoint_linkedblock(mpid){
			let fromblocks = [];
			let toblocks = []
			
			for(var i=0;i<this.links.length;i++){
				if(this.links[i].mergegroup)
					if(this.links[i].mergegroup == mpid){
						let count = 0;
						for(var j=0;j<fromblocks.length;j++){
							if(fromblocks[j] == this.links[i].fromnode){
								count += 1
								break;
							}
						}
						if(count == 0)
							fromblocks.push(this.links[i].fromnode);
						
						count = 0;
						for(var j=0;j<toblocks.length;j++){
							if(toblocks[j] == this.links[i].tonode){
								count += 1
								break;
							}
						}
						
						if(count == 0)
							toblocks.push(this.links[i].tonode);					
					}				
			}
		/*	console.log(this.links, {
				fromnodes: fromblocks,
				tonodes:toblocks
			})*/
			return {
				fromnodes: fromblocks,
				tonodes:toblocks
			}
		}
		
		get_element_byPos(x,y){
			let jselements = $('g.joint-element');
			let that = this;
			for(var i=0;i<jselements.length;i++){
				let ele = jselements[i];
				let rect = ele.getBoundingClientRect(); 
				
				if(rect.x < x &&  x < rect.right && rect.y < y && y < rect.bottom){
				//	console.log(ele,ele.getAttribute('model-id'),that.get_block_byelementid(ele.getAttribute('model-id')))
					return that.get_block_byelementid(ele.getAttribute('model-id')); 				
				}
			}		
				
			return null; 
		}
		get_element_byID(id){
			return this.Graph.getCell(id)
		}
		
		get_block_byelementid(modelid){
			
			for(var i=0;i<this.blocks.length;i++){
				let block = this.blocks[i];
				if(block.node.shape.id == modelid)
					return block;				
			}
		}
		
		get_block_bydataid(id){
			for(var i=0;i<this.blocks.length;i++){
				let block = this.blocks[i];
				if(block.node.id == id)
					return block;				
			}
		}
		get_mergepoint_byelementid(modelid){
			
			for(var i=0;i<this.mergepoints.length;i++){
				let mp = this.mergepoints[i];
				if(mp.node.shape.id == modelid)
					return mp;				
			}
		}
		
		get_object_byelementid(modelid){
			
			let block = this.get_block_byelementid(modelid);
			
			if(!block){
				let mp = this.get_mergepoint_byelementid(modelid)
				if(mp){
					return {
						type: 'mergepoint',
						obj:mp
					}
				}
				else
					return null
			}
			else{
				return {
					type: 'block',
					obj: block
				}
			}
			
			
		}
		
		get_link_bylinkview(linkview){
			for(var i=0;i<this.linklines.length;i++){
				let linkline = this.linklines[i];
			//	console.log(linkline)
				if(linkline._link == linkview.model)
					return linkline;				
			}
			
		}
		get_link_byopr(source,target){
			for(var i=0;i< this.links.length;i++)
				if(this.links[i].fromnode == source && this.links[i].tonode == target)
					return this.links[i];
		}
		
		get_linkview_byopr(source,target){
			
			for(var i=0;i< this.linklines.length;i++){
			//	console.log(this.linklines[i], source, target)
				if(this.linklines[i].data.fromnode == source && this.linklines[i].data.tonode == target)
					return this.linklines[i];
				
			}
			
		}
		
		attach_contextmenu(){
		//	console.log(this.options.flowtype.toUpperCase())
			switch(this.options.flowtype.toUpperCase()){
				case 'PROCESS':
				//	this.attach_process_contextmenu();
					break;
				case 'OPERATION':
				case 'WORKFLOW':
				case 'TRANCODE':
				//	this.attach_workflow_contextmenu();
					this.attach_trancode_contextmenu();
					break;
				case 'FUNCGROUP':
					this.attach_funcgroup_contextmenu();
					break;
			}
		}

		attach_trancode_contextmenu(){
			let that = this;
			/*
				contextmenu for the paper in the trancode flow
			*/
			/*
				let node = {
							id: nodeid,
							name:fgname,
							functiongroupname:functiongroup.name,
							OprSequenceNo: functiongroup.name,
							SequenceNo: 0,
							StepName: functiongroup.name,
							Description: functiongroup.description,
							Elements: [],
							routing:routing,
							type: "FUNCGROUP"
						};
			*/

			$.contextMenu({
				selector: '.joint-paper', 
				build:function($triggerElement,e){
					
					return{
						callback: function(key, options,e){
							console.log(key, options,e)
							switch(key){

								case 'Properties':
									that.build_trancode_properties();
									break;
								case 'AddFuncGroup':
									console.log("add func group")
									that.add_functiongroup();
									break;
								case 'AutoLayout':
									that.auto_layout();
									break;
								case 'Parameters':
									that.build_trancode_parameters();
									break;
							}

						}, 
						items:{
							'Properties':{
								name: 'Properties',
								icon: 'fa-cog',
								disabled: false
							},
							'AddFuncGroup':{
								name: 'Add Function Group',
								icon: 'fa-plus',
								disabled: false
							},
							'AutoLayout':{
								name: 'Auto layout',
								icon: 'fa-plus',
								disabled: false
							},
							'Parameters':{
								name: 'Parameters',
								icon: 'fa-plus',
								disabled: false
							},
							"sep1":'------------',
							'Quit':{
								name: 'Quit',
								icon: function($element, key, item){ return 'context-menu-icon context-menu-icon-quit'; },
							}
						}

					}
				},
								
			})

			/*
				context menu items for the funcgroup block
			*/
			$.contextMenu({
				selector: 'g[data-type="ProcessFlow.StepBlock"]', 
				build:function($triggerElement,e){
					console.log('build the contextmenu:',$triggerElement,e,$triggerElement[0].getAttribute('model-id'))
					let block = that.get_block_byelementid($triggerElement[0].getAttribute('model-id'));

					if(!block)
						return{};
					console.log("selected bock:",block.data)
					let functiongroupname =block.data.functiongroupname;
					let nodeid = block.data.id;
					return{
						callback: function(key, options,e){
							console.log(key, options,e)
							switch(key){

								case 'Properties':
									
									break;
								case 'ChangeName':									
									let newfuncgroupname = prompt('Please input the new function group name',functiongroupname);
									
									if(/[^A-Za-z0-9]_-/.test(newfuncgroupname)){
										alert('The fucntion group name can only contain letters and numbers')
										return;
									}
									console.log(newfuncgroupname)
									if(newfuncgroupname && newfuncgroupname != functiongroupname && !that.validate_funcgroupname(newfuncgroupname)){
										if(that.update_funcgroupname(nodeid,functiongroupname,newfuncgroupname)){											
											$triggerElement.find('text[joint-selector="headerText"]').find('tspan').html(newfuncgroupname);
										}										
									}
									
									break;
								case 'Functions':
									let newoptions = that.options
								//	console.log(newoptions, that.options)
									newoptions.flowtype = 'FUNCGROUP'
									that.destry();
									$.contextMenu('destroy', '.joint-paper');	
									$.contextMenu('destroy', 'g[data-type="ProcessFlow.StepBlock"]');									
									that.setup_objects(newoptions, functiongroupname);
								
									break;
							}

						}, 
						items:{
							'Properties':{
								name: 'Properties',
								icon: 'fa-cog',
								disabled: false
							},
							'ChangeName':{
								name: 'Chaneg Name',
								icon: 'fa-cog',
								disabled: false
							},
							'Functions':{
								name: 'Functions',
								icon: 'fa-cog',
								disabled: false
							},						
							'Delete':{
								name: 'Delete',
								icon: 'fa-trash',
								disabled: false
							},
							"sep1":'------------',
							'Quit':{
								name: 'Quit',
								icon: function($element, key, item){ return 'context-menu-icon context-menu-icon-quit'; },
							}
						}

					}
				},
								
			})

		}
		build_trancode_properties(){
			this.property_panel.innerHTML  = "" 
			var divsToRemove = this.property_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}

			let that = this;
			let flowobj = that.flowobj;
			let control_attrs ={
				class: 'container-fluid',
				style: 'width:100%;height:95%;margin-left:10px;margin-right:10px;'				
			}
			let property_container = (new UI.FormControl(this.property_panel,'div',control_attrs)).control;

			control_attrs ={
				innerHTML: 'Trancode Properties',
			}
			new UI.FormControl(property_container,'h2',control_attrs);
			new UI.FormControl(property_container,'br',{});

			control_attrs ={
				for: 'trancodename',
				innerHTML: 'Trancode Name',
			}
			new UI.FormControl(property_container,'label',control_attrs);

			control_attrs ={
				type: 'text',
				id: 'trancodename',
				name: 'trancodename',
				value: flowobj.trancodename,
				class: 'form-control'
			}
			new UI.FormControl(property_container,'input',control_attrs);
			new UI.FormControl(property_container,'br',{});

			control_attrs ={
				for: 'trancodeversion',
				innerHTML: 'Trancode Version',
			}
			new UI.FormControl(property_container,'label',control_attrs);
			new UI.FormControl(property_container,'br',{});
			control_attrs ={
				type: 'text',
				id: 'trancodeversion',
				name: 'trancodeversion',
				value: flowobj.version,
				class: 'form-control'
			}
			new UI.FormControl(property_container,'input',control_attrs);
			new UI.FormControl(property_container,'br',{});

			control_attrs ={
				for: 'description',
				innerHTML: 'Description',
			}
			new UI.FormControl(property_container,'label',control_attrs);
			new UI.FormControl(property_container,'br',{});

			control_attrs ={
				id: 'description',
				name: 'description',
				value: flowobj.description,
				class: 'form-control'
			}
			new UI.FormControl(property_container,'textarea',control_attrs);
			new UI.FormControl(property_container,'br',{});
			
			control_attrs ={
				class: 'btn btn-primary',
				id: 'savefunction',
				innerHTML: 'Save'
			}
			let save_function =function(){
				let trancodename = document.getElementById('trancodename').value;
				let trancodeversion = document.getElementById('trancodeversion').value;
				let description = document.getElementById('description').value;				
				that.flowobj.trancodename = trancodename;
				that.flowobj.version = trancodeversion;
				that.flowobj.description = description;				
				that.property_panel.style.width = "0px";
				that.property_panel.style.display = "none";
			}
			let events={
				click: save_function
			}
			new UI.FormControl(property_container,'button',control_attrs,events);
			
			control_attrs ={
				class: 'btn btn-danger',
				id: 'cancelfunction',
				innerHTML: 'Cancel'
			}
			events={
				click: function(){
					that.property_panel.style.width = "0px";
					that.property_panel.style.display = "none";
				}
			}
			
			new UI.FormControl(property_container,'button',control_attrs,events);

			that.property_panel.style.width = "350px";
			that.property_panel.style.display = "flex";			

		}
		build_trancode_parameters(){
			this.property_panel.innerHTML  = "" 
			var divsToRemove = this.property_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}

			let that = this;
			let flowobj = that.flowobj;

			let control_attrs ={
				class: 'container-fluid',
				style: 'width:100%;height:95%;margin-left:10px;margin-right:10px;'				
			}
			let property_container = (new UI.FormControl(this.property_panel,'div',control_attrs)).control;

			control_attrs ={
				innerHTML: 'X',
				class: 'btn btn-danger',
				style: 'float:right;top:2px;right:2px;position:absolute;',
				id: 'closefunction'
			}
			let close_function =function(){
				$(".parameter-data").off('change',that.update_trancodeparameter)
				that.property_panel.style.width = "0px";
				that.property_panel.style.display = "none";
				that.property_panel.innerHTML  = ""
			}
			let events={
				click: close_function
			}
			new UI.FormControl(property_container,'button',control_attrs,events);
			
			control_attrs ={
				innerHTML: 'Trancode Inputs',
			}
			new UI.FormControl(property_container,'h2',control_attrs);
			new UI.FormControl(property_container,'br',{});

			this.build_parameters(flowobj.inputs,property_container, 'input');

			control_attrs ={
				innerHTML: 'Trancode Outputs',
			}
			new UI.FormControl(property_container,'h2',control_attrs);
			new UI.FormControl(property_container,'br',{});

			this.build_parameters(flowobj.outputs,property_container, 'output');
			

			that.property_panel.style.width = "350px";
			that.property_panel.style.display = "flex";		
		}
		build_parameters(items,property_container, type){
			let that = this;
			if(!items){
				items =[];
				if(type == 'input')
					that.flowobj.inputs = items;
				else
					that.flowobj.outputs = items;
			}
			let attrs ={
				class: 'btn btn-primary',
				id: 'addfunction_'+type,
				innerHTML: 'Add'
			}
			let events={
				click: function(){
					that.add_trancodeparameter(type);
				}
			}
			new UI.FormControl(property_container,'button',attrs,events);

			attrs ={
				class: 'btn btn-primary',
				id: 'removefunction_'+type,
				innerHTML: 'Remove'
			}
			events={
				click: function(){
					that.remove_trancodeparameter(type);
				}
			}
			new UI.FormControl(property_container,'button',attrs,events);

			/* {
            attrs:{},
            headers: [{
                text: "",
                attrs: []],
            },{}],
            columns: [{
                control: "",
                attrs: [],
            },{}],
            rows:[{},{}]
        } */
			let rows = [];
			
			items.forEach(function(item){
			//	console.log(item)
				let row=[];
				row.push({data:{},attrs:{parameter_id: item.id}})
				row.push(
					{data:{ value: item.name,},attrs:{innerHTML:item.name},}
				)
				row.push({data:{selected:item.type},attrs:{}})
				row.push({data:{value: item.list},attrs:{}})
				rows.push(row);
			});

			let table_data ={
				attrs:{
					class: 'table table-striped',
					id: 'trancodeparametertable_'+type
				},
				headers: [{
					innerHTML: "",
					style: 'width:20px'					
				},{
					innerHTML: "Name",
					style: 'width:220px'
				},{
					innerHTML: "Type",
					style: 'width:80px'
				},{
					innerHTML: "List",
					style: 'width:20px'
				}],
				columns: [{
					control: "input",
					attrs: {type: 'checkbox', 
						style: 'width:20px',
						parameter_type:type,						
						data_type: 'selector',
						class: 'form-control parameter-selector parameter-data',
					}
				},{
					control: "input",
					attrs: {type: 'text',
						style: 'width:220px;',						
						data_type:'name',
						class:'form-control parameter-name parameter-data'
					}
				},{
					control: "select",
					attrs: {
						style: 'width:80px;',
						data_type:'type',
						class:'form-control parameter-type parameter-data'
					},
					options: Function_DataType_List
				},{
					control: "checkbox",
					attrs: {
						style: 'width:20px',
						data_type:'list',
						class:'form-control parameter-list parameter-data'
					}
				}],
				rows:rows
			}
		//	console.log(table_data)
			new UI.HtmlTable(property_container,table_data);
			
			$('#removefunction_'+type).attr('disabled','disabled');

			$(".parameter-data").on('change',function(e){
				that.update_trancodeparameter(e,that);
			})

		}
		add_trancodeparameter(type){
			let newparameter = prompt("please input the parameter name:", "parameter");
			if(/[^A-Za-z0-9_-]/.test(newparameter)){
				alert('The parameter name can only contain letters and numbers')
				return;
			}
			if(type =="input"){
				for(var i=0;i<this.flowobj.inputs.length;i++){

					if(this.flowobj.inputs[i].name == newparameter ){
						alert("the new parameter name cannot be same as existing name!")
						return;
					}
				} 
			}else{
				for(var i=0;i<this.flowobj.outputs.length;i++){

					if(this.flowobj.outputs[i].name == newparameter ){
						alert("the new parameter name cannot be same as existing name!")
						return;
					}
				}  

			}

			let parameter ={
				id: UIFlow.generateUUID(),
				name:newparameter,
				type:0,
				list:false
			}
			if(type == 'input'){
				this.flowobj.inputs.push(parameter);
			}else{
				this.flowobj.outputs.push(parameter);
			}
			
			this.build_trancode_parameters();
			console.log(this.flowobj)
		}
		update_trancodeparameter(e, that){
			//console.log(e.target)
			//let that = this
			let ele = $(e.target);
			let parameter_id = ele.closest('tr').find('.parameter-selector').attr('parameter_id');
			let parameter_type = ele.closest('tr').find('.parameter-selector').attr('parameter_type');
			let newvalue = ele.val();
			let data_type = ele.attr('data_type');
		//	console.log(this,that.flowobj,parameter_id,newvalue,data_type)

			switch(data_type){
				case "name":
					if(/[^A-Za-z0-9_-]/.test(newvalue)){
						alert('The parameter name can only contain letters and numbers')
						return;
					}
					if(parameter_type == "input"){
						for(var i=0;i<this.flowobj.inputs.length;i++){
							if(this.flowobj.inputs[i].name == newparameter ){
								alert("the new parameter name cannot be same as existing name!")
								return;
							}
						}  
						that.flowobj.inputs.forEach(function(item){
							if(item.id == parameter_id){
								item.name = newvalue;
								return;
							}
						})
					}else if(parameter_type == "output"){
						for(var i=0;i<this.flowobj.outputs.length;i++){

							if(this.flowobj.outputs[i].name == newparameter ){
								alert("the new parameter name cannot be same as existing name!")
								return;
							}
						}
						that.flowobj.outputs.forEach(function(item){
							if(item.id == parameter_id){
								item.name = newvalue;
								return;
							}
						})
					}
					break;
				case "type":
					let datatype = -1
					for (var i=0;i<Function_DataType_List.length;i++){
						if(i == newvalue){
							datatype = i;
							break;
						}
					}	
					if(datatype == -1){
						alert("The data type is not correct!")
						return;
					}
					if(parameter_type == "input"){
						that.flowobj.inputs.forEach(function(item){
							if(item.id == parameter_id){
								item.type = datatype;
								return;
							}
						})
					}else if(parameter_type == "output"){
						that.flowobj.outputs.forEach(function(item){
							if(item.id == parameter_id){
								item.type = datatype;
								return;
							}
						})
					}
					break;
				case "list":
					newvalue = ele.is(':checked');
					if(parameter_type == "input"){
						that.flowobj.inputs.forEach(function(item){
							if(item.id == parameter_id){
								item.list = newvalue;
								return;
							}
						})
					}else if(parameter_type == "output"){
						that.flowobj.outputs.forEach(function(item){
							if(item.id == parameter_id){
								item.list = newvalue;
								return;
							}
						})
					}
					break;
				case "selector":
					let table = ele.closest('table');
					let selectors = table.find('.parameter-selector');
					let selectedcout = 0;
					selectors.each(function(index,item){
						if($(item).is(':checked')){
							selectedcout++;
						}
					})

					if(selectedcout == 0)
						$('#removefunction_'+parameter_type).attr('disabled','disabled');
					else 
						$('#removefunction_'+parameter_type).removeAttr('disabled');
					break;
			}
		}
		remove_trancodeparameter(type){
			let that = this;
			let table = $('#trancodeparametertable_'+type);
			let selectors = table.find('.parameter-selector');
			//let selectedcout = 0;
			console.log(table, selectors)
			let selectedparameters = [];
			/*selectors.each(function(item){
				console.log(item)
				if($(item).is(':checked')){
					console.log($(item))
					selectedparameters.push($(item).attr("parameter-id"))
					//selectedcout++;
				}
			}) */
			for(var i=0;i<selectors.length;i++){
				//console.log(i,$(selectors[i]),$(selectors[i]).attr("parameter_id") )
				let item = selectors[i];
				if($(item).is(':checked')){
				//	console.log($(item))
					selectedparameters.push($(item).attr("parameter_id"))
					//selectedcout++;
				}

			}


			if(selectedparameters.length == 0){
				alert("Please select the parameter to be deleted!")
				return;
			}
		//	console.log(that.flowobj,selectedparameters)
			if(confirm("Are you sure to delete the selected parameters?")){
				if(type == 'input'){
					let index = -1;
					for(var i=0;i<selectedparameters.length;i++){
						for(var j=0;j<that.flowobj.inputs.length;j++){
							if(that.flowobj.inputs[j].id == selectedparameters[i]){
								index = j;
								break;
							}
						}
						if(index >=0)
							that.flowobj.inputs = that.flowobj.inputs.splice(index,1);
					}
					
				}else{
					let index = -1;
					for(var i=0;i<selectedparameters.length;i++){
						for(var j=0;j<that.flowobj.outputs.length;j++){
							if(that.flowobj.outputs[j].id == selectedparameters[i]){
								index = j;
								break;
							}
						}
						if(index >=0)
							that.flowobj.outputs = that.flowobj.outputs.splice(index,1);
					}
				}
				that.build_trancode_parameters();
			}
		
		}
		add_functiongroup(){
			let that = this;
			console.log("add function group")
			let newfgname = that.get_funcgroupname();
			let nodeid = UIFlow.generateUUID();
		//	console.log(newfgname,nodeid)
			let node = {
				id: nodeid,
				name: newfgname,
				functiongroupname:newfgname,
				description: newfgname,
				elements: [],
				routerdef:{
					"variable": "",
					"value": [],
					"nextfuncgroups":[],
					"defaultfuncgroup":""
				},
				routing:false,
				type: "FUNCGROUP",
				x: 100,
				y: 100,
				width: this.options.nodewidth,
				height: this.options.nodeheight
			};
		//	console.log(node)
			that.nodes.push(node);
		//	console.log(that.nodes)
			let block = new Block(that, node);
		//	console.log(block)
			that.blocks.push(block);
		//	console.log(that.blocks, that.flowobj)

			that.add_funcgrouptoflowobject(node)
		}
		get_funcgroupname(){
			let that = this;
			let index = 0;
			let find = true;
			let newfgname = "NewFunctiongroupName";
			while(find && index < 100)
			{
				index +=1;
				newfgname = "NewFunctiongroupName" + index.toString().padStart(2, '0');				
				find = this.validate_funcgroupname(newfgname);
			}
			return newfgname;
		}

		validate_funcgroupname(newfgname){
			let that = this;
			let find = false;
			for(var i=0;i< this.nodes.length;i++){
				let node = this.nodes[i];
				if(node.name == newfgname){
					find = true;
					break;
				}
			}
			return find;
		}
		update_funcgroupname(nodeid,oldname,newname){
			let that = this;

			let data={
				name:newname,
				functiongroupname:newname,
			}
			let block = that.get_block_bydataid(nodeid);

			if(block){
				block.update(data,'');
				return true;
			}
			return false;
			
		}
		add_funcgrouptoflowobject(funcgroup){
			let that = this;
			let newfg = {
				id: funcgroup.id,
				name: funcgroup.functiongroupname,
				functiongroupname: funcgroup.functiongroupname,
				functions:[],
				RouterDef:{
					"variable": "",
					"value": [],
					"nextfuncgroups":[],
					"defaultfuncgroup":""
				}
			};

			that.flowobj.functiongroups = that.flowobj.functiongroups.concat(newfg);
		}


		add_funcgrouplinktoflowobject(fromnode, tonode){
			console.log(fromnode, tonode)
			let sourceblock = this.get_block_bydataid(fromnode.id);
			let targetblock = this.get_block_bydataid(tonode.id);

			let that = this;

			if(fromnode.id =='START'){
				that.flowobj.firstfuncgroup = targetblock.data.name;
				return;
			}

			for(var i=0;i<that.flowobj.functiongroups.length;i++){
			//	console.log(that.flowobj.functiongroups[i],fromnode, tonode)
				if(that.flowobj.functiongroups[i].id == fromnode.id){
					
					if(that.flowobj.functiongroups[i].routerdef.variable != "")
					{
						let values = that.flowobj.functiongroups[i].routerdef.value.concat([""]);
						let nextfuncgroups = that.flowobj.functiongroups[i].routerdef.nextfuncgroups.concat([targetblock.data.name]);
						that.flowobj.functiongroups[i].routerdef.value = values;
						that.flowobj.functiongroups[i].routerdef.nextfuncgroups = nextfuncgroups;
					}
					else 
						that.flowobj.functiongroups[i].routerdef.defaultfuncgroup = targetblock.data.name;
					
					break;
				}
			}
		//	console.log(that.flowobj)
		}

		attach_funcgroup_contextmenu(){
			
			let that = this;
			/*
				context menu for the paper
			*/
			$.contextMenu({
				selector: '.joint-paper', 
				build:function($triggerElement,e){
					
					return{
						callback: function(key, options,e){
							console.log(key, options,e)
							switch(key){

								case 'Properties':
									
									break;
								case 'AddFunction':
									var html = "";
									for (var i = 0; i < Function_Type_List.length; i++) {
									  html += '<input type="radio" class="function_type" name="items" value="' + i + '"> ' + Function_Type_List[i] + '<br>';
									}	

									that.property_panel.innerHTML  = "" 
									var divsToRemove = that.property_panel.getElementsByClassName("container-fluid");
									while (divsToRemove.length > 0) {
										divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
									}
									let title = document.createElement('div');
									title.innerHTML = 'Select the function type to add a new Function';
									title.className = 'container-fluid';
									that.property_panel.appendChild(title);

									let property_container = document.createElement('div');
									property_container.setAttribute('class','container-fluid');	
									property_container.style.width = '100%';
									property_container.style.height = '95%';
									property_container.style.marginLeft = '10px';
									property_container.style.marginRight = '10px';
									that.property_panel.appendChild(property_container);									

									property_container.innerHTML = html;
									that.property_panel.style.display = 'block';
									that.property_panel.style.width = '300px';
								//	console.log(property_container.getElementsByClassName('function_type'))
									for(var i=0;i<property_container.getElementsByClassName('function_type').length;i++){
										let ele = property_container.getElementsByClassName('function_type')[i];
										ele.addEventListener('click',	function(e){
											console.log('Select the function type:',e.target.value)
											that.property_panel.style.display = 'none';
											that.property_panel.innerHTML  = "" 
											that.add_function(e.target.value)
										})
									}							
									
									break;
								case 'AutoLayout':
									that.auto_layout();
									break;
								case 'TransCodeFlow':
									let newoptions = that.options
									console.log(newoptions, that.options)
									that.destry();
									newoptions.flowtype = 'TRANCODE'
									$.contextMenu('destroy', '.joint-paper');									
									that.setup_objects(newoptions, "");
							}

						}, 
						items:{
							'Properties':{
								name: 'Properties',
								icon: 'fa-cog',
								disabled: false
							},
							'AddFunction':{
								name: 'Add Function',
								icon: 'fa-plus',
								disabled: false
							},
							'AutoLayout':{
								name: 'Auto layout',
								icon: 'fa-plus',
								disabled: false
							},
							'TransCodeFlow':{
								name: 'TransCode Flow',
								icon: 'fa-back',
								disabled: false
							},
							"sep1":'------------',
							'Quit':{
								name: 'Quit',
								icon: function($element, key, item){ return 'context-menu-icon context-menu-icon-quit'; },
							}
						}

					}
				},
								
			})

			/*
				context for the function block
			*/
			$.contextMenu({
				selector: 'g[data-type="ProcessFlow.StepBlock.Function"]', 
				build:function($triggerElement,e){
					console.log('build the contextmenu:',$triggerElement,e)
					let block = that.get_block_byelementid($triggerElement[0].getAttribute('model-id'));
					console.log("selected bock:",block)
					let functionname =block.data.name;
					let nodeid = block.data.id;
					return{
						callback: function(key, options,e){
							console.log(key, options,e)
							switch(key){

								case 'Properties':
									that.build_function_property_panel(nodeid);
									break;
								case 'ChangeName':
																		
									let newfunctionname = prompt('Please input the new function name',functionname);
									if(/[^A-Za-z0-9_-]/.test(newfunctionname)){
										alert('The newfunction name can only contain letters and numbers')
										return;
									}
									console.log(newfunctionname)
									if(newfunctionname && newfunctionname != functionname){
										if(that.update_functionname(nodeid, functionname,newfunctionname)){
										//	$triggerElement.find('tspan').html(newfunctionname) ;
											$triggerElement.find('text[joint-selector="functionname"]').find('tspan').html(newfunctionname) ;//.attr('functionname',newfunctionname);
											$triggerElement.find('rect[joint-selector="functionheader"]').attr('functionname',newfunctionname);
										}										
									}
									break;
									
								case 'AddInputs':
									let number = $triggerElement.find('circle[port-group="input"]').length;
									let inputname = prompt('Please input the input name','input'+number);
									if(/[^A-Za-z0-9_]/.test(inputname)){
										alert('The input name can only contain letters and numbers')
										
									}else
										that.add_functionInput(block, inputname,number, $triggerElement);
									break;
								case 'AddOutputs':
										let outnumber = $triggerElement.find('circle[port-group="output"]').length;
										let outputname = prompt('Please input the output name','output'+outnumber);
										if(/[^A-Za-z0-9_]/.test(outputname)){
											alert('The output name can only contain letters and numbers')
											
										}else
											that.add_functionOutput(block, outputname,outnumber, $triggerElement);
										break;
							}

						}, 
						items:{
							'Properties':{
								name: 'Properties',
								icon: 'fa-cog',
								disabled: false
							},
							'ChangeName':{
								name: 'Chaneg Function Name',
								icon: 'fa-cog',
								disabled: false
							},
							'AddInputs':{
								name: 'Add Inputs',
								icon: 'fa-plus',
								disabled: false
							},
							'AddOutputs':{
								name: 'Add Outputs',
								icon: 'fa-plus',
								disabled: false
							},
							'Delete':{
								name: 'Delete',
								icon: 'fa-trash',
								disabled: false
							},
							"sep1":'------------',
							'Quit':{
								name: 'Quit',
								icon: function($element, key, item){ return 'context-menu-icon context-menu-icon-quit'; },
							}
						}

					}
				},
								
			})
			/*
				contextmenu for the input/output port
			*/
			$.contextMenu({
				selector: 'circle[joint-selector="portBody"]', 
				build:function($triggerElement,e){
					console.log('build the contextmenu:',$triggerElement,e)
					let node = $triggerElement.attr('functionid');
				//	let port = $triggerElement.attr('port')
					let portid = $triggerElement.attr('port')
					let type = $triggerElement.attr('port-group');
					console.log(node, portid,type)
					return{
						callback: function(key, options,e){
							console.log(key, options,e)
							switch(key){
								case 'ChangeName':
									let block = that.get_block_bydataid(node);
									console.log(block)
									if(!block)
										return;
									let port ="";
									switch (type){
										case 'input':
											for(var i=0; i<block.data.inputs.length;i++){
												if(block.data.inputs[i].id == portid){
													port = block.data.inputs[i].name;
													break;
												}
											}
											break;
										case 'output':
											for(var i=0; i<block.data.outputs.length;i++){
												if(block.data.outputs[i].id == portid){
													port = block.data.outputs[i].name;
													break;
												}
											}
											break;	
									}
									
									let newportname = prompt('Please input the new port name',port);
									
									if(/[^A-Za-z0-9]_-/.test(newportname)){
										alert('The input/output name can only contain letters and numbers')
										return;
									}
									//console.log(newportname)
									if(newportname && newportname != port){
										if(that.update_functioninputoutput(node,portid,newportname,type)){
											$triggerElement.attr('port',newportname) ;
											$triggerElement.parent().find('tspan').html(newportname);
										}										
									}
									break;
								case 'Properties':
									that.build_functionparameter_panel(type,node,portid);
									break;
							}
						}, 
						items:{
							'Properties':{
								name: 'Properties',
								icon: 'fa-cog',
								disabled: false
							},
							'ChangeName':{
								name: 'Change Name',
								icon: 'fa-plus',
								disabled: false
							},							
							'Delete':{
								name: 'Delete',
								icon: 'fa-trash',
								disabled: false
							},
							"sep1":'------------',
							'Quit':{
								name: 'Quit',
								icon: function($element, key, item){ return 'context-menu-icon context-menu-icon-quit'; },
							}
						}

					}

				}
			})

		}
		
		add_functionInput(block,  name, i,element){
			let y= 25 + i*20;
			let id = UIFlow.generateUUID();
			let port = {
				group: 'input',
				id: id,
				args: {x: 0, y: y},
				attrs: { 
					circle: { 									
						fill: Function_DataType_Color_List[0],
						functionid:block.data.id,
						port: id,
						portname: name,
					},
					FunctionInputName:{
						port:id,
						text: name,
					},
					rect:{
						width: 0,
						fill: 'none',
					}
				}
			};	
			let input={
				id: id,
				name: name,
				datatype: 0,
				description: name,
				value:'',
				source:	0,
				aliasname: '',
				defaultvalue: ""
			}	
		/*	block.node.shape.addPort(port)
			
			let blockheight = element.find('rect[joint-selector="body"]').attr('height');
			
			if(blockheight < y+40){
				element.find('rect[joint-selector="body"]').attr('height', y+40);
			}
			console.log(block,  name, i,element, input)
			for(var i=0;i<this.nodes.length;i++){
				if(this.nodes[i].id == block.data.id){
					this.nodes[i].inputs.push(input);
					break;
				}
			} */
		
			for(var i=0;i<this.flowobj.functiongroups.length;i++){
				if(this.flowobj.functiongroups[i].name == this.funcgroupname){
					for(var j=0;j<this.flowobj.functiongroups[i].functions.length;j++){
						if(this.flowobj.functiongroups[i].functions[j].id == block.data.id){
							this.flowobj.functiongroups[i].functions[j].inputs.push(input);
							break;
						}
					}

				}
			}
			this.reload();
		}

		add_functionOutput(block,  name, i, element){
			let y= 25 + i*20;
			let x = block.data.width +6;
			let id = UIFlow.generateUUID();
			let port = {
				group: 'output',
				id: id,
				args: {x: x, y: y},
				attrs: { 
					circle: { 									
						fill: Function_DataType_Color_List[0],
						functionid:block.data.id,
						port: id,
						portname: name,
					},
					FunctionOutputName:{
						port:id,
						text: name,
					},
					rect:{
						width: 0,
						fill: 'none',
					}
				}
			};	
			let output={
				id: id,
				name: name,
				datatype: 0,
				description: name,
				outputdest:	[],
				aliasname: [],
				defaultvalue: ""
			}	
		/*	block.node.shape.addPort(port)
			
			let blockheight = element.find('rect[joint-selector="body"]').attr('height');
			
			if(blockheight < y+40){
				element.find('rect[joint-selector="body"]').attr('height', y+40);
			}
			
			for(var i=0;i<this.nodes.length;i++){
				if(this.nodes[i].id == block.data.id){
					this.nodes[i].outputs.push(output);
					break;
				}
			}
			*/
			for(var i=0;i<this.flowobj.functiongroups.length;i++){
				if(this.flowobj.functiongroups[i].name == this.funcgroupname){
					for(var j=0;j<this.flowobj.functiongroups[i].functions.length;j++){
						if(this.flowobj.functiongroups[i].functions[j].id == block.data.id){
							this.flowobj.functiongroups[i].functions[j].outputs.push(output);
							break;
						}
					}

				}
			}
			this.reload();
		}

		update_functioninputoutput(nodeid, portid,newname, type){
			

			let block = this.get_block_bydataid(nodeid);
			if(block){
				if(!this.validate_functionparametername(block, newname, type))
					return false;

				let data ={
					id: portid,
					name: newname,
				}
				block.update(data, type)
				return true;
			}
			return false;
		
		}
		validate_functionparametername(block,name,type){
			switch (type){
				case 'input':
					for(var i=0;i<block.data.inputs.length;i++){
						if(block.data.inputs[i].name == name)
							return false;
					}
					break;
				case 'output':
					for(var i=0;i<block.data.outputs.length;i++){
						if(block.data.outputs[i].name == name)
							return false;
					}
					break;
				}
			return true;
		}
		update_functionname(nodeid, oldname,newname){
			if(!this.validate_functionname(newname))
				return false;

			let block = this.get_block_bydataid(nodeid);
			if(block){
				let data={
					id: nodeid,
					name: newname,
				}
				block.update(data,'');
				return true;

			}
			return false;
			
		}
		validate_functionname(name){
			for(var i=0;i<this.nodes.length;i++){				
				if(this.nodes[i].functionName == name)
					return false			
			}
			return true;
		}
		add_function(functype){
			
			let nodeid = UIFlow.generateUUID();
			let node = {
				id: nodeid,
				name: Function_Type_List[functype],	
				functionName: Function_Type_List[functype],
				description: Function_Type_List[functype],
				content: "",
				functype: functype,
				inputs: [],
				outputs: [],
				type: "FUNCTION",
				position: {},
				x: 100,
				y: 100,
				width: this.options.nodewidth,
				height: this.options.nodeheight
			};

			this.nodes.push(node);
			let block = new Block(this, node);
			this.blocks.push(block);
			this.add_functiontoflowobj(node)
		}

		add_functiontoflowobj(funcobj){
			
			let that = this
			let functionobj = {
				id: funcobj.id,
				name: funcobj.functionName,
				description: funcobj.description,
				content: funcobj.content,
				functype: funcobj.functype,
				inputs: [],
				outputs: [],
				x: funcobj.x,
				y: funcobj.y,
				width: funcobj.width,
				height: funcobj.height
			}
		//	console.log(that.funcgroupname, funcobj, functionobj)
			
			for(var i=0;i< that.flowobj.functiongroups.length ;i++){
				if(that.flowobj.functiongroups[i].name == that.funcgroupname){
					if(!that.flowobj.functiongroups[i].functions)
						that.flowobj.functiongroups[i].functions=[];
					
					that.flowobj.functiongroups[i].functions.push(functionobj)
				//	console.log(that.flowobj)
					return;
				}

			}
		}
	
		build_function_property_panel(functionid){
			this.property_panel.innerHTML  = "" 
			var divsToRemove = this.property_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}

			let that = this;
			let functionobj = this.get_node(functionid);
			if(!functionobj)
				return;

			let attrs={
				class: 'container-fluid',
				style: 'width: 100%; height: 95%; margin-left: 10px; margin-right: 10px;'
			}
			let property_container = (new UI.FormControl(this.property_panel, 'div', attrs)).control;

			attrs={innerHTML: 'Function Properties'}
			new UI.FormControl(property_container, 'h2', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={
				innerHTML: 'Function Name',
				for: 'functionname'
			}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});
			attrs={
				id: 'functionname',
				type: 'text',
				value: functionobj.functionName,
				class: 'form-control',
				placeholder: 'Enter Function Name',
				style: 'width: 100%;'
			}
			new UI.FormControl(property_container, 'input', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={
				innerHTML: 'Function Description',
				for: 'functiondescription'
			}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});
			attrs={
				id: 'functiondescription',
				type: 'text',
				value: functionobj.description,
				class: 'form-control',
				placeholder: 'Enter Function Description',
				style: 'width: 100%;'
			}
			new UI.FormControl(property_container, 'input', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={
				innerHTML: 'Function type',
				for: 'functiontype'
			}
			new UI.FormControl(property_container, 'label', attrs);

			attrs={
				id: 'functiontype',
				type: 'text',
				style: 'width: 100%;',
			}
			let sel_data ={
				attrs: attrs,
				options: Function_Type_List,
				selected: functionobj.functype
			}
			new UI.Selection(property_container, sel_data);
			new UI.FormControl(property_container, 'br', {});

			if(functionobj.functype == "0"){
				attrs={
					innerHTML: 'Function inputs and outputs mapping',
					for: 'functioncontent'
				}
				new UI.FormControl(property_container, 'label', attrs);
				new UI.FormControl(property_container, 'br', {});

				let options = [];
				console.log('inputs:', functionobj.inputs)
				for(var i=0;i<functionobj.inputs.length;i++){
					options.push({value: functionobj.inputs[i].name, innerHTML: functionobj.inputs[i].name})
				}
				let obj = {};
				if(typeof functionobj.content == 'string'){

					try{
						obj= JSON.parse(functionobj.content)
					}catch{}
				}
				else if (functionobj.content == null)
					obj = {};
				else if(typeof functionobj.content == 'object')
					obj = functionobj.content;

				let rows = [];
				for(var i=0;i<functionobj.outputs.length;i++){
					let cells=[];
					cells.push(
						{data:{},attrs:{innerHTML:functionobj.outputs[i].name},}
					)
					cells.push({data:{selected:obj[functionobj.outputs[i].name]},attrs:{}})					
					rows.push(cells);
				}

				attrs={
					headers: [{innerHTML:'Input', style:'width:120px;'},{innerHTML:'Output', style:'width:120px;'}],
					style: 'width: 100%;',
					id: 'functioncontent',
					columns: [{control:'', attrs:{style:'width:100%;'}}, 
						{control:'select', attrs:{style:'width:100%;'}, options: options }],
					rows:rows
				}

				new UI.HtmlTable(property_container, attrs);
			}
			else{
				attrs={
					innerHTML: 'Function Content',
					for: 'functioncontent'
				}
				new UI.FormControl(property_container, 'label', attrs);
				new UI.FormControl(property_container, 'br', {});
				attrs={
					id: 'functioncontent',
					type: 'text',
					value: functionobj.content,
					class: 'form-control',
					placeholder: 'Enter Function Content',
					style: 'width: 100%;height: 100px;'
				}
				new UI.FormControl(property_container, 'textarea', attrs);
			}
			new UI.FormControl(property_container, 'br', {});

			attrs={
				innerHTML: 'Save',
				id: 'savefunction',
				class: 'btn btn-primary'
			}
			var savefuntion = function(){
				let oldfunctionname = functionobj.functionName;
				let functionname = $('#functionname').val();
				//console.log(oldfunctionname,functionname)
				
				if(oldfunctionname != functionname){
					if(!that.update_functionname(functionobj.id,oldfunctionname,functionname)){	
						alert('function name already exists')				
						return;
					}
				//	let block = that.get_block_bydataid(functionobj.id);
				//	$('g[model-id="'+block.node.shape.id +'"]').find('text').find('tspan').html(functionname);	
				}
				let functiondescription =$('#functiondescription').val();
				let functiontype = $('#functiontype').val();
				let functioncontent ="";
				
				if(functiontype == "0"){
					let fcobj = {};
					$('td.output_parameter').each(function(i){
						let output = $(this).html();
						let input = $('#input'+i).val();
						if(input !='')
							fcobj[output] = input;
					})
					functioncontent = fcobj;
				}
				else
					functioncontent = $('#functioncontent').val();
				
				let block = that.get_block_bydataid(functionobj.id);
				if(block){
					let data={
						content:functioncontent,
						description:functiondescription,
						functype:functiontype
					}
					block.update(data,'');

				}


				that.property_panel.style.width = "0px";
				that.property_panel.style.display = "none";
				that.property_panel.innerHtml = "";
			}
			let events={
				click: savefuntion
			}
			new UI.FormControl(property_container, 'button', attrs,events);

			attrs={
				innerHTML: 'Cancel',
				id: 'cancelfunction',
				class: 'btn btn-danger'
			}
			events={
				click: function(){
					that.property_panel.style.width = "0px";
				that.property_panel.style.display = "none";
				}
			}
			new UI.FormControl(property_container, 'button', attrs,events);
			
			that.property_panel.style.width = "350px";
			that.property_panel.style.display = "flex";
		}

		build_functionparameter_panel(type,functionid,parameterid){

			this.property_panel.innerHTML  = "" 
			var divsToRemove = this.property_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}

			let that = this;
			let fnobj = that.nodes.find(function(obj){
				return obj.id == functionid;
			});

			let parameterobj = {}

			if(type == 'input')
				 parameterobj = fnobj.inputs.find(function(obj){
					return obj.id == parameterid;
				});
			else 
				parameterobj = fnobj.outputs.find(function(obj){
					return obj.id == parameterid;
				});

			let attrs={
				class: 'container-fluid',
				style: 'width:100%; height:95%; margin-left:10px; margin-right:10px;'

			}
			let property_container = (new UI.FormControl(this.property_panel, 'div', attrs)).control;
			
			attrs={innerHTML: 'Function '+type+' Properties'}
			new UI.FormControl(property_container, 'h2', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={ for: 'parametername', innerHTML: 'Parameter Name'}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={class: 'form-control', placeholder: 'Parameter Name', id: 'parametername', value: parameterobj.name, style: 'width: 100%;'}
			new UI.FormControl(property_container, 'input', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={ for: 'parameterdescription', innerHTML: 'Parameter Description'}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={class: 'form-control', placeholder: 'Parameter Description', id: 'parameterdescription', value: parameterobj.description, style: 'width: 100%;'}
			new UI.FormControl(property_container, 'textarea', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={ for: 'parameterdatatype', innerHTML: 'Parameter Data Type'}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});
			attrs={class: 'form-control', placeholder: 'Parameter Data Type', id: 'parameterdatatype', style: 'width: 100%;'}
			let se_data = {
				attrs: attrs,
				selected: parameterobj.datatype,
				options: Function_DataType_List
			}
			new UI.Selection(property_container, se_data);

			attrs={ for: 'parameterlist'}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={id: 'parameterlist', value:parameterobj.list,style: 'width: 70%;'}
			new UI.CheckBox(property_container,'checkbox',attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs ={for: 'parameterdefaultvalue', innerHTML: 'Parameter Default Value'}
			new UI.FormControl(property_container, 'label', attrs);
			new UI.FormControl(property_container, 'br', {});

			attrs={class: 'form-control', placeholder: 'Parameter Default Value', id: 'parameterdefaultvalue', value: parameterobj.defaultvalue, style: 'width: 100%;'}
			new UI.FormControl(property_container, 'input', attrs);
			new UI.FormControl(property_container, 'br', {});
						
			if(type == 'input'){

				attrs={ for: 'parameterdefaultvalue', innerHTML: 'Parameter Default Value'}
				new UI.FormControl(property_container, 'label', attrs);
				new UI.FormControl(property_container, 'br', {});

				attrs={class: 'form-control', placeholder: 'Parameter Value', id: 'parametertvalue', value: parameterobj.value==undefined?'':parameterobj.value, style: 'width: 100%;'}
				new UI.FormControl(property_container, 'input', attrs);
				new UI.FormControl(property_container, 'br', {});

				attrs={ for: 'parametersource', innerHTML: 'Parameter Source'}
				new UI.FormControl(property_container, 'label', attrs);
				new UI.FormControl(property_container, 'br', {});
				attrs={
					attrs:{class: 'form-control', placeholder: 'Parameter Source', id: 'parametersource', style: 'width: 100%;',value:parameterobj.source}, 
					options:Function_Source_List}
				new UI.Selection(property_container, attrs);
				new UI.FormControl(property_container, 'br', {});

				attrs={ for: 'parameteraliasname', innerHTML: 'Parameter Alias Name'}
				new UI.FormControl(property_container, 'label', attrs);
				new UI.FormControl(property_container, 'br', {});
				attrs={class: 'form-control', placeholder: 'Parameter Alias Name', id: 'parameteraliasname', value: parameterobj.aliasname==undefined?'':parameterobj.aliasname, style: 'width: 100%;'}
				new UI.FormControl(property_container, 'input', attrs);
				
		
			}else{
				attrs={ for: 'parameterdest', innerHTML: 'Parameter Destination'}
				new UI.FormControl(property_container, 'label', attrs);
				new UI.FormControl(property_container, 'br', {});

				attrs={class: 'btn btn-primary', id: 'addbtn', innerHTML:'Add',style: 'width: 50%;'}
				let events={click: function(){
					let cells=[];
					let cell={}
					cells.push(cell)
					cells.push({data:{selected:0}})
					cell={data:{value: ""}}
					cells.push(cell)
					console.log(table)
					table.AddRow(cells,table.tbody.control)
					}
				}
				new UI.FormControl(property_container, 'button', attrs, events);
				
				attrs={class: 'btn btn-primary', id: 'removebtn', innerHTML:'Remove',style: 'width: 50%;'};
				events={click: function(){
					$('.parameter_dest_row_selector').each(function(){					
						if($(this).prop('checked')){
							$(this).closest('tr').remove();
						}
					});
				}}
				new UI.FormControl(property_container, 'button', attrs, events);
				new UI.FormControl(property_container, 'br', {});
				let rows=[];
				
				parameterobj.outputdest.forEach(function(obj,index){
					let cells=[];
					cells.push({})
					let cell={data:{selected: obj	}}
					cells.push(cell)
					cell={data:{value: parameterobj.aliasname[index]}}
					cells.push(cell)
					rows.push(cells)
				})


				let table_data={
					attrs:{id:'parameterdesttable', class:'table table-bordered table-hover', style:'width: 100%;'},
					headers:[{innerHTML:'Selector', style:'width:30px;'},{innerHTML:'Destination',style:'width:150px;' },{innerHTML:'Alias Name',style:'width:150px;' }],
					columns:[{control:'checkbox',attrs:{class:'parameter_dest_row_selector'}},
						{control:'select', attrs:{class:'form-control parameterdest_selector', placeholder:'Parameter Destination', style:'width: 100%;'}, 
							options:Function_Dest_List},
						{control:'input',attrs:{class:'form-control parameteraliasname', placeholder:'Parameter Alias Name', style:'width: 100%;'}}],
					rows: rows	
				}
				
				let table=(new UI.HtmlTable(property_container, table_data));

			}
			new UI.FormControl(property_container, 'br', {});

			attrs={class: 'btn btn-primary', id: 'savebutton', innerHTML:'Save',style: 'width: 35%;'}
			let events={click: function(){
				let block = that.get_block_bydataid(functionid);

				parameterobj.name = parametername.value;
				parameterobj.description = parameterdescription.value;
				parameterobj.datatype = parameterdatatype.value;
				parameterobj.list = parameterlist.value;
				parameterobj.defaultvalue = parameterdefaultvalue.value;
				if(type == 'input'){
					parameterobj.source = document.getElementById('parametersource').value;
					parameterobj.aliasname = document.getElementById('parameteraliasname').value;
					parameterobj.value = document.getElementById('parametertvalue').value;
					
					let data = {
						id:parameterobj.id,
						name:parameterobj.name,
						description:parameterobj.description,
						datatype:parameterobj.datatype,
						source:parameterobj.source,
						aliasname:parameterobj.aliasname,
						value:parameterobj.value,				
						defaultvalue:parameterobj.defaultvalue,
						list:parameterobj.list
					};
					if(block){
						block.update(data, type)
					}
				}
				else{
					parameterobj.outputdest = [];
					parameterobj.aliasname = [];
					$('.parameterdest_selector').each(function(){
						if($(this).val() !='' && $(this).closest('tr').find('.parameteraliasname').val() !=''){
							parameterobj.outputdest.push($(this).val());
							parameterobj.aliasname.push($(this).closest('tr').find('.parameteraliasname').val());
						}
					})
					let data = {
						id:parameterobj.id,
						name:parameterobj.name,
						description:parameterobj.description,
						datatype:parameterobj.datatype,
						outputdest:parameterobj.outputdest,
						aliasname:parameterobj.aliasname,
						defaultvalue:parameterobj.defaultvalue,
						list:parameterobj.list
					};
					if(block){
						block.update(data, type)
					}
				}
				parameterobj.defaultvalue = parameterdefaultvalue.value;
				parameterobj.list = parameterlist.checked;
				that.property_panel.innerHTML  = "" 
				that.property_panel.style.width = "0px";
				that.property_panel.style.display = "none";
			}}
			new UI.FormControl(property_container, 'button', attrs, events);
			new UI.FormControl(property_container, 'br', {});

			attrs={class: 'btn btn-primary', id: 'cancelbutton', innerHTML:'Cancel',style: 'width: 35%;'}
			events={click: function(){
				that.property_panel.innerHTML  = ""
				that.property_panel.style.width = "0px";
				that.property_panel.style.display = "none";
			}}
			new UI.FormControl(property_container, 'button', attrs, events);

			property_container.appendChild(cancelbutton);
			that.property_panel.style.width = "350px";
			that.property_panel.style.display = "flex";
		}


		menu_click(menudata) {
			let that = this;
			console.log(menudata, this)
			switch(menudata.type){
				case "Repository":
					this.trigger_event("go_back", []);
					break;
				case "Save":
					this.trigger_event("save_flow", [that.flowobj]);
				break
				case "Sessions":
					this.show_Sessions();
					break;
				case "Parameters":
					this.show_Parameters();
					break;
				case "Export":
					this.export_flowjson();
					break;
				case "Import":
					this.import_flowjson();
					break;
				case "New":
					this.new_flow();
					break;
			}
		}
		new_flow(){
			let that = this;
			let flowobj = {
				name:'New Flow',
				uuid:UIFlow.generateUUID(),
				version:'1.0',
				description:'New Flow',
				functiongroups:[],
			}
			that.flowobj = flowobj;
			let newoptions = that.options					
			newoptions.flowtype = 'TRANCODE'
			that.destry();
			$.contextMenu('destroy', '.joint-paper');									
			that.setup_objects(newoptions, "");
		}
		import_flowjson(){
			let that = this;
			let popup = document.createElement('div')
			popup.setAttribute('class','popupContainer')
			popup.setAttribute('id','popupContainer')

			let popupContent = document.createElement('div')
			popupContent.setAttribute('class','popupContent')
			popupContent.setAttribute('id','popupContent')
			popup.appendChild(popupContent)
			let title = document.createElement('h2')
			title.innerHTML = 'Please select file to import'
			popupContent.appendChild(title)

			let fileInput = document.createElement('input');
			fileInput.type = 'file';
			fileInput.accept = '.json';
			fileInput.addEventListener('change', (event) => {
			  const file = event.target.files[0];
			  that.read_to_import_File(file);
			});	
			popupContent.appendChild(fileInput)

			let closePopupButton = document.createElement('button');
			closePopupButton.setAttribute('id','closePopupButton')
			closePopupButton.innerHTML = 'Close'
			closePopupButton.addEventListener('click', () => {
				popup.style.display = 'none';
				$('#popupContainer').remove();
			});
			popupContent.appendChild(closePopupButton)
			document.body.appendChild(popup)
			popup.style.display = 'block';
		}
		read_to_import_File(file){
			const reader = new FileReader();
			let that = this;
  			reader.onload = (event) => {
				const fileContents = event.target.result;
				try {
					const jsonData = JSON.parse(fileContents);
				// Handle the JSON data
					console.log(jsonData);
					that.flowobj = jsonData;
					let newoptions = that.options					
					newoptions.flowtype = 'TRANCODE'
					that.destry();
					$.contextMenu('destroy', '.joint-paper');									
					that.setup_objects(newoptions, "");

					$('#popupContainer').remove();
					
				} catch (error) {
				console.error('Error parsing JSON file:', error);
				}
			};

			reader.readAsText(file);
		}
		export_flowjson(){
			let that = this;
			let flowjson = this.flowobj;
			let blob = new Blob([JSON.stringify(flowjson)], {type: "text/plain;charset=utf-8"});
			console.log(blob)
			
			let file = new File([blob], this.flowobj.trancodename+'_'+this.flowobj.version+".json", {type: "text/plain;charset=utf-8"});
			
			saveAs(file)
		}

		show_Parameters(){
			let that = this;

			this.item_panel.innerHTML  = "" 
			var divsToRemove = this.item_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}
			let attrs={class: 'container-fluid',style: 'width: 90%;height:95%;margin-left:10px;margin-right:10px;'}
			let container_fluid = (new UI.FormControl(this.item_panel, 'div', attrs)).control;
			
			attrs={class: 'btn btn-danger', id: 'closefunction', innerHTML:'X',style: 'float:right;top:2px;right:2px;position:absolute;'}
			let events={click: function(){
				that.item_panel.style.width = "0px";
				that.item_panel.style.display = "none";
				that.item_panel.innerHTML  = "" }};
			new UI.FormControl(container_fluid, 'button', attrs, events);

			new UI.FormControl(container_fluid, 'h2', {innerHTML:'TranCode Inputs/Outputs Management'});
			new UI.FormControl(container_fluid, 'br', {});

			new UI.FormControl(container_fluid, 'h3', {innerHTML:'TranCode Input Parameters'});
			new UI.FormControl(container_fluid, 'br', {});
			
			attrs={class:'btn btn-primary',id:'addinput',innerHTML:'Add Input',style:'margin-bottom:10px;'}
			events={click: function(){
				let cells=[];
				cells.push({});
				cells.push({data:{value:""}});
				cells.push({data:{selected:0}});
				cells.push({data:{value:false}});
				cells.push({data:{value:""}});
				input_table.AddRow(cells);
			}}
			new UI.FormControl(container_fluid, 'button', attrs, events);
			attrs={class:'btn btn-danger',id:'removeinput',innerHTML:'Remove Input',style:'margin-bottom:10px;'}
			events={click: function(){
				$('#parameter_input_table').find('.parameter_input_line_selector').each(function(){
					if($(this).prop('checked')){
						$(this).closest('tr').remove();
					}
				})
			}}

			new UI.FormControl(container_fluid, 'button', attrs, events);
			
			let rows=[]
			if(!that.flowobj.hasOwnProperty('inputs'))
				that.flowobj.inputs =[];
			if(!that.flowobj.hasOwnProperty('outputs'))
				that.flowobj.outputs =[];

			that.flowobj.inputs.forEach(function(inputparameter){
				let cells=[];
				cells.push({});
				cells.push({data:{value:inputparameter.name}});
				cells.push({data:{selected:inputparameter.datatype}});
				cells.push({data:{value:inputparameter.list}});
				cells.push({data:{value:inputparameter.default}});
				rows.push(cells);
			})
			let table_data={
				attrs:{class: 'table table-bordered', id: 'parameter_input_table', style: 'width: 100%;'},	
				headers:[{innerHTML:'Selector',style:"width:40px;"},
					{innerHTML:'Name', style:'width:120px;'},
					{innerHTML:'Type', style:'width:90px;'},
					{innerHTML:'Array?', style:'width:40px;'},
					{innerHTML:'Default Val', style:'width:120px;'}],
				columns:[{control:'checkbox', attrs:{class:'parameter_input_line_selector',style:'width:100%;'}},
				{control:'input', attrs:{class:'parameter_name',style:'width:100%;'}},
				{control:'select', attrs:{class:'parameter_datatype',style:'width:100%;'},options:Function_DataType_List},
				{control:'checkbox', attrs:{class:'parameter_list',style:'width:100%;'}},
				{control:'input', attrs:{class:'parameter_default',style:'width:100%;'}}],
				tr:{
					attrs:{dragable: true, dragstart:"parameter_dragStart(event, 'input', this)"},
					events:{}
				},
				rows:rows
			}

			let input_table = new UI.HtmlTable(container_fluid, table_data);
			new UI.FormControl(container_fluid, 'br', {});

			new UI.FormControl(container_fluid, 'h3', {innerHTML:'TranCode Output Parameters'});
			new UI.FormControl(container_fluid, 'br', {});

			attrs={class:'btn btn-primary',id:'addoutput',innerHTML:'Add Output',style:'margin-bottom:10px;'}
			events={click: function(){
				let cells=[];
				cells.push({});
				cells.push({data:{value:""}});
				cells.push({data:{selected:0}});
				cells.push({data:{value:false}});
				cells.push({data:{value:""}});
				output_table.AddRow(cells);
			}}
			new UI.FormControl(container_fluid, 'button', attrs, events);
			attrs={class:'btn btn-danger',id:'removeoutput',innerHTML:'Remove Output',style:'margin-bottom:10px;'}
			events={click: function(){
				$('#parameter_output_table').find('.parameter_output_line_selector').each(function(){
					if($(this).prop('checked')){
						$(this).closest('tr').remove();
					}
				})
			}}

			new UI.FormControl(container_fluid, 'button', attrs, events);
			
			rows=[]
			that.flowobj.outputs.forEach(function(parameter){
				let cells=[];
				cells.push({});
				cells.push({data:{value:parameter.name}});
				cells.push({data:{selected:parameter.datatype}});
				cells.push({data:{value:parameter.list}});
				cells.push({data:{value:parameter.default}});
				rows.push(cells);
			})
			
			table_data={
				attrs:{class: 'table table-bordered', id: 'parameter_output_table', style: 'width: 100%;'},	
				headers:[{innerHTML:'Selector',style:"width:40px;"},
					{innerHTML:'Name', style:'width:120px;'},
					{innerHTML:'Type', style:'width:90px;'},
					{innerHTML:'Array?', style:'width:40px;'},
					{innerHTML:'Default Val', style:'width:120px;'}],
				columns:[{control:'checkbox', attrs:{class:'parameter_output_line_selector',style:'width:100%;'}},
				{control:'input', attrs:{class:'parameter_name',style:'width:100%;'}},
				{control:'select', attrs:{class:'parameter_datatype',style:'width:100%;'},options:Function_DataType_List},
				{control:'checkbox', attrs:{class:'parameter_list',style:'width:100%;'}},
				{control:'input', attrs:{class:'parameter_default',style:'width:100%;'}}],
				tr:{
					attrs:{dragable: true,dragstart:"parameter_dragStart(event, 'output', this)"},
					events:{}
				},
				rows:rows
			}

			let output_table = new UI.HtmlTable(container_fluid, table_data);
			new UI.FormControl(container_fluid, 'br', {});

			attrs={class:'btn btn-primary',id:'savefunction',innerHTML:'Save',style:'margin-bottom:10px;'}
			events={
				click: function(){
					let inputs = [];
					let outputs = [];
					$('#parameter_input_table').find('tr').each(function(){
					//	console.log($(this))
						let input = {
							name: $(this).find('.parameter_name').val(),
							datatype: $(this).find('.parameter_datatype').val(),
							list: $(this).find('.parameter_list').is(':checked'),
							default: $(this).find('.parameter_default').val()
						}
					//	console.log(input)
						inputs.push(input);
					})
					$('#parameter_output_table').find('tr').each(function(){
						let output = {
							name: $(this).find('.parameter_name').val(),
							datatype: $(this).find('.parameter_datatype').val(),
							list: $(this).find('.parameter_list').is(':checked'),
							default: $(this).find('.parameter_default').val()
						}
					//	console.log(output)
						outputs.push(output);
					})
					console.log($('#parameter_input_table'),inputs,$('#parameter_output_table'),outputs)
					that.flowobj.inputs = inputs;
					that.flowobj.outputs = outputs;
	
					that.item_panel.innerHTML  = "" 
					that.item_panel.style.width = "0px";
					that.item_panel.style.display = "none";
				}
			}
			new UI.FormControl(container_fluid, 'button', attrs, events);

			attrs={class:'btn btn-danger',id:'closefunction',innerHTML:'Close',style:'margin-bottom:10px;'}
			events={click: function(){
				that.item_panel.innerHTML  = ""
				that.item_panel.style.width = "0px";
				that.item_panel.style.display = "none";
			}}
			new UI.FormControl(container_fluid, 'button', attrs, events);
			
			that.item_panel.style.width = "500px";
			that.item_panel.style.display = "block";
		}
		show_Sessions(){
			//this.item_panel
			let that = this;
			this.item_panel.innerHTML  = "" 
			var divsToRemove = this.item_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}
			let attrs={class: 'container-fluid',style: 'width: 100%;height:95%;margin-left:10px;margin-right:10px;'}
			let container_fluid = (new UI.FormControl(this.item_panel, 'div', attrs)).control;
			
			attrs={class: 'btn btn-danger', id: 'closefunction', innerHTML:'X',style: 'float:right;top:2px;right:2px;position:absolute;'}
			let events={click: function(){
				that.item_panel.style.width = "0px";
				that.item_panel.style.display = "none";
				that.item_panel.innerHTML  = "" }};
			new UI.FormControl(container_fluid, 'button', attrs, events);
			new UI.FormControl(container_fluid, 'h2', {innerHTML:'Sessions Management'});
			new UI.FormControl(container_fluid, 'br', {});
			
			new UI.FormControl(container_fluid, 'h3', {innerHTML:'System Sessions'});
			new UI.FormControl(container_fluid, 'br', {});
			
			let rows=[];
			Function_System_Sessions.forEach(function(session){
				let cells = [];
				cells.push({data:{innerHTML:session}});
				rows.push(cells);
			})
			let table_data={
				attrs:{class: 'table table-bordered', id: 'systemsessiontable', style: 'width: 100%;'},
				headers:[{innerHTML:'Session Name',style:"width:100%;"}],
				columns:[{control:'', attrs:{} }],
				tr:{
					attrs:{dragable: true,dragstart:"session_dragStart(event, 'system')"},
					events:{}
				},
				rows: rows
			}
			new UI.HtmlTable(container_fluid, table_data);
			new UI.FormControl(container_fluid, 'br', {});

			new UI.FormControl(container_fluid, 'h3', {innerHTML:'User Sessions'});
			new UI.FormControl(container_fluid, 'br', {});

			rows=[];
			Object.keys(this.Calculate_UserSessions()).forEach(function(obj,index){
				let cells = [];
				cells.push({data:{innerHTML:obj}});
				rows.push(cells);	
			})
			table_data={
				attrs:{class: 'table table-bordered', id: 'usersessiontable', style: 'width: 100%;'},
				headers:[{innerHTML:'Session Name',style:"width:100%;"}],
				columns:[{control:'', attrs:{} }],
				tr:{
					attrs:{dragable: true,dragstart:"session_dragStart(event, 'user')"},
					events:{}
				},
				rows: rows
			}
			new UI.HtmlTable(container_fluid, table_data);
			
			that.item_panel.style.width = "300px";
			that.item_panel.style.display = "block";
			
		}

		Calculate_UserSessions(){
			let flowobj = this.flowobj;
			let UserSessions = {}

			flowobj.functiongroups.forEach(function(obj,index){
				if(!obj.hasOwnProperty('functions'))
					obj.functions = [];

				obj.functions.forEach(function(obj1,index1){
					if(!obj1.hasOwnProperty('inputs'))
						obj1.inputs = [];

					obj1.inputs.forEach(function(obj2,index2){
						if(obj2.source == '3')
							UserSessions[obj2.aliasname] = obj2.aliasname;
					})

					if(!obj1.hasOwnProperty('outputs'))
						obj1.outputs = [];

					obj1.outputs.forEach(function(obj2,index2){
						if(Array.isArray(obj2.outputdest)){
							console.log(obj2.outputdest)

							if(!obj2.hasOwnProperty('outputdest'))
								obj2.outputdest = [];

							obj2.outputdest.forEach(function(obj3,index3){
								console.log(obj3)
								if(obj3 == '1' && obj2.aliasname[index3] != '')
									UserSessions[obj2.aliasname[index3]] = obj2.aliasname[index3];
							})														
						}
						else if(obj2.outputdest == '1' && obj2.aliasname != '')
							UserSessions[obj2.aliasname] = obj2.aliasname;
					})
				})

			})
			return UserSessions;

		}

		trigger_event(event, args) {
			if (this.options['on_' + event]) {
				this.options['on_' + event].apply(null, args);
			}
		}		
	}

	
	return ProcessFlow;
	
}());

Function_System_Sessions =["UTCTime", "LocalTime", "UserNo","UserID", "WorkSpace"]

function session_dragStart(event, type) {
	event.dataTransfer.effectAllowed='move';
	event.dataTransfer.setData("variable", event.target.textContent);
	event.dataTransfer.setData("type", type);
	event.dataTransfer.setData("category", "session");
}
function parameter_dragStart(event, type, el){

	console.log(event, type, el)
}

const UIFlowlibraryLoadedEvent = new CustomEvent('UIFlow_libraryLoaded');

document.dispatchEvent(UIFlowlibraryLoadedEvent);