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
        FullWidth:0,
        FullHeight:0,
        NodeList:[],
        LeftPanelWidth: 350,
        initialize: function (data=null){
             UI.Log(CustomGridStack)
            let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            LayoutEditor.LeftPanelWidth = document.getElementsByClassName('page_structure_tree')[0].offsetWidth;
            width = width - 10- LayoutEditor.LeftPanelWidth;
            height = height -50;
            
            LayoutEditor.FullWidth = width;
            LayoutEditor.FullHeight = height;

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
            LayoutEditor.ShowPageStructure();
        },
        convertwidthtole:function(width, widthunit){
          if(widthunit == '%')
            return parseInt(width);
          else{
            let lewidth = parseInt(width *100 / LayoutEditor.FullWidth);
            return lewidth;

          }
        },
        convertlewidthtowidth:function(lewidth, widthunit){
          if(widthunit == '%')
            return lewidth;
          else{
            let width = parseInt(lewidth * LayoutEditor.FullWidth / 100);
            return width;
          }
        },        
        convertheighttole:function(height, heightunit){
          if(heightunit == '%')
            return parseInt(height * LayoutEditor.FullHeight / 100);
          else
            return parseInt(height);
        },
        convertleheighttoheight:function(leheight, heightunit){
          if(heightunit == '%')
            return parseInt(leheight * 100 / LayoutEditor.FullHeight);
          else
            return parseInt(leheight);
        },
        getpaneldatafrompaneldata:function(paneldata, level=0){
          let children = [];
          UI.Log('json data:',paneldata)
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

            let widthunit = paneldata.hasOwnProperty('widthunit')? paneldata.widthunit:'px';
            let heightunit = paneldata.hasOwnProperty('heightunit')? paneldata.heightunit:'px';
            let width =  paneldata.hasOwnProperty('width')? paneldata.width:100;
            let height =  paneldata.hasOwnProperty('height')? paneldata.height:100;
            let lewidth = LayoutEditor.convertwidthtole(width, widthunit);
            let leheight = LayoutEditor.convertheighttole(height, heightunit);

            panel = {
              x: paneldata.hasOwnProperty('x')? paneldata.x:0,
              y:  paneldata.hasOwnProperty('y')? paneldata.y:0,
              w:  lewidth,
              h:  leheight,
              width:  width,
              height:  height,
              widthunit:  widthunit,
              heightunit:  heightunit,
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
          UI.Log('panel data:',panel)
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
            UI.Log(subgrid)
            let count = $('.grid-stack-item').length + 1;
            let content = '<div class="layout_panel_operations" style="display:inline-block"><div>panel'+count+'" ></div></div>'            
            if(subgrid == null)
                LayoutEditor.grid.addWidget({x:0, y:100, content:content, w:10, h:100, id:'panel'+count, name:'panel'+count,width:100,height:100,view:'',class:'layout_panel'});
            else{     
              UI.Log(subgrid)         
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
            UI.Log("layout data:", options)
            let panels = [];
            options.forEach(option => {
              panels.push(LayoutEditor.getsubpaneldata(option))
            })
            let panelsnode ={
              panels: panels
            } 
            UI.Log(panelsnode)
            UI.Log(LayoutEditor.JsonObj.getdata(""))
            LayoutEditor.JsonObj.updateNode("", {panels: panels} )     
        },
        getsubpaneldata:function(paneldata){
          let data = JSON.parse(JSON.stringify(paneldata));
          if(data.hasOwnProperty('subGridOpts'))
            delete data.subGridOpts;
          
          if(data.hasOwnProperty('content'))
            delete data.content;

          if(data.hasOwnProperty('w')){
            let widthunit = data.hasOwnProperty('widthunit')? data.widthunit:'px';
            let width = LayoutEditor.convertlewidthtowidth(data.w, widthunit);
            data.width = width;
          }

          if(data.hasOwnProperty('h')){
            let heightunit = data.hasOwnProperty('heightunit')? data.heightunit:'px';
            let height = LayoutEditor.convertleheighttoheight(data.h, heightunit);
            data.height = height;
          }

          let children = [];
          if(paneldata.hasOwnProperty('subGridOpts')){
            if(paneldata.subGridOpts.hasOwnProperty('children')){
              for(var i=0;i<paneldata.subGridOpts.children.length;i++){
                let subpanel = LayoutEditor.getsubpaneldata(paneldata.subGridOpts.children[i]);
                
                UI.Log("child data:", subpanel)
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
              UI.Log('Item moved:', item.el);
              UI.Log('New position:', item.x, item.y);
            });
          });
          
          // Event listener for item resize
          LayoutEditor.grid.on('resizestop', function(event, item) {
            UI.Log('Item resized:', item.el);
            UI.Log('New size:', item.width, item.height);
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
            UI.Log('subgrid', this)
            LayoutEditor.addSelectEvent(this);
            
          })
        },
        window_resize:function(){
          let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            width = width - 10- LayoutEditor.LeftPanelWidth;
            height = height - 50;
            LayoutEditor.FullWidth = width;
            LayoutEditor.FullHeight = height;

            $($('.container-fluid').find('.grid-stack')[0] ).css('width', width+'px');
            $($('.container-fluid').find('.grid-stack')[0] ).css('height', height+'px');
        },
        attachContextEvents: function(){
            $.contextMenu({
              selector: '.grid-stack-item', 
              build:function($triggerElement,e){
                UI.Log($triggerElement,e)
                return{
                  callback: function(key, options,e){
                    UI.Log(key, options,e)
                    switch(key){
                      case 'Add Subpanel':
                        let element = $triggerElement.find('.grid-stack')[0];
                        //let subgrid = LayoutEditor.findGridNodeByelement(LayoutEditor.grid.engine.nodes, $triggerElement[0]);
                        UI.Log(element.gridstack)
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
            UI.Log(datasource, type)
            if(type == 'collection' && datasource !=''){
              let inputs = {
                collectionname:  LayoutEditor.JsonObj.schema.datasource,
                data: LayoutEditor.JsonObj.data,
                keys: ["name"],
                operation: "update"
              }
              UI.Log(inputs)

              let ajax = new UI.Ajax("");
              ajax.post('/collection/update',inputs).then((response) => {
                let result = JSON.parse(response);
                UI.Log(result);
                if(result.data.status == 'Success'){
                  LayoutEditor.JsonObj.data= result.Outputs.data;
                  LayoutEditor.JsonObj.changed = false;
                  alert('Layout saved successfully');
                }

              }).catch((error) => {
                  UI.Log(error);
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
              UI.Log(jsonData);
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
          UI.Log(el)
          LayoutEditor.ShowProperties(el);
        },
        unline_view: function(btn){
          let el = $(btn).closest('.grid-stack-item')[0];
          UI.Log(el)
          LayoutEditor.ShowProperties(el);
        },
        showpropertiesbybtn: function(btn){
          UI.Log(btn)
          let el = $(btn).closest('.grid-stack-item')[0];
          UI.Log(el)
          LayoutEditor.ShowProperties(el);
        },
        CreatePropertySection: function(ItemTitle){
          $('#properties').remove(); 
          let uiview = document.getElementsByClassName('layouteditor_container')[0].parentElement;
          UI.Log(uiview)
          let attrs={
              id: 'properties',
              class: 'properties',
              style: 'position: absolute; display:float; right: 0px; top:45px; width: 350px; height: '+LayoutEditor.FullHeight+'px; background-color: white;',
          }
          let container = (new UI.FormControl(uiview, 'div',attrs)).control;
          UI.Log(container)

          container.style.display = 'block';
          container.style.position = 'absolute';
          container.style.right = '0px';
          container.style.top = '0px';
          container.style.width = '300px';
          container.style.height = '100%';
          container.style.backgroundColor = 'white';

          attrs = {id: 'properties_content',style: 'overflow: auto; height: 100%; width: 100%; padding: 10px;'}
          container = (new UI.FormControl(container, 'div',attrs)).control;

          new UI.FormControl(container, 'h2',{innerHTML: ItemTitle +' Properties'});
          new UI.FormControl(container, 'hr');

          return container;
        },
        ShowRootProperties: function(){
          
          let container = LayoutEditor.CreatePropertySection('Page');

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
            attrs:{id: 'orientation',style: 'width: 100%;'},
            selected: LayoutEditor.JsonObj.data.orientation || '',
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
            UI.Log('click')
            let name = $('#name').val();
            let version = $('#version').val();
            let isdefault = $('#isdefault').is(':checked');
            let orientation = $('#orientation').val();
            let initcode = $('#initcode').val();
            let onloadcode = $('#onloadcode').val();
            let style = $('#style').val();
            UI.Log(name, version, isdefault, orientation)
            LayoutEditor.JsonObj.updateNode("", {name: name, version: version, isdefault: isdefault, orientation: orientation, initcode: initcode, onloadcode: onloadcode, attrs: {style: style}} )     
            $('#properties').remove(); 
          }}
          new UI.FormControl(container, 'button',attrs,events);
          attrs={innerHTML: 'Cancel',class: 'btn btn-primary'}
          events={"click": function(){
            $('#properties').remove(); 
          }}
          new UI.FormControl(container, 'button',attrs, events);
         
        },
        ShowProperties: function(selectedElement){
            UI.Log('ShowProperties',selectedElement) 
          //  let subgrid = selectedElement.hasClass('grid-stack-sub-grid')  
            subgrid = false; 
            let el = selectedElement[0]; 
            UI.Log('grid element',el)
            if(!el)
                return;
            let grid = LayoutEditor.findGridNodeByelement(LayoutEditor.grid.engine.nodes, el);
            UI.Log(grid)
            if(!grid)
                return;
            
            let container = LayoutEditor.CreatePropertySection(grid.name);
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
                for: 'oritenorienorientationtationtion',
                innerHTML: 'orientation'
            }
            new UI.FormControl(container, 'label',attrs);

            attrs={
              attrs:{id: 'orientation',style: 'width: 100%;'},
              selected: grid.orientation || '',
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
                value: grid.width,

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
                value: grid.height,
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
              new UI.FormControl(container, 'br');
              attrs={
                  id: 'view_config',
                  styles:'width: 100%;',
                  value: (grid.view ? grid.view.config : ''),
              }
              new UI.FormControl(container, 'input',attrs);

              let clickevent = {"click": function(){
                let currentpage = Session.CurrentPage
                let popup = new UI.Popup(currentpage.container);
                popup.createPopup();
                popup.title = $('#name').val() + "'s View definition";
                let attrs=[{
                  attrs:{
                    id:"popup",									
                    style:"display:block;min-height:600px;min-width:1200px; width:100%;height:100%"
                  },							
                  children:[
                  
                    {tag:"textarea", attrs:{id:"script-editor", style:"height:80%;width:100%"}},
                  
                    {tag:'div', //attr:{id:'script-editor-buttons', class:'btn-group'},
                      children:[
                        {tag: "button", attrs:{id:"save-script", innerHTML:"Update", class:"btn btn-primary", lngcode:"Update"},events:{click: function(){
                          let scriptcontent = script_editor.getValue();			
                          $('#panel_view_definition').val(scriptcontent);								
                          popup.close()
                        }}},                        
                        {tag: "button", attrs:{id:"cancel-script", innerHTML:"Cancel", class:"btn btn-secondary",lngcode:"Cancel"}, events:{click: function(){popup.close();}}},
                      ]
                    },
                    
                  ]}]
                new UI.Builder(popup.popup, attrs);
                popup.open();

                let script_editor = CodeMirror.fromTextArea(document.getElementById("script-editor"), {
                  styleActiveLine: true,
                  lineNumbers: true,
                  matchBrackets: true,
                  autoCloseBrackets: true,
                  autoCloseTags: true,
                  matchTags: {bothTags: true},
                //   extraKeys: {"Ctrl-J": "toMatchingTag"},
                  mode: "javascript",
                  lineWrapping: true,
                  extraKeys: {"Ctrl-Q": function(cm){ cm.foldCode(cm.getCursor()); }},
                  foldGutter: true,
                  gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"]
                });
                let width = $('#popup').width() - 40;
                let height = $('#popup').height() - 70;
                script_editor.setSize(width, height);

                let jsondata = $('#panel_view_definition').val();
                script_editor.setValue(jsondata);
                try{
                  jsondata = JSON.parse(jsondata);
                  script_editor.setValue(jsondata);
                }
                catch(e){
                  UI.Log(e)                
                }
              //  script_editor.setValue(JSON.parse($('#panel_view_definition').val()));
                

              }}
              new UI.FormControl(container, 'button',{innerHTML: 'Link View',class: 'btn btn-primary',style: 'width: 100%;'},clickevent);
              new UI.FormControl(container, 'br');
              attrs={
                id: 'panel_view_definition',
                type: 'text',
                innerHTML: JSON.stringify(grid.view)  || '',
                style: 'width: 100%; min-height: 100px;'
              }
              new UI.FormControl(container, 'textarea',attrs);

            }

            attrs={
                innerHTML: 'Save',
                class: 'btn btn-primary',
                
            }
            let events={
              "click": function(){
                UI.Log('click')
                let name = $('#name').val();
                let width = $('#width').val();
                let widthunit = $('#widthUnit').val();
                let height = $('#height').val();
                let heightunit = $('#heightUnit').val();
                let style = $('#style').val();
                let orientation = $('#orientation').val();
                let widthmethod = $('#widthmethod').is(':checked');
                let heightmethod = $('#heightmethod').is(':checked');
                let panelview = $('#panel_view_definition').val();
                UI.Log(name, width, height, style,orientation)
                grid.name = name;
                
                let cellwidth = LayoutEditor.grid.cellWidth();
                let pagewidth = cellwidth * 100;
                grid.w = LayoutEditor.convertwidthtole(width, widthunit);
                grid.width = width;
                grid.h = LayoutEditor.convertheighttole(height, heightunit);
                grid.height = height;                
                grid.widthunit = widthunit;
                grid.heightunit = heightunit;
                grid.inlinestyle = style;
                grid.orientation = orientation;
                grid.widthmethod = widthmethod;
                grid.heightmethod = heightmethod;
                if(!subgrid){
                  let view_name = $('#view_name').val();
                  let view_config = $('#view_config').val();
                  let view = {name: view_name, config: view_config};
                  

                  if(panelview !=""){
                    try{
                      let panelviewobj = JSON.parse(panelview);
                      UI.Log(panelviewobj)
                      view = panelviewobj;
                      view_name = view.name;
                      view_config = view.config || view.file;
                    }catch(e){
                      UI.Log(e)
                    }
                  }
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
        },
        ShowViewProperties: function(path){
          view = LayoutEditor.JsonObj.getNode(path); 
          UI.Log(view)
          if(view == null){
            view = {name: '', config: ''};
          }
          let container = LayoutEditor.CreatePropertySection(view.name);
          attrs={
            for: 'name',
            innerHTML: 'Name'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'name',
            type: 'text',
            value: view.name || '',
            placeholder: 'Name',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs={
            for: 'config',
            innerHTML: 'Config'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'config',
            type: 'text',
            value: view.config || '',
            placeholder: 'Config',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs={
            innerHTML: 'Save',
            class: 'btn btn-primary'
          }
          let events={
            "click": function(){
              UI.Log('click')
              let name = $('#name').val();
              let config = $('#config').val();
              UI.Log(name, config)
              view.name = name;
              view.config = config;
              LayoutEditor.JsonObj.updateNode(path, view)
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
        },
        ShowActionProperties: function(path){
          let action = LayoutEditor.JsonObj.getNode(path);
          UI.Log(action)
          let paths = path.split('/');
          let actionname = paths[paths.length-1];
          let container = LayoutEditor.CreatePropertySection(actionname);
          /*
          "actions":{
            "BACK": {"type": "page", "next": "","page":"pages/pagelist.json", "panels":[]}
          }
          */
          attrs={
            for: 'type',
            innerHTML: 'Type'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            attrs:{id: 'type',style: 'width: 100%;'},
            selected: action.type || '',
            options: [
              {attrs:{value: 'Transaction', innerHTML: 'Transaction'}},
              {attrs:{value: 'Home', innerHTML: 'Home'}},
              {attrs:{value: 'Back', innerHTML: 'Back'}},
              {attrs:{value: 'page', innerHTML: 'Page'}},
              {attrs:{value: 'script', innerHTML: 'Script'}}, 
              {attrs:{value: 'view', innerHTML: 'View'}}],
          }
          new UI.Selection(container,attrs);

          attrs={
            for: 'next',
            innerHTML: 'Next'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'next',
            type: 'text',
            value: action.next || '',
            placeholder: 'Next',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs={
            for: 'page',
            innerHTML: 'Page'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'page',
            type: 'text',
            value: action.page || '',
            placeholder: 'Page',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs={
            for: 'script',
            innerHTML: 'Script'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'script',
            type: 'text',
            value: action.script || '',
            placeholder: 'Script',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'textArea',attrs);

          attrs={
            for: 'view',
            innerHTML: 'View'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'view',
            type: 'text',
            value: action.view || '',
            placeholder: 'View',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'textArea',attrs);

          attrs={
            innerHTML: 'Save',
            class: 'btn btn-primary'
          }
          let events={
            "click": function(){
              UI.Log('click')
              let type = $('#type').val();
              let next = $('#next').val();
              let page = $('#page').val();
              let script = $('#script').val();
              let view = $('#view').val();
              UI.Log(type, next, page, script, view)
              action.type = type;
              action.next = next;
              action.page = page;
              action.script = script;
              action.view = view;
              LayoutEditor.JsonObj.updateNode(path, action)
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
            }
          }
          new UI.FormControl(container, 'button',attrs, events);
        },
        ShowPageStructure: function(){

          let container = document.getElementsByClassName('page_structure_tree')[0];
          container.innerHTML = "";
          new UI.FormControl(container, 'div', {id:'ui-json-object-tree',class:'tree',style:'width:100%;height:100%;'});


          
          let rootid = LayoutEditor.JsonObj.data.id ? LayoutEditor.JsonObj.data.id : UI.generateUUID();
          let rootdata ={
            text: LayoutEditor.JsonObj.data.name,
            id: rootid,
            parent: "#",
            state: { opened: true },
            a_attr: {"node-type": "page", "data-key":LayoutEditor.JsonObj.data.name} ,
            icon: "fa fa-newspaper",          
          }
          let nodelist =[];
          nodelist.push(rootdata);
          if(LayoutEditor.JsonObj.data.hasOwnProperty('panels')){
            for(var i=0;i<LayoutEditor.JsonObj.data.panels.length;i++){
              let panel = LayoutEditor.JsonObj.data.panels[i];
              let childpanels = LayoutEditor.getChildPanels(rootid, panel, "panels");
              nodelist = nodelist.concat(childpanels);
            }
          }
          LayoutEditor.NodeList = nodelist;
          UI.Log(nodelist)
        //  let pagetree= document.getElementById('ui-json-object-tree');
       //   pagetree.setData(nodelist);
          
          $(function() {
            $('#ui-json-object-tree').jstree({
            'core': {
              'data': nodelist
            }
            });		
          });  
          $('#ui-json-object-tree').on("select_node.jstree", function (e, data) {
            const selectedNodeData = data.node;
            let nodeId = selectedNodeData.id;
            let nodeText = selectedNodeData.text;
            let nodetype = selectedNodeData.a_attr["node-type"];
            let nodekey = (selectedNodeData.a_attr["data-key"]).replace(/'/g, '"');;
            switch(nodetype){
              case 'page':
                LayoutEditor.ShowRootProperties();
                break;
              case 'panel':
                let element = LayoutEditor.getGridbyPanelID(nodeId, LayoutEditor.grid.engine.nodes)
                UI.Log(element)
                LayoutEditor.ShowProperties($(element));
                break;
              case 'view':                               
                UI.Log(nodekey)
                LayoutEditor.ShowViewProperties(nodekey);
                break;
              case 'panelview':
                let panelview = LayoutEditor.findPanelView(LayoutEditor.JsonObj.data, nodekey);
                UI.Log(panelview)
                LayoutEditor.ShowProperties(panelview);
                break;
              case 'action':
                let action = LayoutEditor.findAction(LayoutEditor.JsonObj.data, nodekey);
                UI.Log(action)
                LayoutEditor.ShowProperties(action);
                break;              
            }
          });
          LayoutEditor.AttachContextMenutoTree();
        },
        AttachContextMenutoTree: function(){
          $.contextMenu({
            selector: '.jstree-anchor[node-type="panel"]', 
            build:function($triggerElement,e){
              UI.Log($triggerElement,e)
              let node = $triggerElement.closest('li.jstree-node');
              let panelid = node.attr('id');
              let element = LayoutEditor.getGridbyPanelID(panelid, LayoutEditor.grid.engine.nodes)
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(element,node,panelid,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Add Subpanel':               
                     
                      LayoutEditor.addPanel(element.gridstack);
                      LayoutEditor.generateJson();
                      layoutEditor.render();
                      layoutEditor.ShowPageStructure();
                      break;
                    case 'Properties':
                      LayoutEditor.ShowProperties($(element));
                      break;
                    case 'Remove':
                      let gridEl = element;
                      LayoutEditor.grid.removeWidget(gridEl);
                      LayoutEditor.generateJson();
                      layoutEditor.render();
                      layoutEditor.ShowPageStructure();
                      break;
                  }

                }, 
                items:{
                  'Properties':{
                    name: 'Properties',
                    icon: 'fa-cog',
                    disabled: false
                  },
                  'Add Subpanel':{
                    name: 'Add Subpanel',
                    icon: 'fa-plus',
                    disabled:function(){
                      let nodedata = (LayoutEditor.JsonObj.getNode(nodekey)).value;
                      if(nodedata.hasOwnProperty('view')|| nodedata.hasOwnProperty('panelviews')){
                        return true;
                      }else{
                        return false;
                      } 
                    }
                  }, 
                  'Link View':{
                    name: 'Link View',
                    icon: 'fa-plus',
                    disabled:function(){
                      let nodedata = (LayoutEditor.JsonObj.getNode(nodekey)).value;
                      if(nodedata.hasOwnProperty('panels')){
                        if (nodedata.panels.length > 0)
                            return true;
                      }
                      return false;
                    }
                  }, 
                  'Add Panel View':{
                    name: 'Add Panel View',
                    icon: 'fa-plus',
                    disabled:function(){
                      let nodedata = (LayoutEditor.JsonObj.getNode(nodekey)).value;
                      if(nodedata.hasOwnProperty('panels')){
                        if (nodedata.panels.length > 0)
                            return true;
                      }
                      return false;
                    }

                  }, 
                  'Remove':{
                    name: 'Remove',
                    icon: 'fa-minus',
                    disabled: function(){
                      let nodedata = (LayoutEditor.JsonObj.getNode(nodekey)).value;
                      if(nodedata.hasOwnProperty('panels')){
                        if (nodedata.panels.length > 0)
                            return true;
                      }
                    }
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
          $.contextMenu({
            selector: '.jstree-anchor[node-type="view"]', 
            build:function($triggerElement,e){
              UI.Log($triggerElement,e)
              let node = $triggerElement.closest('li.jstree-node');
              let panelid = node.attr('id');
              let element = LayoutEditor.getGridbyPanelID(panelid, LayoutEditor.grid.engine.nodes)
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(element,node,panelid,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Add Action':               
                        let actionname = prompt("Please enter action name", "");
                        if (actionname != null) {
                          let viewnode = LayoutEditor.JsonObj.getNode(nodekey);
                          let actions = {};
                          if(viewnode.value.hasOwnProperty('actions')){
                            actions = viewnode.value.actions;
                            for(key in actions.value){
                              if(key == actionname){
                                UI.ShowError("Action name "+actionname+ " already exists");
                                return;
                              }
                            }
                          }else{
                          //  UI.Log('insert actions',nodekey)
                          //  LayoutEditor.JsonObj.inserNodeKey(nodekey,'actions')
                            actions = {};                             
                          }
                          let action = {type: 'page', next: '', page: '', script: '', view: ''};
                          actions[actionname] = action;
                          UI.Log(nodekey + "/actions", actions)
                          LayoutEditor.JsonObj.setNodewithKey(nodekey, 'actions', actions);
                          
                          LayoutEditor.ShowPageStructure();
                          LayoutEditor.ShowActionProperties(nodekey + "/actions/" + actionname);
                        }
                      break;
                    case 'Properties':
                      LayoutEditor.ShowViewProperties(nodekey);
                      break;
                    case 'Remove':
                      
                      break;
                  }

                }, 
                items:{
                  'Properties':{
                    name: 'Properties',
                    icon: 'fa-cog',
                    disabled: false
                  },
                  'Add Action':{
                    name: 'Add Action',
                    icon: 'fa-plus',
                    disabled:false
                  },                   
                  'Remove':{
                    name: 'Remove',
                    icon: 'fa-minus',
                    disabled: function(){
                      let nodedata = (LayoutEditor.JsonObj.getNode(nodekey)).value;
                      if(nodedata.hasOwnProperty('actions')){
                        if (nodedata.actions.length > 0)
                            return true;
                      }
                    }
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
          $.contextMenu({
            selector: '.jstree-anchor[node-type="action"]', 
            build:function($triggerElement,e){
              UI.Log($triggerElement,e)
              let node = $triggerElement.closest('li.jstree-node');
              let panelid = node.attr('id');
              let element = LayoutEditor.getGridbyPanelID(panelid, LayoutEditor.grid.engine.nodes)
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(element,node,panelid,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Properties':
                      LayoutEditor.ShowActionProperties(nodekey);
                      break;
                    case 'Remove':
                      
                      break;
                  }

                }, 
                items:{
                  'Properties':{
                    name: 'Properties',
                    icon: 'fa-cog',
                    disabled: false
                  },                                   
                  'Remove':{
                    name: 'Remove',
                    icon: 'fa-minus',
                    disabled: false,
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

          $.contextMenu({
            selector: '.jstree-anchor[node-type="page"]', 
            build:function($triggerElement,e){
              UI.Log($triggerElement,e)
              let node = $triggerElement.closest('li.jstree-node');
              let panelid = node.attr('id');
              let element = LayoutEditor.getGridbyPanelID(panelid, LayoutEditor.grid.engine.nodes)
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(element,node,panelid,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Add Panel':            
                      LayoutEditor.addPanel();
                      break;
                    case 'Properties':
                      LayoutEditor.ShowRootProperties();
                      break;
                    case 'Save':
                      LayoutEditor.savelayout();
                      break;
                    case 'Redlines':
                      LayoutEditor.showRedlines();
                      break;
                  }

                }, 
                items:{
                  'Properties':{
                    name: 'Properties',
                    icon: 'fa-cog',
                    disabled: false
                  },
                  'Add Panel':{
                    name: 'Add Panel',
                    icon: 'fa-plus',
                    disabled:false
                  },               
                  'Save':{
                    name: 'Save',
                    icon: 'fa-save',
                    disabled:false
                  },
                /*   'Save As':{
                    name: 'Save As',
                    icon: 'fa-save',
                    disabled:false
                  }, 
                  'Export':{
                    name: 'Export',
                    icon: 'fa-export',
                    disabled:false
                  },  */
                  'Redlines':{
                    name: 'Redlines Change',
                    icon: 'fa-code-compare',
                    disabled:false
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
        getGridbyPanelID: function(panelid, nodes){
          
          for(var i=0;i<nodes.length;i++){
            if(nodes[i].id == panelid)
              return nodes[i].el;
          }
          for(var i=0;i<nodes.length;i++){
            if(nodes[i].subGrid)
              if(nodes[i].subGrid.engine)
                if(nodes[i].subGrid.engine.nodes){
                  let el = LayoutEditor.getGridbyPanelID(panelid, nodes[i].subGrid.engine.nodes)
                  if(el != null)
                    return el;
                }
          }
          return null;          
        },
        getChildPanels: function(parent, panel, parentkey){             
          let panellist =[];
          UI.Log("get the panel tree node:", panel, parent)
          
          if(panel.hasOwnProperty('panels')){
            let key = parentkey + "/{'id':'" + panel.id + "'}";
            let data = {
              id: panel.id,
              parent: parent,
              text: panel.name,              
              a_attr: {"node-type": "panel", "data-key":key} ,
              icon: "fa fa-solid fa-square", 
              state: { opened: true },
            }
            panellist.push(data); 

            for(var i=0;i<panel.panels.length;i++){
              let childpanels = LayoutEditor.getChildPanels(panel.id, panel.panels[i], key + "/panels");
              panellist = panellist.concat(childpanels);
            }
          }else if(panel.hasOwnProperty('view')){
            let datakey = parentkey + "/{'id':'" + panel.id + "'}";
            let data = {
              id: panel.id,
              parent: parent,
              text: panel.name,
              type: "panel",
              a_attr: {"node-type": "panel", "data-key":datakey} ,
              icon: "fa fa-regular fa-square",
              state: { opened: true },
            }
            panellist.push(data); 
            let viewID = UI.generateUUID();
            let viewkey = datakey + "/view";
            data = {
              id: viewID,
              parent: panel.id,
              text: panel.view.name ? panel.view.name : "-----",
              a_attr: {"node-type": "view", "data-key":viewkey} ,
              icon: "fa fa-regular fa-flag",
              state: { opened: true },
            }
            panellist.push(data);

            if(panel.view.hasOwnProperty('actions')){
              let actions = panel.view.actions;
              for(key in actions){
                let data = {
                  id: UI.generateUUID(),
                  parent: viewID,
                  text: key,
                  a_attr: {"node-type": "action", "data-key": viewkey + "/actions/" + key} ,
                  icon: "fa fa-regular fa-diamond",
                  state: { opened: true },
                }
                panellist.push(data); 
              }
            }
            let panelviewnodeid = UI.generateUUID()
            let panelviewkey = datakey + "/panelviews";
            data = {
              id: UI.generateUUID(),
              parent: panel.id,
              text: "Panel Views",
              a_attr: {"node-type": "panelviewcontainer", "data-key":panelviewkey} ,
              icon: "fa fa-regular fa-layer-group",
              state: { opened: true },
            }
            panellist.push(data);

            if (panel.hasOwnProperty("panelviews")){
              for(var i=0;i<panel.panelviews.length;i++){
                let panelview = panel.panelviews[i];
                let panelviewid = UI.generateUUID();
                let data = {
                  id: panelviewid,
                  parent: panelviewnodeid,
                  text: panelview.name,
                  a_attr: {"node-type": "panelview", "data-key":panelviewkey + "/{'name':'" + panelview.name + "'}"} ,
                  icon: "fa fa-regular fa-expand",
                  state: { opened: true },
                }
                panellist.push(data); 

                if(panelview.hasOwnProperty('actions')){
                  let actions = panelview.actions;
                  for(key in actions){
                    let data = {
                      id: UI.generateUUID(),
                      parent: panelviewid,
                      text: key,
                      type: "action",
                      a_attr: {"node-type": "action", "data-key":panelviewkey + "/{'name':'" + panelview.name + "'}/actions/"+key} ,
                      icon: "fa fa-regular fa-diamond",
                      state: { opened: true },
                    }
                    panellist.push(data); 
                  }
                }

              }
            }
          }
          return panellist;
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