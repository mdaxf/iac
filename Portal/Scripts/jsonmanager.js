var UI = UI || {};
(function (UI) {

    class JSONManager {
        constructor(jsonObject, options = { allowChanges: true, schemafile:"", schema:null }) {
            this.originalObject = JSON.parse(JSON.stringify(jsonObject));        
          this.data = (typeof jsonObject == 'string')? JSON.parse(jsonObject) : jsonObject;
          this.options = options;
          UI.Log(options)
          this.changed = false;
          this.schemafile = options.schemafile || '';
          this.schema = this.options.schema || null;
          this.jschema = new JSONSchema(this.schema);
          this.allowChanges = options.allowChanges || true;
          this.options.allowChanges = this.allowChanges;

          this.loadschema();
        //  UI.Log(this.schemafile,this.schema)
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
                 //   UI.Log(this.schemafile,response)
                    that.getSchemaDefinitions();
                }).catch((error) => {
                    UI.Log(error);
                })
            }
            else if(this.schema !={})
                that.getSchemaDefinitions();
            
        }
        getSchemaDefinitions(){
        //    UI.Log(this.schema)
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
                UI.Log("Modifications are not allowed.");
                return;
            } 
            const node = this.getNode(path);
            if (node) {
            //    UI.Log(node,key)
                if (node.isArrayElement) {
                    node.value.push({});
                } else {
                    node.value[key] = null;
                }
            //    UI.Log(node)
                this.changed = true;
            }
        }
        insertNode(path, Value) {
            if (!this.options.allowChanges) {
                UI.Log("Modifications are not allowed.");
                return;
            } 
           
            const node = this.getNode(path);
            let isvalidJson = this.isValidJSON(Value);
            if (node) {
             //   UI.Log('insert node value:',node, isvalidJson, path, Value)
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
                        
                     //   UI.Log(valueobj)
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
        simpleUpdateNode(key, newValue) {
            if (!this.options.allowChanges) {
                UI.Log("Modifications are not allowed.");
                return;
            }
            const node = this.getNode(key);
            UI.Log("simpleUpdateNode", node, key, newValue)
            if (node) {
                this.data[key] = newValue;
                this.changed = true;
            }
        //    UI.Log(this.data)
        }
        setNodewithKey(path, key, value)  {
            if (!this.options.allowChanges) {
                UI.Log("Modifications are not allowed.");
                return;
            }
            const node = this.getNode(path);
            UI.Log('set node value:',node)
            if (node) {
                node.value[key]= value;
                this.changed = true;
                UI.Log(this.data, node)
            }
        }

        updateNodeValue(path, newValue) {
          if (!this.options.allowChanges) {
            UI.Log("Modifications are not allowed.");
            return;
          }
      //    UI.Log('updateNodeValue',path, newValue )
          let isValueObj =this.isValidJSON(newValue);
        
          if(isValueObj){
            path = path
          }else{
            let paths = path.split('/');
            let lastPath = paths[paths.length-1];
            paths.pop();
            path = paths.join('/');            
            newValue = JSON.parse(`{"${lastPath}":"${newValue}"}`);
            
         //   UI.Log(newValue);
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
            UI.Log('find node:',node, path)
            if (node.isArrayElement) {
              const arrayNode = this.getNode(node.arrayPath);
              if (arrayNode) {
                    let valueobj = null;
                    if(typeof newValue == 'string')
                        valueobj = JSON.parse(newValue);
                    else
                        valueobj = newValue;
                //    UI.Log(arrayNode, newValue)
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

                    UI.Log(node, valueobj)
                    for (const key in valueobj) {
                        if (valueobj.hasOwnProperty(key)) {
                            const element = valueobj[key];
                            node.value[key] = element;
                        }
                    }
            } */
            this.changed = true;
          } else {
            UI.Log("Invalid path:", path);
          }
        }

        deleteNode(path) {
            if (!this.options.allowChanges) {
                UI.Log("Modifications are not allowed.");
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
                UI.Log("Invalid path:", path);
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

        //  UI.Log('getnode:',path)
          const keys = path.toString().includes("/")? path.toString().split("/"): [path];
          let currentNode = this.data;
          let arrayPath = null;
          let isArrayElement = false;
          let index = -1;
        //  UI.Log(keys)
          for (let i = 0; i < keys.length; i++) {
            const key = keys[i];
        //    UI.Log(i,currentNode,key, this.isValidJSON(key))
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
            //    UI.Log(currentNode,key)
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
        //    UI.Log(str)
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
        //    UI.Log('findNodebykey',jsonObject,key, Array.isArray(jsonObject),jsonObject.hasOwnProperty(key))
            if(Array.isArray(jsonObject)){

                if(this.isValidJSON(key)){
                    for (var i=0;i<jsonObject.length;i++){
                    //    UI.Log('check the node:',i,jsonObject[i] )
                        let node = jsonObject[i];
                        let found = true;
                        for (const searchKey in key) {
                        //    UI.Log("check the key:", searchKey, key[searchKey], node[searchKey])
                            if (key.hasOwnProperty(searchKey) && node[searchKey] != key[searchKey]) {
                              found = false;
                              break;
                            }
                        }
                        if (found) {
                          return {
                            value: node,
                            isArrayElement: Array.isArray(node),
                            arrayPath: '',
                            index: i,
                          };
                        }
                    }
                    return null;
                }else{
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
        //  UI.Log(schemaNode, path, keys)
          let isArray = false;
          for(var i=0;i<keys.length;i++){
            let key = keys[i]
        //    UI.Log(this.schema,schemaNode,key)
            if(!schemaNode.properties.hasOwnProperty(key))
                return null;

            Properties = schemaNode.properties[key];
         //   UI.Log(key, Properties)
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
            //    UI.Log(paths)
                let currentNode = this.schema
                for(var j=0;j<paths.length;j++){
                    let path1 = paths[j];
                    currentNode =currentNode[path1]                    
                }
           //    UI.Log(i, key, keys,schemaNode,currentNode)
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
                    //    UI.Log(key, path, pro, Properties)
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
        //  UI.Log('get the schema definition:',path)

          for(var i=0;i<keys.length;i++){
            let key = keys[i]
         //   UI.Log(schemaNode)
            if(!schemaNode.properties.hasOwnProperty(key))
                return null;

            Properties = schemaNode.properties[key];
          //  UI.Log(key, Properties)
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
          //      UI.Log(paths)
                let currentSchemaDefinition = this.schema
                for(var j=0;j<paths.length;j++){
                    let path1 = paths[j];
                    currentSchemaDefinition =currentSchemaDefinition[path1]                    
                }
             //   UI.Log(currentSchemaDefinition, i, keys)

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
        //    UI.Log(lng,lngcode, lngdefault)

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
                
            //    UI.Log(key,requiredfields,hiddenfields, unchangablefields, invalidateNode, schemadata)
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
                    
                //    UI.Log('parsed value:', key,required, hidden,unchangable)
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
                
            //    UI.Log(lng,lngcode, lngdefault)
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
                         //   UI.Log(node[key],values[n],optionlngcodes[n], optiondefaults[n])
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
               
            //    UI.Log(schemaoptions,editable, unchangable, showlabelonly,key,node[key], nodelabelvalue, nodeeditablevalue)
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
        showdetailpage(wrapper){
            let that = this;
            let attrs ={};
            let srcdata = this.jschema.getdatadetail();
            
            if(srcdata == null && srcdata =={})
                return;

            if(!srcdata.hasOwnProperty("detailpage"))
                return;

            let data = srcdata['detailpage'];
            let defaulttab ="";

            if(data.hasOwnProperty("tabs")){
                attrs={
                    "name": "ui-json-detail-page-tabs",
                    "id": "ui-json-detail-page-tabs",
                    "class": "ui-json-detail-page-tabs",
                    "style": "width:100%;height:30px;"
                }
                if(wrapper == null)
                    wrapper = document.body;
                let tabs = (new UI.FormControl(wrapper, 'div',attrs)).control;

                let tabitems = data['tabs']
                for(const key in tabitems){
                    if(defaulttab == "")
                        defaulttab = key

                    let item = tabitems[key]; 
                    let lngcode =""
                    let lngdefault = key
                    if(item.hasOwnProperty("lng"))
                    {
                        let lng = item['lng']
                        if(lng.hasOwnProperty('code'))
                            lngcode = lng['code'];
                        if(lng.hasOwnProperty('default'))
                            lngdefault=lng["default"];
                    }
                    attrs={
                        "name": "ui-json-detail-page-tab-"+key,
                        "id": "ui-json-detail-page-tab-"+key,
                        "tab-key": key,
                        "class": "ui-json-detail-page-tab",
                        "style": "width:100px;height:30px;float:left;",
                        "lngcode": lngcode,
                        "innerHTML": lngdefault
                    }

                    if(key == defaulttab)
                        attrs["class"] = "ui-json-detail-page-tab ui-json-detail-page-tab-active"

                    let events={
                        "click": function(){
                            let id = $(this).attr('id');
                            let tabid = id.replace('ui-json-detail-page-tab-','');
                            let tabkey = $(this).attr('tab-key');
                            $('.ui-json-detail-page-tab').removeClass('ui-json-detail-page-tab-active');
                            $('#'+id).addClass('ui-json-detail-page-tab-active');
                            $('.ui-json-detail-page-tab-content').hide();
                            $('.ui-json-detail-page-tab-content[tab-key="'+tabkey +'"]').show();
                        }
                    }
                    let tab = (new UI.FormControl(tabs, 'div',attrs,events)).control;
                }
                
            }
            let actionbar = (new UI.FormControl(wrapper, 'div',{"style":"display:inline-block;height:30px; float:right", "class": "ui_actions_section"})).control;

            let events={}
            UI.Log(this.allowChanges)

            if(this.allowChanges){
                
                attrs={
                    "name": "ui-json-detail-page-save",
                    "id": "ui-json-detail-page-save",
                    "class": "ui-json-detail-page-save btn btn-primary",
                    "style": "width:100px;height:30px;float:right;",
                    "innerHTML": "Save",
                    "value": "Save",
                    "lngcode": "Save"
                }
                events={
                    "click": function(){
                        that.getdetailsavedata();
                        UI.Log("save", that.data, that.nullvalues)
                        that.trigger_event("save", [that.data, that.nullvalues]);
                    }}
                new UI.FormControl(actionbar, 'button',attrs,events)
            }
            attrs={
                "name": "ui-json-detail-page-cancel",
                "id": "ui-json-detail-page-cancel",
                "class": "ui-json-detail-page-save btn btn-secondary",
                "style": "width:100px;height:30px;float:right;",
                "innerHTML": "Cancel",
                "value": "Cancel",
                "lngcode": "Cancel"
            }
            events={
                "click": function(){
                    that.canceldetail();
                }}
            new UI.FormControl(actionbar, 'button',attrs,events)            

            attrs={
                "name": "ui-json-detail-page-tab-content",
                "id": "ui-json-detail-page-tab-content",
                "class": "ui-json-detail-page-tab-content-container",
                "style": "width:100%;min-height:500px; height:calc(100% - 60px);"
            }
            let tabcontent = (new UI.FormControl(wrapper, 'div',attrs)).control;

            for(const key in data){
                if(key == "tabs" || key == "Query")
                    continue;
                
                let item = data[key];

                attrs = {
                    "name": "ui-json-detail-page-tab-content-"+key,
                    "id": "ui-json-detail-page-tab-content-"+key,
                    "class": "ui-json-detail-page-tab-content",
                    "tab-key": key
                  //  "style": "width:100%;height:100%;display:block;"
                }

                if(key == defaulttab)
                    attrs["style"] = "width:100%;height:100%;display:block;"
                else 
                    attrs["style"] = "width:100%;height:100%;display:none;"

                let tabcontentitem = (new UI.FormControl(tabcontent, 'div',attrs)).control;

                let tables = item["tables"]
            //    UI.Log(tables)
                if(tables != null && tables != undefined && tables !={})
                for(var i=0;i<tables.length;i++){
                    let key1 = tables[i];
                    
                    
                    let style = "";
                    if(key1.hasOwnProperty("style"))
                        style = key1["style"]

           //         UI.Log(tables,key1, style)
                    let table = (new UI.FormControl(tabcontentitem, 'table',{"class":"table table-bordered table-hover", "style":style})).control;

                    let cols = 1
                    if(key1.hasOwnProperty("cols"))
                        cols = key1["cols"]
                    
                    let rows = 0;
                    if(key1.hasOwnProperty("rows"))
                        rows = tables[i]["rows"]

                    let fields = key1["fields"]
                    
                    let cellnumber = 0;
                    let tr = (new UI.FormControl(table, 'tr',{"class": "ui-json-detail-page-tab-content-tr-"+cols,})).control;
                    for(var j=0;j<fields.length;j++){
                        let field = fields[j];

                        if(cellnumber >= cols){
                            let width = 100/cols;
                            rows += 1;
                            tr = (new UI.FormControl(table, 'tr',{"class": "ui-json-detail-page-tab-content-tr-"+cols,})).control;
                            cellnumber = 0;
                        }
                        cellnumber += 1

                        if(typeof field == "object"){
                            let fieldkey = Object.keys(field)[0];
                            let fieldvalue = field[fieldkey];

                        //    UI.Log(fieldkey, fieldvalue)
                            let type =""
                            if(fieldvalue.hasOwnProperty("type"))
                                type = fieldvalue["type"]
                            
                            if(type == "link" || type == "singlelink"){
                                let attrs ={
                                    "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,                                    
                                    "innerHTML": fieldkey,
                                }
                                let td = (new UI.FormControl(tr, 'td',{})).control;
                                let link = (new UI.FormControl(td, "a",attrs)).control;
                                
                                if(type == "link")
                                    $(link).click(function(){                                                                     
                                        that.displayhyperlinks(wrapper,fieldvalue,fieldkey);
                                    })
                                else if(type == "singlelink"){
                                    $(link).click(function(){                                                                     
                                        that.displaysubdetail(wrapper,fieldvalue);
                                    })
                                }
                            }
                            else{

                                let tag = "input"
                                if(fieldvalue.hasOwnProperty("tag"))
                                    tag = fieldvalue["tag"]

                                if(fieldvalue.hasOwnProperty("link")){
                                    this.createcellelement(tr, fieldkey, key, i,rows,cellnumber,wrapper, fieldvalue)
                                }
                                else{                                
                                    let attrs = fieldvalue["attrs"]                                
                                    let td = (new UI.FormControl(tr, 'td',attrs)).control;
        
                                    let node = that.getNode(fieldkey);
                                    
                                    let value = "";
                                    if(node.hasOwnProperty("value")){
                                        value = node.value
                                    }
                                    attrs = {};
                                    UI.Log(field,fieldvalue)
                                    if(fieldvalue.hasOwnProperty("nodeattrs")){
                                        attrs = fieldvalue["nodeattrs"]
                                        const regex = /\{([^}]+)\}/g;
                                        var matches  =[];

                                        if(typeof attrs == "string")
                                            matches = attrs.match(regex);
                                        else if(typeof attrs == "object"){
                                            for(const key in attrs){
                                                let value = attrs[key];
                                                if(typeof value == "string")
                                                    matches = matches.concat(value.match(regex));
                                            }
                                        }
                                        UI.Log(matches)
                                        if (matches.length > 0) {
                                            var extractedValues = matches.map(function(match) {
                                                return match.slice(1, -1); // Remove the curly braces
                                            });
                                            UI.Log(extractedValues, node)

                                            for(var n=0;n<extractedValues.length;n++){
                                                let node1 = that.getNode(extractedValues[n]);
                                                if(node1 && node1.hasOwnProperty("value")){
                                                    let value1 = node1.value
                                                    attrs = JSON.parse(JSON.stringify(attrs).replaceAll('{'+extractedValues[n]+'}', value1))
                                                }                                         
                                            }   
                                        }                                        
                                    }
                                    UI.Log(field,attrs)
                                    if(tag != "input" ||tag != "select"){
                                       
                                        attrs.name = "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber;
                                        attrs.innerHTML = value
                                    }
                                    else{

                                        attrs.name = "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber;
                                        attrs.innerHTML = value
                                        attrs["data-key"] = fieldkey
                                    }
                                    UI.Log(fieldkey, field,attrs,that.jschema)
                                    let schemadata = that.jschema.getPropertiesFromSchema(fieldkey);
                                    let setnullvalue = false;
                                    let schemanullvalue = "";
                                    if(schemadata.properties.hasOwnProperty("nullvalue")){
                                     setnullvalue = true;
                                     schemanullvalue = schemadata.properties["nullvalue"]                       
                                    }

                                    let control = (new UI.FormControl(td, tag,attrs)).control;

                                    if(setnullvalue){
                                       control.setAttribute("nullvalue", schemanullvalue)
                                    }
                                }
                            }
                            
                        }else{
                            this.createcellelement(tr, field, key, i,rows,cellnumber,wrapper,null)
                        }
                    }

                }
            }

            $('.ui-json-detail-page-tab-content-container').find('input').change(function(){
                let key = $(this).attr('data-key');
                if(this.getAttribute('type') == 'checkbox'){
                    that.simpleUpdateNode(key,this.checked);
                }else{
                    let value = $(this).val();
                    that.simpleUpdateNode(key, value);
                }
            })
            $('.ui-json-detail-page-tab-content-container').find('select').change(function(){
                let key = $(this).attr('data-key');
                let value = $(this).val();
                that.simpleUpdateNode(key, value);
            })

            UI.translate(wrapper);
        }

        getdetailsavedata(){
            let that = this
            that.nullvalues = {};
            $('.ui-json-detail-page-tab-content-container').find('input').each(function(){
                let key = $(this).attr('data-key');
                if(this.getAttribute('type') == 'checkbox'){
                    that.simpleUpdateNode(key,this.checked);
                }else{
                    let value = $(this).val();
                    that.simpleUpdateNode(key, value);
                }
                if($(this).attr("nullvalue") != undefined){
                    that.nullvalues[key] =   $(this).attr("nullvalue")   
                }
            })

            $('.ui-json-detail-page-tab-content-container').find('select').each(function(){
                let key = $(this).attr('data-key');
                let value = $(this).val();
                that.simpleUpdateNode(key, value);
                if($(this).attr("nullvalue") != undefined){
                    that.nullvalues[key] = $(this).attr("nullvalue")     
                }
            })
            UI.Log(that.nullvalues)

        }
        createcellelement(tr, field,key, i,rows,cellnumber,wrapper,linkobj){
            let that = this;
            let attrs = {};
            let td = (new UI.FormControl(tr, 'td',attrs)).control;
        //    UI.Log(that.allowChanges)
            UI.Log("create cell element:", tr, field, key, i,rows,cellnumber,wrapper,linkobj)

            if(field != "dummy" && field != "save" && field != "cancel"){
                let node = that.getNode(field);
            //    UI.Log(field, node)
                let value = "";
                   if(node)
                   if(node.hasOwnProperty("value")){
                       value = node.value
                   }
                   let datatype = "string"
                   let lng={};
                   let schemaoptions ={}
                   
                   let schemadata = that.jschema.getPropertiesFromSchema(field);
                   let schemaformate ="";
                   let schemareadonly = false;
               //    UI.Log(field, value, schemadata)
                   if(schemadata.properties.hasOwnProperty("type"))
                       datatype = schemadata.properties["type"]
               
                   if(schemadata.properties.hasOwnProperty("lng"))
                       lng = schemadata.properties["lng"]
                   
                   if(schemadata.properties.hasOwnProperty("options"))
                       schemaoptions = schemadata.properties["options"] 

                   if(schemadata.properties.hasOwnProperty("readonly"))
                       schemareadonly = schemadata.properties["readonly"]
                  
                   if(schemadata.properties.hasOwnProperty("format"))
                       schemaformate = schemadata.properties["format"]

                   let setnullvalue = false;
                   let schemanullvalue = "";
                   if(schemadata.properties.hasOwnProperty("nullvalue")){
                    setnullvalue = true;
                    schemanullvalue = schemadata.properties["nullvalue"]                       
                   }
                    
                   
                   let lngcode = lng['code']
                   let lngdefault = lng['default']    
                   attrs={
                    "styles": "width:100%"
                    }
                   let div = "" ;//(new UI.FormControl(td, 'div',attrs)).control;
                   attrs ={
                       lngcode: lngcode,
                       "for": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                       "innerHTML": lngdefault,
                   }

                   new UI.FormControl(td, 'label',attrs);


                   attrs={
                    "styles": "width:100%"
                   }
                //   div = (new UI.FormControl(td, 'div',attrs)).control;
                   let control = null
                   div = td;
                   if(datatype == 'boolean'){
                       try{
                           if(value == 'true' || value == true)
                               attrs = {
                                   "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                                   "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                                   "type": "checkbox",
                                   "data-key": field,
                                   "checked": true,
                               }
                           else 
                               attrs = {
                                   "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                                   "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                                   "data-key": field,
                                   "type": "checkbox",
                               }
                       }catch(e){
                           attrs = {
                               "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                               "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                               "data-key": field,
                               "type": "checkbox",
                           }
                       }

                       if(schemareadonly || !that.allowChanges)
                           attrs["disabled"] = true;

                       control = (new UI.FormControl(div, 'input',attrs)).control;

                   
                   }
                   else if(schemaoptions && schemaoptions !={} ){
                       let values = schemaoptions['value']
                       let optionlngcodes  = schemaoptions['lngcode']  
                       let optiondefaults = schemaoptions['default']
                       let nodeeditablevalue ='';
                       let nullstring = "";

                       if(setnullvalue){
                        nullstring = 'nullvalue="'+schemanullvalue+'"'
                        
                        }
                       if(Array.isArray(values)){
                           if(schemareadonly)
                               nodeeditablevalue =  '<select class="node_input" '+nullstring+' disabled data-key="'+field+'" id="ui-json-detail-page-tab-content-'+key+'-table-'+i+'-row-'+rows+'-cell-'+cellnumber+'" >'
                           else
                               nodeeditablevalue =  '<select class="node_input" '+nullstring+' data-key="'+field+'" id="ui-json-detail-page-tab-content-'+key+'-table-'+i+'-row-'+rows+'-cell-'+cellnumber+'" >'
                           for(var n=0;n<values.length;n++){
                            //   UI.Log(node[key],values[n],optionlngcodes[n], optiondefaults[n])
                               nodeeditablevalue += '<option value="'+values[n]+'" lngcode="'+optionlngcodes[n]+'" '+(value == values[n]? 'selected':'') +' >' + optiondefaults[n] + '</option>'
                           }
                           nodeeditablevalue += '</select>'
                                                    

                           td.innerHTML =td.innerHTML + nodeeditablevalue;
                       }
                       else{
                           attrs = {
                               "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                               "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                               "data-key": field,
                               "value": value,
                           }

                           if(schemareadonly || !that.allowChanges)
                            attrs["disabled"] = true;

                           if(schemaformate == "datetime")
                               attrs["type"] = "datetime-local"

                               control = (new UI.FormControl(div, 'input',attrs)).control;   
                       }         

                   }else{
                       attrs = {
                           "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                           "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber,
                           "data-key": field,
                           "value": value,
                       }
                       
                       if(schemaformate == "datetime")
                           attrs["type"] = "datetime-local"

                       if(schemareadonly || !that.allowChanges)
                           attrs["disabled"] = true;
                           control = (new UI.FormControl(div, 'input',attrs)).control;   
                   }
                   
                   if(setnullvalue && control != null){
                        control.setAttribute('nullvalue', schemanullvalue)
                   }

                   if(linkobj != null){
                        attrs = {
                            "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber+ "-link",
                            "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber+ "-link",
                            "data-key": field,
                            "class": "fa-solid fa-link",
                        }
                        if(schemareadonly || !that.allowChanges)
                           attrs["disabled"] = true;

                        let link = (new UI.FormControl(div, 'i',attrs)).control;
                        $(link).click(function(){
                            let inputid = $(this).closest('td').find('input').attr('id');
                            let schema = linkobj.schema;
                         //   let field = linkobj.field;
                            that.displaylinkeditem(wrapper,inputid, schema, field);
                        })
                        attrs = {
                            "name": "ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber+ "-unlink",
                            "id":"ui-json-detail-page-tab-content-"+key+"-table-"+i+"-row-"+rows+"-cell-"+cellnumber+ "-unlink",
                            "data-key": field,
                            "class": "fa-solid fa-link-slash",
                        }
                        if(schemareadonly || !that.allowChanges)
                           attrs["disabled"] = true;

                        let unlink = (new UI.FormControl(div, 'i',attrs)).control;
                        $(unlink).click(function(){
                            let datakey = $('#'+fieldid).attr('data-key');
                            $(this).closest('td').find('input').val('');
                            that.simpleUpdateNode(datakey, '');
                        })
                        
                   }
               } 
        }
        // display the field link and unlink icon to show the data list to select
        displaylinkeditem(wrapper, fieldid, schema, field){
         /*   let attrs = {
                "name": "ui-json-detail-page-linked-item-section",
                "id": "ui-json-detail-page-linked-item-section",
                "style": "width:100%;height:100%; display:float; left:0px; top:80px; position:absolute; background-color:white; z-index:10;",
                "class": "popup"
            }
            let section = (new UI.FormControl(wrapper, 'div',attrs)).control;
            let that = this;
            let panel = {};
            panel.panelElement = section; */
            // let div = document.createElement('div');
           // div.innerHTML = "<h3>Linked Item</h3>"
           // panel.panelElement.appendChild(div);
           let page =Session.CurrentPage;
           let that = this;
            let cfg = {
            //    "file":"templates/datalist.html", 
                "name": "Data List",
                "type": "document",
                "title": field+ " list", 
                "actions": {
                    "SELECT":{"type": "script", "next": "","page":"","panels":[], "script": "selectitem"},
                    "CANCEL":{"type": "script", "next": "","page":"","panels":[], "script": "cancelitem"},
                }
            }
         //   UI.Log(cfg)
            let org_schema = Session.snapshoot.sessionData.ui_dataschema
            let org_entity = Session.snapshoot.sessionData.entity
            let org_selectedKey = Session.snapshoot.sessionData.selectedKey
            let inputs = {}
            inputs.ui_dataschema = schema
        //    UI.Log(inputs)
            cfg.inputs = inputs;
            cfg.actions.SELECT.script = function(data){
                UI.Log("execute the action:",field, data)
                $('#'+fieldid).val(data.selectedKey);
                let datakey = $('#'+fieldid).attr('data-key');
                UI.Log(fieldid, datakey, data.selectedKey)
                that.simpleUpdateNode(datakey, data.selectedKey);
                Session.snapshoot.sessionData.entity = org_entity;
                Session.snapshoot.sessionData.selectedKey = org_selectedKey;
                Session.snapshoot.sessionData.ui_dataschema = org_schema;
                page.popupClose();
            }
            cfg.actions.CANCEL.script = function(data){
                UI.Log("execute the action:", data)

                Session.snapshoot.sessionData.entity = org_entity;
                Session.snapshoot.sessionData.selectedKey = org_selectedKey;
                Session.snapshoot.sessionData.ui_dataschema = org_schema;
                page.popupClose();
            }
            cfg.onloadedscript = function(){
                $('.iac-ui-popup .ui_actions_section button').hide();
                $('.iac-ui-popup .ui_actions_section button[value="Select"]').show();
                $('.iac-ui-popup .ui_actions_section button[value="Cancel"]').show();
              }
            Session.snapshoot.sessionData.ui_dataschema = schema
            page.popupOpen(cfg);  	
    	    page.popup.onClose(function(){
                Session.snapshoot.sessionData.selectedKey = org_selectedKey;
                Session.snapshoot.sessionData.ui_dataschema = org_schema;
            }) 
        }
        displayhyperlinkmaster(wrapper,data){
            
            //var data = Session.snapshoot.sessionData
            UI.Log("displayhyperlinkmaster:",wrapper,data)
            var schema = data.ui_linkedjdata.masterschema;
            let cfg = {
                "file":"templates/datalist.html", 
                "title": ""+ " Master list", 
                "name": "Data List",
                "type": "document",
                "actions": {
                    "MSELECT":{"type": "script", "next": "","page":"","panels":[], "script": "selectitem"},
                    "CANCEL":{"type": "script", "next": "","page":"","panels":[], "script": "cancelitem"},
                }
            }
            let page =Session.CurrentPage;
            UI.Log(page)
            let org_schema = Session.snapshoot.sessionData.ui_dataschema
            let org_entity = Session.snapshoot.sessionData.entity
            let org_selectedKey = Session.snapshoot.sessionData.selectedKey

            let inputs = {}
            inputs.ui_dataschema = schema
        //    UI.Log(inputs)
            cfg.inputs = inputs;
            cfg.actions.MSELECT.script = function(data){
                UI.Log("execute the action:", data)
                
                if(data.selectedRows.length == 0){
                    UI.ShowError("Please select one record at least.")
                    return;
                }

                let insertdata ={}
                UI.Log(data.ui_linkedjdata)
                for(var i=0;i<data.ui_linkedjdata.linkfields.length;i++){
                    if(data.ui_linkedjdata.linkfields[i].hasOwnProperty(data.ui_linkedjdata.keyfield)){
                        let valuefield = data.ui_linkedjdata.linkfields[i][data.ui_linkedjdata.keyfield];
                        let vfs = valuefield.split('.')
                        if(vfs.length == 2)
                            insertdata[vfs[1]] = data.ui_linkedjdata.keyvalue;
                        else    
                            insertdata[valuefield] = data.ui_linkedjdata.keyvalue;  
                    } 
                }

                let keyfield = data.ui_linkedjdata.keyfield;
                if(keyfield == undefined || keyfield == "" || keyfield == null)
                    keyfield = "ID";

                for(var i=0;i<data.selectedRows.length;i++){
                    insertdata[data.ui_linkedjdata.masterdatafield] = data.selectedRows[i][data.ui_linkedjdata.keyfield];
                    let url = '/sqldata/insert'
                    insertdata.createdby = UI.userlogin.username
                    insertdata.createdon = (new Date()).toISOString().slice(0, 19).replace('T', ' ');
                    insertdata.updatedby = UI.userlogin.username
                    insertdata.updatedon = (new Date()).toISOString().slice(0, 19).replace('T', ' ');

                    let inputdata = {
                        "tablename": data.ui_linkedjdata.tablename,
                        "data": insertdata
                    }
                    UI.Log(url,inputdata)
                    UI.ajax.post(url,inputdata).then((response) => {
                        data = JSON.parse(response); 
                        Session.CurrentPage.panels[0].view.fireOnLoaded();
                    }).catch((error) => {
                        UI.ShowError(error);
                    });
                }              
                Session.snapshoot.sessionData.entity = org_entity;
                Session.snapshoot.sessionData.selectedKey = org_selectedKey;
                Session.snapshoot.sessionData.ui_dataschema = org_schema;

                page.popupClose();
                //id = data.data.id
                page.Refresh();
 
                
            }
            cfg.actions.CANCEL.script = function(data){
                UI.Log("execute the action:", data)

                Session.snapshoot.sessionData.entity = org_entity;
                Session.snapshoot.sessionData.selectedKey = org_selectedKey;
                Session.snapshoot.sessionData.ui_dataschema = org_schema;
                page.popupClose();
            }
            Session.snapshoot.sessionData.ui_dataschema = schema
            cfg.onloadedscript = function(){
                $('.iac-ui-popup .ui_actions_section button').hide();
                $('.iac-ui-popup .ui_actions_section button[value="MultiSelect"]').show();
                $('.iac-ui-popup .ui_actions_section button[value="Cancel"]').show();
              }
            
            page.popupOpen(cfg);
            cfg.onloadedscript();
            page.popup.onClose(function(){
                Session.snapshoot.sessionData.entity = org_entity;
                Session.snapshoot.sessionData.selectedKey = org_selectedKey;
                Session.snapshoot.sessionData.ui_dataschema = org_schema;
            })
        }
        //display the link section item with the <a > link
        displayhyperlinks(wrapper,fieldvalue, field){
            let that = this;
            let configuration = {};
            configuration.name = "Data Link List";
            if(fieldvalue.hasOwnProperty("lng"))
                configuration.title = fieldvalue["lng"]["default"] + " List";

            let panel = {}
            panel.name = "datalink_content_panel";
            let view ={}
            view.title = this.schema["datasource"] +"_"+field+"__datalinklist";
            view.name = "Data Link List"
            view.type = "document";
        //    view.file ="templates/datalinklist.html";
            let inputs={}       
            
            let keyfield ="";
            if(fieldvalue.hasOwnProperty("keyfield")){
                keyfield = fieldvalue["keyfield"];
            }
            else{
                UI.Log("There is a linked field defined.")
                return;
            }                

            let keyvalue = this.data[keyfield];
            let where = {};
            if(fieldvalue.hasOwnProperty("linkfields")){
                let linkfields = fieldvalue["linkfields"];
                let wherestr = "";
                for(var i=0;i<linkfields.length;i++){
                    UI.Log(linkfields[i],keyvalue)
                    if(linkfields[i].hasOwnProperty(keyfield)){
                        wherestr += linkfields[i][keyfield] + " = '" + keyvalue + "'";                        
                    }
                }
                if(wherestr !="")
                    where[wherestr] = "";
            }
                        
            fieldvalue["where"] = where;
            fieldvalue["keyvalue"] = keyvalue;
            inputs.ui_linkedjdata = fieldvalue;

            view.inputs = inputs;
            Session.snapshoot.sessionData.ui_linkedjdata = fieldvalue;
            view.outputs = {};
            view.outputs.ui_linkedjdata ={};

            if(that.allowChanges){
                let actions = {}
                actions.ADD = {"type": "script", "next": "","page":"","panels":[], "script": "addnewitem"}
                view.actions = actions;
                view.actions.ADD.script = function(data){
                    that.displayhyperlinkmaster(wrapper,data);
                }
            }
            panel.view = view;
            configuration.panels = [panel];

            UI.Log("display the list screen",configuration,inputs)
            new UI.Page(configuration);
        
        }
        displaysubdetail(wrapper,fieldvalue){
            let attrs = {
                "name": "ui-json-detail-page-linked-item-section",
                "id": "ui-json-detail-page-linked-item-section",
                "style": "width:100%;height:100%; display:float; left:0px; top:0px; position:absolute; background-color:white; z-index:10;"
            }
            let section = (new UI.FormControl(wrapper, 'div',attrs)).control;
            let that = this;
            let panel = {};
            panel.panelElement = section;
            let cfg = {
                "file":"templates/datadetail.html",
                "name": "detail page",
                "actions": {
                    "SAVE":{"type": "script", "next": "","page":"","panels":[], "script": "saveitem"},
                    "CANCEL":{"type": "script", "next": "","page":"","panels":[], "script": "cancelitem"},
                }
            }
            let inputs = {}
            inputs.ui_dataschema = fieldvalue['schema']
            inputs.ui_data = fieldvalue['data']
            inputs.selectedKey = (that.getNode(fieldvalue['keyfield'])).value
            cfg.inputs = inputs;
            cfg.actions.SAVE.script = function(data){             
                $('#ui-json-detail-page-linked-item-section').remove();
            }
            cfg.actions.CANCEL.script = function(data){
                $('#ui-json-detail-page-linked-item-section').remove();
            }
            Session.snapshoot.sessionData.ui_dataschema = fieldvalue['schema']
            new UI.View(panel,cfg)
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
        trigger_event(event, args) {
            UI.Log(event,args)
			if (this.options['on_' + event]) {
				this.options['on_' + event].apply(null, args);
			}
		}
    }
    UI.JSONManager = JSONManager

    class JSONSchema{
        constructor(schema){
            this.schema = schema || {};
            this.getSchemaRootDefinitions();
        }
        loadschema(file){
            let that = this;
            $.ajax({
                url: file,
                dataType: 'json',
                async: false,
                success: function(data) {
                    that.schema = data;
                    that.getSchemaRootDefinitions();
                }
            });
        }
        getSchemaRootDefinitions(){
            //    UI.Log(this.schema)
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
        getfields(){
            let fields = [];
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
            if(this.schemaRootNode != null){
                for(const key in this.schemaRootNode.properties){
                    fields.push(key);
                }
            }
            return fields;
        }
        parsedata(data){
           for(const key in this.schemaRootNode.properties){
                if(this.schemaRootNode.properties[key]['type'] == "string" ){
                    if(this.schemaRootNode.properties[key].hasOwnProperty('format')){
                        if(this.schemaRootNode.properties[key]['format'] == 'uuid'){
                            if(!data.hasOwnProperty(key)){
                                data[key] = this.generateUUID();
                            }
                        }
                    }
                }

                if(data.hasOwnProperty(key)){
                    switch (this.schemaRootNode.properties[key]['type']){
                        case "integer":
                            try{
                                data[key] = parseInt(data[key]);
                            }catch(e){
                                data[key] = 0;
                            }
                            break;
                        case "number":
                            try{
                                data[key] = parseFloat(data[key]);
                            }catch(e){
                                data[key] = 0.0;
                            }
                            break;
                        case "boolean":
                            try{
                                data[key] = (data[key] == 'true' || data[key] == true)? 1:0;
                            }catch(e){
                                data[key] = 0;
                            }
                            break;
                        case "string":
                            try{
                                if(this.schemaRootNode.properties[key].hasOwnProperty('format')){
                                    if(this.schemaRootNode.properties[key]['format'] == 'datetime'){
                                        data[key] = new Date(data[key]).toISOString().slice(0, 19).replace('T', ' ');
                                        if(data[key] == '1970-01-01 00:00:00')
                                            delete data[key];
                                    }
                                    else
                                        data[key] = data[key].toString();
                                }                                
                                else
                                        data[key] = data[key].toString();
                            }catch(e){
                                if(this.schemaRootNode.properties[key].hasOwnProperty('format')){
                                    if(this.schemaRootNode.properties[key]['format'] == 'datetime'){                                       

                                        data[key] = new Date().toISOString().slice(0, 19).replace('T', ' ');

                                        if(data[key] == '1970-01-01 00:00:00')
                                            delete data[key];
                                    }
                                    else
                                        data[key] = "";
                                }
                                else
                                        data[key] = "";
                            }
                            break;
                        case "object":
                        case "Object":
                            if(data[key] == undefined || data[key] == null || data[key] == "")
                                data[key] = {};
                            else
                                data[key] = JSON.parse(data[key]);

                                data[key] = JSON.stringify(data[key]);
                            
                            data[key] = JSON.stringify(data[key]);
                            break;
                        default:
                            if(data[key] == undefined || data[key] == null)
                                data[key] = "";
                            else
                                data[key] = data[key].toString();

                    }
                }
           }
           return data;
        }
        createemptydata(){
            let data ={};
            for(const key in this.schemaRootNode.properties){
                var value = "";
                if(this.schemaRootNode.properties[key].hasOwnProperty('type')){
                    switch (this.schemaRootNode.properties[key]['type'])
                    {
                        case "integer":
                            value = 0;
                            break;
                        case "number":
                            value = 0.0;
                            break;
                       // case "boolean":
                        //    value = false;
                        //    break;
                        case "string": 
                            if(this.schemaRootNode.properties[key].hasOwnProperty('format')){
                                if(this.schemaRootNode.properties[key]['format'] == 'datetime')
                                    value = "NULL"
                            }
                            else 
                                value = "";
                            break;
                        default:    
                            value = "";
                            break;
                    }

                }
                if(value != "NULL")
                    data[key] = value

            }

            return data;
        }
        getdatadetail(){
            let data={};
            if(this.schema !=null && this.schema !={}){
                if(this.schema.hasOwnProperty("datasourcetype"))
                    data.datasourcetype = this.schema["datasourcetype"];
                if(this.schema.hasOwnProperty("datasource"))
                    data.datasource = this.schema["datasource"];
                if(this.schema.hasOwnProperty("listfields"))
                    data.listfields = this.schema["listfields"];
                if(this.schema.hasOwnProperty("keyfield"))
                    data.keyfield = this.schema["keyfield"];
                if(this.schema.hasOwnProperty("detailpage"))
                    data.detailpage = this.schema["detailpage"];                    
            }
            return data;
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
          //  UI.Log(schemaNode, path, keys)
            let isArray = false;
            for(var i=0;i<keys.length;i++){
              let key = keys[i]
            //  UI.Log(this.schema,schemaNode,key)
              if(!schemaNode.properties.hasOwnProperty(key))
                  return null;
  
              Properties = schemaNode.properties[key];
           //   UI.Log(key, Properties)
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
              //    UI.Log(paths)
                  let currentNode = this.schema
                  for(var j=0;j<paths.length;j++){
                      let path1 = paths[j];
                      currentNode =currentNode[path1]                    
                  }
             //    UI.Log(i, key, keys,schemaNode,currentNode)
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
                      //    UI.Log(key, path, pro, Properties)
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
          //  UI.Log('get the schema definition:',path)
  
            for(var i=0;i<keys.length;i++){
              let key = keys[i]
           //   UI.Log(schemaNode)
              if(!schemaNode.properties.hasOwnProperty(key))
                  return null;
  
              Properties = schemaNode.properties[key];
            //  UI.Log(key, Properties)
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
            //      UI.Log(paths)
                  let currentSchemaDefinition = this.schema
                  for(var j=0;j<paths.length;j++){
                      let path1 = paths[j];
                      currentSchemaDefinition =currentSchemaDefinition[path1]                    
                  }
               //   UI.Log(currentSchemaDefinition, i, keys)
  
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
