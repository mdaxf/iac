var UI;
(function (UI) {
 /*
        common UI functions and classes
        Ajax Call
    */
        UI.CONTROLLER_URL = "api/ui";
        class Ajax {
            constructor(token) {
              this.token = token;
            }
          
            initializeRequest(method, url, stream) {
              return new Promise((resolve, reject) => {
                const xhr = new XMLHttpRequest();
                xhr.open(method, `${url}`, true);                
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
          
            get(url, stream) {
              return this.initializeRequest('GET', url, stream);
            }
          
            post(url, data) {
              return new Promise((resolve, reject) => {
                const xhr = new XMLHttpRequest();
                xhr.open('POST', `${url}`, true);
            //    xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);
                xhr.setRequestHeader('Content-Type', 'application/json');
                xhr.onload = () => {
                  if (xhr.status >= 200 && xhr.status < 300) {
                    resolve(xhr.response);
                  } else {
                    reject(xhr.statusText);
                  }
                };
                xhr.onerror = () => reject(xhr.statusText);
                xhr.onabort = () => reject('abort');
                xhr.send(JSON.stringify(data));
              });
            }
          
            delete(url) {
              return this.initializeRequest('DELETE', url);
            }
          }
        UI.Ajax = Ajax; 
        UI.ajax = new Ajax();   
})(UI || (UI = {}));
(function (UI) {    
    function generateUUID(){
        var d = new Date().getTime();
        var uuid = 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            var r = (d + Math.random()*16)%16 | 0;
            d = Math.floor(d/16);
            return (c=='x' ? r : (r&0x3|0x8)).toString(16);
        });
        return uuid;    
    }
    UI.generateUUID = generateUUID;
    function safeName(name){
        return name.replace(/[^a-zA-Z0-9]/g, "_");
    }
    UI.safeName = safeName;
    function safeId(id){
        return id.replace(/[^a-zA-Z0-9]/g, "_");
    }   
    UI.safeId = safeId; 
    function safeClass(className){
        return className.replace(/[^a-zA-Z0-9]/g, "_");
    }   
    UI.safeClass = safeClass;
    function replaceAll(target, search, replacement) {
        return target.replace(new RegExp(search, "g"), replacement);
    }
    UI.replaceAll = replaceAll;

    function createFragment(html) {
        if (html == null)
            html = "";
        let range = document.createRange();
        if (range.createContextualFragment) {
            try {
                return range.createContextualFragment(html);
            }
            catch (e) {
                // createContextualFragment is not supported on Safari (ios)
            }
        }
        let div = document.createElement("div");
        div.innerHTML = html;
        let inlineScripts = parseInlineScripts(div);
        if (inlineScripts.length > 0) {
            let inlineScriptsEl = document.createElement("script");
            inlineScriptsEl.type = "text/javascript";
            inlineScriptsEl.textContent = inlineScripts.join("; ");
            document.body.appendChild(inlineScriptsEl);
        }
        let fragment = document.createDocumentFragment(), child;
        while ((child = div.firstChild)) {
            fragment.appendChild(child);
        }
        return fragment;
    }
    UI.createFragment = createFragment;

})(UI || (UI = {}));


(function (UI){



    class UISession{
        constructor(configurator){
            let defaultconfig = {
                "name": "ui-root",
                "level": 0
            }
            this.configurator = this.configurator || defaultconfig;

            this.stack = [];
            this.snapshoot ={
                "stack":[],
                "configurator":this.configurator,
                "sessionData":{},
                "immediateData":{}
            };
            this._item = {};
            this._inputs = {};
            this._outputs = {};
            this.model = {};
            this.children = [];
            this.views = {};
            this.panels={};
            this.pages={};
            this.viewResponsitory = {};
            this.pageResponsitory = {};
            this.fileResponsitory = {};
        }
        popFromStack(sliceIdx) {
            if (typeof (sliceIdx) !== "undefined") {
                this.stack = this.stack.slice(0, sliceIdx);
            }
            else if (this._item === this.stack[this.stack.length - 1])
                this.stack.pop();
            this._item = this.stack.length > 0 ? this.stack[this.stack.length - 1] : null;
            if (this._item) {
                delete this._item.panelViews[UI.Layout.POPUP_PANEL_ID];
                this.model = this._item.model;
            }
            else {
                this.model = null;
            }
        }
        pushToStack(stackItem, replaceCurrentScreen) {
            if (stackItem.screenNavigationType === UI.NavigationType.Home)
                this.stack = [];
            if (stackItem.screenNavigationType !== UI.NavigationType.Immediate) {
                if (this.currentItem == null || this.stack.length === 0 || (this.stack[this.stack.length - 1].screenInstance !== stackItem.screenInstance && !replaceCurrentScreen))
                    this.stack.push(stackItem);
                else
                    this.stack[this.stack.length - 1] = stackItem;
            }
            this._item = stackItem;
        }
        joinSnapshoot(snapshoot) {
            Session.joinObject(this.snapshoot.sessionObject, snapshoot.sessionObject);
            Session.joinObject(this.snapshoot.immediateObject, snapshoot.immediateObject);
        }
        joinObject(target, source) {
            return Object.assign({}, target, source);    
        }
    }
    UI.Session = UISession;
    Session = new UI.Session();

    class EventDispatcher {
        constructor() {
            this.handlers = {};
        }
        addEventListener(eventName, func) {
            this.handlers[eventName] = this.handlers[eventName] || [];
            let eventHandlers = this.handlers[eventName];
            if (eventHandlers.indexOf(func) === -1) {
                eventHandlers.push(func);
                return true;
            }
            return false;
        }
        fireEvent(eventName, param, cancellable = false) {
            let eventHandlers = this.handlers[eventName];
            if (!eventHandlers)
                return true;
            for (let handle of eventHandlers) {
                try {
                    if (handle.call(this, param) === false && cancellable)
                        return false;
                }
                catch (e) {
                    console.log(e);
                }
            }
            return true;
        }
        removeEventListener(eventName, func) {
            let eventHandlers = this.handlers[eventName];
            if (!eventHandlers)
                return null;
            let idx = eventHandlers.indexOf(func);
            return idx > -1 ? eventHandlers.splice(idx, 1)[0] : null;
        }
        clearListeners(eventName) {
            if (eventName) {
                let h = this.handlers[eventName];
                this.handlers[eventName] = [];
                return {
                    [eventName]: h
                };
            }
            let h = this.handlers;
            this.handlers = {};
            return h;
        }
    }
    UI.EventDispatcher = EventDispatcher;
})(UI || (UI = {}));

(function (UI) {
  

    /*
        UI Stucture:
        UI - > Page -> Panels -> View

    */

    const unitStyle = {
        0: "%",
        1: "px",
    };
    const orientationClass = {
        0: "vertical",
        1: "horizontal",
        2: "floating",
    };
    class Panel{
        /*
            {
                name: "panel-name",
                orientation: 0, // 0: vertical, 1: horizontal, 2: floating  
                view: {
                    name: "view-name",
                    type: "view-type",
                    file: "view-content",
                    code: "view-code",
                    script: "view-script",
                    style: "view-style"
                } 

            }
        */
        constructor(page,configuration){
            this.page = page;
            this.configuration  = configuration;
            this.view = null;
            this.panel = null;
            this.name = UI.safeName(this.configuration.name);
        }
        create(){
            this.panel = document.createElement("div");
            this.panel.className = "ui-panel";
            this.panel.classList.add(UI.safeClass(`panel_${this.configuration.name}`));
            this.panel.classList.add(orientationClass[this.configuration.orientation]);
            let paneliId = 'panel_'+UI.generateUUID();
            this.panel.setAttribute("id", paneliId);
            this.panel.id = paneliId;
            this.id = paneliId;
            this.panelElement = this.panel;
            if(this.configuration.inlinestyle){
                this.panelElement.setAttribute("style", this.configuration.inlinestyle);
            }   
            this.displayview();
            this.page.pageElement.appendChild(this.panel);
            this.panelElement.addEventListener("SUBMIT_EVENT_ID", this.submitPanel.bind(this, this.panelId));
            Session.panels[UI.safeId(this.name)] = this;
            return this;
        }
        submitPanel(panelId) {
            let panel = this.page.panels[panelId];
            if (panel) {
                if(panel.view)
                    panel.view.compute();
            } 
        }
        clear(){
            if(this.view)
                this.view.clear();
            this.view = null;
            this.panel.innerHTML = "";
        }
        changeview(view){
            this.clear();
            this.configuration.view = {
                "config": view
            };
            this.displayview();
        }
        displayview(){
            if(this.configuration.view){
                let view = new View(this,this.configuration.view);                
                this.view = view;
            }
        }
    }
    class View extends UI.EventDispatcher{
        /*
            {
                name: "view-name",
                type: "view-type",
                content: "view-content",
                file: "view-content",
                code: "view-code",
                script: "view-script",
                style: "view-style"

            }

        */
        constructor(Panel, configuration){
            super();
            console.log(Panel,configuration)
            this.Panel = Panel;
            if(configuration.name){
                if(configuration.name in Session.viewResponsitory){
                    this.configuration = Session.viewResponsitory[configuration.name];
                    this.builview();
                    return; 
                }
            } 
            if(configuration.config) {

                this.loadconfiguration(configuration);
                
            }
            else
            {    
                this.configuration = configuration; 
                this.builview();  
            }  
            
        }
        async loadconfiguration(configuration){
            await this.loadviewconfiguration(configuration);   
        }
        builview(){
            console.log(this.configuration)
            if(!this.configuration.name)
            {
                return;
            }
            Session.viewResponsitory[UI.safeName(this.configuration.name)] = this.configuration;

            this.id = 'view_'+UI.generateUUID();
            this.view = document.createElement("div");
            this.name = UI.safeName(this.configuration.name);
            this.view.className = "ui-view";
            this.view.classList.add(UI.safeClass(`view_${this.configuration.name}`));
            this.view.setAttribute("id", this.id);
            this.view.setAttribute("viewID", this.id);
            this.Panel.panelElement.appendChild(this.view);
            this.Context = this.id;
            this.inputs={};
            this.outputs ={};
            this.create();
            this.fireOnLoaded();
        }
        clear(){
            this.fireOnUnloading();
            console.log("clear view",this);
            const elements = document.querySelectorAll(`[viewID="${this.id}"]`);
            console.log(elements)
            for (let i = 0; i < elements.length; i++) {
                elements[i].remove();
            }
            //delete Session.view[UI.safeId(this.id)];
        }
        create(){
          //  console.log(this, this.Panel.panelElement);
            Session.views[UI.safeId(this.id)] = this;

            if(this.configuration.inputs)        
                this.createinputs(this.configuration.inputs);
     
            if(this.configuration.content){                
                this.content = this.configuration.content;
                this.view.innerHTML = this.createcontext(this.content);
            }
            if(this.configuration.file)
                this.loadfile(this.configuration.file)
            
            if(this.configuration.style){
                this.createStyleContent(this.configuration.style);                
            }  

            if(this.configuration.inlinestyle){
                this.view.setAttribute("style", this.configuration.inlinestyle);
            } 
                        
            if(this.configuration.script){
                this.loadfile(this.configuration.script).then((response) => {    
                    this.script = response.data;
                    this.createScriptContent(this.script);
                }).catch((error) => {   
                    console.log(error);
                })  
            }

            if(this.configuration.inlinescript){
                let script = this.configuration.inlinescript;
                this.createScriptContent(script);
            }

            Session.views[UI.safeId(this.id)] = this;

            console.log(this,this.onLoaded)
            
            return this;
        }
        async loadviewconfiguration(configuration){
            
            if(configuration.config in Session.fileResponsitory){
                this.configuration = JSON.parse(Session.fileResponsitory[configuration.config]);
                this.configuration = Object.assign({}, this.configuration, configuration); 
            
                this.builview(); 
                return;                
            }


            let ajax = new UI.Ajax(""); 
        
            ajax.get(configuration.config,false).then((response) => {  
              //  return JSON.parse(response);
                Session.fileResponsitory[configuration.config] = response;

                this.configuration = JSON.parse(response);//Object.assign(configuration,response);
                this.configuration = Object.assign({}, this.configuration, configuration); 
                
                this.builview(); 
                             
            }).catch((error) => {
                console.log("error:",error);
            })
        }

        async loadfile(file){

            if(file in Session.fileResponsitory){

                this.buildviewwithresponse(Session.fileResponsitory[file]);
                return;
            }

            let ajax = new UI.Ajax("");
            ajax.get(file,false).then((response) => {
            //    console.log(response)
                Session.fileResponsitory[file] = response;
                //return response;
                this.buildviewwithresponse(response);

            }).catch((error) => {
                console.log(error);
            })
        }
        buildviewwithresponse(response){
            const parser = new DOMParser();
            const doc = parser.parseFromString(response, 'text/html');
            const head = doc.querySelector('head');
            const body = doc.querySelector('body');

            const styles = head.querySelectorAll('link');
            let scripts = doc.querySelectorAll('script[src]');
            

            for (let i = 0; i < styles.length; i++) {
                if(styles[i].href){
                    this.createStyle(styles[i].href);
                    continue;
                }
                else if(styles[i].textContent)
                    this.createStyleContent(styles[i].textContent);
            }
            console.log(scripts)
            for (let i = 0; i < scripts.length; i++) {
                if(scripts[i].src){
                    this.createScript(scripts[i].src);
                    continue;
                }
                else if(scripts[i].textContent)
                    this.createScriptContent(scripts[i].textContent);
            }
            
            scripts = doc.querySelectorAll('script:not([src])');
            
            body.querySelectorAll("script").forEach((script) => {
                script.remove();
            }) 
            
            this.view.appendChild(UI.createFragment(body.innerHTML));
            console.log(scripts)
            for (let i = 0; i < scripts.length; i++) {
                if(scripts[i].src){
                    this.createScript(scripts[i].src);
                    continue;
                }
                else if(scripts[i].textContent)
                    this.createScriptContent(scripts[i].textContent);
            }
        }
        createinputs(inputs){
            let inputscript = "";
        //    console.log(Session.snapshoot.sessionData,inputs);
            Object.keys(inputs).reduce((acc, key) => {
                if(Session.snapshoot.sessionData.hasOwnProperty(key)   ){
                    inputs[key] = Session.snapshoot.sessionData[key];                      
                }
            }, {})

            this.inputs = inputs;
            console.log(inputs)
            return inputscript;
        }
        createoutputs(outputs){
            let outputscript = "";
            Object.keys(outputs).reduce((acc, key) => {  
          
                return acc;
            }, {})

         //   this.createScriptContent(outputscript);
            this.outputs = outputs;
            return outputscript;
        }
        createcommonfunctions(){
          let s = 'object_'+UI.safeId(this.id) + ' = ' + `Session.views[`+this.id+`]` + ';';
          console.log(s)
          this.createScriptContent(s);
        }
        getoutputs(){
            let outputs = this.outputs;
            Object.keys(outputs).reduce((acc, key) => {                
                acc[outputs[key] ] = this.Context+'.outputs.'+key;
                return acc;
            }, {})
            this.outputs = outputs;
            return outputs;
        }
        createScript(path) {
            var s = document.createElement("script");
            s.src = path;
            s.setAttribute("viewid", this.id);
            document.head.appendChild(s);
            return s;
        }
        createScriptContent(Content) {
            var s = document.createElement("script");
            s.type = "text/javascript";
            s.setAttribute("viewid", this.id);
            
            s.textContent =this.createcontext(Content)
            console.log(s);
            this.view.appendChild(s);
           // document.head.appendChild(s);
            return s;
        }
        createStyle(link) {
            var s = document.createElement("link");
            s.href = link;
            s.rel = "stylesheet";
            s.setAttribute("viewid", this.id);
            document.head.appendChild(s);
            return s;
        }
        createStyleContent(Content) {
            var s = document.createElement("style");
            s.setAttribute("type", "text/css");
            s.setAttribute("viewid", this.id);
            s.textContent  = Content;
            document.head.appendChild(s);
            return s;
        }        
        createcontext(content){     
               
            let newcontent = content.replaceAll("$Context", 'Session.views["'+UI.safeId(this.id)+'"]');
            
            return newcontent;
                    
        }
        submit(){
        //    console.log("submit",this.inputs,this.outputs);
            Session.snapshoot.sessionData =  Object.assign({},Session.snapshoot.sessionData, this.getoutputs());
        //    console.log(Session.snapshoot.sessionData)

            if(this.outputs.action){
            //    console.log(this.configuration.actions[this.outputs.action])
                if(this.configuration.actions[this.outputs.action]){
                    var action = this.configuration.actions[this.outputs.action];
                //    console.log("selected action:",this.outputs.action,action)
                    if(action.type == "view"){
                        if(action.panels){
                            console.log("actions:",action.panels)
                            for (var i=0; i<action.panels.length; i++){
                                let viewpanel = action.panels[i].panel;
                                let panel = Session.panels[UI.safeId(viewpanel)];
                                if(panel){
                                    panel.changeview(action.panels[i].view);
                                }
                            }
                        }
                    }                    
                    else if(action.type == "page"){
                    //    console.log("page:",action)
                    //    window.location.href = "UIPage.html?page="+ action.page;
                        let page = new Page({"file":action.page});

                    }

                }

            }
        }
        executeTransaction(url,inputs, func, fail){
            UI.ajax.post(url, inputs).then((response) => {
                   if(typeof(func) == 'function')
                        func(response); 
    
                }).catch((error) => {
                    if(typeof(fail) == 'function')
                        fail(error);
                    console.log(error);
                });
        }
        onLoaded(func) {   
            console.log(func)        
            this.addEventListener("loaded", func);
        }
        fireOnLoaded() {
            this.fireEvent("loaded");
            this.clearListeners("loaded");
        }
        onUnloading(func) {
            this.addEventListener("unloading", func);
        }
        onUnloaded(func) {
            this.addEventListener("unloaded", func);
        }
        fireOnUnloading() {
            this.isUnloading = true;
            this.fireEvent("unloading");
            this.clearListeners("unloading");
            this.node = null;
            this.fireEvent("unloaded");
            this.clearListeners("unloaded");
        }
    }
    
    class Page{
            /*
            sample configuration
            {
                "name": "root page",
                "panels": [{
                        "name": "header",
                        "orientation": 1,
                        "view": {
                            "name": "generic header",
                            "type": "view-type",
                            "content": "<div> header content </div>",
                            "style": "{ \"background-color\": \"blue\", \"height\": \"20%\"}"
                        }
                    },
                    {
                        "name": "content",
                        "orientation": 1,
                        "view": {
                            "name": "generic header",
                            "type": "view-type",
                            "content": "<div> this is the content </div>",
                            "style": "{\"background-color\": \"blue\", \"height\": \"80%\"}"
                        }
                    }
                ]
            }
        */    
        constructor(configuration){
            console.log(configuration)
            this.configuration = configuration;
            this.page={};
            this.panels = [];
            const elements = document.getElementsByClassName('ui-page');
            for (let i = 0; i < elements.length; i++) {
                elements[i].remove();
            }
            
            if(configuration.name in Session.pages)
            {
                this.configuration = Session.pages[configuration.name];
                this.create();
            }
            else if(configuration.file){
                this.loadfile(configuration);
            }
            else{
                Session.pageResponsitory[this.configuration.name] = this.configuration;
                this.create();
            }
           // console.log(this);
        }
        loadfile(configuration){
            if(configuration.file in Session.pageResponsitory){
                this.configuration =  Session.pageResponsitory[configuration.file];
                this.create();
                return;
            }

            let ajax = new UI.Ajax("");
            ajax.get(configuration.file,false).then((response) => {
                
                Session.pageResponsitory[configuration.file] = JSON.parse(response);
                
                this.configuration = JSON.parse(response);
                this.create();
            }).catch((error) => {
                console.log(error);
            })

        }
        create(){
            
            console.log(this.configuration)
            let page = document.createElement("div");
            page.className = "ui-page";
            let id = 'page_'+UI.generateUUID();
            page.setAttribute("id", id);
            page.setAttribute("style", "width:100%;height:100%;position:absolute;top:0;left:0;")
            this.page.id = id;
            this.id = id;
            this.page.element = page;
            this.pageElement = page;
            this.buildpagepanels();
            document.body.appendChild(page);
            Session.pages[this.id] = this;
            return page;
        }
        buildpagepanels(){
            this.page.panels = [];
            for (let i = 0; i < this.configuration.panels.length; i++) {
                let panel = new Panel(this,this.configuration.panels[i]);
                this.page.panel = panel.create();
                this.panels.push(panel);
            }
        }
        clear(){
            /*this.page.innerHTML = "";
            this.panels.each((panel) => {
                panel.clear();
            }); */
            document.getElementById(this.id).remove();
            this.panels = [];
            this.page={};
        }          

    }
    UI.Page = Page;

})(UI || (UI = {}));

(function (UI) {
    function startpage(pagefile){
        console.log(pagefile);
        let page = new UI.Page({file:pagefile});

       /*
        let ajax = new UI.Ajax("");
        ajax.get(pagefile,false).then((response) => {
            console.log(response)
            let page = new UI.Page(JSON.parse(response));

            //page.create();            
        }).catch((error) => {
            console.log(error);
        })
        */
    }
    UI.startpage = startpage;
    function startbyconfig(configuration){
        
        let page = new UI.Page(configuration);
     
    }
    UI.startbyconfig = startbyconfig;
})(UI || (UI = {}));

console.log("UI loaded");
console.log(UI.Ajax);

