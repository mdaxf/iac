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
        console.log("ui-tabulator", this)

    }
    
    static get observedAttributes() {
        return ["schema", "datakey_field", "datakey_value", "data_viewonly"];
    }

    loaddatabyschema(){
        let ajax = new UI.Ajax("");
      
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
          if(schema.hasOwnProperty('datasourcetype'))
              type = schema.datasourcetype;
          if(schema.hasOwnProperty('datasource'))
              datasource = schema.datasource;
          if(schema.hasOwnProperty('listfields'))
              listfields = schema.listfields;
             
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
                
              inputs["tablename"] = datasource;                 
              inputs["operation"] = "list";
              data ={}                    
              data[datasource] = {
                  "fields":listfields
              }
              inputs["data"] = data;
              inputs["where"] = {};
                  
              url = '/sqldata/get'
          }
          else {
              UI.ShowError('Data Source Not Found');
              return;
          }
          if(this.data_url != null)
              url = this.data_url;
  
          this.url = url;
          this.inputs = inputs;
  
          let Tabulator_Columns = [];
          
          Tabulator_Columns.push(
              {formatter:"rowSelection", titleFormatter:"rowSelection", hozAlign:"center", width:30, headerSort:false, cellClick:function(e, cell){
                  cell.getRow().toggleSelect();
                  console.log(cell, e)
                }}
          )
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
              //    column.lngcode = lng.code;
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
      //    console.log(Tabulator_Columns, fieldpropertiesobj, listfields)
          let height = this.parentElement.clientHeight;
          if(height == 0)
              height = 400;
          else
              height = height - 50;
          this.uitabulator.style.height = height + "px";
  
          let Tabulator_Data = [];
          this.Table = new Tabulator(this.uitabulator, {
              height: height + "px",
              layout:"fitColumns",
              resizableColumnFit:true,
              responsiveLayout:"hide",
          //    autoColumns:true,
          //    layout: "fitColumns",
              columns: Tabulator_Columns,
              data: Tabulator_Data
          });
  
          if(this.data_method.toLowerCase() == "get"){
              ajax.get(url, inputs).then((response) => {
                  Tabulator_Data = JSON.parse(response)["data"];
              //    console.log(response,Tabulator_Data)
                 this.Table.setData(Tabulator_Data);
              }).catch((error) => {
                  UI.ShowError(error);
              })
          }else{
              ajax.post(url, inputs).then((response) => {
                  Tabulator_Data = JSON.parse(response)["data"];
                  //    console.log(response,data)
                 this.Table.setData(Tabulator_Data);
              }).catch((error) => {
                  UI.ShowError(error);
              })
          }
  
  
        }).catch((error) => {
          UI.ShowError(error);
        });
  
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
    connectedCallback() {
        this.schema = this.getAttribute("schema");
        this.datakey_field = this.getAttribute("datakey_field");
        this.datakey_value = this.getAttribute("datakey_value");
        this.data_viewonly = this.getAttribute("data_viewonly");
        this.data_url = this.getAttribute("data_url");
        this.data_method = this.getAttribute("data_method");

        if(!this.data_method)
            this.data_method = "";
        this.uitabulator = document.createElement('div');
        this.shadow.appendChild(this.uitabulator);
      
        if(this.schema != null && this.schema != "")
            this.loaddatabyschema();
        else
            this.createemptytable();
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
            this.Table = new Tabulator(this.uitabulator, {autoColumns:true, data:data});
            //this.Table.setData(data);
        }).catch((error) => {
            UI.ShowError(error);
        })


    }

    getSelectedKeys(){
        let rows = this.Table.getSelectedRows();
        let data = [];        
        for(var i=0;i<rows.length;i++){
            let row = rows[i].getData();
        //    console.log(row)
            data.push(row[this.keyfield]);
        }

        return data
    }

    getTableSchema(){
        return this.schemadata;
    }

    refresh(){
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
            query += " FROM Machine_States As MS "
            query += " LEFT JOIN Reason_Codes As RC ON MS.State = RC.Name "
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