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

var UI = UI || {};
(function () {
    class FormControl {
        constructor(parent, tag, attrs=[], events=[], data={}){
            this.parent = parent;
            this.events = events;
            this.attrs = attrs;
            this.data = data;
            this.tag = tag;
            
        //    console.log(this.parent,this.attrs, this.tag, this.events, this.data)
            this.build();
            this.setevents();
        }
        build(){
            this.control = document.createElement(this.tag);
            for(const key in this.attrs){
                if(key.toLowerCase() == 'innerhtml')
                    this.control.innerHTML = this.attrs[key];
                else if(key.toLowerCase() == 'innertext')
                    this.control.innerText = this.attrs[key];
                else if(this.tag =='select' && (key.toLowerCase() == 'selected' || key.toLowerCase() == 'options' || key.toLowerCase() == 'value')){
                }
                else if(this.attrs[key] != null && this.attrs[key] != undefined && this.attrs[key] != '' && typeof this.attrs[key] != 'object') 
                    this.control.setAttribute(key, this.attrs[key]);
            }
        //    this.control.classList.add("ui");
            this.parent.appendChild(this.control);
        }
        setevents(){
            for(const key in this.events){
                this.control.addEventListener(key, this.events[key]);
            }        
        }
        destroy(){
            for(const key in this.events){
                this.control.removeEventListener(key, this.events[key]);
            }
            this.parent.removeChild(this.control);
        }
    }
    UI.FormControl = FormControl;

    class Option extends FormControl{
        build(){
        //    console.log('build option:',this.parent,this.attrs, this.tag, this.events, this.data)
            this.control = document.createElement("option");            
            this.control.classList.add("ui");
            for(const key in this.attrs){
                var value;
                if(key.toLowerCase() == 'innerhtml')
                    this.control.innerHTML = this.attrs[key];
                else if(key.toLowerCase() == 'innertext')
                    this.control.innerText = this.attrs[key];
                else if(key.toLowerCase() != 'value' && key.toLowerCase()!= 'selected')
                    this.control.setAttribute(key, this.attrs[key]);

                if(key.toLowerCase() == "value"){
                    value = this.attrs[key];
                }
                
            }
            
            //this.control.innerHTML = this.data.text;
        //    console.log(value, this.data.selected)
            if(value == this.data.selected){
                this.control.setAttribute("selected", "selected");
            }
            
            this.parent.appendChild(this.control);
        }
    }

    UI.Option = Option;

    class CheckBox extends FormControl{
        build(){
         //   console.log('build checkbox:',this.parent,this.attrs, this.tag, this.events, this.data)
            this.control = document.createElement("input");
            this.control.setAttribute("type", "checkbox");
            this.control.classList.add("ui");
            for(const key in this.attrs){
                if(key.toLowerCase() == 'innerhtml')
                    this.control.innerHTML = this.attrs[key];
                else if(key.toLowerCase() == 'innertext')
                    this.control.innerText = this.attrs[key];
                else if(key != 'value')
                    this.control.setAttribute(key, this.attrs[key]);
            }

            if(this.attrs['value']){
               this.control.setAttribute("checked", "checked");
            }

            this.parent.appendChild(this.control);
        }
    }
    UI.CheckBox = CheckBox;

    class HtmlTable{
       /* {
            attrs:{},
            headers: [{
                innerHTML: "",
                style,
            },{}],
            columns: [{
                control: "",
                attrs: [],
            },{}],
            rows:[{},{}]
        } */
        constructor(parent, data){
            this.parent = parent;
            this.data = data;
            this.build();
        }
        build(){
            let that = this;
            let attrs = {};// {"class":"table table-striped table-bordered table-hover table-sm"};
            attrs = Object.assign(attrs, this.data.attrs);
            this.table = new FormControl(this.parent, "table", attrs, {});
            this.thead = new FormControl(this.table.control, "thead");
            let htr = new FormControl(this.thead.control, "tr");
           // console.log('table data:',this.data.headers)
            this.data.headers.forEach((header)=>{
            //    console.log('header:',header);
                new FormControl(htr.control, "th", header, {});
            });
            this.tbody = new FormControl(this.table.control, "tbody");
            this.AddRows(this.data.rows);         
        }
        AddFilters(filters){
            this.AddRow(filters, this.thead.control);
        }
        AddRows(rows){
            let that = this;
            rows.forEach((row)=>{
                that.AddRow(row,that.tbody.control);       
            });
        }
        AddRow(row, control=null){
            let that = this;
            if(!control)
                control = that.tbody.control;

            let rowattrs = {};
            if(that.data.hasOwnProperty('tr'))
                if(that.data.tr.hasOwnProperty('attrs'))
                    rowattrs = Object.assign(rowattrs, that.data.tr.attrs);
            let rowevents = {};
            if(that.data.hasOwnProperty('tr'))
                if(that.data.tr.hasOwnProperty('events'))
                    rowevents = Object.assign(rowevents, that.data.tr.events);
            let tr = new FormControl(control, "tr",rowattrs, rowevents);
         //   console.log(row, control)
            that.data.columns.forEach((column, index)=>{
             //   console.log(column,index,column.control)
                let attrs = Object.assign({},column.attrs);
                attrs = Object.assign(attrs, row[index].data);
                let events = Object.assign({}, row[index].events);
                events = Object.assign(events, column.events);
                if(row[index].attrs)
                    attrs = Object.assign(attrs, row[index].attrs);

            //    console.log('control type:',column.control, index)
                let control = !column.control? '':  column.control;
                
                if(control == "")
                    new FormControl(tr.control, "td", attrs, events);
                else{
                    let td = new FormControl(tr.control, "td",[] ); 
                 //   console.log(td, column,index,control)
                    if(control == 'select'){
                        let select = new FormControl(td.control, 'select',attrs, events);
                        column.options.forEach((option,i)=>{
                            let optionel = document.createElement("option");
                            optionel.classList.add("ui");
                            if(typeof option == 'string'){
                                
                                optionel.setAttribute("value", i);
                                optionel.innerHTML = option;
                                if(row[index].data.selected == i)
                                    optionel.setAttribute("selected", "selected");
                                
                                
                            }else if(typeof option == 'object'){
                                optionel.setAttribute("value", option.value);
                                optionel.innerHTML = option.innerHTML;
                                if(row[index].data.selected == option.value)
                                    optionel.setAttribute("selected", "selected");
                            }
                            select.control.appendChild(optionel);
                            
                        });                 
                    }
                    else if(control== 'checkbox'){
                        let checkbox = (new CheckBox(td.control,'input', attrs,events)).control;
                        /*checkbox.addEventListener('change', function(){
                            row[index].attrs = row[index].attrs || {};
                            row[index].attrs["value"] = checkbox.checked;
                        }); */
                    }
                    else
                        new FormControl(td.control, control, attrs, events);
                }
            });
        
        }
    } 
    
    UI.HtmlTable = HtmlTable;

    class Selection{
        /*
        {
            attrs:{},
            options:[{},{}],
            selected: "",
        }
        */
        constructor(parent, data){
         //   console.log('create select:',parent,data)
            this.parent = parent;
            this.data = data;
            this.build();
        }
        build(){
            let that = this;
            let attrs = {};
            attrs = Object.assign(attrs, this.data.attrs);
            this.control = (new FormControl(this.parent, "select", attrs, this.data.events? this.data.events:{})).control;  
        //    console.log('data:',this.data.options)
            this.data.options.forEach((option, index)=>{
                if(typeof option == "string"){
                    let attrs = {value: index, innerHTML: option};                    
                    let optioncell =(new FormControl(that.control, 'option',attrs)).control;
                    if(that.data.selected == index)
                        optioncell.setAttribute("selected", "selected");
                }
                else if(typeof option == "object"){
                    let attrs = Object.assign(option,option.attrs);
                    let events = Object.assign({},option.events);
                    let optioncell = null
                    if(attrs)
                        optioncell =(new FormControl(that.control, "option", attrs,events)).control;
                    else
                        optioncell =(new FormControl(that.control, "option", option,events)).control;
                    if(that.data.selected == option.value)
                        optioncell.setAttribute("selected", "selected");
                }
            });

        }
    }

    UI.Selection = Selection;

    class Builder{
        constructor(parent, data){
            this.parent = parent || document.body;
            this.data = data || [];
            this.build();
        }
        build(){
            this.data.forEach((control)=>{
                
                this.AddControl(this.parent,control);
            });
        }
        AddControl(parent, control){
            let that = this;
            let attrs = {};
            if(control.hasOwnProperty('attrs'))
                attrs = Object.assign(attrs, control.attrs);
            else 
                attrs = Object.assign(attrs, control);
            let events = {};
            if(control.hasOwnProperty('events'))
                events = Object.assign(events, control.events);
            let controlcell = null;
            if(control.hasOwnProperty('tag')){
                if(control.tag == 'selection')
                    controlcell = new Selection(parent, control);
                else if(control.tag == 'htmltable')
                    controlcell = new HtmlTable(parent, control);
                else if(control.tag == 'checkbox')
                    controlcell = new CheckBox(parent, control);
                else
                    controlcell = new FormControl(parent, control.tag, attrs, events);
            }
            else
                controlcell = new FormControl(parent, "div", attrs, events);

            if(control.hasOwnProperty('children'))
                control.children.forEach((child)=>{
                    that.AddControl(controlcell.control, child);
                });
        }
    }
    UI.Builder = Builder
})(UI || (UI = {}));