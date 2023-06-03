class CustomGridStack extends GridStack {
    constructor(...args) {
        super(...args);
    }

    // Custom method to add new properties to the panel
    addCustomProperty(panel, customProperty) {
        panel.customProperty = customProperty;
    }
}

var LayoutEditor = {
        grid: null,
        Options: null,
        subOptions: null,
        cellHeight: 1,
        cellWidth: 1,
        JsonObj:null,
        schema:null,
        initialize: function (data=null){
             console.log(CustomGridStack)
            let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            width = width - 50;
            height = height - 10;
            LayoutEditor.subOptions = {
              cellHeight: LayoutEditor.cellHeight, // should be 50 - top/bottom
              cellWidth: LayoutEditor.cellWidth, // should be 50 - left/right
            //  column: 'auto', // size to match container. make sure to include gridstack-extra.min.css
              acceptWidgets: true, // will accept .grid-stack-item by default
              margin: 1,
              column:100,
              subGridDynamic: true, 
            };
            var layoutdata={};
            var children=[];
            if(data){
              LayoutEditor.JsonObj = new UI.JSONManager(data, {allowChanges:true, schema:LayoutEditor.schema})
              layoutdata = LayoutEditor.getpaneldatafrompaneldata(data, 0)
            /*  panels = data.panels;
              for(var i=0;i<panels.length;i++){
                let panel = LayoutEditor.getpaneldatafrompaneldata(panels[i]);
                children.push(panel);
              } */
            }
            else{
                let sampledata = {
                  "name": "PageDefinition",
                  "version": "1.0.0",
                  "isdefault": true,
                  "orientation": 0,
                  "initcode":"",
                  "onloadcode": "",
                  "attrs": {
                    "style": "width: 100%; height: 100%;"
                  },
                  "panels": []
                }
                LayoutEditor.JsonObj = new UI.JSONManager(sampledata, {allowChanges:true})
            }


            LayoutEditor.Options = { // main grid options
                cellHeight: LayoutEditor.cellHeight, // should be 50 - top/bottom
                cellWidth: LayoutEditor.cellWidth, // should be 50 - left/right
                verticalMargin: 1,
                horizontalMargin: 1,
                margin: 1,
                minRow: 2, // don't collapse when empty
                column:100,
                disableOneColumnMode: true,
                acceptWidgets: true,
                subGridOpts: LayoutEditor.subOptions,
                subGridDynamic: true,
                id: 'main_layout_panel',
                children: children
            };
            
            LayoutEditor.Options = Object.assign(LayoutEditor.Options,layoutdata);

            // create and load it all from JSON above
            LayoutEditor.grid = CustomGridStack.addGrid(document.querySelector('.container-fluid'), LayoutEditor.Options);

            let gridEls = CustomGridStack.getElements('.grid-stack-item');
            gridEls.forEach(gridEl => {
                LayoutEditor.addSelectEvent(gridEl);
            })
            $($('.container-fluid').find('.grid-stack')[0] ).css('background-color', 'lightgrey');
            $($('.container-fluid').find('.grid-stack')[0] ).css('width', width+'px');
            $($('.container-fluid').find('.grid-stack')[0] ).css('height', height+'px');
            $('.selected').length > 0 ? $('.btn_removepanel').show() : $('.btn_removepanel').hide();

            window.addEventListener('resize', LayoutEditor.window_resize);
            LayoutEditor.attachContextEvents();
        },
        getpaneldatafrompaneldata:function(paneldata, level=0){
          let children = [];
          console.log('json data:',paneldata)
          if(paneldata.hasOwnProperty('panels')){
            let panels = paneldata.panels;
            for(var i=0;i<panels.length;i++){
              let panel = LayoutEditor.getpaneldatafrompaneldata(panels[i], level+1);
              children.push(panel);
            }
            //children = LayoutEditor.getpaneldatafrompaneldata(paneldata.panels);
          }
          if(level == 0){
            let rootpanel = JSON.parse(JSON.stringify(paneldata));
            if(rootpanel.hasOwnProperty('panels'))
                delete  rootpanel.panels;
            panel = {
              layoutdata: rootpanel,
              children:children
            }
          }
          else{
            panel = {
              x: paneldata.hasOwnProperty('x')? paneldata.x:0,
              y:  paneldata.hasOwnProperty('y')? paneldata.y:0,
              w:  paneldata.hasOwnProperty('w')? paneldata.w:1,
              h:  paneldata.hasOwnProperty('h')? paneldata.h:100,
              width:  paneldata.hasOwnProperty('width')? paneldata.width:100,
              height:  paneldata.hasOwnProperty('height')? paneldata.height:100,
              widthunit:  paneldata.hasOwnProperty('widthunit')? paneldata.widthunit:'%',
              heightunit:  paneldata.hasOwnProperty('heightunit')? paneldata.heightunit:'%',
              id:  paneldata.hasOwnProperty('id')? paneldata.id: UI.generateUUID(),
              name: paneldata.hasOwnProperty('name')? paneldata.name:'panel',
              content: paneldata.hasOwnProperty('name')?paneldata.name:'panel',
              view: paneldata.hasOwnProperty('view')?paneldata.view:{},
              class: paneldata.hasOwnProperty('class')?paneldata.class:'',
              orientation: paneldata.hasOwnProperty('orientation')?paneldata.orientation: 1,
              inlinestyle: paneldata.hasOwnProperty('inlinestyle')?paneldata.inlinestyle:'',
              widthmethod: paneldata.hasOwnProperty('widthmethod')?paneldata.widthmethod:false,
              heightmethod: paneldata.hasOwnProperty('heightmethod')?paneldata.heightmethod:false,
              subGridOpts: {children:children}
            }
          }
          console.log('panel data:',panel)
          return panel;
        },
        addPanelContainer: function (){
          LayoutEditor.Options = LayoutEditor.grid.save(true, false);
          let node={
            x:0,
            y:0,
            w:10,
            h:100,
            content: 'The panel can include other panels',
            id: 'sub_grid'+ (LayoutEditor.Options.length+1),
            name: 'sub_grid'+ (LayoutEditor.Options.length+1),
            width: 10,
            height: 100,
            subGridOpts: {children: [], id:'sub_grid'+ (LayoutEditor.Options.length+1), class: 'sub_grid', ...LayoutEditor.subOptions}
          }
          
          LayoutEditor.Options.push(node); 
          LayoutEditor.grid.removeAll();
          LayoutEditor.load(LayoutEditor.Options,false);    
          LayoutEditor.capatureevents();
        /*  let subgrid = LayoutEditor.grid.makeSubGrid(node);
          LayoutEditor.addSelectEvent(subgrid)      */  
        },

        addPanel: function (subgrid = null){
            console.log(subgrid)
            let count = $('.grid-stack-item').length + 1;
            let content = '<div class="layout_panel_operations" style="display:inline-block"><div>panel'+count+'" ></div></div>'            
            if(subgrid == null)
                LayoutEditor.grid.addWidget({x:0, y:100, content:content, w:10, h:100, id:'panel'+count, name:'panel'+count,width:100,height:100,view:'',class:'layout_panel'});
            else{     
              console.log(subgrid)         
              subGrid.addWidget({x:0, y:100, content:content, w:10, h:100, id:'panel'+count, name:'panel'+count,width:100,height:100,view:'',class:'layout_panel'});
            }

        //    LayoutEditor.addSelectEvent(cell)
        },
        render:function() {
          LayoutEditor.Options = LayoutEditor.grid.save(true, false);
          LayoutEditor.grid.removeAll();
          LayoutEditor.load(LayoutEditor.Options,false); 
          LayoutEditor.attachContextEvents();
        },
        save:function(content = true, full = true) {},
        generateJson:function(content = true, full = true) {
            let options = LayoutEditor.grid.save(true, false)
            let panels = [];
            options.forEach(option => {
              panels.push(LayoutEditor.getsubpaneldata(option))
            })
            let panelsnode ={
              panels: panels
            } 
            console.log(panelsnode)
            console.log(LayoutEditor.JsonObj.getdata(""))
            LayoutEditor.JsonObj.updateNode("", {panels: panels} )     
        },
        getsubpaneldata:function(paneldata){
          let data = JSON.parse(JSON.stringify(paneldata));
          if(data.hasOwnProperty('subGridOpts'))
            delete data.subGridOpts;
          
          if(data.hasOwnProperty('content'))
            delete data.content;

          let children = [];
          if(paneldata.hasOwnProperty('subGridOpts')){
            if(paneldata.subGridOpts.hasOwnProperty('children')){
              for(var i=0;i<paneldata.subGridOpts.children.length;i++){
                let subpanel = LayoutEditor.getsubpaneldata(paneldata.subGridOpts.children[i]);
                children.push(subpanel);
              }
            }
          }
          if(children.length > 0)
            data.panels = children;

          return data;
          
        },
        addSelectEvent: function(gridEl){
            $(gridEl).on('click', function(event, items) {
                if($(this).hasClass('selected'))
                {
                  $(this).removeClass('selected');
                  $('.selected').length > 0 ? $('.btn_removepanel').show() : $('.btn_removepanel').hide();
                }
                else
                {
                  $(this).addClass('selected');
                  $('.btn_removepanel').show();
                }
            });
            
        },
        capatureevents(){
          LayoutEditor.grid.on('change', function(event, items) {
            items.forEach(function(item) {
              console.log('Item moved:', item.el);
              console.log('New position:', item.x, item.y);
            });
          });
          
          // Event listener for item resize
          LayoutEditor.grid.on('resizestop', function(event, item) {
            console.log('Item resized:', item.el);
            console.log('New size:', item.width, item.height);
          });

        },
        removeSelected: function(){
            let gridEls = CustomGridStack.getElements('.grid-stack-item.selected');
            gridEls.forEach(gridEl => {
                LayoutEditor.grid.removeWidget(gridEl);
            })
        },
        destroy:function (full = true) {
          if (full) {
            LayoutEditor.grid.destroy();
            LayoutEditor.grid = undefined;
          } else {
            LayoutEditor.grid.removeAll();
          }
        },
        load:function(options, full = true) {
          if (full) {
            LayoutEditor.grid = CustomGridStack.addGrid(document.querySelector('.container-fluid'), options);
          } else {
            LayoutEditor.grid.load(options);
          }
          let gridEls = CustomGridStack.getElements('.grid-stack-item');
            gridEls.forEach(gridEl => {
                LayoutEditor.addSelectEvent(gridEl);
            })  
          $('.sub-grid').each(function(){
            console.log(this)
            LayoutEditor.addSelectEvent(this);
            
          })
        },
        window_resize:function(){
          let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            width = width - 50;
            height = height - 100;
            $($('.container-fluid').find('.grid-stack')[0] ).css('width', width+'px');
            $($('.container-fluid').find('.grid-stack')[0] ).css('height', height+'px');
        },
        attachContextEvents: function(){
            $.contextMenu({
              selector: '.grid-stack-item', 
              build:function($triggerElement,e){
                console.log($triggerElement,e)
                return{
                  callback: function(key, options,e){
                    console.log(key, options,e)
                    switch(key){
                      case 'Add Subpanel':
                        let element = $triggerElement.find('.grid-stack')[0];
                        //let subgrid = LayoutEditor.findGridNodeByelement(LayoutEditor.grid.engine.nodes, $triggerElement[0]);
                        console.log(element.gridstack)
                        LayoutEditor.addPanel(element.gridstack);
                        break;
                      case 'Properties':
                        LayoutEditor.ShowProperties($triggerElement);
                        break;
                      case 'Remove':
                        let gridEl = $triggerElement[0];
                        LayoutEditor.grid.removeWidget(gridEl);
                        break;
                    }

                  }, 
                  items:{
                    'Properties':{
                      name: 'Properties',
                      icon: 'fa-cog',
                      disabled: false
                    },
                  /*  'Add Subpanel':{
                      name: 'Add Subpanel',
                      icon: 'fa-plus',
                      disabled: !$triggerElement.hasClass('grid-stack-sub-grid')
                    }, */
                    'Remove':{
                      name: 'Remove',
                      icon: 'fa-minus',
                      disabled: $triggerElement.hasClass('grid-stack-sub-grid') && $triggerElement.find('.grid-stack-item').length > 0
                    },
                    "sep1":'------------',
                    'Quit':{
                      name: 'Quit',
                      icon: function($element, key, item){ return 'context-menu-icon context-menu-icon-quit'; },
                    }
                  }

                }
              },
                      
            })
        },
        savelayout:function(){
          LayoutEditor.generateJson();
          if(!LayoutEditor.JsonObj.data.hasOwnProperty('uuid'))
              LayoutEditor.JsonObj.data["uuid"] = UI.generateUUID();

          if(LayoutEditor.JsonObj.schema != null && LayoutEditor.JsonObj.schema !={}){
            let type = ''
            if(LayoutEditor.JsonObj.schema.hasOwnProperty('datasourcetype'))
              type = LayoutEditor.JsonObj.schema.datasourcetype;
            
            let datasource ="";
            if(LayoutEditor.JsonObj.schema.hasOwnProperty('datasource'))
              datasource= LayoutEditor.JsonObj.schema.datasource;
            console.log(datasource, type)
            if(type == 'collection' && datasource !=''){
              let inputs = {
                collectionname:  LayoutEditor.JsonObj.schema.datasource,
                data: LayoutEditor.JsonObj.data,
                keys: ["name"],
                operation: "update"
              }
              console.log(inputs)

              let ajax = new UI.Ajax("");
              ajax.post('/collection/update',inputs).then((response) => {
                let result = JSON.parse(response);
                console.log(result);
                if(result.Outputs.status == 'Success'){
                  LayoutEditor.JsonObj.data= result.Outputs.data;
                  LayoutEditor.JsonObj.changed = false;
                  alert('Layout saved successfully');
                }

              }).catch((error) => {
                  console.log(error);
              })

            }           
          }          
        },
        loadLayout:function(){
          $('#popupContainer').remove();
          if(LayoutEditor.JsonObj.changed){
            let result = confirm("Do you want to save the layout?");
            if(result){
              LayoutEditor.savelayout();
            }
          }
          let popup = document.createElement('div')
          popup.setAttribute('class','ui-popup-panel-container')
          popup.setAttribute('id','popupContainer')

          let popupContent = document.createElement('div')
          popupContent.setAttribute('class','ui-popup-panel-content')
          popupContent.setAttribute('id','popupContent')
          popup.appendChild(popupContent)
          let title = document.createElement('h2')
          title.innerHTML = 'Please select file to import'
          popupContent.appendChild(title)

          let fileInput = document.createElement('input');
          fileInput.type = 'file';
          fileInput.accept = '.json';
          fileInput.addEventListener('change', (event) => {
            const file = event.target.files[0];
            LayoutEditor.read_to_import_File(file);
            popup.style.display = 'none';
            $('.ui-popup-panel-container').remove();

          });	
          popupContent.appendChild(fileInput)

          let closePopupButton = document.createElement('button');
          closePopupButton.setAttribute('class','ui-popup-panel-closebtn')
          closePopupButton.innerHTML = 'Close'
          closePopupButton.addEventListener('click', () => {
            popup.style.display = 'none';
            $('.ui-popup-panel-container').remove();
          });
          popupContent.appendChild(closePopupButton)
          document.body.appendChild(popup)
          popup.style.display = 'block';
        
        },
        read_to_import_File:function(file){
			
          const reader = new FileReader();
          let that = this;
            reader.onload = (event) => {
            const fileContents = event.target.result;
            try {
              const jsonData = JSON.parse(fileContents);
            // Handle the JSON data
              console.log(jsonData);
              LayoutEditor.destroy(true)
              LayoutEditor.initialize(jsonData,LayoutEditor.JsonObj.options);
              
            } catch (error) {
            console.error('Error parsing JSON file:', error);
            }
          };
    
          reader.readAsText(file);
        },

        findGridNodeByelement: function(nodes, el){
          //LayoutEditor.grid.engine.
          for(var i=0;i<nodes.length;i++){
            let node = nodes[i];
          
            if(node.el == el)
              return node;

            if(node.subGrid){
              let subnode = LayoutEditor.findGridNodeByelement(node.subGrid.engine.nodes, el);
              if(subnode != null)
                return subnode;
            }          
          }
          return null
        },
        link_view: function(btn){
          let el = $(btn).closest('.grid-stack-item')[0];
          console.log(el)
          LayoutEditor.ShowProperties(el);
        },
        unline_view: function(btn){
          let el = $(btn).closest('.grid-stack-item')[0];
          console.log(el)
          LayoutEditor.ShowProperties(el);
        },
        showpropertiesbybtn: function(btn){
          console.log(btn)
          let el = $(btn).closest('.grid-stack-item')[0];
          console.log(el)
          LayoutEditor.ShowProperties(el);
        },
        ShowRootProperties: function(){
          let uiview = document.getElementsByClassName('container-fluid')[0].parentElement;
          console.log(uiview)
          let attrs={
              id: 'properties',
              class: 'properties',
              style: 'position: absolute; display:float; right: 0px; top:0px; width: 300px; height: 100%; background-color: white;',
          }
          let container = (new UI.FormControl(uiview, 'div',attrs)).control;
          console.log(container)
          new UI.FormControl(container, 'h1',{innerHTML: 'Page Properties'});
          new UI.FormControl(container, 'hr');
          attrs={for: 'name',innerHTML: 'Name'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'name',type: 'text',value: LayoutEditor.JsonObj.data.name || '',placeholder: 'Name',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);
          attrs={for: 'version',innerHTML: 'Version'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'version',type: 'text',value: LayoutEditor.JsonObj.data.version || '',placeholder: 'Version',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);
          attrs={for: 'isdefault',innerHTML: 'Is Default'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'isdefault',type: 'checkbox',checked: LayoutEditor.JsonObj.data.isdefault || '',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);
          attrs={for: 'orientation',innerHTML: 'Orientation'}
          new UI.FormControl(container, 'label',attrs);
          attrs={
            attrs:{id: 'oritention',style: 'width: 100%;'},
            selected: grid.oritention || '',
            options: [{attrs:{value: '0', innerHTML: 'Vertical'}},{attrs:{value: '1', innerHTML: 'Horizontal'}}],              
          }
          new UI.Selection(container,attrs);

          attrs={for: 'initcode',innerHTML: 'Initialize Code'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'initcode',type: 'text',value: LayoutEditor.JsonObj.data.initcode || '',placeholder: 'Initialize Code',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);
          attrs={for: 'onloadcode',innerHTML: 'OnLoad Code'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'onloadcode',type: 'text',value: LayoutEditor.JsonObj.data.onloadcode || '',placeholder: 'OnLoad Code',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);
          attrs={for: 'style',innerHTML: 'inline Style'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'style',type: 'text',value: LayoutEditor.JsonObj.data.attrs.style || '',placeholder: 'inline Style',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);
          attrs={innerHTML: 'Save',class: 'btn btn-primary'}
          let events={"click": function(){
            console.log('click')
            let name = $('#name').val();
            let version = $('#version').val();
            let isdefault = $('#isdefault').is(':checked');
            let orientation = $('#orientation').val();
            let initcode = $('#initcode').val();
            let onloadcode = $('#onloadcode').val();
            let style = $('#style').val();
            console.log(name, version, isdefault, orientation)
            LayoutEditor.JsonObj.updateNode("", {name: name, version: version, isdefault: isdefault, orientation: orientation, initcode: initcode, onloadcode: onloadcode, attrs: {style: style}} )     
            $('#properties').remove(); 
          }}
          new UI.FormControl(container, 'button',attrs,events);
          attrs={innerHTML: 'Cancel',class: 'btn btn-primary'}
          events={"click": function(){
            $('#properties').remove(); 
          }}
          new UI.FormControl(container, 'button',attrs, events);
          container.style.display = 'block';

        },
        ShowProperties: function(selectedElement){
            console.log('ShowProperties',selectedElement) 
          //  let subgrid = selectedElement.hasClass('grid-stack-sub-grid')  
            subgrid = false; 
            let el = selectedElement[0]; 
            console.log('grid element',el)
            if(!el)
                return;
            let grid = LayoutEditor.findGridNodeByelement(LayoutEditor.grid.engine.nodes, el);
            console.log(grid)
            if(!grid)
                return;
            let uiview = document.getElementsByClassName('container-fluid')[0].parentElement;
            console.log(uiview)
            let attrs={
                id: 'properties',
                class: 'properties',
                style: 'position: absolute; display:float; right: 0px; top:0px; width: 300px; height: 100%;',
            }
            let container = (new UI.FormControl(uiview, 'div',attrs)).control;
            console.log(container)
            new UI.FormControl(container, 'h3',{innerHTML: 'Properties'});
            new UI.FormControl(container, 'hr');
            attrs={
                for: 'name',
                innerHTML: 'Name'
            }
            new UI.FormControl(container, 'label',attrs);
            attrs={
                id: 'name',
                type: 'text',
                value: grid.name || '',
                placeholder: 'Name',
                style: 'width: 100%;'
            }
            new UI.FormControl(container, 'input',attrs);
            
            attrs={
                for: 'oritention',
                innerHTML: 'Oritention'
            }
            new UI.FormControl(container, 'label',attrs);

            attrs={
              attrs:{id: 'oritention',style: 'width: 100%;'},
              selected: grid.oritention || '',
              //options: [{value: '0', innerHTML: 'Vertical'},{value: '1', innerHTML: 'Horizontal'}, {value: '2', innerHTML: 'Floating'}],
              options: [{attrs:{value: '0', innerHTML: 'Vertical'}},{attrs:{value: '1', innerHTML: 'Horizontal'}}, {attrs:{value: '2', innerHTML: 'Floating'}}],
                
            }
            new UI.Selection(container,attrs);

            let rowdiv = (new UI.FormControl(container, 'div',{style: "display: row;"})).control;
            attrs={
                for: 'width',
                innerHTML: 'Width', 
                style: 'display: col'
            }
            new UI.FormControl(rowdiv, 'label',attrs);
            attrs={
              for: 'widthUnit',
              innerHTML: '', 
              style: 'display: col'
          }
          new UI.FormControl(rowdiv, 'label',attrs);
          
          rowdiv = (new UI.FormControl(container, 'div',{style: "display: row;"})).control;
          attrs = {id: 'widthmethod',style: 'width: 30%; display: col;', checked: grid.widthmethod || ''},
  
          new UI.CheckBox(rowdiv,'checkbox',attrs);
          attrs={
                id: 'width',
                type: 'text',
                value: grid.w * LayoutEditor.cellWidth || '',

                style: 'width: 40%; display: col;'
           }
           new UI.FormControl(rowdiv, 'input',attrs);
           attrs={
            attrs:{id: 'widthUnit',style: 'width: 30%; display: col;'},
            selected: grid.widthunit || '',
            //options: [{value: '0', innerHTML: 'Vertical'},{value: '1', innerHTML: 'Horizontal'}, {value: '2', innerHTML: 'Floating'}],
            options: [{attrs:{value: '%', innerHTML: '%'}},{attrs:{value: 'px', innerHTML: 'px'}}],              
          }
          new UI.Selection(rowdiv,attrs);
          rowdiv = (new UI.FormControl(container, 'div',{style: "display: row;"})).control;
          attrs={
                for: 'height',
                innerHTML: 'Height',
                style: 'width: 60%;display: col;'
          }
            new UI.FormControl(rowdiv, 'label',attrs);
            attrs={
              for: 'heightUnit',
              innerHTML: '', 
              style: 'display: col'
          }
          new UI.FormControl(rowdiv, 'label',attrs);
          rowdiv = (new UI.FormControl(container, 'div',{style: "display: row;"})).control;
          attrs = {id: 'heightmethod',style: 'width: 30%; display: col;', checked: grid.heightmethod || ''};  
          new UI.CheckBox(rowdiv,'checkbox',attrs);
            attrs={
                id: 'height',
                type: 'text',
                value: grid.h * LayoutEditor.cellHeight || '',
                style: 'width: 40%;'
            }
            new UI.FormControl(rowdiv, 'input',attrs);
            attrs={
              attrs:{id: 'heightUnit',style: 'width: 30%; display: col;'},
              selected: grid.heightunit || '',
              //options: [{value: '0', innerHTML: 'Vertical'},{value: '1', innerHTML: 'Horizontal'}, {value: '2', innerHTML: 'Floating'}],
              options: [{attrs:{value: '%', innerHTML: '%'}},{attrs:{value: 'px', innerHTML: 'px'}}],              
            }
            new UI.Selection(rowdiv,attrs);

            attrs={
                for: 'style',
                innerHTML: 'inline Style'
            }
            new UI.FormControl(container, 'label',attrs);
            attrs={
                id: 'style',
                type: 'text',
                innerHTML: grid.inlinestyle  || '',
                style: 'width: 100%; min-height: 100px;'
            }
            new UI.FormControl(container, 'textarea',attrs);

            if(!subgrid){
              attrs={
                innerHTML: 'Link View',
                
              }
              new UI.FormControl(container, 'h3',attrs);
              attrs={
                  id: 'view_name',
                  styles:'width: 100%;',
                  value:  (grid.view ?  grid.view.name : ''),
              }
              new UI.FormControl(container, 'input',attrs);
              attrs={
                  id: 'view_config',
                  styles:'width: 100%;',
                  value: (grid.view ? grid.view.config : ''),
              }
              new UI.FormControl(container, 'input',attrs);
              /*
              attrs={
                  innerHTML: 'Link View',
                  class: 'btn btn-primary',
                  style: 'width: 100%;',
                  events: {
                    click: function(){
                      LayoutEditor.link_view(el);
                    }
                  }
              }
              new UI.FormControl(container, 'button',attrs);
              attrs={
                  innerHTML: 'Unlink View',
                  class: 'btn btn-primary',
                  style: 'width: 100%;',
                  events: {
                    click: function(){
                      LayoutEditor.unline_view(el);
                    }
                  }
              }
              new UI.FormControl(container, 'button',attrs);
              */
            }

            attrs={
                innerHTML: 'Save',
                class: 'btn btn-primary',
                
            }
            let events={
              "click": function(){
                console.log('click')
                let name = $('#name').val();
                let width = $('#width').val();
                let widthunit = $('#widthUnit').val();
                let height = $('#height').val();
                let heightunit = $('#heightUnit').val();
                let style = $('#style').val();
                let oritention = $('#oritention').val();
                let widthmethod = $('#widthmethod').is(':checked');
                let heightmethod = $('#heightmethod').is(':checked');
                console.log(name, width, height, style,oritention)
                grid.name = name;
                
                let cellwidth = LayoutEditor.grid.cellWidth();
                let pagewidth = cellwidth * 100;
                if(widthunit == '%'){
                  grid.width = width;
                  grid.w = Math.round(width);
                }
                else{
                  grid.width = width;
                  grid.w = Math.round(width / cellwidth);
                }
                grid.widthunit = widthunit;
                grid.heightunit = heightunit;
                grid.h = height;
                grid.height = height;
                grid.inlinestyle = style;
                grid.oritention = oritention;
                grid.widthmethod = widthmethod;
                grid.heightmethod = heightmethod;
                if(!subgrid){
                  let view_name = $('#view_name').val();
                  let view_config = $('#view_config').val();
                  let view = {name: view_name, config: view_config};
                  grid.view = view;
                  let content = '<div class="layout_panel_operations" style="display:inline-block"><div>'+name+'" ></div>'
                  content += '<div>'+ view_name +'</div>'
                  content += '<div>'+ view_config +'</div></div>'   
                  grid.content = content;
                }else{
                  grid.content = name
                }

                LayoutEditor.generateJson();
                LayoutEditor.render();
                $('#properties').remove(); 
              }
            }
            new UI.FormControl(container, 'button',attrs,events);
            attrs={
                innerHTML: 'Cancel',
                class: 'btn btn-primary'
            }
            events={
              "click": function(){
                $('#properties').remove(); 
              }}
            new UI.FormControl(container, 'button',attrs, events);
            container.style.display = 'block';
            container.style.position = 'absolute';
            container.style.right = '0px';
            container.style.top = '0px';
            container.style.width = '300px';
            container.style.height = '100%';
            container.style.backgroundColor = 'white';
        },
        ShowTree(){
          if(LayoutEditor.JsonObj)
            LayoutEditor.JsonObj.ShowTree();
        },
        ShowRedlines(){
          if(LayoutEditor.JsonObj)
            LayoutEditor.JsonObj.showRedlines();
        }
        
}