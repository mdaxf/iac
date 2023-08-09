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
HTMLElement.prototype.getElementByClassName = function (cls) {
    var list = this.getElementsByClassName(cls);
    if (list.length > 0)
        return list[0];
    return null;
};
HTMLElement.prototype.clearChilds = function () {
    while (this.firstChild)
        this.removeChild(this.firstChild);
};

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
                CurrentPage: Session.cloneObject(this.CurrentPage),
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
    class Popup extends UI.EventDispatcher {
        constructor(container) {
            super();
            this.container = container;
            this.modal = true;
            this.body = document.createElement("div");
        }
        open() {
            if (!this.popup) {
                this.createPopup();
            }
            this.titleEl.textContent = this.title;
            if (this.modal) {
                if (this.overlay == null) {
                    this.overlay = new HTMLOverlay(this.container);
                    this.overlay.content = this.popup;
                }
                this.overlay.open();
            }
            else {
                this.container.appendChild(this.popup);
            }
        }
        createPopup() {
            const div = document.createElement("div");
            div.className = "iac-ui-popup";
            const head = document.createElement("div");
            head.className = "iac-ui-popup-head";
            this.titleEl = document.createElement("span");
            this.titleEl.className = "iac-ui-popup-title";
            head.appendChild(this.titleEl);
            const button = document.createElement("button");
            button.className = "iac-ui-popup-close fa fa-close";
            button.type = "button";
            button.onclick = ev => {
                this.close();
            };
            head.appendChild(button);
            div.appendChild(head);
            this.body.classList.add("iac-ui-popup-body");
            div.appendChild(this.body);
            this.popup = div;
        }
        close() {
            this.fireEvent("close");
            if (this.modal) {
                this.overlay.remove();
                this.overlay = null;
            }
            else {
                this.container.removeChild(this.popup);
            }
        }
        onClose(func) {
            this.addEventListener("close", func);
        }
        offClose(func) {
            this.removeEventListener("close", func);
        }
    }
    UI.Popup = new Popup();

    class HTMLOverlay extends UI.EventDispatcher {
        constructor(root = document.body) {
            super();
            this.overlayElement = HTMLOverlay.createOverlay();
            if (root)
                root.appendChild(this.overlayElement);
            this.overlayElement.aprOverlay = this;
            var thisOverlay = this;
            this.overlayElement.addEventListener("DOMNodeRemoved", function (ev) {
                if (ev.target === this && thisOverlay.visible) {
                    thisOverlay.fireEvent("close", this);
                }
            });
            this.overlayElement.addEventListener("DOMNodeInserted", function (ev) {
                var newelem = ev.target;
                if (newelem.nodeType === Node.ELEMENT_NODE) {
                    var form = newelem.closest('form');
                    if (form) {
                        var autofocus = form.querySelector('input[autofocus]');
                        if (autofocus) {
                            return;
                        }
                    }
                    var popupElements = newelem.querySelectorAll("input:not([type=hidden]),button:not(.close)");
                   // if (popupElements.length > 0)
                   //     SF.Forms.focusElement(popupElements[0]);
                }
            });
            this.overlayElement.addEventListener("keydown", function (event) {
                if (event.key != "Tab")
                    return;
                if (this.querySelectorAll) {
                    var popupElements = this.querySelectorAll("input:not([type=hidden]),button:not(.close)");
                    if (popupElements.length > 0) {
                        if (!event.shiftKey && event.target == popupElements[popupElements.length - 1]) {
                         //   SF.Forms.focusElement(popupElements[0]);
                            event.preventDefault();
                        }
                        else if (event.shiftKey && event.target == popupElements[0]) {
                         //   SF.Forms.focusElement(popupElements[popupElements.length - 1]);
                            event.preventDefault();
                        }
                    }
                }
            });
        }
        get visible() {
            return this._visible;
        }
        get content() {
            return this._content;
        }
        set content(value) {
            if (value) {
                if (this._content)
                    this.overlayElement.replaceChild(value, this._content);
                else
                    this.overlayElement.appendChild(value);
            }
            else if (this._content && this._content.parentElement === this.overlayElement) {
                this.overlayElement.removeChild(this._content);
            }
            this._content = value;
        }
        static createOverlay() {
            var baseDiv = document.createElement("div");
            baseDiv.className = "iac-ui-overlay";
            baseDiv.style.display = "none";
            var overlayFrame = document.createElement("iframe"); //used to overlay over popups etc;
            overlayFrame.attributes["allowTransparency"] = true;
            overlayFrame.src = "about:blank";
            baseDiv.appendChild(overlayFrame);
            return baseDiv;
        }
        onOpen(func) {
            this.addEventListener("open", func);
            return this;
        }
        onClose(func) {
            this.addEventListener("close", func);
            return this;
        }
        open() {
            if (!this.overlayElement)
                throw "can not reopen removed ovelay";
            if (!this._visible) {
                this.prevFocusElement = document.activeElement;
                if (this.prevFocusElement && this.prevFocusElement.blur && this.prevFocusElement != document.body)
                    this.prevFocusElement.blur();
                SF.Forms.focusElement(this.overlayElement);
                this.fireEvent("open", this);
                if (this.content)
                    this.overlayElement.style.display = "";
                this._visible = true;
            }
            return this;
        }
        close() {
            if (this._visible) {
                if (this.prevFocusElement) {
                    SF.Forms.focusElement(this.prevFocusElement);
                    this.prevFocusElement = null;
                }
                this.fireEvent("close", this);
                if (this.content)
                    this.overlayElement.style.display = "none";
                this._visible = false;
            }
            return this;
        }
        remove() {
            if (this._visible) {
                if (this.prevFocusElement) {
                    SF.Forms.focusElement(this.prevFocusElement);
                    this.prevFocusElement = null;
                }
                this.fireEvent("close", this);
                this._visible = false;
            }
            if (this.overlayElement) {
                this.overlayElement.parentElement.removeChild(this.overlayElement);
                this.overlayElement = null;
            }
            return;
        }
        static getCurrent(cont = document.body) {
            const ch = cont.children;
            for (var idx = ch.length - 1; idx > -1; --idx) {
                if (ch[idx].classList.contains("apr-overlay"))
                    return ch[idx].aprOverlay;
            }
            return null;
        }
    }
    UI.HTMLOverlay = HTMLOverlay;
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
                            // clear the crumbs
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
                          //  let stackitem = Session.createStackItem(this);                 
                          //  Session.pushToStack(stackitem);
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
            page.style.flexDirection = "column";
            this.page.id = id;
            this.id = id;
            let pagecontent = document.createElement("div");
            pagecontent.setAttribute("id", id + "_content");
            pagecontent.className = "ui-page-content";
            pagecontent.style.width = "100%";
            pagecontent.style.height = (height-45) + "px";
            pagecontent.style.overflow = "auto";
            pagecontent.style.display = "flex";
            pagecontent.style.flexWrap = "nowrap";
            pagecontent.style.alignItems = "flex-start";
            switch (this.configuration.orientation) {                
                case 0:
                    pagecontent.style.flexDirection = "row";
                    break;
                case 1:
                    pagecontent.style.flexDirection = "column";
                    break;
                case 3:
                    pagecontent.style.flexDirection = "floating";
                default:
                    pagecontent.style.flexDirection = "row";
                    break;
            }
            this.page.element = pagecontent;
            this.pageElement = pagecontent;
            page.appendChild(pagecontent);
            document.title = this.configuration.title || this.configuration.name;
            this.buildpagepanels();
            document.body.appendChild(page);

            this.PageID = id;
            this.PageTitle = this.configuration.title || this.configuration.name;

            Session.pages[this.id] = this;
            Session.CurrentPage = this;
            
            this.setevents();

            new Pageheader(page)

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

    class Pageheader{
        constructor(root){
            this.root = root;
            console.log('create header for page:',root)
            this.element = this.getHeaderContainer();
        }
        removeHeader() {
            let list = this.root.getElementsByClassName("iac-ui-page-header");
            if (list.length !== 0) {
                let p = list[0].parentElement;
                if (p)
                    p.removeChild(list[0]);
            }
        }
        createEl(tag, className = "", textContent = "") {
            const element = document.createElement(tag);
            if (className)
                element.className = className;
            if (textContent)
                element.textContent = textContent;
            return element;
        }
        createElAndAppend(parent, tag, className = "", textContent = "") {
            const element = this.createEl(tag, className, textContent);
            parent.appendChild(element);
            return element;
        }
        getHeaderContainer() {
            let root = this.root;
            let list = root.getElementsByClassName("iac-ui-page-header");
            if (list.length === 0) {
                let header = this.createEl("div", "iac-ui-page-header");
                let headerLeft = this.createElAndAppend(header, "div", "iac-ui-page-header-left");
                let headerCenter = this.createElAndAppend(header, "div", "iac-ui-page-header-center");
                this.createElAndAppend(headerLeft, "span", "iac-ui-icon-logo");
                this.createElAndAppend(headerLeft, "span", "iac-app-name").innerHTML = `<b>IACF</b>`;
                this.headercrumbs = this.createElAndAppend(headerLeft, "span", "iac-ui-crumbs");
                
                this.createElAndAppend(headerCenter, "span", "iac-ui-header-searchIcon");

                let headerRight = this.createElAndAppend(header, "div", "iac-ui-page-header-right");
                this.headerClock = this.createElAndAppend(headerRight, "span", "iac-ui-header-clock clock");
                this.headerUserinfo = this.createElAndAppend(headerRight, "span", "iac-ui-header-userinfo");
                this.headerUserimage = this.createElAndAppend(headerRight, "span", "iac-ui-header-userimage");                                
                this.headerMenuicon = this.createElAndAppend(headerRight, "span", "iac-ui-header-menuIcon");
                root.insertBefore(header, root.firstChild);
                this.rightElementRenderer();
                return header;
            }
            else
                return list[0];
        }
        rightElementRenderer(){
            this.headerBreadcrumbs();
            this.clockRenderer();
            this.userInfoRenderer();
            
        };
        getUserLocalTime(offset, culture) {
            let d = new Date();
            let utc = d.getTime() + d.getTimezoneOffset() * 60000;
            let nd = new Date(utc + offset);
            let formatedTime = nd.toLocaleTimeString(culture);
            return formatedTime;
        }
        getClientLocalTime(){
            let d = new Date();
            return d.toLocaleTimeString();
        }
        clockRenderer(){
            let interval = null;
            let that = this;
            let init = true;
                console.log('clock renderer:', init)
                if (init) {
                    
                    let element = that.createElAndAppend(this.headerClock,"span", "iac-ui-header-clock clock");
                    let updateTime = () => {                       
                        element.textContent = that.getClientLocalTime();
                    };
                    updateTime();
                    interval = setInterval(updateTime, 1000);
                    return element;
                }
                else {
                    clearInterval(interval);
                }
        
        };
        headerIconRender(actionCode, imageUrl, tag = "a") {
            let element = this.createEl(tag);
            if (imageUrl) {
                let img = document.createElement("img");
                img.src = (actionCode != "user") ? ("Images/" + imageUrl) : imageUrl;
                element.appendChild(img);
            }
            element.className = actionCode;
            return element;
        }
        userInfoRenderer() {
            let that = this;
            let init = true;
                if (init) {
                    let userelement = that.createElAndAppend(this.headerUserinfo, "span", "iac-ui-header-userinfo");

                    if(UI.userlogin.username){
                        userelement.textContent = UI.userlogin.username;
                        let element = that.headerIconRender("user", `user/image?username=${UI.userlogin.username}`);
                        element.classList.add("iac-ui-header-action");
                        element.classList.add("active");
                        let list = document.createElement("ul");
                        let li = that.createElAndAppend(list, "li");
                        let a = that.createElAndAppend(li, "a");
                        const icon = that.headerIconRender("logout", "", "span");
                        icon.classList.add("iac-ui-header-action");
                        a.appendChild(icon);
                        let item = that.createElAndAppend(a, "span", "", "Logout");
                        item.textContent = "Logout";
                        item.setAttribute("lngcode", "Logout");    
                        li.addEventListener("click", function () {
                            UI.userlogin.logout(location.href);
                        });
                        this.headerUserimage.appendChild(element);
                        let popup = UI.Popup.createPopup(list);
                        popup.attach(element);
                    }
                    else{
                        userelement.textContent = "Guest";
                        let element = that.headerIconRender("user","portal/images/guest.png");
                        element.classList.add("iac-ui-header-action");
                        element.classList.add("active");
                        let list = document.createElement("ul");
                        let li = that.createElAndAppend(list, "li");
                        let a = that.createElAndAppend(li, "a");
                        const icon = that.headerIconRender("login", "", "span");
                        icon.classList.add("iac-ui-header-action");
                        a.appendChild(icon);
                        let item = that.createElAndAppend(a, "span", "", "Login");
                        item.textContent = "Login";
                        item.setAttribute("lngcode", "Login");    
                        li.addEventListener("click", function () {
                            UI.userlogin.login(location.href);
                        });
                        this.headerUserimage.appendChild(element);
                        let popup = UI.Popup.createPopup(list);
                        popup.attach(element);
                    }
                    return element;

                }

        };
        headerBreadcrumbs(){
            let list;
            let that = this;
            let init = true;
            let headerView;
            const createItem = (item, idx) => {
                let li = document.createElement("li");
                li.title = item.CurrentPage.PageID;
                li.dataset["crumbs"] = idx++ + "";
                that.createElAndAppend(li, "span", "", item.CurrentPage.PageTitle);
                return li;
            };
            const crumbsFunc = (param) => {
                const stack = Session.stack;
                list.clearChilds();
                let idx = 0;
                stack.forEach(item => {
                    list.appendChild(createItem(item, idx++));
                });
             //   Apr.Breadcrumbs.init();
            };
            const viewFunc = (param) => {
                //headerView = app.layout.getCurrentView(SF.HeaderView.RENDERER_PANEL_ID);
            };
          //  return init => {
                if (init) {
                    let container = this.headercrumbs ;// that.createEl("div", "iac-ui-crumbs");
                    list = that.createElAndAppend(container, "ul");
                    const stack = Session.stack;
                    list.clearChilds();
                    let idx = 0;
                    stack.forEach(item => {
                        list.appendChild(createItem(item, idx++));
                    });
                    list.appendChild(createItem(Session, idx++));

                    container.addEventListener("click", ev => {
                        let item = ev.target;
                        while (item.dataset["crumbs"] == null && item != container)
                            item = item.parentElement;
                        let idx = item.dataset["crumbs"];
                        if (idx !== undefined) {
                          //  if (headerView != null) {
                          //      headerView.submitHeaderAction({ toStackIndex: Number(idx) });
                          //  }
                        }
                    });
                

                 //   app.onScreenLoad(crumbsFunc);
                 //   app.onScreenReady(viewFunc);
                 //   return container;
                }
              //  app.offScreenLoad(crumbsFunc);
              //  app.offScreenReady(viewFunc);
             //   headerView = null;
              //  Apr.Breadcrumbs.deinit();
         //   };
        };
    }
    UI.Pageheader = Pageheader;

})(UI || (UI = {}));




