var UIJEditor= UIJEditor || {};
(function(UIJEditor){
    class JsonEditor{
        constructor(container, data, title, options){
            this.container = container || "json-editor-container";
            this.options = options || {};
            this.data = data || {};
            this.title = title || "JSON Editor";
            
            this.jdata = (new UI.JSONManager(this.data, this.options));
            this.create_document();
            this.create_tree();
        }
        create_document(){
            let that =this
            this.jstreecontainer = document.getElementById(this.container);
            if(!this.jstreecontainer)
                this.jstreecontainer = (new UI.FormControl(document.body, 'div', {id: this.container, style:"width:100%;height:100%;overflow:auto;"})).control;           
           

            let attrs = {
                id:"json-editor-actions",
                display:"inline-block",
                class:"row",
            }
            let action_bar = (new UI.FormControl(this.jstreecontainer, 'div', attrs)).control;
            new UI.FormControl(action_bar, 'input', {id: "json-editor-search", placeholder:"Search"});
            let events={
                click: function(){
                    that.showRedlines();
                }
            }
            new UI.FormControl(action_bar, 'button', {id: "json-editor-saveredline", innerHTML:"Changes"},events);
            new UI.FormControl(action_bar, 'button', {id: "json-editor-save", innerHTML:"Save"}, {});
            
            
        }
        create_tree(){
            let that = this           
            new UI.FormControl(this.jstreecontainer, 'div', {id: "json-editor-tree"}, {});

            let rootdata ={           
                   text: this.title,
                   state: { opened: true },
                   children: this.jdata.formatJSONforjstree(this.options),
             }

             let plugins = [];
             let editable = this.options.editable || false;
             let showlabelonly = this.options.showlabelonly || false;

            
             if(!showlabelonly && editable)
                plugins = ["contextmenu", "dnd", "search","state", "types"]
             else
                plugins = ["search","state", "types"]

                plugins = ["contextmenu"]
            this.tree = $('#json-editor-tree').jstree({
                  'core': {
                    "animation" : 0,
                    "check_callback" : true,
                    "themes" : { "stripes" : true },
                    'data': rootdata,
                   /* check_callback: function (operation, node, node_parent, node_position, more) {
                        console.log('callback:', operation, node, node_parent, node_position, more)
                        if (operation === 'select_node') {
                            console.log('click node', node)
                          return false; // Disable click event
                        }
                        return true; // Allow other events
                      } */
                  },
                  "types" : {
                      "root" : {
                      "valid_children" : ["default"]
                      },
                      "default" : {
                      "valid_children" : ["default","file"]
                      },
                      "file" : {
                      "icon" : "glyphicon glyphicon-file",
                      "valid_children" : []
                      }
                  },
                  'plugins': plugins
                });
                $('#json-editor-tree').on('select_node.jstree', function (e, data) {
                    e.preventDefault(); // Prevent the default select_node event behavior
                    console.log('select node:',data)
                    // Handle your custom logic for the node click event here
                    // For example, you can retrieve the node information using data.node and perform your desired actions
                    let nodeid = data.node.id;
                    $('li.jstree-node').removeClass('jstree_selected_node');
                    $('#'+nodeid).addClass('jstree_selected_node')
                    if($('#'+nodeid).find('input').length > 0)
                        $('#'+nodeid).find('input').focus();
                    else if($('#'+nodeid).find('select').length > 0)
                        $('#'+nodeid).find('select').focus();
                  });

                $('#json-editor-tree').on("loaded.jstree", function() {
                    console.log('loaded.jstree'); 
                    that.attachEvents();
                  });
                

                  $('#json-editor-tree').on("open_node.jstree", function() {
                    console.log('open node.jstree'); 
                    that.attachEvents();
                  });

                $('#json-editor-tree').on("changed.jstree", function (e, data) {
                  console.log('changed.jstree',data); 
                  return false                 
                });
                
                // Handle node edit event
                $('#json-editor-tree').on('rename_node.jstree', function(e, data) {
                  var node = data.node;
                  console.log('Edited node:', data);
                  that.edit_node(data);
                });
                $('#json-editor-tree').on('move_node.jstree', function(e,data) {
                    
                    console.log('move_node:', data);
                    that.move_node(data);
                    
                  });
                $('#json-editor-tree').on('paste.jstree', function(e, data) {
                    
                    console.log('paste.jstree:', data);
                    that.paste_node(data);
                    
                  });
                $('#json-editor-tree').on('cut.jstree', function(e, data) {
                   
                    console.log('cut.jstree:', data);
                    
                  });
                $('#json-editor-tree').on('delete_node.jstree', function(e, data) {                    
                    console.log('delete_node:', data);
                    that.delete_node(data);
                    
                  });
                $('#json-editor-tree').on('create_node.jstree', function(e, data) {                    
                    console.log('create_node:', data);
                    that.create_node(data);
                    
                });
         
        }
        refresh(){
            $('#json-editor-tree').remove();

            this.create_tree();
        }
        move_node(data){
        //    let oldparentid = data.old_parent;
            let parentid = data.parent;
            let nodeid = data.node.id;
         //   let oldparentpath = $('#'+oldparentid).attr('path');
            let newparentpath = $('#'+parentid).attr('path');
            if(newparentpath == undefined)
                newparentpath ="";

            let nodepath = $('#'+nodeid).attr('path');
            let jnode = this.jdata.getNode(nodepath).value;
            console.log(jnode)
            let key = nodepath.includes("/")? nodepath.split("/").pop(): nodepath;
            let newnode = {}
            newnode[key] = jnode;
            console.log('remove node', nodepath, 'insert node', newparentpath, newnode)
            this.jdata.deleteNode(nodepath);
            this.jdata.insertNode(newparentpath, newnode);
            this.refresh();
        }
        delete_node(data){
            let that = this;
            let node = data.node;
            let path = $('#'+node.id).attr('path');
            console.log(path)
            that.jdata.deleteNode(path);
            that.refresh();
        }
        paste_node(data){
            let that =this;
            let change = false;
            console.log(data)
            let parent = data.parent;
            if(data.mode== 'copy_node'){
                
                data.node.forEach(function(node){
                    console.log(node)
                    let nodepath = $('#'+node.id).attr('path');
                    let jnode = that.jdata.getNode(nodepath).value;
                    let path = $('#'+parent).attr('path');
                    if(path == undefined)
                        path ="";
                    console.log(jnode)
                    that.jdata.insertNode(path, jnode);
                    change = true;
                });
            }else if(data.mode== 'move_node'){
                data.node.forEach(function(node){
                    console.log(node)
                    let nodepath = $('#'+node.id).attr('path');
                    let jnode = that.jdata.getNode(nodepath).value;
                    let path = $('#'+parent).attr('path');
                    if(path == undefined)
                        path ="";
                    
                    that.jdata.deleteNode(nodepath);
                    that.jdata.insertNode(path, jnode);
                    
                    change = true;
                });

            }

            if(change)
                that.refresh();
        }
        create_node(data){
            console.log(data)
            let that = this;
            let parent = data.parent;
            let node = data.node;
            let path = $('#'+parent).attr('path')
            if(path == undefined)
                path = "";
            
            let jnode = that.jdata.getNode(path);
            console.log(path, jnode)
            if(jnode){
                if(jnode.isArrayElement){
                    that.jdata.inserNodeKey(path, '');
                    that.refresh();
                    return;
                }
                else if(typeof jnode.value == 'object')
                    that.newnodeparent = parent;
                else
                { 
                    alert('Cannot add child to this node');                   
                    return;
                }
            }
            
        }
        edit_node(data){
            let that = this;
        //    console.log(that, data, that.newnodeparent)
            let parent =  that.newnodeparent;
            let node = data.node;
            let text = node.text;
            let path = $('#'+node.id).attr('path');

            if((path ==undefined || path =='') && parent !=''){
                path = $('#'+parent).attr('path');
                if(path == undefined)
                    path ="";
            }
            else{
          //      console.log(path)
                paths = path.split('/');
                paths.pop();
                path = paths.join('/');
            }

        //    console.log('path ', path, text)
            if(parent !='')
                that.jdata.inserNodeKey(path, text);

            that.refresh();

            that.newnodeparent = ""
        }
        attachEvents(){
            let that = this;
            $('.node_input[type!="checkbox"]').off('change')
            $('.node_input[type="checkbox"]').off('click')
            $('.node_input[type!="checkbox"]').on('change', (event) => {
                
                const element = event.target;
            //    console.log('node change',element, event)
                var $input = $(element);
                var $node = $input.closest('.jstree-node');

                var path = $node.attr('path');
            //    console.log($input, $node)
                if($input.attr('unchangable')== "true"){
                    let orgvalue = that.jdata.getdata(path)
                    $input.val(orgvalue);
                }
                else{
                    var value
                    value = $input.val();
                    $node.addClass('changed');
              //      console.log(path, value)
                    that.jdata.updateNode(path, value);
                }
            })
            $('.node_input[type="checkbox"]').on('click', (event) => {
                const element = event.target;
             //   console.log('node change',element, event)
                var $input = $(element);
                var $node = $input.closest('.jstree-node');
                var path = $node.attr('path');
                var value
                if($input.prop('checked'))
                    value = true;
                else
                    value = false;
                $node.addClass('changed');
               // console.log(path, value)

                that.jdata.updateNode(path, value);
                that.refresh();
            })
/*
            $.contextMenu({
				selector: '.jstree-node', 
				build:function($triggerElement,e){
					that.disable_paperevents();
					console.log('build the contextmenu:',$triggerElement,e,$triggerElement[0].getAttribute('id'))
					let modelid = $triggerElement[0].getAttribute('id');
					return{
						callback: function(key, options,e){
							console.log(key, options,e)
							switch(key){

								
							}

						}, 
						items:{
							'Create':{
								name: 'Create',
								icon: 'fa-plus',
								disabled: false
							},
							'Delete':{
								name: 'Deletee',
								icon: 'fa-minus',
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
								
			}) */
        }
        
        getChanges(){
            return this.jdata.getChanges();
        }

        showRedlines(){
            this.jdata.showRedlines();
        }
    }
    UIJEditor.JsonEditor = JsonEditor;
})(UIJEditor || {})