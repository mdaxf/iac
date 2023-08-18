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
            this.userID = 0;
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
            this.loginurl = "login.html"
        }
        checkiflogin(success, fail){
            let userdata = sessionStorage.getItem(this.sessionkey);
        //    // console.log(userdata)
            if(userdata){
                let userjdata = JSON.parse(userdata);
                this.userID = userjdata.ID;
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

           //     // console.log(this.token, parsedDate, new Date(), (parsedDate > new Date()))

                if(parsedDate > new Date()){
                  //  // console.log("renew")
                    UI.ajax.post(UI.CONTROLLER_URL+"/user/login", {"username":this.username, "password":this.password, "token":this.token, "clientid": this.clientid, "renew": true}).then((response) => {
                        userjdata = JSON.parse(response);
                        this.userID = userjdata.ID;
                        this.username = userjdata.username;
                        this.password = userjdata.password;
                        this.token = userjdata.token;
                        this.islogin = userjdata.islogin;
                        this.clientid = userjdata.clientid;
                        this.createdon = userjdata.createdon;
                        this.expirateon = userjdata.expirateon;
                        userjdata.updatedon = new Date();

                        // console.log(userjdata)
                        sessionStorage.setItem(this.sessionkey, JSON.stringify(userjdata));
                        
                        if(success){
                            success();                        
                        }
                        return true;
    
                    }).catch((error) => {
                        // console.log(error)
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
       //     // console.log(userdata)
            if(userdata){
                
                let userjdata = JSON.parse(userdata);
                this.userID = userjdata.ID;
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
                    this.userID = userjdata.ID;
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
                // console.log(this.clientid,username,password)
                UI.ajax.post(UI.CONTROLLER_URL+"/user/login", {"username":username, "password":password, "token":this.token, "clientid": this.clientid, "renew": false}).then((response) => {
                    let userjdata = JSON.parse(response);
                    this.userID = userjdata.ID;
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
                    // console.log(error)
                    if(fail)
                        fail();
                })
            }
        }
        logout(success, fail){
            let userdata = sessionStorage.getItem(this.sessionkey);
            

            if(userdata){
                let userjdata = JSON.parse(userdata);
         //       // console.log(userjdata)
                let username = userjdata.username;
                let token = userjdata.token;
                let clientid = userjdata.clientid;

                UI.ajax.post(UI.CONTROLLER_URL+"/user/logout", {"username":username, "token":token, "clientid": clientid}).then((response) => {
                    sessionStorage.removeItem(this.sessionkey);
                    this.userID = 0;
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
            sessionStorage.removeItem(this.sessionkey);
            this.username = "";
            this.userID = 0;
            this.password = "";
            this.token = "";
            this.islogin = false;
            window.clearTimeout(this.tokenupdatetimer);
            window.location.href = this.loginurl;
        }
    }
    UI.UserLogin = UserLogin;
    UI.userlogin = new UserLogin();

    function tokencheck(){
        UI.userlogin.checkiflogin(function(){
            // console.log("token updated success:", UI.userlogin.username);
            UI.userlogin.tokenupdatetimer = window.setTimeout(tokencheck, UI.userlogin.tokenchecktime);
        }, function(){
            // console.log("token updated fail:", UI.userlogin.username);
            // console.log(UI.Page);
            if(UI.Page)
                window.location.href = UI.userlogin.loginurl;
             //   new UI.Page({file:'pages/logon.json'});
            
        })
    }

    UI.tokencheck = tokencheck;

})(UI || (UI = {}));

(function (UI) {
    class HomeMenu{
        constructor(){
            this.sessionmenuprefix= window.location.origin+"_"+ "menu_";
        }
        loadMenus(parentID,Success, Fail){
            let userdata = sessionStorage.getItem(UI.userlogin.sessionkey);
        //     console.log(userdata)
            if(userdata){
                
                let userjdata = JSON.parse(userdata);
                let userID = userjdata.id;
                let isMobile = this.isMobile()? "1":"0";
                sessionStorage.removeItem(this.sessionmenuprefix+userID)
                UI.ajax.get(UI.CONTROLLER_URL+"/user/menus?userid="+userID +"&mobile="+isMobile + "&parentid="+parentID).then((response) => {
                    let menus = JSON.parse(response);
                    sessionStorage.setItem(this.sessionmenuprefix+userID+ "_"+parentID, JSON.stringify(menus));
                    if(Success)
                        Success(menus);
                }).catch((error) => {
                    console.log(error);
                    if(Fail)
                        Fail();
                })
            }

        }
        isMobile(){
            var userAgent = navigator.userAgent;
            var isMobile = /Mobi|Android/i.test(userAgent);
            return isMobile;
        }
    }

    UI.HomeMenu = new HomeMenu();
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
        // console.log(scriptSrc, arr);
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

                // console.log(src, checkedsrc, check);

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
        console.log(error);
        alert(error);
         
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
            let item = null;
            if (typeof (sliceIdx) !== "undefined") {
                item = this.stack[sliceIdx];
                this.stack = this.stack.slice(0, sliceIdx);
            }
            else if(this.stack.length > 0){
                item = this.stack[this.stack.length - 1]
                this.stack.pop();
                
            }
            this._item = this.stack.length > 0 ? this.stack[this.stack.length - 1] : null;
           /* if (this._item) {
                delete this._item.panelViews[UI.Layout.POPUP_PANEL_ID];
                this.model = this._item.model;
            }
            else {
                this.model = null;
            } */
        //    // console.log(item, this.stack, this._item)
            return item;
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
        clearstack(){
            this.stack = [];
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
                    // console.log(e);
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

function rAFThrottle(func) {
    var _busy = false;
    return function () {
        if (!_busy) {
            _busy = true;
            var args = arguments;
            window.requestAnimationFrame(() => {
                _busy = false;
                func.apply(this, args);
            });
        }
    };
}

(function (UI) {
    class ContextPopup {
        constructor(menu) {
            this.visible = false;
            var dropdown = document.createElement("div");
            dropdown.classList.add("iac-ui-header-popup");
            var triangle = document.createElement("div");
            triangle.classList.add("triangle");
            dropdown.appendChild(triangle);
            dropdown.appendChild(menu);
            dropdown.addEventListener("click", ev => this.close());
            this.element = dropdown;
            ContextPopup.attachPopupEvents();
        }
        static hidePopups(exceptElement) {
            while (exceptElement != null && exceptElement.classList != null && !exceptElement.classList.contains("iac-ui-header-popup")) {
                exceptElement = exceptElement.parentElement;
            }
            for (var idx = 0; idx < this.popups.length; ++idx) {
                if (this.popups[idx].element != exceptElement) {
                    this.popups[idx].close();
                }
            }
        }
        attach(element) {
            element.addEventListener("click", ev => {
                this.open();
                ev.stopPropagation();
            });
            element.appendChild(this.element);
        }
        static attachPopupEvents() {
            if (!this._popupEventHandlers) {
                this._popupEventHandlers = [];
                var mouseDownHandler = ev => {
                    this.hidePopups(ev.target);
                };
                document.addEventListener("mousedown", mouseDownHandler);
                this._popupEventHandlers.push(mouseDownHandler);
                var resizeHandler = rAFThrottle(ev => {
                    this.hidePopups(null);
                });
                window.addEventListener("resize", resizeHandler);
                this._popupEventHandlers.push(resizeHandler);
            }
        }
        open() {
            if (this.visible)
                return;
            this.visible = true;
            this.element.parentElement.classList.add('iac-ui-active');
            this.element.style.visibility = "visible";
            var rightEdge = this.element.offsetLeft + this.element.offsetWidth;
            var diff = window.innerWidth - rightEdge;
            if (diff < 0) {
                var offsideX;
                if (this.element.parentElement.classList.contains('user')) {
                    offsideX = -(this.element.offsetWidth) + 44;
                }
                else {
                    offsideX = -(this.element.offsetWidth / 2) + (this.element.parentElement.offsetWidth / 2);
                }
                var tr = this.element.querySelector(".triangle");
                this.element.style.transform = "translateX(" + offsideX + "px)";
                tr.style.left = this.element.offsetWidth - (this.element.parentElement.offsetWidth / 2) - 8 + "px";
            }
            else if (!this.element.parentElement.classList.contains('iac-ui-crumbs')) {
                this.element.style.transform = "translateX(calc(-50% + 22px)";
            }
        }
        close() {
            if (!this.visible)
                return;
            this.visible = false;
            this.element.style.visibility = "hidden";
            this.element.parentElement.classList.remove('iac-ui-active');
        }
        remove() {
            ContextPopup.popups.splice(ContextPopup.popups.indexOf(this), 1);
            if (this.element && this.element.parentElement) {
                this.element.parentElement.removeChild(this.element);
            }
            if (ContextPopup.popups.length == 0)
                ContextPopup.detachPopupEvents();
        }
        static detachPopupEvents() {
            if (this._popupEventHandlers) {
                document.removeEventListener("mousedown", this._popupEventHandlers[0]);
                window.removeEventListener("resize", this._popupEventHandlers[1]);
                this._popupEventHandlers = null;
            }
        }
        static createPopup(menu) {
            var popup = new ContextPopup(menu);
            this.popups.push(popup);
            return popup;
        }
    }
    ContextPopup.popups = [];
    ContextPopup._popupEventHandlers = null;
    UI.ContextPopup = ContextPopup;

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
                if(this.overlay)
                    this.overlay.remove();
                this.overlay = null;
                this.modal = null;
            }
            else {
               // this.container.removeChild(this.popup);
               this.popup.remove();
            }
        }
        onClose(func) {
            this.addEventListener("close", func);
        }
        offClose(func) {
            this.removeEventListener("close", func);
        }

    }
    UI.Popup = Popup;

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
                //UI.Forms.focusElement(this.overlayElement);
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
                    //UI.Forms.focusElement(this.prevFocusElement);
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
                    //UI.Forms.focusElement(this.prevFocusElement);
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
    const POPUP_PANEL_ID = "-ui-page-popup-panel";

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
          //  console.log(page,configuration)
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
        create(popup = false){
            this.panel = document.createElement("div");
            this.panel.className = "ui-panel";
            this.panel.classList.add(UI.safeClass(`panel_${this.configuration.name}`));
            this.panel.classList.add(orientationClass[this.configuration.orientation]);
            if(popup)
                this.panel.classList.add("ui-page-popup-panel");

            let paneliId = "";

            if(popup)
                paneliId = POPUP_PANEL_ID;
            else
                paneliId = 'panel_'+UI.generateUUID();

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
            if(!popup)
                this.page.pageElement.appendChild(this.panel);
            else
            {
                
                this.page.popup.body.appendChild(this.panel);
                if(this.configuration.view){
                    let title = this.configuration.view.title || this.configuration.view.name || this.configuration.name;
                    $(".iac-ui-popup-title").html(title);
                }
                else{
                    $(".iac-ui-popup-title").html(this.configuration.name);
                }
            }
            //this.page.pageElement.appendChild(this.panel);
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
            // console.log(Panel,configuration)
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
                // console.log("login success:", UI.userlogin.username);  
                this.initialize(Panel,configuration);
              }, function(){
                //  UI.startpage("pages/logon.json");
                 // // console.log(pagefile);
                  // console.log("there is no validated login user!");
                //  new UI.Page({file:"pages/logon.json"});
                  window.location.href = UI.userlogin.loginurl;
                //  return;
              });
        }
        async loadconfiguration(configuration){
            await this.loadviewconfiguration(configuration);   
        }
        builview(){
            this.loaded = false;
            // console.log(this.configuration)
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
            // console.log("clear view",this);
            delete Session.views[this.id];
       //     const elements = document.querySelectorAll(`[viewID="${this.id}"]`);
            const elements = document.querySelectorAll(`[id="${this.id}"]`);
            // console.log(elements)
            for (let i = 0; i < elements.length; i++) {
                elements[i].remove();
            }
            //delete Session.view[UI.safeId(this.id)];
        }
        create(){
          //  // console.log(this, this.Panel.panelElement);
            Session.views[UI.safeId(this.id)] = this;
            // console.log(this.configuration)
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

            // console.log(this,this.onLoaded)
            
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
                // console.log("error:",error);
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
            //    // console.log(response)
                Session.fileResponsitory[file] = response;
                //return response;
                that.buildviewwithresponse(response);

                that.promiseCount = that.promiseCount-1;
                this.Promiseitems[file] = false;
                if(that.promiseCount == 0)
                    that.fireOnLoaded();

            }).catch((error) => {
                // console.log(error);
            })
        }
        createwithCode(code){
              
            let that = this;
            // load codecontent from the database

            let ajax = new UI.Ajax("");
            ajax.get(code,false).then((response) => {
                that.buildviewwithresponse(response.data);
            }).catch((error) => {
                // console.log(error)
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
            // console.log('scripts:',scripts)
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
            
            
            // console.log(scripts)
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
            // console.log(Session.snapshoot.sessionData,inputs);
            Object.keys(inputs).reduce((acc, key) => {
                if(Session.snapshoot.sessionData.hasOwnProperty(key)   ){
                    inputs[key] = Session.snapshoot.sessionData[key];                      
                }
            }, {})

            this.inputs = inputs;
            // console.log(inputs)
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
          // console.log(s)
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
        //    // console.log(s);
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
            newcontent = newcontent.replaceAll("$PageID", UI.safeId(this.id));
            newcontent = newcontent.replaceAll("$ViewID", UI.safeId(this.id));
            newcontent = newcontent.replaceAll("$View", 'Session.views["'+UI.safeId(this.id)+'"]');
            return newcontent;
        }
        submit(){
        //    // console.log("submit",this.inputs,this.outputs);
            Session.snapshoot.sessionData =  Object.assign({},Session.snapshoot.sessionData, this.getoutputs());

            if(this.outputs.action){
                this.executeactionchain();
            }
        }
        executeactionchain(){
            if(Session.snapshoot.sessionData.action){
                //    // console.log(this.configuration.actions[this.outputs.action])
                    if(this.configuration.actions[this.outputs.action]){
                        var action = this.configuration.actions[this.outputs.action];
                        // console.log("selected action:",this.outputs.action,action)
                        if(action.type == "Transaction"){
                            let url = "/exetrancode";
                            let data = {
                                "code":action.code,
                                "inputs":Session.snapshoot.sessionData,
                            }
                            this.executeTransaction(url,data, this.updateoutputs, function(error){ console.log(error)});
                            if(Session.snapshoot.sessionData.action !="")
                                this.executeactionchain();
                        }
                        else if(action.type == "Home"){
                            this.Panel.page.clear();
                            Session.clearstack();
                            // clear the crumbs
                            let page = new Page({"file":"page/home.json"});
                        }
                        else if(action.type == "Back"){
                            if(Sesion.stack.length > 0){
                                let stackitem = Session.popFromStack();
                                this.Panel.page.clear();
                                let page = new Page(stackitem.page.configuration);
                                
                            }
                        }
                        else if(action.type == "view"){
                            if(action.panels){
                                // console.log("actions:",action.panels)
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
                            this.Panel.page.clear();
                            if(action.page.toLowerCase().indexOf("home.json") !=-1){
                                Session.clearstack();
                            }
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
                    // console.log(error);
                });
        }
        executeLoadData(url,inputs, func, fail){
            // console.log('execute loading data')
            UI.ajax.get(url, inputs, false).then((response) => {
                if(typeof(func) == 'function')
                    func(response); 
     
                }).catch((error) => {    
                    if(typeof(fail) == 'function')
                        fail(error);
                        // console.log(error);
                });
        }
        onLoaded(func) {   
            // console.log(func)        
            this.addEventListener("loaded", func);
            this.loaded = true;
            let readyevent = new CustomEvent("Viewready");
            this.view.dispatchEvent(readyevent);
        }
        fireOnLoaded() {
            let that = this
            if(that.loaded){
                // console.log("fireOnLoaded",document.readyState)
                this.fireEvent("loaded");
                this.clearListeners("loaded");
            }
            else{
                this.view.addEventListener("Viewready", function() {
                    // console.log("fireOnLoaded with Viewready event")
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
    
    class Page extends UI.EventDispatcher{
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
        constructor(configuration, pageID = null) {
            super();
            // console.log(configuration)
            this.configuration = configuration;
            this.page={};
            this.panels = [];

            const elements = document.getElementsByClassName('ui-page');
            for (let i = 0; i < elements.length; i++) {
                elements[i].remove();
            }
            UI.tokencheck();
                        
            if(pageID != null && pageID in Session.pages && pageID != undefined){
                // console.log("pageID:",pageID)
                this.configuration = Session.pages[pageID].configuration;
                this.id = pageID;
                this.create();
            }
            else if(configuration.name in Session.pageResponsitory)
            {
                this.configuration = Session.pageResponsitory[configuration.name];
                let found = false;
                
                for(var key in Session.pages){
                  //  console.log(key, this.configuration)
                    if(Session.pages[key].configuration.name == this.configuration.name){
                        // console.log(Session.pages[key], configuration.name)
                        this.id = key;
                        found = true;
                        this.create();
                        return;
                    }
                }
                if(!found)
                    this.init();
            }
            else if(configuration.file){
                this.loadfile(configuration);
            }
            else{
                Session.pageResponsitory[this.configuration.name] = this.configuration;
                this.init();
            }

        }
        loadfile(configuration){
            if(configuration.file in Session.pageResponsitory){
                this.configuration =  Session.pageResponsitory[configuration.file];
                let found = false;
                if(Session.pages && Session.pages.length > 0){
                    for(var key in Session.pages){
                        if(Session.pages[key].configuration.name == this.configuration.name){
                            this.id = key;
                            found = true;
                            this.create();
                            return;
                        }
                    }
                }
                if(!found)
                    this.init();
                return;
            }

            let ajax = new UI.Ajax("");
            ajax.get(configuration.file,false).then((response) => {
                let pagedata = JSON.parse(response);
                Session.pageResponsitory[configuration.file] = pagedata;
                
                this.configuration = pagedata;
                this.init();
            }).catch((error) => {
                // console.log(error);
            })

        }
        async init(){
            let id = 'page_'+UI.generateUUID();
            this.id = id;

            if(this.configuration.onInitialize){

                let url = "/exetrancode";
                let data = {
                    "code":this.configuration.onInitialize,
                    "inputs":Session.snapshoot.sessionData,
                }
                await this.executeTransaction(url,data, this.updatesession, function(error){
                        console.log(error)
                });   
            }

            this.create();
        }
        async create(){  
            this.configuration.title = this.configuration.title || this.configuration.name;     

            if(this.configuration.name == "IAC Home"){
                Session.clearstack();
            }

            if(this.configuration.onLoad){

                let url = "/exetrancode";
                let data = {
                    "code":this.configuration.onLoad,
                    "inputs":Session.snapshoot.sessionData,
                }
                await this.executeTransaction(url,data, this.updatesession, function(error){
                     console.log(error)
                });   
            }
            
            // console.log(this.configuration)
            let id = this.id;
            let page = document.createElement("div");
            page.className = "ui-page";
            
            page.setAttribute("id", this.id);
            this.container = page;

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
            this.page.id = this.id;
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
         //   document.title = this.configuration.title || this.configuration.name;
            this.buildpagepanels();
            document.body.appendChild(page);

            this.PageID = id;
            this.PageTitle = this.configuration.title || this.configuration.name;

           // console.log(Session.snapshoot.sessionData,this.configuration.title)
            for(var key in Session.snapshoot.sessionData){
              //  console.log(key, this.configuration.title)
              this.PageTitle = this.PageTitle.replaceAll('{'+key+'}' , Session.snapshoot.sessionData[key])
            }
            document.title = this.PageTitle ;

            Session.pages[this.id] = this;
            Session.CurrentPage = this;
                        
            let stackitem = Session.createStackItem(this);                 
            Session.pushToStack(stackitem);

            new Pageheader(page)
            
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
        updatesession(data){
            Session.snapshoot.sessionData =  Object.assign({},Session.snapshoot.sessionData, data);
        }
        async executeTransaction(url,inputs, func, fail){
            UI.ajax.post(url, inputs).then((response) => {
                   if(typeof(func) == 'function')
                        func(response); 
    
                }).catch((error) => {
                    if(typeof(fail) == 'function')
                        fail(error);
                    // console.log(error);
                });
        }
        Refresh(){
           this.clear();
           this.create(); 
        }
        popupOpen(view) {
            if (!this.popup) {
                this.popup = new UI.Popup(this.container);
                this.initializePopup(view);
                this.popup.onClose(() => {
                    this.clearPopup();
                });
            }
            this.popup.title = view.title;
            this.popup.open();
        }
        popupClose() {
            if (this.popup)
                this.popup.close();
            this.popup = null;
        }
        initializePopup(view) {         
            if($('#'+ POPUP_PANEL_ID).length == 0){  
                const panel = new Panel(this,{
                    "name": POPUP_PANEL_ID, 
                    "view": view});
                panel.create(true);
                this.panels.push(panel);
            }
        }
        clearPopup() {
            if (this.popup)
                this.popup.close();
            this.popup = null;
        }
        back(){
            if(Session.stack.length > 0){
                let stackitem = Session.popFromStack();
                // console.log("page back action:", stackitem)

                if(!stackitem)
                    return;
                
                if(stackitem.CurrentPage)
                    new Page(stackitem.CurrentPage.configuration);
                    
            }
        }
        home(){
            if(this.configuration.name == "IAC Home")
                return;
            this.clear();
            Session.clearstack();
            new Page({"file":"pages/home.json"});
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
            // console.log('set events')
            let that = this;
            window.addEventListener("resize", that.resize)
        }
        resize(){
         //   // console.log('start to resize')
            let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            $('.ui-page').css('width',width+'px');
            $('.ui-page').css('height',height+'px');
            $('.ui-page-content').css('width',width+'px');
            $('.ui-page-content').css('height',(height-45)+'px');
        }

    }
    UI.Page = Page;

    class Pageheader{
        constructor(root){
            this.root = root;
            // console.log('create header for page:',root)
            this.element = this.getHeaderContainer();
            this.headerBreadcrumbs();
            this. headerMenuActions();
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
            else{
                this.headercrumbs = root.getElementsByClassName("iac-ui-crumbs")[0];
                this.headerMenuicon = root.getElementsByClassName("iac-ui-header-menuIcon")[0];
                return list[0];
            }
        }
        rightElementRenderer(){
            
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
                // console.log('clock renderer:', init)
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
                        let ajax = new UI.Ajax();
                        let element = that.headerIconRender("user", "images/avatardefault.png");
                        ajax.get(`../user/image?username=${UI.userlogin.username}`, function (data) {
                            let imageurl = JSON.parse(data);
                            // console.log(imageurl)
                                if (imageurl) {
                                    element.children[0].src = imageurl;
                                }
                            }, function (error) {
                                 console.log(error);
                            });

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
                                // console.log("logout")
                                UI.userlogin.logout();
                            });
                            this.headerUserimage.appendChild(element);
                            let popup = UI.ContextPopup.createPopup(list);
                            // console.log(popup)
                            popup.attach(element);
                    }
                    else{
                        userelement.textContent = "Guest";
                        let element = that.headerIconRender("user","images/avatardefault.png");
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
                            
                            window.location.href = UI.userlogin.loginurl;
                        });
                        this.headerUserimage.appendChild(element);
                        let popup = UI.UI.ContextPopup.createPopup(list);
                        // console.log(popup)
                        popup.attach(element);
                    }
                //    return element;

                }

        };
        headerBreadcrumbs(){
            let list;
            let that = this;
            let init = true;
            let headerView;
            const createItem = (item, idx) => {
                let li = document.createElement("li");
                li.setAttribute("pageid",item.CurrentPage.PageID);
                li.title = item.CurrentPage.PageTitle;
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
                   // list.appendChild(createItem(Session, idx++));

                    container.addEventListener("click", ev => {
                        let item = ev.target;
                        while (item.dataset["crumbs"] == null && item != container)
                            item = item.parentElement;
                        let idx = item.dataset["crumbs"];
                        if (idx !== undefined) {
                            let pageid = item.attributes["pageid"].value;
                            Session.popFromStack(idx);
                            let page = new UI.Page("",pageid);
                        }
                        // console.log("crumbs clicked:", item, idx);



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
        headerMenuActions(){


            let that = this;
            if(Session.stack.length > 1){
                if(that.root.getElementsByClassName("ui-page-header-icon-back").length ==0 ){
                    let element = that.createElAndAppend(this.headerMenuicon, "span","ui-page-header-icon-back");
                    element.classList.add("ui-page-header-icon-back");
                    element.addEventListener("click",  Session.CurrentPage.back)
                }
            }else if(that.root.getElementsByClassName("ui-page-header-icon-back").length > 0 ){
                let element = that.root.getElementsByClassName("ui-page-header-icon-back")[0];
                element.removeEventListener("click",  Session.CurrentPage.back)
                element.remove();
            }

            if(Session.CurrentPage.configuration.name== "IAC Home"){
                if(that.root.getElementsByClassName("ui-page-header-icon-home").length >0 ){
                    let element2 = that.root.getElementsByClassName("ui-page-header-icon-home")[0];
                    element2.removeEventListener("click", Session.CurrentPage.home)
                    element2.remove();
                }
            } else{
                if(that.root.getElementsByClassName("ui-page-header-icon-home").length ==0 ){
                    let element2 = that.createElAndAppend(this.headerMenuicon, "a","ui-page-header-icon-home");
                    
                    element2.addEventListener("click", function(){
                        // console.log("home clicked")
                        Session.CurrentPage.home();})
                }
            }
        };
    }
    UI.Pageheader = Pageheader;

})(UI || (UI = {}));

  
(function (UI) {
    function startpage(pagefile){
        // console.log(pagefile);
        let page = new UI.Page({file:pagefile});

       /*
        let ajax = new UI.Ajax("");
        ajax.get(pagefile,false).then((response) => {
            // console.log(response)
            let page = new UI.Page(JSON.parse(response));

            //page.create();            
        }).catch((error) => {
            // console.log(error);
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
// console.log("UI loaded");
// console.log(UI.Ajax);
*/
