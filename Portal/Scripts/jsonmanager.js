var UI = UI || {};
(function (UI) {

    class JSONManager {
        constructor(jsonObject, options = { allowChanges: true, schemafile:"", schema:null }) {
            this.originalObject = JSON.parse(JSON.stringify(jsonObject));        
          this.data = (typeof jsonObject == 'string')? JSON.parse(jsonObject) : jsonObject;
          this.options = options;
          this.changed = false;
          this.schemafile = options.schemafile || '';
          this.schema = this.options.schema || null;
          this.loadschema();
        //  console.log(this.schemafile,this.schema)
        }
        get(url, stream) {
            return new Promise((resolve, reject) => {
                const xhr = new XMLHttpRequest();
                xhr.open('GET', `${url}`, true);                
            //    xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);
                if (stream) {
                  xhr.responseType = 'stream';
                }
                xhr.onload = () => {
                  if (xhr.status >= 200 && xhr.status < 300) {
                    resolve(xhr.response);
                  } else {
                    reject(xhr.statusText);
                  }
                };
                xhr.onerror = () => reject(xhr.statusText);
                xhr.onabort = () => reject('abort');
                xhr.send();
              });
        }
        loadschema(){
            let that = this;
            if(!this.schema &&  this.schemafile !=''){
                let ajax = new UI.Ajax("");
                ajax.get(this.schemafile,false).then((response) => {
                    that.schema = JSON.parse(response);
                 //   console.log(this.schemafile,response)
                    that.getSchemaDefinitions();
                }).catch((error) => {
                    console.log(error);
                })
            }
            else if(this.schema !={})
                that.getSchemaDefinitions();
            
        }
        getSchemaDefinitions(){
        //    console.log(this.schema)
            this.schemaRootNode = null
            if(this.schema !=null && this.schema !={}){
              if(this.schema.hasOwnProperty('definitions') && this.schema.hasOwnProperty("$ref")){
                this.schemaDefinitions = this.schema['definitions'];
                let rootpath = this.schema["$ref"];
                if(rootpath.startsWith('#/'))
                    rootpath = rootpath.replace('#/','')
                
                let paths = rootpath.includes('/')? rootpath.split('/'):[rootpath];
                let currentNode = this.schema
                for(var i=0;i<paths.length;i++){
                    let path = paths[i];
                    currentNode =currentNode[path]        
                }
                this.schemaRootNode = currentNode;
                }
            }
               
        }
        addNode(path, Value) {
            this.insertNode(path, Value);
        }
        inserNodeKey(path, key){
            if (!this.options.allowChanges) {
                console.log("Modifications are not allowed.");
                return;
            } 
            const node = this.getNode(path);
            if (node) {
            //    console.log(node,key)
                if (node.isArrayElement) {
                    node.value.push({});
                } else {
                    node.value[key] = null;
                }
            //    console.log(node)
                this.changed = true;
            }
        }
        insertNode(path, Value) {
            if (!this.options.allowChanges) {
                console.log("Modifications are not allowed.");
                return;
            } 
           
            const node = this.getNode(path);
            let isvalidJson = this.isValidJSON(Value);
            if (node) {
             //   console.log('insert node value:',node, isvalidJson, path, Value)
                if (Array.isArray(node.value)) {  
                                
                    if(isvalidJson){
                        let valueobj = null;                     

                        if(typeof Value == 'string'){
                            valueobj = JSON.parse(Value);
                        }else{
                            valueobj = Value;
                        }  
                
                        node.value.push(valueobj)
                    }
                    else{
                        node.value.push(Value)
                    }
                } else {
                    if(isvalidJson){
                        let valueobj = null;
                        if(typeof Value == 'string')
                            valueobj = JSON.parse(Value);
                        else
                            valueobj = Value;
                        
                     //   console.log(valueobj)
                        for (const key in valueobj) {
                           // if (valueobj.hasOwnProperty(key)) {
                                const element = valueobj[key];
                                node.value[key] = element;
                           // }
                            

                        }
                    }
                    else{
                        node.value = Value;
                    }              
                }

                this.changed = true;
            }
        }
         // {}
        updateNode(path, newValue) {
            this.updateNodeValue(path,newValue) ;  
        }
        updateNodeValue(path, newValue) {
          if (!this.options.allowChanges) {
            console.log("Modifications are not allowed.");
            return;
          }
        //  console.log('updateNodeValue',path, newValue )
          let isValueObj =this.isValidJSON(newValue);
        
          if(isValueObj){
            path = path
          }else{
            let paths = path.split('/');
            let lastPath = paths[paths.length-1];
            paths.pop();
            path = paths.join('/');            
            newValue = JSON.parse(`{"${lastPath}":"${newValue}"}`);
            
         //   console.log(newValue);
         }
          const node = this.getNode(path);
          if (node) {
            let valueobj = null;
            if(typeof newValue == 'string')
                valueobj = JSON.parse(newValue);
            else
                valueobj = newValue;
            for (const key in valueobj) {
                if (valueobj.hasOwnProperty(key)) {
                   const element = valueobj[key];
                    node.value[key] = element;
                }
            }

            /*
            console.log('find node:',node, path)
            if (node.isArrayElement) {
              const arrayNode = this.getNode(node.arrayPath);
              if (arrayNode) {
                    let valueobj = null;
                    if(typeof newValue == 'string')
                        valueobj = JSON.parse(newValue);
                    else
                        valueobj = newValue;
                //    console.log(arrayNode, newValue)
                    for (const key in valueobj) {
                        if (valueobj.hasOwnProperty(key)) {
                            const element = valueobj[key];
                            arrayNode.value[key] = element;
                        }
                    }
              }
            } else {
                              
                    let valueobj = null;
                    if(typeof newValue == 'string')
                        valueobj = JSON.parse(newValue);
                    else
                        valueobj = newValue;

                    console.log(node, valueobj)
                    for (const key in valueobj) {
                        if (valueobj.hasOwnProperty(key)) {
                            const element = valueobj[key];
                            node.value[key] = element;
                        }
                    }
            } */
            this.changed = true;
          } else {
            console.log("Invalid path:", path);
          }
        }

        deleteNode(path) {
            if (!this.options.allowChanges) {
                console.log("Modifications are not allowed.");
                return;
            }
            
            const node = this.getNode(path);
            if (node) {   
                        
                let parentpath = path.includes("/")? path.split('/').slice(0, path.split('/').length-1).join('/'): '';
                
                let pnode = this.getNode(parentpath);
               
                if (node.isArrayElement) {
                    pnode.value.splice(node.index, 1);
                    
                } else {
                    let key = path.includes("/")? path.split('/').pop() : path;                    
                    delete pnode.value[key];
                    
                }
                this.changed = true;
            } else {
                console.log("Invalid path:", path);
            }
        }

        getdata(path){
            let that = this;
            let node = that.getNode(path);
            if(node){
                return node.value;
            }
            else
                return null;
        }
        getNode(path) {
          if(path.startsWith('#/'))
              path = path.replace('#/','')

          if(path =='')
            return {
                value: this.data,
                isArrayElement: Array.isArray(this.data),
                arrayPath: '',
                index: -1,
              };

       //   console.log('getnode:',path)
          const keys = path.toString().includes("/")? path.toString().split("/"): [path];
          let currentNode = this.data;
          let arrayPath = null;
          let isArrayElement = false;
          let index = -1;
      //    console.log(keys)
          for (let i = 0; i < keys.length; i++) {
            const key = keys[i];
         //   console.log(i,currentNode,key, this.isValidJSON(key))
            isArrayElement = Array.isArray(currentNode)
            if(this.isValidJSON(key)){

                let keyobj = JSON.parse(key);
                let findobj = this.findNodeByKeys(currentNode, keyobj);
               
                if(findobj){
                    currentNode = findobj.value;
                    arrayPath = keys.slice(0,i).join('/');
                    isArrayElement = isArrayElement
                    index = findobj.index;
                }
                else{
                    return null;
                }

            }
            else{
            //    console.log(currentNode,key)
                let findobj = this.findNodebykey(currentNode,key);
                if(findobj){
                    currentNode = findobj.value;
                    arrayPath = keys.slice(0,i).join('/');
                    isArrayElement = findobj.isArrayElement
                    index = findobj.index;
                }
                else{
                    return null;
                }
            }
          }
          
          return {
            value: currentNode,
            isArrayElement: isArrayElement,
            arrayPath: arrayPath,
            index: index,
          };
        }

        isValidJSON(str){
            if(str == null || str == undefined)
                return false;
            if(typeof str == 'object')
                return true;
            let obj = null;
            try {
               obj =  JSON.parse(str);
            } catch (e) {
                return false;
            }

            return typeof obj == 'object';
        }
        findNodebykey(jsonObject,key){
        //    console.log('findNodebykey',jsonObject,key, Array.isArray(jsonObject),jsonObject.hasOwnProperty(key))
            if(Array.isArray(jsonObject)){
                try {
                    let index = parseInt(key);
                    if(!isNaN(index))
                        return {
                            value: jsonObject[index],
                            isArrayElement: true,
                            arrayPath: '',
                            index: index,
                        };
                }
                catch(e){
                    return null;
                }

            }
            else if(jsonObject.hasOwnProperty(key))
                return {
                    value: jsonObject[key],
                    isArrayElement: false,
                    arrayPath: '',
                    index: -1,
                };
            else
                return null;
        }
        findNodeByKeys(jsonObject, criteria) {
            if(Array.isArray(jsonObject)){
                for (const key in jsonObject) {
                    if (jsonObject.hasOwnProperty(key)) {
                        const node = jsonObject[key];
                        let found = true;
                  
                        for (const searchKey in criteria) {
                          if (criteria.hasOwnProperty(searchKey) && node[searchKey] !== criteria[searchKey]) {
                            found = false;
                            break;
                          }
                        }
                  
                        if (found) {
                          return {
                            value: node,
                            isArrayElement: Array.isArray(node),
                            arrayPath: parseInt(key),
                            index: parseInt(key),
                          };
                        }
                    }
                }

            }
            else{
                let found = true;
                for (const searchKey in criteria) {
                    if (criteria.hasOwnProperty(searchKey) && jsonObject[searchKey] !== criteria[searchKey]) {
                      found = false;
                      break;
                    }
                }
                if (found) {
                    return {
                      value: jsonObject,
                      isArrayElement: false,
                      arrayPath: null,
                      index: -1,
                    };
                } 

            }
         
            return null;
          }
        getPropertiesFromSchema(path){
          if(!this.schema || this.schema =={})
            return null;

          if(this.schemaRootNode == null){
                this.getSchemaDefinitions();
          }
          if(path.startsWith('#/'))
              path = path.replace('#/','')

          if(path =='')
            return {
                node:this.schemaRootNode
            }
         //  functiongroups/functions/inputs/datatype
          const keys = path.toString().includes("/")? path.toString().split("/"): [path];
          let Properties = {};
          let schemaNode = this.schemaRootNode
        //  console.log(schemaNode, path, keys)
          let isArray = false;
          for(var i=0;i<keys.length;i++){
            let key = keys[i]
            console.log(this.schema,schemaNode,key)
            if(!schemaNode.properties.hasOwnProperty(key))
                return null;

            Properties = schemaNode.properties[key];
         //   console.log(key, Properties)
            if(Properties.hasOwnProperty('$ref') || (Properties.hasOwnProperty('type') && Properties['type'] == 'array')){
                let nodepath = '';
                
                if(Properties.hasOwnProperty('type') && Properties['type'] == 'array')
                    isArray =true;
                else 
                    isArray = false;

                if(Properties.hasOwnProperty('$ref'))
                    nodepath = Properties['$ref']                
                else if(Properties['items'].hasOwnProperty('$ref') && (i<keys.length-1))
                    nodepath = Properties['items']['$ref'];
                else 
                    return {
                        node:schemaNode,
                        properties: Properties,
                        isArray: isArray
                    }

                if(nodepath.startsWith('#/'))
                    nodepath = nodepath.replace('#/','')
                
                let paths = nodepath.includes('/')? nodepath.split('/'):[nodepath];
            //    console.log(paths)
                let currentNode = this.schema
                for(var j=0;j<paths.length;j++){
                    let path1 = paths[j];
                    currentNode =currentNode[path1]                    
                }
           //    console.log(i, key, keys,schemaNode,currentNode)
                if(currentNode.hasOwnProperty('type')){
                    if(currentNode['type'] == 'object'){
                        if(i == keys.length-1){
                            let pro =  currentNode.hasOwnProperty('properties')? currentNode['properties']: currentNode
                            pro = Object.assign(pro, Properties)
                            return {
                                node: schemaNode,
                                properties: pro,
                                isArray: isArray 
                            }
                        }else
                            schemaNode = currentNode;
                    }
                    else{
                        let pro =  currentNode.hasOwnProperty('properties')? currentNode['properties']: currentNode
                    //    console.log(key, path, pro, Properties)
                        pro = Object.assign(pro, Properties)
                        return{
                            node: schemaNode,
                            properties: pro,
                            isArray: isArray
                        }
                    }
                }else if(i == keys.length-1)
                    return {
                        node: schemaNode,
                        properties: pro,
                        isArray: isArray 
                    }

                schemaNode = currentNode;
            }
            else{
                return{
                    node: schemaNode,
                    properties: Properties,
                    isArray: isArray
                }
            }
          }
          return null;

        }
        getSchemaDefinition(path){

          if(!this.schema || this.schema =={})
            return null;
          
          if(path.startsWith('#/'))
              path = path.replace('#/','')

          if(path =='')
            return this.schemaRootNode

          const keys = path.toString().includes("/")? path.toString().split("/"): [path];
          
          let Properties = {};
          let schemaNode = this.schemaRootNode
        //  console.log('get the schema definition:',path)

          for(var i=0;i<keys.length;i++){
            let key = keys[i]
         //   console.log(schemaNode)
            if(!schemaNode.properties.hasOwnProperty(key))
                return null;

            Properties = schemaNode.properties[key];
          //  console.log(key, Properties)
            if(Properties.hasOwnProperty('$ref') || (Properties.hasOwnProperty('type') && Properties['type'] == 'array')){
                let nodepath = '';
                if(Properties.hasOwnProperty('$ref'))
                    nodepath = Properties['$ref']                
                else if(Properties['items'].hasOwnProperty('$ref'))
                    nodepath = Properties['items']['$ref'];
                else 
                    return  schemaNode              

                if(nodepath.startsWith('#/'))
                    nodepath = nodepath.replace('#/','')
                
                let paths = nodepath.includes('/')? nodepath.split('/'):[nodepath];
          //      console.log(paths)
                let currentSchemaDefinition = this.schema
                for(var j=0;j<paths.length;j++){
                    let path1 = paths[j];
                    currentSchemaDefinition =currentSchemaDefinition[path1]                    
                }
             //   console.log(currentSchemaDefinition, i, keys)

                if(currentSchemaDefinition.hasOwnProperty('type')){
                    if(currentSchemaDefinition['type'] == 'object'){
                        if(i== keys.length -1)
                            return currentSchemaDefinition;
                        else
                            schemaNode = currentSchemaDefinition;
                    }
                    else{
                        return schemaNode
                    }
                }
                else if(i== keys.length -1)
                    return currentSchemaDefinition;

                schemaNode = currentSchemaDefinition;
            }
            else{
                return schemaNode
            }
          }
          return null;           
        }
        
        formatJSONforjstree(options ={} ) {
            return this.formatJsonfortreebyNode(this.data,'#',0, options, '','');
        }
        formatJsonfortreebyNode(node, parent, level, inoptions, path, schemapath){
            if(node == null || node == undefined || typeof node != 'object')
                return [];
            let options = inoptions || {}; 
            let openlevel = options.openlevel || -1;
            let editable = options.editable || false;
            let showlabelonly = options.showlabelonly || !editable;

            let fmtdata =[];
            let nodedefinition = this.getSchemaDefinition(schemapath)
            let requiredfields = [];
            let hiddenfields =[];
            let unchangablefields = [];

            if(nodedefinition){
                if(nodedefinition.hasOwnProperty('required'))
                    requiredfields = nodedefinition['required']
                if(nodedefinition.hasOwnProperty('hidden'))
                    hiddenfields = nodedefinition['hidden']
                if(nodedefinition.hasOwnProperty('unchangable'))
                    unchangablefields = nodedefinition['unchangable']     

            }
            let nodekeys =[];
            let isarray = Array.isArray(node)

            for (const key in node) { 

                nodekeys.push(key)

                let newpath = path ==''? key: path + '/'+ key; 
                
                
                let newschemapath = isarray? schemapath : (schemapath ==''? key : schemapath +'/'+ key)

                let item = this.build_treenode(newschemapath, key, node[key],requiredfields,hiddenfields, unchangablefields,isarray,level,options,newpath,showlabelonly,editable,openlevel , false)
                
                
                fmtdata.push(item);

            }
            if(!isarray){
                for(var i=0;i<requiredfields.length;i++){
                    var find = false;
                    for(var j=0;j<nodekeys.length;j++)
                        if(nodekeys[j] == requiredfields[i]){
                            find = true;
                            break;
                        }
                    let key = requiredfields[i]
                    if(!find){
                        isarray = false
                        if(nodedefinition.properties.hasOwnProperty(key))
                            if(nodedefinition.properties[key].hasOwnProperty('type'))
                                if(nodedefinition.properties[key]['type'] == 'array')
                                    isarray = true 
                        
                        let newpath = path ==''? key: path + '/'+ key; 
                        let newschemapath = isarray? schemapath : (schemapath ==''? key : schemapath +'/'+ key)
                        let item = this.build_treenode(newschemapath, key, '',requiredfields,hiddenfields, unchangablefields,isarray,level,options,newpath,showlabelonly,editable,openlevel, true)               
                    
                        fmtdata.push(item);
                    }

                }
            }
            
            return fmtdata;
        }
        build_treenode(newschemapath,key, value,requiredfields,hiddenfields, unchangablefields,isarray,level,options,newpath,showlabelonly, editable,openlevel, ismissed = false){
            let schemadata =null;

            let required = true;
            let hidden = false;
            let unchangable = false; 
            let invalidateNode = false;         
            

            let datatype = "string";
            let lng = {};
            let schemaoptions = null;    
            
            let lngcode =""; 
            let lngdefault = key
        //    console.log(lng,lngcode, lngdefault)

            if(this.schema && this.schema!={}){
                if(!isarray )
                    schemadata = this.getPropertiesFromSchema(newschemapath)

                if(!isarray && schemadata == null ){
                    invalidateNode = true;
                }                    

                if(schemadata !=null){               

             
                    if(schemadata.properties.hasOwnProperty("type"))
                    datatype = schemadata.properties["type"]
                
                    if(schemadata.properties.hasOwnProperty("lng"))
                        lng = schemadata.properties["lng"]
                    
                    if(schemadata.properties.hasOwnProperty("options"))
                        schemaoptions = schemadata.properties["options"] 
                    
                    if(schemadata.isArray)
                        isarray = true;
                }
                
            //    console.log(key,requiredfields,hiddenfields, unchangablefields, invalidateNode, schemadata)
                if(schemadata !=null){
                    if(requiredfields.find(item => item == key) != undefined)
                        required = true;
                    else 
                        required = false;
                    
                    if(hiddenfields.find(item => item == key) != undefined)
                        hidden = true;
                    else   
                        hidden = false;
                    
                    if(unchangablefields.find(item => item == key) != undefined)
                        unchangable = true;
                    else
                        unchangable = false;
                    
                //    console.log('parsed value:', key,required, hidden,unchangable)
                }

                if(lng !=null && lng !={}){
                    if(lng.hasOwnProperty('code'))
                        lngcode = lng['code'];
                    if(lng.hasOwnProperty('default'))
                        lngdefault=lng["default"];
                }
            }
            let id = this.generateUUID();
                let newparent = id

                let children = [];
                if(!ismissed)
                    children = this.formatJsonfortreebyNode(value, newparent, level+1, options, newpath, newschemapath);
               
                 
                let nodeeditablevalue ='';
                
            //    console.log(lng,lngcode, lngdefault)
                let nodelabelvalue = '<label for="node_'+id+'" key="'+ key +'" lngcode="'+lngcode+'">'+lngdefault+'</label>'

                if(children.length > 0){
                    nodeeditablevalue ='';
                }else if(showlabelonly){
                    nodeeditablevalue = ':'+value;
                }
                else if(datatype == 'boolean'){
                    try{
                        if(value == 'true' || value == true)
                            nodeeditablevalue = '<input class="node_input"  type="checkbox" '+ ((!editable || unchangable)? 'disabled' : '') +' id="node_'+id+'" checked></input>'
                        else 
                            nodeeditablevalue = '<input class="node_input"  type="checkbox" '+((!editable || unchangable)? 'disabled' : '' )+' id="node_'+id+'"></input>'                    
                    }catch(e){
                        nodeeditablevalue = '<input class="node_input"  type="checkbox" '+((!editable || unchangable)? 'disabled' : '') +' id="node_'+id+'"></input>'  
                    }
                }else if(schemaoptions && schemaoptions !={} ){
                    let values = schemaoptions['value']
                    let optionlngcodes  = schemaoptions['lngcode']  
                    let optiondefaults = schemaoptions['default']

                    if(Array.isArray(values)){
                        nodeeditablevalue =  '<select class="node_input"  '+((!editable || unchangable)? 'disabled' : '') +' id="node_'+id+'" >'
                        for(var n=0;n<values.length;n++){
                         //   console.log(node[key],values[n],optionlngcodes[n], optiondefaults[n])
                            nodeeditablevalue += '<option value="'+values[n]+'" lngcode="'+optionlngcodes[n]+'" '+(value == values[n]? 'selected':'') +' >' + optiondefaults[n] + '</option>'
                        }
                        nodeeditablevalue += '</select>'
                    }
                    else
                        nodeeditablevalue =  '<input  type="text" class="node_input" '+((!editable || unchangable)? 'disabled' : '') +' id="node_'+id+'" value="'+value+'"></input>'     
                }else if(datatype != "array" && datatype != "object"){ 
                    nodeeditablevalue = '<input type="text" class="node_input" unchangable="'+unchangable+'" '+((!editable || unchangable)? 'disabled' : '') +' id="node_'+id+'" value="'+value+'"></input>' 
                }else
                    nodeeditablevalue = '';
               
            //    console.log(schemaoptions,editable, unchangable, showlabelonly,key,node[key], nodelabelvalue, nodeeditablevalue)
                let nodevalue = nodelabelvalue + nodeeditablevalue;

                let item ={
                    id:id,
                    text: nodevalue ,
                    parent:parent,
                    state:{
                        opened:level < openlevel || openlevel<0,
                    },
                    children: children,
                    li_attr: {path:newpath, nodestatus: invalidateNode? 'invalidate_node':'validate_node', hidenode: hidden, missed: ismissed, nodetype: datatype },
                 //   a_attr: {data:nodevalue}
                }

                return item;

        }
        ShowTree(){
            
            $('#ui_left_float_panel').remove();

            let attrs = {
				'class':'ui_left_float_panel',
				'id':'ui_left_float_panel',
				'style':'width:0px;height:100%;float:left;position:absolute;top:0px;left:0px;background-color:lightgrey;overflow:auto;' +
								'border-left:2px solid #ccc;resize:horizontal;z-index:9'
			}
            this.item_panel = (new UI.FormControl(document.body, 'div', attrs)).control;

            let that = this;
			this.item_panel.innerHTML  = "" 
			var divsToRemove = this.item_panel.getElementsByClassName("container-fluid");
			while (divsToRemove.length > 0) {
				divsToRemove[0].parentNode.removeChild(divsToRemove[0]);
			}
			attrs={class: 'container-fluid',style: 'width: 90%;height:95%;margin-left:10px;margin-right:10px;'}
			let container_fluid = (new UI.FormControl(this.item_panel, 'div', attrs)).control;
			
			attrs={class: 'btn btn-danger', id: 'closefunction', innerHTML:'X',style: 'float:right;top:2px;right:2px;position:absolute;'}
			let events={click: function(){
				that.item_panel.style.width = "0px";
				that.item_panel.style.display = "none";
				that.item_panel.innerHTML  = "" }};
			new UI.FormControl(container_fluid, 'button', attrs, events);
			new UI.FormControl(container_fluid, 'div', {id:'ui-json-object-tree',class:'tree',style:'width:100%;height:100%;'});
			that.item_panel.style.width = "350px";
			that.item_panel.style.display = "flex";
			var options = {
				showlabelonly:true,
				editable:true,
				openlevel: -1
			}
            let title = 'root'
            if(that.data.hasOwnProperty('name')){
                title = that.data['name']
                if(that.data.hasOwnProperty('version'))
                    title += ' - '+that.data['version']

            }else if(that.schema && that.schema !={}){
                if(that.schema.hasOwnProperty('$ref'))
                {
                    let ref = that.schema['$ref']
                    refs = ref.split('/')
                    title = refs[refs.length-1]
                }
            }
			let rootdata ={
				text: title,
				state: { opened: true },
				children: that.formatJSONforjstree(that.data),
			}
			
			$(function() {
			  $('#ui-json-object-tree').jstree({
				'core': {
				  'data': rootdata
				}
			  });		
			});  

        }
        ExportJSON(){          
            return JSON.stringify(this.data);
        }
        generateUUID(){
            var d = new Date().getTime();
            var uuid = 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
                var r = (d + Math.random()*16)%16 | 0;
                d = Math.floor(d/16);
                return (c=='x' ? r : (r&0x3|0x8)).toString(16);
            });
            return uuid;    
        }
        showRedlines(){
			let attrs = {
                id:"json-redlines-pop",
                class:"flow-popup-panel",
                style:"max-width: 100%;max-height: 100%;width: 80%;height: 80%;z-index:9;"
            }
            let panel = (new UI.FormControl(document.body, 'div', attrs)).control;
            
            new UI.FormControl(panel, 'h3', {innerHTML:"Changes redlines"});
            let events ={
                click: function(){                    
                    $('#json-redlines-pop').remove();
                }
            }
            attrs={class: 'btn btn-danger', id: 'closefunction', innerHTML:'X',style: 'float:right;top:2px;right:2px;position:absolute;'}

            new UI.FormControl(panel, 'button', attrs, events);

            new UI.FormControl(panel, 'div', {id:"json-diff-editor", style:"height: 90%; width: 100%;"});
            // Initialize CodeMirror with diff highlighting

              const diffEditor = CodeMirror.MergeView(document.getElementById("json-diff-editor"), {
                value: JSON.stringify(this.originalObject, null, 2),
                orig: JSON.stringify(this.data, null, 2),
                lineNumbers: true,
                readOnly: true,
                mode: "application/json"
            });
            let width = $('#json-redlines-pop').width();
            let height = $('#json-redlines-pop').height()-50;
            var container = diffEditor.editor().display.wrapper.parentElement;
            $('.CodeMirror-merge').height('100%')
            $('.CodeMirror').height('100%')
            /*
            // Set the width and height of the container element
            container.style.width = (width-50) + 'px';
            container.style.height = (height-50)+'px'; */
		}
        getChanges() {
            const changes = {};
            this.compareObjects(this.originalObject, this.data, changes);
            return changes;
          }
        
          compareObjects(original, updated, changes, path = "") {
            for (let key in updated) {
              if (updated.hasOwnProperty(key)) {
                const updatedValue = updated[key];
                const originalValue = original[key];
        
                if (updatedValue !== originalValue) {
                  const currentPath = path ? `${path}.${key}` : key;
                  if (typeof updatedValue === "object" && typeof originalValue === "object") {
                    if (Array.isArray(updatedValue)) {
                      this.compareArray(originalValue, updatedValue, changes, currentPath);
                    } else {
                      this.compareObjects(originalValue, updatedValue, changes, currentPath);
                    }
                  } else {
                    changes[currentPath] = {
                      oldValue: originalValue,
                      newValue: updatedValue,
                    };
                  }
                }
              }
            }
          }
        
          compareArray(original, updated, changes, path) {
            if (original.length !== updated.length) {
              changes[path] = {
                oldValue: original,
                newValue: updated,
              };
              return;
            }
        
            for (let i = 0; i < updated.length; i++) {
              if (typeof updated[i] === "object" && typeof original[i] === "object") {
                this.compareObjects(original[i], updated[i], changes, `${path}[${i}]`);
              } else if (original[i] !== updated[i]) {
                changes[`${path}[${i}]`] = {
                  oldValue: original[i],
                  newValue: updated[i],
                };
              }
            }
          }
    }
    UI.JSONManager = JSONManager

    class JSONSchema{
        constructor(schema){
            this.schema = schema || {};
            this.getSchemaRootDefinitions();
        }
        getSchemaRootDefinitions(){
            //    console.log(this.schema)
            this.schemaRootNode = null
            if(this.schema !=null && this.schema !={}){
                if(this.schema.hasOwnProperty('definitions') && this.schema.hasOwnProperty("$ref")){
                    this.schemaDefinitions = this.schema['definitions'];
                    let rootpath = this.schema["$ref"];
                    if(rootpath.startsWith('#/'))
                        rootpath = rootpath.replace('#/','')
                    
                    let paths = rootpath.includes('/')? rootpath.split('/'):[rootpath];
                    let currentNode = this.schema
                    for(var i=0;i<paths.length;i++){
                        let path = paths[i];
                        currentNode =currentNode[path]        
                    }
                    this.schemaRootNode = currentNode;
                }
            }
                   
        }
        getPropertiesFromSchema(path){
            if(!this.schema || this.schema =={})
              return null;
  
            if(this.schemaRootNode == null){
                  this.getSchemaRootDefinitions();
            }
            if(path.startsWith('#/'))
                path = path.replace('#/','')
  
            if(path =='')
              return {
                  node:this.schemaRootNode
              }
           //  functiongroups/functions/inputs/datatype
            const keys = path.toString().includes("/")? path.toString().split("/"): [path];
            let Properties = {};
            let schemaNode = this.schemaRootNode
          //  console.log(schemaNode, path, keys)
            let isArray = false;
            for(var i=0;i<keys.length;i++){
              let key = keys[i]
              console.log(this.schema,schemaNode,key)
              if(!schemaNode.properties.hasOwnProperty(key))
                  return null;
  
              Properties = schemaNode.properties[key];
           //   console.log(key, Properties)
              if(Properties.hasOwnProperty('$ref') || (Properties.hasOwnProperty('type') && Properties['type'] == 'array')){
                  let nodepath = '';
                  
                  if(Properties.hasOwnProperty('type') && Properties['type'] == 'array')
                      isArray =true;
                  else 
                      isArray = false;
  
                  if(Properties.hasOwnProperty('$ref'))
                      nodepath = Properties['$ref']                
                  else if(Properties['items'].hasOwnProperty('$ref') && (i<keys.length-1))
                      nodepath = Properties['items']['$ref'];
                  else 
                      return {
                          node:schemaNode,
                          properties: Properties,
                          isArray: isArray
                      }
  
                  if(nodepath.startsWith('#/'))
                      nodepath = nodepath.replace('#/','')
                  
                  let paths = nodepath.includes('/')? nodepath.split('/'):[nodepath];
              //    console.log(paths)
                  let currentNode = this.schema
                  for(var j=0;j<paths.length;j++){
                      let path1 = paths[j];
                      currentNode =currentNode[path1]                    
                  }
             //    console.log(i, key, keys,schemaNode,currentNode)
                  if(currentNode.hasOwnProperty('type')){
                      if(currentNode['type'] == 'object'){
                          if(i == keys.length-1){
                              let pro =  currentNode.hasOwnProperty('properties')? currentNode['properties']: currentNode
                              pro = Object.assign(pro, Properties)
                              return {
                                  node: schemaNode,
                                  properties: pro,
                                  isArray: isArray 
                              }
                          }else
                              schemaNode = currentNode;
                      }
                      else{
                          let pro =  currentNode.hasOwnProperty('properties')? currentNode['properties']: currentNode
                      //    console.log(key, path, pro, Properties)
                          pro = Object.assign(pro, Properties)
                          return{
                              node: schemaNode,
                              properties: pro,
                              isArray: isArray
                          }
                      }
                  }else if(i == keys.length-1)
                      return {
                          node: schemaNode,
                          properties: pro,
                          isArray: isArray 
                      }
  
                  schemaNode = currentNode;
              }
              else{
                  return{
                      node: schemaNode,
                      properties: Properties,
                      isArray: isArray
                  }
              }
            }
            return null;
  
          }
          getSchemaDefinition(path){
  
            if(!this.schema || this.schema =={})
              return null;
            
            if(path.startsWith('#/'))
                path = path.replace('#/','')
  
            if(path =='')
              return this.schemaRootNode
  
            const keys = path.toString().includes("/")? path.toString().split("/"): [path];
            
            let Properties = {};
            let schemaNode = this.schemaRootNode
          //  console.log('get the schema definition:',path)
  
            for(var i=0;i<keys.length;i++){
              let key = keys[i]
           //   console.log(schemaNode)
              if(!schemaNode.properties.hasOwnProperty(key))
                  return null;
  
              Properties = schemaNode.properties[key];
            //  console.log(key, Properties)
              if(Properties.hasOwnProperty('$ref') || (Properties.hasOwnProperty('type') && Properties['type'] == 'array')){
                  let nodepath = '';
                  if(Properties.hasOwnProperty('$ref'))
                      nodepath = Properties['$ref']                
                  else if(Properties['items'].hasOwnProperty('$ref'))
                      nodepath = Properties['items']['$ref'];
                  else 
                      return  schemaNode              
  
                  if(nodepath.startsWith('#/'))
                      nodepath = nodepath.replace('#/','')
                  
                  let paths = nodepath.includes('/')? nodepath.split('/'):[nodepath];
            //      console.log(paths)
                  let currentSchemaDefinition = this.schema
                  for(var j=0;j<paths.length;j++){
                      let path1 = paths[j];
                      currentSchemaDefinition =currentSchemaDefinition[path1]                    
                  }
               //   console.log(currentSchemaDefinition, i, keys)
  
                  if(currentSchemaDefinition.hasOwnProperty('type')){
                      if(currentSchemaDefinition['type'] == 'object'){
                          if(i== keys.length -1)
                              return currentSchemaDefinition;
                          else
                              schemaNode = currentSchemaDefinition;
                      }
                      else{
                          return schemaNode
                      }
                  }
                  else if(i== keys.length -1)
                      return currentSchemaDefinition;
  
                  schemaNode = currentSchemaDefinition;
              }
              else{
                  return schemaNode
              }
            }
            return null;           
          }
    }
    UI.JSONSchema = JSONSchema
    
})(UI || (UI = {}));
