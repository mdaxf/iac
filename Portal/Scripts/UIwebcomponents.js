// Line Chart Web Component
customElements.define('ui-tabulator', class extends HTMLElement {
    constructor() {
        super();
        this.shadow = this.attachShadow({ mode: 'open' });
        let link = document.createElement('link');
        link.setAttribute('rel', 'stylesheet');
        link.setAttribute('href', 'styles/tabulator/tabulator_iac.css');
        this.shadow.appendChild(link);
   /*    link = document.createElement('link');
        link.setAttribute('rel', 'stylesheet');
        link.setAttribute('href', 'styles/uitabulator.css'); */
        this.shadow.appendChild(link);
      //  console.log("ui-tabulator", this)
 
    }
    connectedCallback() {
        this.schema = this.getAttribute("schema");
        this.datakey_field = this.getAttribute("datakey_field");
        this.datakey_value = this.getAttribute("datakey_value");
        this.data_viewonly = this.getAttribute("data_viewonly");
        this.data_url = this.getAttribute("data_url");
        this.data_method = this.getAttribute("data_method");
        this.condition = this.getAttribute("condition");

        if(!this.data_method)
            this.data_method = "";
        this.uitabulator = document.createElement('div');
        this.shadow.appendChild(this.uitabulator);
      
        this.loadthetabulator();
        
        if(this.schema != null && this.schema != "")
            this.loaddatabyschema();
        else
            this.createemptytable();

        const UITabulatorLoadedEvent = new CustomEvent('uitabulator_loaded');

        this.dispatchEvent(UITabulatorLoadedEvent);
    }

    static get observedAttributes() {
      //  return ["schema", "datakey_field", "datakey_value", "data_viewonly", "data_url", "data_method", "condition"];
      return ["condition"];
    }
    loadthetabulator(){
        this.columns = [];
        this.data = [];
        this.lngcodes={};
        this.langs={
            "data":{
                "loading":"Loading", //data loader text
                "error":"Error", //data error text
            },
            "groups":{ //copy for the auto generated item count in group header
                "item":"item", //the singular  for item
                "items":"items", //the plural for items
            },
            "pagination":{
            	"page_size":"Page Size", //label for the page size select element
                "page_title":"Show Page",//tooltip text for the numeric page button, appears in front of the page number (eg. "Show Page" will result in a tool tip of "Show Page 1" on the page 1 button)
                "first":"First", //text for the first page button
                "first_title":"First Page", //tooltip text for the first page button
                "last":"Last",
                "last_title":"Last Page",
                "prev":"Prev",
                "prev_title":"Prev Page",
                "next":"Next",
                "next_title":"Next Page",
                "all":"All",
                "counter":{
                    "showing": "Showing",
                    "of": "of",
                    "rows": "rows",
                    "pages": "pages",
                }
            },
            "headerFilters":{
                "default":"filter column...", //default header filter placeholder text
            }
        }
        //this.locallangs ={}

        if(UI.userlogin.language != "en" || UI.userlogin.language != "en-US" || UI.userlogin.language != ""){

            let tablelangscodelist = [];
            let tablelangs = {};
            let that = this;
            let getLeafNodes = function(obj, path = []) {
                const leafNodes = [];
            
                for (const key in obj) {
                if (obj.hasOwnProperty(key)) {
                    const value = obj[key];
                    const currentPath = [...path, key];
            
                    if (typeof value === "object" && !Array.isArray(value)) {
                    // If the value is an object (but not an array), recursively traverse it
                    const childLeafNodes = getLeafNodes(value, currentPath);
                    leafNodes.push(...childLeafNodes);
                    } else {
                    // If the value is not an object, it's a leaf node
                    leafNodes.push({ path: currentPath.join("-"), value });
                    }
                }
                }
            
                return leafNodes;
            }
            let setObjectValue = function(obj, paths, value) {   
                for(var i=0;i<paths.length;i++){
                    let path = paths[i];
                    if(i == paths.length -1){
                        obj[path] = value;
                    }else{
                        if(!obj.hasOwnProperty(path)){
                            obj[path] = {};
                        }
                        obj = obj[path];
                    }
                }

            }

            const leafNodes = getLeafNodes(this.langs);

            leafNodes.forEach((node) => {
                tablelangscodelist.push(node.path);
                tablelangs[node.path] = node.value;
            });

            UI.translatebycodes(tablelangscodelist, function(data){
                for(var key in data){
                    tablelangs[key] = data[key];
                    let paths = key.split("-");
                    setObjectValue(that.langs, paths, data[key]);
                }
            }, function(){
                UI.Log("language code translation completed")
            });
        }
    }
    loaddatabyschema(){
        let that = this;
        let ajax = new UI.Ajax("");
        this.lngcodes = {}; 
        let languagecodelist = [];
        if(this.schema == null)
          return;
        
        console.log(this.schema, this.datakey_field, this.datakey_value, this.data_viewonly)
        ajax.get('/portal/datasets/schemas/'+ this.schema + '.json').then((response) => {
          let schema = JSON.parse(response);
          this.schemadata = schema;
  
          if(schema == null){
              UI.ShowError('Data Schema Not Found');
              return;
          }
          let type = ""; 
          let datasource = "";
          let listfields = [];
          let querystr ="";
          if(schema.hasOwnProperty('datasourcetype'))
              type = schema.datasourcetype;
          if(schema.hasOwnProperty('datasource'))
              datasource = schema.datasource;
          if(schema.hasOwnProperty('listfields'))
              listfields = schema.listfields;
          if(schema.hasOwnProperty('query'))
            querystr = schema.query;

          let data ={}
          for(var i=0;i<listfields.length;i++){
              data[listfields[i]] = 1;
          }
          let jdataschema =new UI.JSONSchema(schema)
          //    console.log(jdataschema)
  
              
          let fieldpropertiesobj = {};
          for(var i=0;i<listfields.length;i++){
              let fieldschema = jdataschema.getPropertiesFromSchema(listfields[i].replace('.','/'));
              if(fieldschema)
                  fieldpropertiesobj[listfields[i]] =  fieldschema.properties;
              else 
                  fieldpropertiesobj[listfields[i]] = null;
                  
          }
  
           //   console.log(fieldpropertiesobj)
          this.keyfield = "";
          if(schema.hasOwnProperty('keyfield'))
              this.keyfield = schema.keyfield;
          let url = "";
  
          let inputs={};
          if(type == "collection" && datasource !='' && data != '{}'){
                  
              inputs["collectionname"] = datasource;                 
              inputs["data"] = data;
              inputs["operation"] = "list";
              url = '/collection/list'
          }
          else if(type == "table" && datasource !=''){
            if(querystr ==""){
              inputs["tablename"] = datasource.toLowerCase();                 
              inputs["operation"] = "list";
              data ={}                    
              data[datasource] = {
                  "fields":listfields
              }
              inputs["data"] = data;
              inputs["where"] = {};
                  
              url = '/sqldata/get'
            }else{
                url = "/sqldata/query"                
                inputs["querystr"] = querystr;
                inputs["operation"] = "query";
            }

          }
          else {
              UI.ShowError('Data Source Not Found');
              return;
          }
          if(this.data_url != null)
              url = this.data_url;
  
          if(this.condition){
                let where={}
                where[this.condition] = "";
                inputs["where"] = where;
          }
          this.url = url;
          this.inputs = inputs;
          
          let Tabulator_Columns = [];
          
          Tabulator_Columns.push(
              {formatter:"rowSelection", titleFormatter:"rowSelection", hozAlign:"center", width:30, headerSort:false, cellClick:function(e, cell){
                  cell.getRow().toggleSelect();
                  console.log(cell, e)
                }}
          )
          let mappedfields = {};
          for(var i=0;i<listfields.length;i++){
              let fieldschema = fieldpropertiesobj[listfields[i]];
              let lng = null;
              let type = "string"
              if(fieldschema){
                  if(fieldschema.hasOwnProperty('lng'))
                      lng = fieldschema.lng;
              
                  if(fieldschema.hasOwnProperty('type'))
                      type = fieldschema.type;
              }
              
              let column = {title:listfields[i], field:listfields[i],headerSort: true};
              if(lng){
                  column.title = lng.default;
                  mappedfields[listfields[i]] = lng.code;
                  languagecodelist.push(lng.code);
                  this.lngcodes[lng.code] = lng.default;
              }else{
                mappedfields[listfields[i]] = listfields[i];
                this.lngcodes[listfields[i]] = listfields[i];
              }
  
              if(type == "boolean"){
                  column.formatter = "tickCross";
                  column.hozAlign = "center";
                  column.width = 50;
                  column.sorttype = "boolean";
                  column.headerFilter ="tickCross"
              }
              else if(type == "integer" || type == "number"){                
                  column.sorter = "number";
                  column.headerFilter ="input"
              }else {
                  column.sorter = "string";
                  column.headerFilter ="input"
              }
              Tabulator_Columns.push(column);
          }
          this.columns = Tabulator_Columns;
          
          this.buldtable();
            
          UI.translatebycodes(languagecodelist, function(data){
                that.lngcodes = Object.assign(that.lngcodes, data);
                UI.Log(that.lngcodes, data)
                UI.Log(that.columns)  
                that.columns.forEach(function(column){
                    let field = column.field;
                    let mappedfield = mappedfields[field];
                    if(that.lngcodes.hasOwnProperty(mappedfield) ){
                        column.title = that.lngcodes[mappedfield];
                    }
                })
                UI.Log("applied translation: ",that.columns)
                that.Table.setColumns(that.columns);
            }, function(){
                UI.Log("language code translation completed")
            });

      //    console.log(Tabulator_Columns, fieldpropertiesobj, listfields)
          let Tabulator_Data = [];
          let userlanguage = UI.userlogin.language || "en";
          UI.Log("tableBuilt,", userlanguage);
          if(this.data_method.toLowerCase() == "get"){
              ajax.get(url, inputs).then((response) => {
                  Tabulator_Data = JSON.parse(response)["data"];
              //    console.log(response,Tabulator_Data)
                 this.data = Tabulator_Data;
                 this.Table.setData(Tabulator_Data);                 
                 this.Table.setLocale(userlanguage);
              }).catch((error) => {
                  UI.ShowError(error);
              })
          }else{
              ajax.post(url, inputs).then((response) => {
                  Tabulator_Data = JSON.parse(response)["data"];
                  //    console.log(response,data)
                 this.data = Tabulator_Data;
                 this.Table.setData(Tabulator_Data);
                 this.Table.setLocale(userlanguage);
              }).catch((error) => {
                  UI.ShowError(error);
              })
          }
  
  
        }).catch((error) => {
          UI.ShowError(error);
        });
  
    }
    buldtable(){
        let that = this;
        let height = this.parentElement.clientHeight;
        if(height == 0)
            height = 400;
        else
            height = height - 50;
        this.uitabulator.style.height = height + "px";
        let userlanguage = UI.userlogin.language || "en";
        let langs = {}
        langs[userlanguage] = this.langs;
        this.Table = new Tabulator(this.uitabulator, {
            height: height + "px",
            layout:"fitColumns",
            pagination:"local",
            paginationSize:16,
            paginationSizeSelector:[16, 30, 50, 100],
            resizableColumnFit:true,
            responsiveLayout:"hide",
            movableColumns:true,
            paginationCounter:"rows",
            clipboard:true,
            clipboardPasteAction:"replace",
        //    locale:true,
        //    autoColumns:true,
        //    layout: "fitColumns",
            langs: langs,
            columns: this.columns,
            data: this.data,
            tableBuilt:function(){
                UI.Log("tableBuilt");
                this.setLocale(userlanguage);
            }
        });

      //  console.log(this.Table.getLocale());
      //  this.Table.setLocale(userlanguage);
    }
    createemptytable(){
        let Tabulator_Data = [];
        let height = this.parentElement.clientHeight;
        if(height == 0)
            height = 400;
        else
            height = height - 50;
        this.uitabulator.style.height = height + "px";

    //    this.Table = new Tabulator(this.uitabulator, {autoColumns:true, data:[]});
       /* this.Table = new Tabulator(this.uitabulator, {
              height: height + "px",
              layout:"fitColumns",
              resizableColumnFit:true,
              responsiveLayout:"hide",
          //    autoColumns:true,
          //    layout: "fitColumns",
              data: Tabulator_Data
          }); */
    }
    

    loaddatabyQuery(query){
        let ajax = new UI.Ajax("");
        let url = "/sqldata/query"
        let inputs = {}        
        inputs["querystr"] = query;
        inputs["operation"] = "query";
        console.log(url, inputs )
        ajax.post(url,inputs,false).then((response) => {
            let data = JSON.parse(response)["data"];
        //    console.log(data)
            this.data = data;
            let userlanguage = UI.userlogin.language || "en";
            let langs = {}
            langs[userlanguage] = this.langs;
            this.Table = new Tabulator(this.uitabulator, {
                layout:"fitColumns",
                paginationSize:16,
                paginationSizeSelector:[16, 30, 50, 100],
                resizableColumnFit:true,
                responsiveLayout:"hide",
                movableColumns:true,
                paginationCounter:"rows",
                clipboard:true,
                clipboardPasteAction:"replace",
                langs: langs,
                autoColumns:true, 
                data:data,
                tableBuilt:function(){
                    UI.Log("tableBuilt");
                    this.setLocale(userlanguage);
                }
            });

        //    this.Table.setLocale(userlanguage);
            this.translateheader()
            //this.Table.setData(data);
        }).catch((error) => {
            UI.ShowError(error);
        })


    }
    translateheader(){
        let that = this;
        if(this.columns ==[]){
            this.columns = this.Table.getColumns();
        }
        let languagecodelist = [];
        this.lngcodes={};
        this.columns.forEach(function(column){
            if(that.lngcodes.hasOwnProperty(column.title) ){
                languagecodelist.push(that.lngcodes[column.title]);
                that.lngcodes[column.title] = column.title;
            }
        })
        UI.translatebycodes(languagecodelist, function(data){
            that.lngcodes = Object.assign(that.lngcodes, data);
            UI.Log(that.lngcodes, data)
            UI.Log(that.columns)  
            that.columns.forEach(function(column){
                if(that.lngcodes.hasOwnProperty(column.title) ){
                    column.title = that.lngcodes[column.title];
                }
            })
            UI.Log("applied translation: ",that.columns)
            that.Table.setColumns(that.columns);
        }, function(){
            UI.Log("language code translation completed")
        });

    }
    getSelectedKeys(){
        let rows = this.Table.getSelectedRows();
        let data = [];        
        for(var i=0;i<rows.length;i++){
            let row = rows[i].getData();
        //    console.log(row)
            data.push(row[this.keyfield]);
        }
        JSON.stringify(data);
        return data
    }

    getTableSchema(){
        return this.schemadata;
    }

    refresh(){
        if(this.schemadata == null){
            UI.ShowError('Data Schema Not Found');
            return;
        }
        let ajax = new UI.Ajax("");
        if(this.data_method.toLocaleLowerCase() == "get"){
            ajax.get(this.url, this.inputs).then((response) => {
                let Tabulator_Data = JSON.parse(response)["data"];
                //    console.log(response,data)
               this.Table.setData(Tabulator_Data);
            }).catch((error) => {
                UI.ShowError(error);
            })
        }else{
            ajax.post(this.url, this.inputs).then((response) => {
                let Tabulator_Data = JSON.parse(response)["data"];
                //    console.log(response,data)
               this.Table.setData(Tabulator_Data);
            }).catch((error) => {
                UI.ShowError(error);
            })
        }
    }
});

