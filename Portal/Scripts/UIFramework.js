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

var UI;
(function (UI) {
 /*
        common UI functions and classes
        Ajax Call
    */
        UI.CONTROLLER_URL = "";
        class Ajax {
            constructor(token) {
              this.token = token;
              if(!token || token == ''){
                let sessionkey= window.location.origin+"_"+ "user";
                var userdata = sessionStorage.getItem(sessionkey);
                if(userdata){
                    var userjdata = JSON.parse(userdata);
                    this.token = userjdata.token;
                }
              }
            }
          
            initializeRequest(method, url, stream) {
              return new Promise((resolve, reject) => {
                const xhr = new XMLHttpRequest();
                xhr.open(method, `${url}`, true);
                
                if(this.token && this.token !='')
                    xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);

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
          
            getbyurl(url, stream) {
              return this.initializeRequest('GET', url, stream);
            }

            get(url, data, stream= false) {
                return new Promise((resolve, reject) => {
                    const xhr = new XMLHttpRequest();
                    xhr.open('GET', `${url}`, true);                
                
                    if(this.token && this.token !='')
                        xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);

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
                    xhr.send(JSON.stringify(data));
                  });
            }
          
            post(url, data) {
              return new Promise((resolve, reject) => {
                const xhr = new XMLHttpRequest();
                xhr.open('POST', `${url}`, true);
            
                if(this.token && this.token !='')
                    xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);

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
        UI.ajax = new Ajax("");   
})(UI || (UI = {}));

(function (UI) {

    class UserLogin{
        constructor(){
            this.username = "";
            this.password = "";
            this.token = "";
            this.islogin = false;
            this.clientid = "";
            this.createdon = "";
            this.expirateon = "";
            this.tokenchecktime = 1000*60*10;
            this.logonpage = "Logon page";
            this.tokenupdatetimer = null;
            this.updatedon = null;
            this.sessionkey= window.location.origin+"_"+ "user";
        }
        checkiflogin(success, fail){
            let userdata = sessionStorage.getItem(this.sessionkey);
            console.log(userdata)
            if(userdata){
                let userjdata = JSON.parse(userdata);
                this.username = userjdata.username;
                this.password = userjdata.password;
                this.token = userjdata.token;
                this.islogin = userjdata.islogin;
                this.clientid = userjdata.clientid;
                this.createdon = userjdata.createdon;
                this.expirateon = userjdata.expirateon;               
                this.updatedon = userjdata.updatedon;
                this.updatedon = new Date(this.updatedon);

                let checkedtime = new Date(this.updatedon.getTime() + this.tokenchecktime);
                if(checkedtime > new Date()){
                    if(success){
                        success();
                    }
                    return true;
                }

                let parsedDate = new Date(this.expirateon);

                console.log(this.token, parsedDate, new Date(), (parsedDate > new Date()))

                if(parsedDate > new Date()){
                    console.log("renew")
                    UI.ajax.post(UI.CONTROLLER_URL+"/user/login", {"username":this.username, "password":this.password, "token":this.token, "clientid": this.clientid, "renew": true}).then((response) => {
                        userjdata = JSON.parse(response);
                        this.username = userjdata.username;
                        this.password = userjdata.password;
                        this.token = userjdata.token;
                        this.islogin = userjdata.islogin;
                        this.clientid = userjdata.clientid;
                        this.createdon = userjdata.createdon;
                        this.expirateon = userjdata.expirateon;
                        userjdata.updatedon = new Date();

                        console.log(userjdata)
                        sessionStorage.setItem(this.sessionkey, JSON.stringify(userjdata));
                        
                        if(success){
                            success();                        
                        }
                        return true;
    
                    }).catch((error) => {
                        console.log(error)
                        if(fail)
                            fail();
                            return false;
                    })
                }
                    
            }
            if(fail) 
                fail();

            return false;
        }
        login(username, password, success, fail){
            
            let userdata = sessionStorage.getItem(this.sessionkey);
            console.log(userdata)
            if(userdata){
                let userjdata = JSON.parse(userdata);
                this.username = userjdata.username;
                this.password = userjdata.password;
                this.token = userjdata.token;
                this.islogin = userjdata.islogin;
                this.clientid = userjdata.clientid;
                this.createdon = userjdata.createdon;
                this.expirateon = userjdata.expirateon;
            }
            

            if(this.username == username && this.islogin){

                UI.ajax.post(UI.CONTROLLER_URL+"/user/login", {"username":username, "password":password, "token":this.token, "clientid": this.clientid, "renew": true}).then((response) => {
                    userjdata = JSON.parse(response);
                    this.username = userjdata.username;
                    this.password = userjdata.password;
                    this.token = userjdata.token;
                    this.islogin = userjdata.islogin;
                    this.clientid = userjdata.clientid;
                    this.createdon = userjdata.createdon;
                    this.expirateon = userjdata.expirateon;
                    userjdata.updatedon = new Date();

                    sessionStorage.setItem(this.sessionkey, JSON.stringify(userjdata));

                    if(success){
                        success();
                    }   

                }).catch((error) => {
                    if(fail)
                        fail();
                })
            }
            else{
                

                if(this.clientid == ""){
                    this.clientid = UI.generateUUID();
                }
                console.log(this.clientid,username,password)
                UI.ajax.post(UI.CONTROLLER_URL+"/user/login", {"username":username, "password":password, "token":this.token, "clientid": this.clientid, "renew": false}).then((response) => {
                    let userjdata = JSON.parse(response);
                    this.username = userjdata.username;
                    this.password = userjdata.password;
                    this.token = userjdata.token;
                    this.islogin = userjdata.islogin;
                    this.clientid = userjdata.clientid;
                    this.createdon = userjdata.createdon;
                    this.expirateon = userjdata.expirateon;
                    userjdata.updatedon = new Date();

                    sessionStorage.setItem(this.sessionkey, JSON.stringify(userjdata));
                    
                    if(success){
                        success();
                    }                

                }).catch((error) => {
                    console.log(error)
                    if(fail)
                        fail();
                })
            }
        }
        logout(success, fail){
            let userdata = sessionStorage.getItem(this.sessionkey);

            username = userdata.username;
            token = userdata.token;
            clientid = userdata.clientid;

            UI.ajax.post(UI.CONTROLLER_URL+"/user/login", {"username":username, "token":this.token, "clientid": this.clientid}).then((response) => {
                sessionStorage.removeItem(this.sessionkey);
                this.username = "";
                this.password = "";
                this.token = "";
                this.islogin = false;

                if(success){
                    success();
                } 
            
            }).catch((error) => {
                if(fail)
                    fail();
            })

        }
    }
    UI.UserLogin = UserLogin;
    UI.userlogin = new UserLogin();

    function tokencheck(){
        UI.userlogin.checkiflogin(function(){
            console.log("token updated success:", UI.userlogin.username);
            UI.userlogin.tokenupdatetimer = window.setTimeout(tokencheck, UI.userlogin.tokenchecktime);
        }, function(){
            console.log("token updated fail:", UI.userlogin.username);
            console.log(UI.Page);
            if(UI.Page && UI.Page.configuration.name != UI.userlogin.logonpage)
                new UI.Page({file:'pages/logon.json'});
            
        })
    }

    UI.tokencheck = tokencheck;

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

    function isScriptLoaded(scriptSrc) {
        let arr = Array.from(document.getElementsByTagName('script'))
        console.log(scriptSrc, arr);
        for(var i=0;i<arr.length;i++)
        {
            let script = arr[i];
           
            if(script.src == null || script.src == undefined || script.src=="")
            {
                continue;
            }else{

                let src = script.src.toLowerCase();
                let checkedsrc = scriptSrc.toLowerCase();
                let check = src.indexOf(checkedsrc) != -1 || checkedsrc.indexOf(src) != -1;

                console.log(src, checkedsrc, check);

                if (check)
                    return true;
            }
        };
        return false;
        //return Array.from(document.getElementsByTagName('script'))
        //  .some(script => (script.src.toLowerCase().indexOf(scriptSrc.toLowerCase()) !=-1 || scriptSrc.toLowerCase().indexOf(script.src.toLowerCase()) !=-1) );
    }
    UI.isScriptLoaded = isScriptLoaded;

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

    ShowError = function (error) {
        alert(error);
        console.log(error);
    }
    UI.ShowError = ShowError;
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
        createStackItem(instance) {
            return {
                sessionObject: Session.cloneObject(this.snapshoot.sessionObject),
                inputs: Session.cloneObject(this._inputs),
                outputs: Session.cloneObject(this._outputs),
                children: Session.cloneObject(this.children),
                views: Session.cloneObject(this.views),
                panels: Session.cloneObject(this.panels),
                pages: Session.cloneObject(this.pages),
                viewResponsitory: Session.cloneObject(this.viewResponsitory),
                pageResponsitory: Session.cloneObject(this.pageResponsitory),
                fileResponsitory: Session.cloneObject(this.fileResponsitory),
                configuration: Session.cloneObject(this.configurator),
                Instance: instance,
                model: this.model
            };
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
        pushToStack(stackItem) {
            /*if (stackItem.screenNavigationType === UI.NavigationType.Home)
                this.stack = [];
            if (stackItem.screenNavigationType !== UI.NavigationType.Immediate) {
                if (this.currentItem == null || this.stack.length === 0 || (this.stack[this.stack.length - 1].screenInstance !== stackItem.screenInstance && !replaceCurrentScreen))
                    this.stack.push(stackItem);
                else
                    this.stack[this.stack.length - 1] = stackItem;
            } */
            this.stack.push(stackItem);
            this._item = stackItem;
        }
        joinSnapshoot(snapshoot) {
            Session.joinObject(this.snapshoot.sessionObject, snapshoot.sessionObject);
            Session.joinObject(this.snapshoot.immediateObject, snapshoot.immediateObject);
        }
        joinObject(target, source) {
            return Object.assign({}, target, source);    
        }
        cloneObject(targetObject) {
            let temp = {};
            for (var key in targetObject)
                if (Array.isArray(targetObject[key]))
                    temp[key] = targetObject[key].slice();
                else
                    temp[key] = targetObject[key];
            return temp;
        }
        clear(){
            this.stack = [];
            this.snapshoot ={
                "stack":[],
                "configurator":{},
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
            this.configuration.orientation = this.configuration.orientation || 0;
            this.configuration.inlinestyle = this.configuration.inlinestyle || "";
            this.configuration.view = this.configuration.view || {};
            this.configuration.panels = this.configuration.panels || [];
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
            
            if(this.configuration.panels.length > 0){
                for(let panel of this.configuration.panels){
                    let p = new Panel(this.page,panel);
                    p.create();
                    this.panel.appendChild(p.panel);
                }
            }
            else{
                this.displayview();
            }
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
            console.log(Panel,configuration)
            super();
           /* if(Panel.page.configuration.name !="Logon page"){

                this.validelogin(Panel,configuration);
            }
            else
            {
                this.initialize(Panel,configuration);
            }; */
            this.initialize(Panel,configuration);
        }
        initialize(Panel,configuration){            
            
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
        async validelogin(Panel,configuration){
            await UI.userlogin.checkiflogin(function(){
                console.log("login success:", UI.userlogin.username);  
                this.initialize(Panel,configuration);
              }, function(){
                //  UI.startpage("pages/logon.json");
                 // console.log(pagefile);
                  console.log("there is no validated login user!");
                  new UI.Page({file:"pages/logon.json"});
                  return;
              });
        }
        async loadconfiguration(configuration){
            await this.loadviewconfiguration(configuration);   
        }
        builview(){
            this.loaded = false;
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
            this.promiseCount =0;
            this.Promiseitems = {};
            this.create();
                        
        }
        clear(){
            this.fireOnUnloading();
            console.log("clear view",this);
            delete Session.views[this.id];
       //     const elements = document.querySelectorAll(`[viewID="${this.id}"]`);
            const elements = document.querySelectorAll(`[id="${this.id}"]`);
            console.log(elements)
            for (let i = 0; i < elements.length; i++) {
                elements[i].remove();
            }
            //delete Session.view[UI.safeId(this.id)];
        }
        create(){
          //  console.log(this, this.Panel.panelElement);
            Session.views[UI.safeId(this.id)] = this;
            console.log(this.configuration)
            if(this.configuration.inputs)        
                this.createinputs(this.configuration.inputs);
     
            if(this.configuration.content){                
                this.content = this.configuration.content;
                this.view.innerHTML = UI.createFragment(this.createcontext(this.content));
            }
            if(this.configuration.file)
                this.loadfile(this.configuration.file)
            
            if(this.configuration.form)
                this.buildform(this.configuration.form);

            if(this.configuration.code){
                this.createwithCode(this.configuration.code);
            }

            if(this.configuration.style){
                this.createStyleContent(this.configuration.style);                
            }  

            if(this.configuration.inlinestyle){
                this.view.setAttribute("style", this.configuration.inlinestyle);
            } 
                        
            if(this.configuration.script){
                this.createScript(this.configuration.script);
            }

            if(this.configuration.inlinescript){
                let script = this.configuration.inlinescript;
                this.createScriptContent(script);
            }

            Session.views[UI.safeId(this.id)] = this;

            console.log(this,this.onLoaded)
            
            if(this.promiseCount == 0)
                this.fireOnLoaded();

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
        
            await ajax.get(configuration.config,false).then((response) => {  
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
            let that = this
            
            if(file in Session.fileResponsitory){

                this.buildviewwithresponse(Session.fileResponsitory[file]);
                return;
            }

            this.promiseCount = this.promiseCount+1;
            this.Promiseitems[file] = true;

            let ajax = new UI.Ajax("");
            await ajax.get(file,false).then((response) => {
            //    console.log(response)
                Session.fileResponsitory[file] = response;
                //return response;
                that.buildviewwithresponse(response);

                that.promiseCount = that.promiseCount-1;
                this.Promiseitems[file] = false;
                if(that.promiseCount == 0)
                    that.fireOnLoaded();

            }).catch((error) => {
                console.log(error);
            })
        }
        createwithCode(code){
              
            let that = this;
            // load codecontent from the database

            let ajax = new UI.Ajax("");
            ajax.get(code,false).then((response) => {
                that.buildviewwithresponse(response.data);
            }).catch((error) => {
                console.log(error)
            })
        }
        buildform(form){
            // the inputs which needs to be replaced will be liked as {key1}
            let that = this;
            let text = JSON.stringify(form);

            let result = text.replace(/\{([^}]+)\}/g, function(match, key) {
                if (that.inputs.hasOwnProperty(key)) {
                  return that.inputs[key];
                }
                return match;
              });

            let response = JSON.parse(result);  
            new UI.Builder(this.view,response);
        }
        buildviewwithresponse(response){
            let that = this;
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
            console.log('scripts:',scripts)
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
            
            
            console.log(scripts)
            for (let i = 0; i < scripts.length; i++) {
                if(scripts[i].src){                

                    this.createScript(scripts[i].src);
                    continue;
                }
                else if(scripts[i].textContent)
                    this.createScriptContent(scripts[i].textContent);
            }
            
            this.view.appendChild(UI.createFragment(this.createcontext(body.innerHTML)));

        }
        createinputs(inputs){
            let inputscript = "";
            console.log(Session.snapshoot.sessionData,inputs);
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
            let that = this;

            if(UI.isScriptLoaded(path))
                return;

            this.Promiseitems[path] = true;
            that.promiseCount = that.promiseCount+1; 
           
            var s = document.createElement("script");
            s.src = path;
            s.async = false;
            s.setAttribute("viewid", this.id);

            s.onload = function () {
                that.Promiseitems[path] = false;
                that.promiseCount = that.promiseCount-1;
                if(that.promiseCount == 0)
                    that.fireOnLoaded();
            };
            document.head.appendChild(s);            
            return s;
        }
        createScriptContent(Content) {
            
            var s = document.createElement("script");
            s.type = "text/javascript";
            s.setAttribute("viewid", this.id);
            
            s.textContent =this.createcontext(Content)
        //    console.log(s);
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

            if(this.outputs.action){
                this.executeactionchain();
            }
        }
        executeactionchain(){
            if(Session.snapshoot.sessionData.action){
                //    console.log(this.configuration.actions[this.outputs.action])
                    if(this.configuration.actions[this.outputs.action]){
                        var action = this.configuration.actions[this.outputs.action];
                    //    console.log("selected action:",this.outputs.action,action)
                        if(action.type == "Transaction"){
                            let url = "/exetrancode";
                            let data = {
                                "code":action.code,
                                "inputs":Session.snapshoot.sessionData,
                            }
                            this.executeTransaction(url,data, this.updateoutputs, function(error){console.log(error)});
                            if(Session.snapshoot.sessionData.action !="")
                                this.executeactionchain();
                        }
                        else if(action.type == "Home"){
                            Session.clear();
                            let page = new Page({"file":"page/home.json"});
                        }
                        else if(action.type == "view"){
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
                            if(action.page.toLowerCase().indexOf("page/home.json") !=-1){
                                Session.clear();
                            }   
                            let stackitem = Session.createStackItem(this);                 
                            Session.pushToStack(stackitem);
                            let page = new Page({"file":action.page});    
                        }
                        else if(action.type == "script"){
                            if(action.script){
                                action.script(Session.snapshoot.sessionData);
                            }
                        }
    
                    }
    
            }  

        }
        updateoutputs(outputs){
            Session.snapshoot.sessionData =  Object.assign({},Session.snapshoot.sessionData, outputs);
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
        executeLoadData(url,inputs, func, fail){
            console.log('execute loading data')
            UI.ajax.get(url, inputs, false).then((response) => {
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
            this.loaded = true;
            let readyevent = new CustomEvent("Viewready");
            this.view.dispatchEvent(readyevent);
        }
        fireOnLoaded() {
            let that = this
            if(that.loaded){
                console.log("fireOnLoaded",document.readyState)
                this.fireEvent("loaded");
                this.clearListeners("loaded");
            }
            else{
                this.view.addEventListener("Viewready", function() {
                    console.log("fireOnLoaded with Viewready event")
                    that.fireEvent("loaded");
                    that.clearListeners("loaded");
                    that.view.removeEventListener("Viewready",this);
                });
            }
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
    UI.View = View;
    
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
            this.configuration.title = this.configuration.title || this.configuration.name;

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
            if(configuration.name !=UI.userlogin.logonpage && UI.userlogin.tokenupdatetimer == null){
                UI.tokencheck();
            }
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
            
            this.configuration.attrs = this.configuration.attrs || {};
            this.configuration.orientation = this.configuration.orientation || 0;

            for (var key in this.configuration.attrs) {
                page.setAttribute(key, this.configuration.attrs[key]);
            }
            let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            
            page.style.width = width + "px";
            page.style.height = height + "px";
            page.style.display = "flex";
            page.style.flexWrap = "nowrap";
            page.style.overflow = "hidden";
            page.style.alignItems = "flex-start";
            switch (this.configuration.orientation) {                
                case 0:
                    page.style.flexDirection = "row";
                    break;
                case 1:
                    page.style.flexDirection = "column";
                    break;
                case 3:
                    page.style.flexDirection = "floating";
                default:
                    page.style.flexDirection = "row";
                    break;
            }
            this.page.id = id;
            this.id = id;
            this.page.element = page;
            this.pageElement = page;
            document.title = this.configuration.title || this.configuration.name;
            this.buildpagepanels();
            document.body.appendChild(page);
            Session.pages[this.id] = this;
            this.setevents();
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
            
            let that =this;
            window.removeEventListener("resize", that.resize)
            document.getElementById(this.id).remove();
            this.panels = [];
            this.page={};
        }  
        setevents(){
            console.log('set events')
            let that = this;
            window.addEventListener("resize", that.resize)
        }
        resize(){
         //   console.log('start to resize')
            let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            $('.ui-page').css('width',width+'px');
            $('.ui-page').css('height',height+'px');

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

/*
console.log("UI loaded");
console.log(UI.Ajax);
*/