(function (UI) {
    
    class ItemController {
        constructor(createFunction) {
            this.createFunction = createFunction;
        }
        get element() {
            if (!this._element) {
                this.toggle(true);
            }
            return this._element;
        }
        toggle(show) {
            if (show !== false) {
                if (!this._element) {
                    this._element = this.createFunction(true, this.setting);
                    return this._element;
                }
            }
            else {
                if (this._element) {
                    this.createFunction(false, this.setting);
                    const p = this._element.parentElement;
                    if (p)
                        p.removeChild(this._element);
                    this._element = null;
                }
            }
            return null;
        }
    }
    class HeaderRenderer {
        constructor(app) {
            this.items = {};
            this.previousItems = [];
            this.items["breadcrumbs"] = new ItemController(HeaderRenderer.bcRenderer(app));
            this.items["clock"] = new ItemController(HeaderRenderer.clockRenderer);
            this.items["search"] = new ItemController(HeaderRenderer.searchRenderer(app));
            this.items["actions"] = new ItemController(HeaderRenderer.actionRenderer(app));
            this.items["userinfo"] = new ItemController(HeaderRenderer.userInfoRenderer);
            this.items["userimage"] = new ItemController(HeaderRenderer.userImageRenderer);
            this.items["customInfoOperation"] = new ItemController(HeaderRenderer.customInfoOperationRenderer);
            this.items["searchIcon"] = new ItemController(HeaderRenderer.searchIconRenderer);
            this.items["menuIcon"] = new ItemController(HeaderRenderer.menuIconRenderer);
        }
        static canShowLogoutButton() {
            const params = new URLSearchParams(window.location.search);
            return params.get('Context') === null && params.get('InvocationContextName') === null;
        }
        static removeHeader(root) {
            let list = root.getElementsByClassName("iac-ui-page-header");
            if (list.length !== 0) {
                let p = list[0].parentElement;
                if (p)
                    p.removeChild(list[0]);
            }
        }
        static createEl(tag, className = "", textContent = "") {
            const element = document.createElement(tag);
            if (className)
                element.className = className;
            if (textContent)
                element.textContent = textContent;
            return element;
        }
        static createElAndAppend(parent, tag, className = "", textContent = "") {
            const element = HeaderRenderer.createEl(tag, className, textContent);
            parent.appendChild(element);
            return element;
        }
        static getHeaderContainer(root) {
            let list = root.getElementsByClassName("iac-ui-page-header");
            if (list.length === 0) {
                let header = HeaderRenderer.createEl("div", "iac-ui-page-header");
                let headerLeft = HeaderRenderer.createElAndAppend(header, "div", "iac-ui-page-header-left");
                HeaderRenderer.createElAndAppend(header, "div", "iac-ui-page-header-center");
                HeaderRenderer.createElAndAppend(headerLeft, "span", "iac-ui-icon-logo");
                HeaderRenderer.createElAndAppend(headerLeft, "span", "iac-app-name").innerHTML = `<b>IACF</b>`;
                root.insertBefore(header, root.firstChild);
                HeaderRenderer.createElAndAppend(header, "div", "iac-ui-page-header-right");
                return header;
            }
            else
                return list[0];
        }
        static getClientLocalTime(offset, culture) {
            let d = new Date();
            let utc = d.getTime() + d.getTimezoneOffset() * 60000;
            let nd = new Date(utc + offset);
            let formatedTime = nd.toLocaleTimeString(culture);
            return formatedTime;
        }
        static getActionClass(action) {
            let list = ["iac-ui-header-action"];
            if (action.cssClasses) {
                list.push(action.cssClasses);
            }
            else if (!action.imageUrl) {
                list.push("default-icon");
            }
            list.push(action.code.toLowerCase());
            return list.join(" ");
        }
        static headerIconRender(actionCode, imageUrl, tag = "a") {
            let element = HeaderRenderer.createEl(tag);
            if (imageUrl) {
                let img = document.createElement("img");
                img.src = (actionCode != "user") ? ("Images/" + imageUrl) : imageUrl;
                element.appendChild(img);
            }
            element.className = actionCode;
            return element;
        }
        initializeSections(root, view) {
            let renderingArray;
            let leftEl = root.querySelector('.iac-ui-page-header-left');
            let centerEl = root.querySelector('.iac-ui-page-header-center');
            let rightEl = root.querySelector('.iac-ui-page-header-right');
            let items = {
                "breadcrumbs": leftEl,
                "search": centerEl,
                "customInfoOperation": leftEl,
                "clock": rightEl,
                "userinfo": rightEl,
                "userimage": rightEl,
                "actions": rightEl,
                "searchIcon": rightEl,
                "menuIcon": rightEl,
            }, baseArray = Object.keys(items);
            if (view) {
                if (view.entity.id === this.currentId)
                    return;
                this.currentId = view.entity.id;
                let props = view.entity.customProperties.replace(/(\w+):/g, '"$1":');
                let json = props ? JSON.parse(props) : {};
                renderingArray = baseArray.slice();
                if (json.ViewOperationPosition > 0 && json.ViewOperationPosition < renderingArray.length - 1) {
                    if (json.ViewOperationPosition > 2) {
                        items["customInfoOperation"] = rightEl;
                    }
                    let ci = renderingArray.splice(renderingArray.indexOf("customInfoOperation"), 1);
                    renderingArray.splice(json.ViewOperationPosition - 1, 0, ci[0]);
                }
                for (var idx = 5; idx > 0; --idx) {
                    let prop = "CustomInfo" + idx;
                    if (json[prop]) {
                        renderingArray.splice(idx - 1, 0, prop);
                        this.items[prop].setting = json[prop];
                    }
                }
                if (json.BreadCrumbs === false)
                    renderingArray.splice(renderingArray.indexOf("breadcrumbs"), 1);
                if (json.Clock === false)
                    renderingArray.splice(renderingArray.indexOf("clock"), 1);
                if (json.Search === false) {
                    renderingArray.splice(renderingArray.indexOf("search"), 1);
                    renderingArray.splice(renderingArray.indexOf("searchIcon"), 1);
                }
                if (json.UserInformation === false) {
                    renderingArray.splice(renderingArray.indexOf("userinfo"), 1);
                    renderingArray.splice(renderingArray.indexOf("userimage"), 1);
                }
                if (!view.entity.actions.length && !HeaderRenderer.canShowLogoutButton()) {
                    renderingArray.splice(renderingArray.indexOf("menuIcon"), 1);
                }
                if (!view.renderHeaderPanel)
                    renderingArray.splice(renderingArray.indexOf("customInfoOperation"), 1);
                if (renderingArray.length === this.previousItems.length && renderingArray.every((item, idx) => item === this.previousItems[idx])) {
                    return;
                }
            }
            else {
                renderingArray = [];
                this.currentId = null;
            }
            if (leftEl.querySelector('iac-ui-crumbs')) {
                root.removeChild(leftEl.querySelector('iac-ui-crumbs'));
            }
            centerEl.clearChilds();
            rightEl.clearChilds();
            for (var idx = 0; idx < renderingArray.length; ++idx) {
                let item = renderingArray[idx];
                let elem = this.items[item].element;
                if (elem) {
                    items[item].appendChild(elem);
                }
                this.previousItems.splice(this.previousItems.indexOf(item), 1);
            }
            this.headerOperationElement = renderingArray.indexOf("customInfoOperation") != -1 ? this.items["customInfoOperation"].element : null;
            for (let item of this.previousItems) {
                this.items[item].toggle(false);
            }
            this.previousItems = renderingArray;
        }
        static get plugin() {
            let headerRenderer = (app, log) => {
                let instance = new HeaderRenderer(app);
                this.isFullScreen = window.location.href.indexOf("fs=true") !== -1;
                if (this.isFullScreen) {
                    //Immediate header container render - no flashing;
                    HeaderRenderer.getHeaderContainer(app.container);
                }
                app.onScreenLoad(p => {
                    if (p.screen.hasHeader || this.isFullScreen) {
                        const view = p.views.find(v => v.panel === SF.HeaderView.RENDERER_PANEL_ID);
                        if (view) {
                            instance.initializeSections(HeaderRenderer.getHeaderContainer(app.container), view);
                            app.layout.initializeHeader(view, app, instance.headerOperationElement);
                        }
                        else
                            log.error("No Header definition!");
                    }
                    else {
                        instance.initializeSections(HeaderRenderer.getHeaderContainer(app.container), null);
                        HeaderRenderer.removeHeader(app.container);
                    }
                });
                return instance;
            };
            headerRenderer.pluginName = "Header";
            return headerRenderer;
        }
    }
    HeaderRenderer.isFullScreen = false;
    HeaderRenderer.clockRenderer = (() => {
        let interval;
        return init => {
            if (init) {
                let element = HeaderRenderer.createEl("span", "iac-ui-header-clock clock");
                let localization = Apr["Localization"];
                if (localization !== undefined) {
                    let options = {
                        hour: "numeric",
                        minute: "numeric",
                        second: "numeric",
                        timeZone: localization.TimeZoneName.replace(" ", "_")
                    };
                    try {
                        var timeFormat = new Intl.DateTimeFormat(localization.UICulture, options);
                    }
                    catch (e) {
                        //for browsers without Intl.DateTimeFormat support
                    }
                }
                let updateTime = () => {
                    let d = new Date();
                    if (timeFormat !== undefined) {
                        element.textContent = timeFormat.format(d);
                    }
                    else if (localization !== undefined) {
                        //for browsers without Intl.DateTimeFormat support
                        element.textContent = HeaderRenderer.getClientLocalTime(localization.TimeZoneOffset, localization.UICulture);
                    }
                    else {
                        element.textContent = d.toLocaleTimeString();
                    }
                };
                updateTime();
                interval = setInterval(updateTime, 1000);
                return element;
            }
            else {
                clearInterval(interval);
            }
        };
    })();
    HeaderRenderer.userInfoRenderer = (() => {
        return init => {
            if (init) {
                return HeaderRenderer.createEl("span", "iac-ui-header-user-name", SF.properties["employeeName"]);
            }
        };
    })();
    HeaderRenderer.userImageRenderer = (() => {
        let popup;
        return init => {
            if (init) {
                let element = HeaderRenderer.headerIconRender("user", `${SF.BADGE_CONTROLLER_URL}/employee?employeeNo=${encodeURIComponent(SF.properties["employeeNo"])}`);
                element.classList.add("iac-ui-header-action");
                if (HeaderRenderer.canShowLogoutButton()) {
                    element.classList.add("active");
                    let list = document.createElement("ul");
                    let li = HeaderRenderer.createElAndAppend(list, "li");
                    let a = HeaderRenderer.createElAndAppend(li, "a");
                    const icon = HeaderRenderer.headerIconRender("logout", "", "span");
                    icon.classList.add("iac-ui-header-action");
                    a.appendChild(icon);
                    let item = HeaderRenderer.createElAndAppend(a, "span", "", "Logout");
                    SF.App.literals.getOne(SF.properties["errorPrefix"] + ".Logout", literal => {
                        if (literal.extendedTranslation)
                            item.textContent = literal.extendedTranslation;
                    });
                    li.addEventListener("click", function () {
                        Apr.LogoutService.logout(location.href);
                    });
                    popup = Apr.ContextPopup.createPopup(list);
                    popup.attach(element);
                }
                return element;
            }
            if (popup)
                popup.remove();
        };
    })();
    HeaderRenderer.bcRenderer = (app) => {
        let list;
        let headerView;
        const createItem = (item, idx) => {
            let li = document.createElement("li");
            li.title = item.screenCode;
            li.dataset["crumbs"] = idx++ + "";
            HeaderRenderer.createElAndAppend(li, "span", "", item.screenTitle);
            return li;
        };
        const crumbsFunc = (param) => {
            const stack = app.session.stack;
            list.clearChilds();
            let idx = 0;
            stack.forEach(item => {
                list.appendChild(createItem(item, idx++));
            });
            Apr.Breadcrumbs.init();
        };
        const viewFunc = (param) => {
            headerView = app.layout.getCurrentView(SF.HeaderView.RENDERER_PANEL_ID);
        };
        return init => {
            if (init) {
                let container = HeaderRenderer.createEl("div", "iac-ui-crumbs");
                list = HeaderRenderer.createElAndAppend(container, "ul");
                container.addEventListener("click", ev => {
                    let item = ev.target;
                    while (item.dataset["crumbs"] == null && item != container)
                        item = item.parentElement;
                    let idx = item.dataset["crumbs"];
                    if (idx !== undefined) {
                        if (headerView != null) {
                            headerView.submitHeaderAction({ toStackIndex: Number(idx) });
                        }
                    }
                });
                app.onScreenLoad(crumbsFunc);
                app.onScreenReady(viewFunc);
                return container;
            }
            app.offScreenLoad(crumbsFunc);
            app.offScreenReady(viewFunc);
            headerView = null;
            Apr.Breadcrumbs.deinit();
        };
    };
    HeaderRenderer.searchRenderer = (app) => {
        let input;
        let div;
        let headerView;
        const keyHandler = (ev) => {
            if (ev.keyCode === 13) {
                search(ev.target);
                ev.preventDefault();
                ev.stopPropagation();
            }
        };
        const search = (elem) => {
            const name = elem.value;
            if (name) {
              /*  const screenKey = { name: name, projectCode: app.currentScreen.projectCode };
                SF.App.screens.getOne(screenKey, s => {
                    if (s.baseScreen)
                        headerView.submitHeaderAction({ toScreen: { name: name, projectCode: app.currentScreen.projectCode } });
                    else
                        throw new SF.SfError("ScreenNotLanding", [name]);
                });  */
            }
            setTimeout(() => elem.focus(), 0);
        };
        const viewFunc = (param) => {
         //   headerView = app.layout.getCurrentView(SF.HeaderView.RENDERER_PANEL_ID);
        };
        const screenLoad = (p) => {
            if (input)
                input.value = p.screen.code;
        };
        return init => {
            if (init) {
              //  app.onScreenReady(viewFunc);
              //  app.onScreenLoad(screenLoad);
                div = HeaderRenderer.createEl("div", "iac-ui-searchbox");
                input = HeaderRenderer.createElAndAppend(div, "input");
                input.placeholder = "Search";
                input.type = "text";
                input.addEventListener("keydown", keyHandler);
                const icon = HeaderRenderer.createElAndAppend(div, "div", "iac-ui-icon-search");
                icon.addEventListener("mousedown", () => {
                    // let input = document.querySelector(".apr-delmia-header .apr-searchbox input") as HTMLInputElement;
                    input["noHideOnBlur"] = true;
                    search(input);
                });
                return div;
            }
        //    app.offScreenLoad(screenLoad);
        //   app.offScreenReady(viewFunc);
            headerView = null;
            input.removeEventListener("keydown", keyHandler);
            return null;
        };
    };
    HeaderRenderer.actionRenderer = (app) => {
        let actionsCont;
        const actionRenderingHandler = (v) => {
           /* if (v.panel === SF.HeaderView.RENDERER_PANEL_ID) {
                actionsCont.clearChilds();
                v.entity.actions.forEach(action => {
                    if (action.type === SF.ActionType.ButtonPrimary || action.type === SF.ActionType.ButtonSecondary) {
                        const element = HeaderRenderer.headerIconRender(HeaderRenderer.getActionClass(action), action.imageUrl);
                        element.setAttribute("style", action.inlineStyle);
                        element.title = action.title || action.code;
                        element.dataset["code"] = action.code;
                        element.addEventListener("click", function (ev) {
                            v.submitHeaderAction({ actionCode: this.dataset["code"] });
                        });
                        actionsCont.appendChild(element);
                    }
                });
            }  */
        };
        return init => {
            if (init) {
                actionsCont = HeaderRenderer.createEl("div", "iac-ui-header-actions");
            //    app.onViewLoaded(actionRenderingHandler);
                return actionsCont;
            }
           // app.offViewLoaded(actionRenderingHandler);
        };
    };
    HeaderRenderer.searchIconRenderer = (() => {
        return init => {
            if (init) {
                let element = HeaderRenderer.createEl("a", "iac-ui-header-action search");
                element.title = "Search";
            /*    SF.App.literals.getOne(SF.properties["errorPrefix"] + ".Search", literal => {
                    if (literal.extendedTranslation)
                        element.title = literal.extendedTranslation;
                });  */
                element.addEventListener("click", window["onHeaderSearchIconClick"]);
                return element;
            }
        };
    })();
    HeaderRenderer.menuIconRenderer = (() => {
        let list;
        const actionRenderingHandler = (v) => {
            if (v.panel === SF.HeaderView.RENDERER_PANEL_ID) {
                list.clearChilds();
                v.entity.actions.slice().reverse().forEach(action => {
                    if (action.type === SF.ActionType.ButtonPrimary || action.type === SF.ActionType.ButtonSecondary) {
                        const li = HeaderRenderer.createElAndAppend(list, "li");
                        const a = HeaderRenderer.createElAndAppend(li, "a");
                        const icon = HeaderRenderer.headerIconRender(HeaderRenderer.getActionClass(action), action.imageUrl, "span");
                        icon.setAttribute("style", action.inlineStyle);
                        a.appendChild(icon);
                        HeaderRenderer.createElAndAppend(a, "span", "", action.title || action.code);
                        a.dataset["code"] = action.code;
                        a.addEventListener("click", function (ev) {
                            v.submitHeaderAction({ actionCode: this.dataset["code"] });
                        });
                    }
                });
                if (HeaderRenderer.canShowLogoutButton()) {
                    const li = HeaderRenderer.createElAndAppend(list, "li", 'logout');
                    const a = HeaderRenderer.createElAndAppend(li, "a");
                    const icon = HeaderRenderer.headerIconRender("logout", "", "span");
                    icon.classList.add("iac-ui-header-action");
                    a.appendChild(icon);
                    const span = HeaderRenderer.createElAndAppend(a, "span", "", "Logout");
                  /*  SF.App.literals.getOne(SF.properties["errorPrefix"] + ".Logout", literal => {
                        if (literal.extendedTranslation)
                            span.textContent = literal.extendedTranslation;
                    });  */
                    li.addEventListener("click", function () {
                        //Apr.LogoutService.logout(location.href);
                    });  
                }
                const li = HeaderRenderer.createElAndAppend(list, "li", "group-header");
                const icon = HeaderRenderer.headerIconRender("user", `${SF.BADGE_CONTROLLER_URL}/employee?employeeNo=${encodeURIComponent(SF.properties["employeeNo"])}`, "span");
                icon.classList.add("iac-ui-header-action");
                li.appendChild(icon);
                HeaderRenderer.createElAndAppend(li, "span", "", SF.properties["employeeName"]);
            }
        };
        return init => {
            if (init) {
                const element = HeaderRenderer.createEl("a", "iac-ui-header-action menu");
                element.title = "Menu";
                /*SF.App.literals.getOne(SF.properties["errorPrefix"] + ".Menu", literal => {
                    if (literal.extendedTranslation)
                        element.title = literal.extendedTranslation;
                }); */
                list = document.createElement("ul");
                list.setAttribute("title", "");
              //  const popup = Apr.ContextPopup.createPopup(list);
              //  popup.attach(element);
              //  SF.app.onViewLoaded(actionRenderingHandler);
                return element;
            }
            //SF.app.offViewLoaded(actionRenderingHandler);
            return null;
        };
    })();
    HeaderRenderer.customInfoOperationRenderer = (() => {
        return init => {
            if (init) {
                let cont = document.createElement("div");
                cont.classList.add("iac-ui-header-panel");
                return cont;
            }
        };
    })();
    UI.HeaderRenderer = HeaderRenderer; 

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