/*
 a web component to display a block chart for the machine states
*/
customElements.define('ui-machine-state-chart', class extends HTMLElement {
    constructor() {
        super();

        // Create a shadow DOM
        this.shadow = this.attachShadow({ mode: 'open' });
        let link = document.createElement('link');
        link.setAttribute('rel', 'stylesheet');
        link.setAttribute('href', 'https://visjs.github.io/vis-timeline/styles/vis-timeline-graph2d.min.css');
        this.shadowRoot.appendChild(link);
        this.container =null;
    }

    refresh() {
        this.timeline.redraw();
    }

    render() {

        if(this.timeline != null)
            this.timeline.destroy();

        if(this.container != null)
            this.shadowRoot.removeChild(this.container);

        let Machines = this.getAttribute("machinelist");
        this.StartTime = this.getAttribute("starttime");
        this.EndTime = this.getAttribute("endtime");

        if(this.EndTime == null || this.EndTime == "")
            this.EndTime = new Date();

        this.MachineList = Machines.split(",");

        if(this.StartTime == null || this.StartTime == "")
            this.StartTime = new Date(this.EndTime - 3600000);

        if(this.MachineList.length == 0)
            return; 

        let list = [];
        for(var i=0;i<this.MachineList.length;i++){
            list.push({
                id: i+1,
                content: '<div class="machine_state_group_label">'+this.MachineList[i] + "</div>",
            })
        }
        this.groups = new vis.DataSet(list); 
          
    //    console.log(this.MachineList, this.StartTime, this.EndTime  )
    //    console.log(this.groups)
        // create items
        this.items = new vis.DataSet();

        let url = "/sqldata/query"
        let query = "SELECT MS.Machine As Machine, MS.State As State, MS.StartTime As StartTime, IFNULL(MS.EndTime, Now()) As EndTime, IFNuLL(RC.BGColor,'light') As BGColor, IFNULL(RC.Color, 'black') As Color "
            query += " FROM machine_states As MS "
            query += " LEFT JOIN reason_codes As RC ON MS.State = RC.Name "
            query += " WHERE MS.Machine IN ('" + this.MachineList.join("','") + "') "
       //     query += " AND MS.StartTime >= '" + this.StartTime + "' "
       //     query += " AND ((MS.EndTime <= '" + this.EndTime + "') OR (MS.EndTime IS NULL)) "
        
        let inputs = {}        
        inputs["querystr"] = query;
        inputs["operation"] = "query";
        console.log(url, inputs )
        let ajax = new UI.Ajax("");
        ajax.post(url,inputs,false).then((response) => {
            let data = JSON.parse(response)["data"];
            console.log(data)
            let start = new Date(this.StartTime);
            var date = new Date();
            for(var i=0;i<data.length;i++){
                let item = data[i];
                date = item.StartTime;
                let start = new Date(date); //this.parseDate(item.StartTime);
                date=item.EndTime;
                let end = new Date(date);
                let group = 1
                for(var j=0;j<this.MachineList.length;j++){
                    if(this.MachineList[j] == item.Machine){
                        group = j+1;
                        break;
                    }
                }
                let content = item.State;
                let d = {
                    id: i+1,
                    group: group,
                    start: start,
                    end: end,
                    content: content,
                    style:"height:50px;color: " + item.Color + "; background-color: " + item.BGColor + ";",
                    type:'range',
                    title: '<div><b>' +item.State + '</b></div><div>' + item.StartTime + ' - ' + item.EndTime + '</div>'
                }
                
            //    console.log(d)
                this.items.add(d);
            }

            var options = {
                stack: false,
                start: this.StartTime,
                end: this.EndTime,
                editable: false,
                showMajorLabels: true, 
                showMinorLabels: true,
                showCurrentTime:true,
                margin: {
                item: 10, // minimal margin between items
                axis: 5   // minimal margin between items and the axis
                },
                orientation: 'top'
            };    
           
            this.container = document.createElement('div');
            //   container.style.width = '100%';
            //   container.style.height = '400px';
            //console.log(this.groups, this.items, options)
            this.shadowRoot.appendChild(this.container);
            
            this.timeline = new vis.Timeline(this.container, null, options);
            this.timeline.setGroups(this.groups);
            this.timeline.setItems(this.items);
        }).catch((error) => {
            UI.ShowError(error);
        })
    }
    connectedCallback() {
        this.render();
    }

    static get observedAttributes() {
        return ["machinelist", "starttime", "endtime"];
    }

    attributeChangedCallback(name, oldValue, newValue) {
   //    this.render();
    }

    addData(item){
        let start = new Date(item.StartTime);
        let end = new Date(item.EndTime);
        let group = item.Machine;
        let content = item.State;
        this.items.add({
            id: i+1,
            group: group,
            start: start,
            end: end,
            content: content,
            style:"color: " + item.Color + "; background-color: " + item.BGColor + ";"
        }); 
        this.timeline.setItems(this.items);
        
    }


});