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
                  "panels": [{
                    x: 0,y:0, width: 100, height:LayoutEditor.cellWidth, 
                    widthunit: '%', heightunit: 'px', 
                    iscontainer: false, noMove: true, 
                    id: UI.generateUUID(), 
                    name: 'main_layout_panel', 
                    content: 'main_layout_panel', 
                    class: 'main_layout_panel', 
                    orientation: 0, inlinestyle: '', widthmethod: false, heightmethod: false, 
                    subGridOpts: {children: []}
                  }]
                }
                LayoutEditor.JsonObj = new UI.JSONManager(sampledata, {allowChanges:true, schema:LayoutEditor.schema})
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


 
            $($('.container-fluid').find('.grid-stack')[0] ).css('background-color', 'lightgrey');
            $($('.container-fluid').find('.grid-stack')[0] ).css('width', LayoutEditor.FullWidth+'px');
            $($('.container-fluid').find('.grid-stack')[0] ).css('height', LayoutEditor.FullHeight+'px');
            LayoutEditor.render();

          /*  let gridEls = CustomGridStack.getElements('.grid-stack-item');
            gridEls.forEach(gridEl => {
                LayoutEditor.addSelectEvent(gridEl);
            }) */

          //  $('.selected').length > 0 ? $('.btn_removepanel').show() : $('.btn_removepanel').hide();

            window.addEventListener('resize', LayoutEditor.window_resize);
          //  LayoutEditor.attachContextEvents();
            LayoutEditor.ShowPageStructure();
        },
        generateLayout:function(){          
          LayoutEditor.destroy(true);
          document.querySelector('.container-fluid').innerHTML = '';
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
            children: []
          };

          layoutdata = LayoutEditor.getpaneldatafrompaneldata(LayoutEditor.JsonObj.data, 0)
          UI.Log(LayoutEditor.Options, layoutdata)
          LayoutEditor.Options = Object.assign(LayoutEditor.Options,layoutdata);
          // create and load it all from JSON above
          LayoutEditor.grid = CustomGridStack.addGrid(document.querySelector('.container-fluid'), LayoutEditor.Options);

          $($('.container-fluid').find('.grid-stack')[0] ).css('background-color', 'lightgrey');
          $($('.container-fluid').find('.grid-stack')[0] ).css('width', LayoutEditor.FullWidth+'px');
          $($('.container-fluid').find('.grid-stack')[0] ).css('height', LayoutEditor.FullHeight+'px');
          LayoutEditor.render();

          let gridEls = CustomGridStack.getElements('.grid-stack-item');
        /*  gridEls.forEach(gridEl => {
              LayoutEditor.addSelectEvent(gridEl);
          })
          */
        //  $('.selected').length > 0 ? $('.btn_removepanel').show() : $('.btn_removepanel').hide();

          window.addEventListener('resize', LayoutEditor.window_resize);
       //   LayoutEditor.attachContextEvents();     

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
        getpaneldatafrompaneldata:function(paneldata, level=0, parent = null){
          let children = [];
          UI.Log('json data:',paneldata)
          /*
          0: flex-direction: column; vertical
          1: flex-direction: row;
          */
          let orientation= paneldata.hasOwnProperty('orientation')?paneldata.orientation: 0
          if(paneldata.hasOwnProperty('panels')){
            let panels = paneldata.panels;

            for(var i=0;i<panels.length;i++){
              let panel = LayoutEditor.getpaneldatafrompaneldata(panels[i], level+1, paneldata);
              children.push(panel);
            }
          /*
            let wcalcpanels = [];
            let hcalcpanels =[];
            let subtotalw = 0;
            let subtotalh = 0;
            for(var i=0;i<children.length;i++){
              if(children[i].heightmethod  && orientation == 0){                
                hcalcpanels.push(children[i]);
                children = children.splice(i,1);
              }else if(children[i].widthmethod && orientation == 1){                
                 wcalcpanels.push(children[i]);
                 children = children.splice(i,1);
              }else{
                if(orientation == 0){
                  subtotalh += children[i].h;
                }else{
                  subtotalw += children[i].w;
              }
              }
            }
            
            if(hcalcpanels.length == 1){
              let h = LayoutEditor.FullHeight - subtotalh;
              for(var i=0;i<hcalcpanels.length;i++){
                hcalcpanels[i].h = h > 0? h:0;
                children.push(hcalcpanels[i]);
              }
            }else if (hcalcpanels.length > 1){
              children = children.concat(hcalcpanels);
            }

            if(wcalcpanels.length ==1){
              let w = 100 - subtotalw;
              for(var i=0;i<wcalcpanels.length;i++){
                wcalcpanels.w = w > 0? w:0;
                children.push(wcalcpanels[i]);
              }
            }else if(wcalcpanels.length > 1){
              children = children.concat(wcalcpanels);
            }
            */          
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
            let iscontainer = paneldata.hasOwnProperty('iscontainer')? paneldata.iscontainer:false;
            let subGridOpts = {children:children}
            if(!iscontainer){
              let content = '<div class="fa fa-link">' + paneldata.hasOwnProperty('view')?paneldata.view.name : "" + '</div>'
              if(paneldata.hasOwnProperty('view')){
                if(!paneldata.view.hasOwnProperty('ID'))
                  paneldata.view.id = UI.generateUUID();
              }

              if(paneldata.hasOwnProperty('panelviews')){
                for(var i=0;i<paneldata.panelviews.length;i++){
                  content += '<div class="fa fa-link">' + paneldata.panelviews[i].name +'</div>'
                  if(!paneldata.panelviews[i].hasOwnProperty('ID'))
                    paneldata.panelviews[i].id = UI.generateUUID();
                }

              }
            }else{
              subGridOpts.acceptWidgets = true;
            }

            
            panel = {
              x: paneldata.hasOwnProperty('x')? paneldata.x:0,
              y:  paneldata.hasOwnProperty('y')? paneldata.y:0,
              w:  lewidth,
              h:  leheight,
              width:  width,
              height:  height,
              widthunit:  widthunit,
              heightunit:  heightunit,
              iscontainer: iscontainer,
              locked: !iscontainer,
              noMove: true,
              id:  paneldata.hasOwnProperty('id')? paneldata.id: UI.generateUUID(),
              name: paneldata.hasOwnProperty('name')? paneldata.name:'panel',
              content: paneldata.hasOwnProperty('name')?paneldata.name:'panel',
              view: paneldata.hasOwnProperty('view')?paneldata.view:{},
              panelviews :paneldata.hasOwnProperty('panelviews')?paneldata.panelviews:[],
              class: paneldata.hasOwnProperty('class')?paneldata.class:'',
              orientation: paneldata.hasOwnProperty('orientation')?paneldata.orientation: 1,
              inlinestyle: paneldata.hasOwnProperty('inlinestyle')?paneldata.inlinestyle:'',
              widthmethod: paneldata.hasOwnProperty('widthmethod')?paneldata.widthmethod:false,
              heightmethod: paneldata.hasOwnProperty('heightmethod')?paneldata.heightmethod:false,
              subGridOpts: subGridOpts
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
        //  LayoutEditor.capatureevents();
        /*  let subgrid = LayoutEditor.grid.makeSubGrid(node);
          LayoutEditor.addSelectEvent(subgrid)      */  
        },

        addPanel: function (grid){
            UI.Log(grid)
            let count = $('.grid-stack-item').length + 1;
            let content = '<div class="layout_panel_operations" style="display:inline-block"><div>panel'+count+'" ></div></div>'            
            if(grid == null || grid == undefined || grid == "")
                LayoutEditor.grid.addWidget({x:0, y:100, content:content, w:10, h:100, id:UI.generateUUID(), name:'panel'+count,width:100,height:100,view:{},class:'layout_panel'});
            else{     
              
              UI.Log(grid)         
              grid.addWidget({x:0, y:100, content:content, w:10, h:100, id:UI.generateUUID(), name:'panel'+count,width:100,height:100,view:{},class:'layout_panel'});
            }

        //    LayoutEditor.addSelectEvent(cell)
        },
        render:function() {
          LayoutEditor.Options = LayoutEditor.grid.save(true, false);
          LayoutEditor.grid.removeAll();
          LayoutEditor.load(LayoutEditor.Options,false); 
        //  LayoutEditor.attachContextEvents();
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
          UI.Log("panel data:", paneldata)
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

          UI.Log("parsed panel data:", data)
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
              LayoutEditor.generateJson();              
            });
          });
          
          // Event listener for item resize
          LayoutEditor.grid.on('resizestop', function(event, item) {
            UI.Log('Item resized:', item.el);
            UI.Log('New size:', item.width, item.height);
            LayoutEditor.generateJson();
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
                //LayoutEditor.JsonObj.data= result.Outputs.data;
                LayoutEditor.JsonObj.changed = false;
                
                UI.ShowMessage('Layout saved successfully','Success');

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
            LayoutEditor.ShowPageStructure();
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
              for: 'panelcontainer',
              innerHTML: 'Is a Parent Panel?', 
          }
          new UI.FormControl(container, 'label',attrs);

          attrs = {id: 'panelcontainer',style: 'width: 100%; display: col;', value: grid.hasOwnProperty('iscontainer')? grid.iscontainer : false };
          new UI.CheckBox(container,'checkbox',attrs);


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
          attrs = {id: 'widthmethod',style: 'width: 30%; display: col;', value: grid.widthmethod || false},
  
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
          attrs = {id: 'heightmethod',style: 'width: 30%; display: col;', value: grid.heightmethod || false};  
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
                
                UI.Log(name, width, height, style,orientation)
                grid.name = name;

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
                grid.iscontainer = $('#panelcontainer').is(':checked');                
                grid.content = name         
                LayoutEditor.grid.update(el, grid.x, grid.y, grid.w, grid.h);       
                LayoutEditor.generateJson();
                LayoutEditor.ShowPageStructure();
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
          let view = LayoutEditor.JsonObj.getNode(path); 
          UI.Log("view node:", path,view)
          UI.Log(view)

          if(view == null){
            view = {name: '', config: '', inputs:{}, outputs:{}, actions:{}};
          }else{
            if(view.value =="")
              view = {name: '', config: '', inputs:{}, outputs:{}, actions:{}};
            else
              view = view.value
          }
          let container = LayoutEditor.CreatePropertySection(view.name);

          attrs={
            for: 'name',  innerHTML: 'Name'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'name', type: 'text', value: view.name || '',   placeholder: 'Name',       style: 'width: 100%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs={
            for: 'config',    innerHTML: 'Config'
          }
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'config',     type: 'text',     value: view.config || '',       placeholder: 'Config',         style: 'width: 80%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs={id: 'linkconfig', class:"fa-solid fa-link", style: "width: 20%;", onclick: "LayoutEditor.load_viewConfigue('"+path+"')"} 
          new UI.FormControl(container, 'i',attrs);
          attrs ={for: "file", innerHTML: "File"}
          new UI.FormControl(container, 'label',attrs);
          attrs={
            id: 'file',          type: 'text',         value: view.file || '',       placeholder: 'file to show the content',       style: 'width: 100%;'
          }
          new UI.FormControl(container, 'input',attrs);
          attrs ={for: "View_inputs", innerHTML: "Inputs"}
          new UI.FormControl(container, 'label',attrs);
          let parameterstr = "";
          if(view.hasOwnProperty("inputs")){
            UI.Log(view.inputs)
            for(key in view.inputs){
              UI.Log(parameterstr, key)
              parameterstr = parameterstr + key + ";"
            }
          }  
          attrs={
            id: 'View_inputs',
            type: 'text',
            innerHTML: parameterstr || '',
            placeholder: 'inputs to render the view',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'textarea',attrs);
          attrs ={for: "View_outputs", innerHTML: "Outputs"}
          new UI.FormControl(container, 'label',attrs);
          parameterstr = "";
          if(view.hasOwnProperty("outputs")){
            for(key in view.outputs)
                parameterstr = parameterstr + key + ";"
          }          
          attrs={
            id: 'View_outputs',
            type: 'text',
            innerHTML: parameterstr || '',
            placeholder: 'Outputs that view generated',
            style: 'width: 100%;'
          }
          new UI.FormControl(container, 'textarea',attrs);

          attrs={
            innerHTML: 'Save',
            class: 'btn btn-primary'
          }
          let events={
            "click": function(){
              UI.Log('click', view)
              let name = $('#name').val();
              let config = $('#config').val();             
              view.name = name;
              view.config = config;
              view.file = $('#file').val();
              let parameterstr = $('#View_inputs').val()
              let parameterlist = parameterstr.split(';')
              let inputs={}
              for(var i=0;i<parameterlist.length;i++){
                parameterlist[i] = parameterlist[i].trim()
                if(parameterlist[i]!="")
                  inputs[parameterlist[i]] = ""
              }
              view.inputs =inputs;
              parameterstr = $('#View_outputs').val()
               parameterlist = parameterstr.split(';')
              let outputs={}
              for(var i=0;i<parameterlist.length;i++){
                parameterlist[i] = parameterlist[i].trim()
                if(parameterlist[i]!="")
                  outputs[parameterlist[i]] = ""
              }
              view.outputs =outputs;
              UI.Log(path, view)
              LayoutEditor.JsonObj.updateNodeValue(path,view)
              LayoutEditor.generateLayout();
              LayoutEditor.ShowPageStructure();
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
        load_viewConfigue: function(path){
          
          let cfg = {
            "file":"templates/datalist.html", 
            "name": "View List", 
            "actions": {
                "SELECT":{"type": "script", "next": "","page":"","panels":[], "script": "selectitem"},
                "CANCEL":{"type": "script", "next": "","page":"","panels":[], "script": "cancelitem"},
            }
          }
          let page =Session.CurrentPage;
          UI.Log(page)
          let org_schema = Session.snapshoot.sessionData.ui_dataschema
          let org_entity = Session.snapshoot.sessionData.entity
          let org_selectedKey = Session.snapshoot.sessionData.selectedKey

          let inputs = {}
          inputs.ui_dataschema = 'uiview'
        //    UI.Log(inputs)
          cfg.inputs = inputs;
          cfg.actions.SELECT.script = function(data){
            UI.Log(data)
           
          //  Session.snapshoot.sessionData.selectedKey = data.key;
          //  Session.snapshoot.sessionData.ui_dataschema = data.schema;
            let url = "/collection/id"
            let inputs={
              collectionname:'UI_View',
              data: {"_id": data.selectedKey},
              operation: "detail"
            }
            UI.Log("get the detail view:",inputs)
            UI.Post(url, inputs, function(response){
              UI.Log(response)
              let result = JSON.parse(response);
              let view = LayoutEditor.JsonObj.getNode(path);
              let data = {};
              data.name = result.data.name;
              data.config = result.data.name;
              data.title = result.data.title;
              data.revision = result.data.revision;
              data.inputs = result.data.inputs;
              data.outputs = result.data.outputs;
              data.actions = result.data.actions;
              view.value.type = "document"
              let viewdata = Object.assign(view.value, data);
              view.value = viewdata;
              LayoutEditor.generateLayout();
              LayoutEditor.ShowPageStructure();
              page.popupClose();  
              Session.snapshoot.sessionData.ui_dataschema = org_schema;
              Session.snapshoot.sessionData.selectedKey = org_selectedKey;
            }, function(error){
              UI.ShowError(error);
                            
              page.popupClose(); 
              Session.snapshoot.sessionData.ui_dataschema = org_schema;
              Session.snapshoot.sessionData.selectedKey = org_selectedKey; 
            })                     
          }
          cfg.actions.CANCEL.script = function(data){
            UI.Log("execute the action:", data)
            Session.snapshoot.sessionData.selectedKey = org_selectedKey;
            Session.snapshoot.sessionData.ui_dataschema = org_schema;
            page.popupClose();
          }
          Session.snapshoot.sessionData.ui_dataschema = "uiview"
          //UI.Log(cfg)
          //new UI.View(panel,cfg) 
          page.popupOpen(cfg);
          page.popup.onClose(function(){
              Session.snapshoot.sessionData.selectedKey = org_selectedKey;
              Session.snapshoot.sessionData.ui_dataschema = org_schema;
          })

        },
        ShowActionProperties: function(path){
          let action = LayoutEditor.JsonObj.getNode(path).value;
          UI.Log(action)
          let paths = path.split('/');
          let actionname = paths[paths.length-1];
          let container = LayoutEditor.CreatePropertySection(actionname);
          let attrs={for: 'name', innerHTML: 'Name', lngcode: 'Name'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'name', type: 'text', value: actionname, placeholder: 'Name', style: 'width: 100%; disabled'}
          new UI.FormControl(container, 'input',attrs);

          attrs={for: 'type', innerHTML: 'Type', lngcode: 'Type'}
          new UI.FormControl(container, 'label',attrs);
          attrs={
            attrs:{id: 'type',style: 'width: 100%;'}, selected: action.type || '',
            options: [
              {attrs:{value: 'Transaction', innerHTML: 'Transaction', lngcode: 'Transaction'}},
              {attrs:{value: 'Home', innerHTML: 'Home', lngcode: 'Home'}},
              {attrs:{value: 'Back', innerHTML: 'Back',lngcode: 'Back'}},
              {attrs:{value: 'page', innerHTML: 'Page',lngcode: 'Page'}},
              {attrs:{value: 'script', innerHTML: 'Script',lngcode: 'Script'}}, 
              {attrs:{value: 'view', innerHTML: 'View',lngcode: 'View'}}],
          }
          new UI.Selection(container,attrs);

          attrs={for: 'trancode',innerHTML: 'TranCode',lngcode: 'TranCode'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'trancode',    type: 'text',      value: action.code || '',   placeholder: 'Page',       style: 'width: 100%;'      }
          new UI.FormControl(container, 'input',attrs);
          attrs={for: 'actionpage',innerHTML: 'Page'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'actionpage',    type: 'text',      value: action.page || '',   placeholder: 'Page',       style: 'width: 100%;'      }
          new UI.FormControl(container, 'input',attrs);
          attrs={      for: 'script',         innerHTML: 'Script'       }
          new UI.FormControl(container, 'label',attrs);
          attrs={     id: 'script',            type: 'text',          value: action.script || '',            placeholder: 'Script',            style: 'width: 100%;'          }
          new UI.FormControl(container, 'textArea',attrs);

          attrs={   for: 'view',      innerHTML: 'View'    }
          new UI.FormControl(container, 'label',attrs);
          attrs={      id: 'view',  type: 'text',  value: action.view || '',  placeholder: 'View',       style: 'width: 100%;' }
          new UI.FormControl(container, 'textArea',attrs);

          attrs={for: 'next',innerHTML: 'Next'}
          new UI.FormControl(container, 'label',attrs);
          attrs={id: 'next', type: 'text', value: action.next || '', placeholder: 'Next',style: 'width: 100%;'}
          new UI.FormControl(container, 'input',attrs);

          attrs={
            innerHTML: 'Save',
            class: 'btn btn-primary'
          }
          let events={
            "click": function(){
              UI.Log('click')
              
              let type = $('#type').val();
              let next = $('#next').val();
              let page = $('#actionpage').val();
              let script = $('#script').val();
              let view = $('#view').val();
              let trancode = $('#trancode').val();
              UI.Log(type, next, page, script, view)
              action.type = type;
              action.next = next;
              action.page = page;
              action.script = script;
              action.view = view;
              action.code = trancode;
              let paths = path.split('/');
              let actionpath =[];
              let viewtype = "view";
              for(var i=path.length-1;i>=0;i--){
                actionpath.push(path[i]);
                if(path[i] == 'panelviews'){
                  viewtype = "panelviews";
                  break;
                }else if(path[i] == 'view'){
                  break;
                }
              }
            //  LayoutEditor.JsonObj.updateNode(path, action)
              LayoutEditor.generateLayout();
              LayoutEditor.ShowPageStructure();
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
            a_attr: {"node-type": "page", "panelid":"#" ,"data-key":LayoutEditor.JsonObj.data.name} ,
            icon: "fa-brands fa-delicious",          
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
       /*   $('#ui-json-object-tree').on("select_node.jstree", function (e, data) {
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
                let element = (LayoutEditor.getGridNodebyPanelID(nodeId, LayoutEditor.grid.engine.nodes)).el
                UI.Log(element)
                LayoutEditor.ShowProperties($(element));
                break;
              case 'view':                               
                UI.Log(nodekey)
                LayoutEditor.ShowViewProperties(nodekey);
                break;
              case 'panelview':
                LayoutEditor.ShowViewProperties(nodekey);
                break;
              case 'action':
                let action = LayoutEditor.ShowActionProperties(nodekey);
            
                break;              
            }
          });   */
          LayoutEditor.AttachContextMenutoTree();
        },
        AttachContextMenutoTree: function(){
          $.contextMenu({
            selector: '.jstree-anchor[node-type="panel"]', 
            build:function($triggerElement,e){
              UI.Log($triggerElement,e)
              let node = $triggerElement.closest('li.jstree-node');
              let panelid = node.attr('id');
              let gridNode = (LayoutEditor.getGridNodebyPanelID(panelid, LayoutEditor.grid.engine.nodes))
              let element = (gridNode).el
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(element,node,panelid,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Add Subpanel':   
                      UI.Log('Add Subpanel:', nodekey, gridNode,gridNode.grid)
                      let subpanelname = prompt("Please enter subpanel name", "");
                      if (subpanelname != null) {
                        let subpanel = {x:0, y:100, content:subpanelname, w:10, h:100, id:UI.generateUUID(), name:subpanelname,width:100,height:100,view:{},class:'layout_panel'}
                        let panelpath = nodekey + "/panels";
                        let panels = LayoutEditor.JsonObj.addNode(panelpath, subpanel);
                        LayoutEditor.generateLayout();

                        LayoutEditor.ShowPageStructure();
                      }
                      break;
                    case 'Properties':
                      LayoutEditor.ShowProperties($(element));
                      break;
                    case 'Remove':                      
                      /*let gridEl = CustomGridStack.getElements('.grid-stack-item[gs-id="'+panelid+'"]');
                      UI.Log("remove the panel:",gridEl)
                      LayoutEditor.grid.removeWidget(gridEl[0]); */
                      LayoutEditor.removeGridNodebyPanelID(panelid,LayoutEditor.grid.engine.nodes);
                      LayoutEditor.generateJson();
                      LayoutEditor.render();
                      LayoutEditor.ShowPageStructure();
                      break;
                    case 'Add Panel View':
                      let viewname = prompt("Please enter view name", "");
                      if (viewname != null) {
                        let view = {name: viewname, config: '', file: '', inputs: {}, outputs: {}};
                        let panelviewspath = nodekey + "/panelviews";
                        let panelviews = LayoutEditor.JsonObj.addNode(panelviewspath, view);
                      /*  gridNode.panelviews.push(view);
                        LayoutEditor.generateJson();
                        LayoutEditor.render(); */
                        LayoutEditor.generateLayout();
                        LayoutEditor.ShowPageStructure();
                      }
                      break;
                    case 'Link View':
                      LayoutEditor.load_viewConfigue(nodekey+"/view");
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
                      if(nodedata.hasOwnProperty('iscontainer')){
                        if (nodedata.iscontainer == true)
                            return false;
                      }
                      return true;
                    }
                  }, 
                  'Link View':{
                    name: 'Link View',
                    icon: 'fa-plus',
                    disabled:function(){
                     
                      let nodedata = (LayoutEditor.JsonObj.getNode(nodekey)).value;
                      if(nodedata.hasOwnProperty('iscontainer')){
                        if (nodedata.iscontainer == true)
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
                      if(nodedata.hasOwnProperty('iscontainer')){
                        if (nodedata.iscontainer == true)
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
                    icon: function(){ return 'context-menu-icon context-menu-icon-quit'; },
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
              let viewpanelid = $triggerElement.attr('panelid');
              let gridNode = (LayoutEditor.getGridNodebyPanelID(viewpanelid, LayoutEditor.grid.engine.nodes))
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');
              let viewtype = $triggerElement.attr('viewtype');
              UI.Log(node,nodekey,viewpanelid,gridNode, viewtype)
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
                          LayoutEditor.generateLayout();
                          LayoutEditor.ShowPageStructure();
                          
                        } 
                      break;
                    case 'Properties':
                      LayoutEditor.ShowViewProperties(nodekey);
                      break;
                    case 'Link View':
                      LayoutEditor.load_viewConfigue(nodekey);
                    
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
                  'Link View':{
                    name: 'Link View',
                    icon: 'fa-link',
                    disabled:false
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
                    icon: function(){ return 'context-menu-icon context-menu-icon-quit'; },
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
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(node,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Properties':
                      LayoutEditor.ShowActionProperties(nodekey);
                      break;
                    case 'Remove':

                      LayoutEditor.JsonObj.deleteNode(nodekey);
                      LayoutEditor.generateLayout();
                      LayoutEditor.ShowPageStructure();

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
                    icon: function(){ return 'context-menu-icon context-menu-icon-quit'; },
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
              let nodekey = $triggerElement.attr('data-key').replace(/'/g, '"');;
              UI.Log(node,nodekey)
              return{
                callback: function(key, options,e){
                  UI.Log(key, options,e)
                  switch(key){
                    case 'Add Panel':            
                      LayoutEditor.addPanel();
                      LayoutEditor.generateJson();
                      LayoutEditor.render();
                      LayoutEditor.ShowPageStructure();
                      break;
                    case 'Properties':
                      LayoutEditor.ShowRootProperties();
                      break;
                    case 'Save':
                      LayoutEditor.savelayout();
                      break;
                    case 'Export':
                     
                      let pagejson = LayoutEditor.JsonObj.data;
                      let blob = new Blob([JSON.stringify(pagejson)], {type: "text/plain;charset=utf-8"});
                      let file = new File([blob], 'page_'+pagejson.name+'_'+pagejson.version+".json", {type: "text/plain;charset=utf-8"});                      
                      
                      saveAs(file)

                    break;
                    case 'Redlines':
                      LayoutEditor.JsonObj.showRedlines();
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
                 'Import':{
                    name: 'Import',
                    icon: 'fa-file-import',
                    disabled:false
                  }, 
                  'Export':{
                    name: 'Export',
                    icon: 'fa-export',
                    disabled:false
                  },  
                  'Redlines':{
                    name: 'Redlines Change',
                    icon: 'fa-code-compare',
                    disabled:false
                  },
                  "sep1":'------------',
                  'Quit':{
                    name: 'Quit',
                    icon: function(){ return 'context-menu-icon context-menu-icon-quit'; },
                  }
                }

              }
            },
                    
          })
        },
        removeGridNodebyPanelID: function(panelid, nodes){
          for(var i=0;i<nodes.length;i++){
            if(nodes[i].id == panelid){
              nodes.splice(i,1);
              return;              
            }
          }
          for(var i=0;i<nodes.length;i++){
            if(nodes[i].subGrid)
              if(nodes[i].subGrid.engine)
                if(nodes[i].subGrid.engine.nodes){
                  LayoutEditor.removeGridNodebyPanelID(panelid, nodes[i].subGrid.engine.nodes)
                }
          }
        },
        getGridNodebyPath: function(path){
          let paths = path.split('/');
          UI.Log(path, paths)
          for(var i=paths.length-1;i>=0;i--){
            if(paths[i] == 'panels'){
              UI.Log(i,paths[i-1], paths[i])
              let node = JSON.parse(paths[i+1]);
              if(node.hasOwnProperty('id')){
                return LayoutEditor.getGridNodebyPanelID(node.id, LayoutEditor.grid.engine.nodes);
              }
              break;
            }
          }
          
        },
        getGridNodebyPanelID: function(panelid, nodes){
          
          for(var i=0;i<nodes.length;i++){
            if(nodes[i].id == panelid)
              return nodes[i];
          }
          for(var i=0;i<nodes.length;i++){
            if(nodes[i].subGrid)
              if(nodes[i].subGrid.engine)
                if(nodes[i].subGrid.engine.nodes){
                  let node = LayoutEditor.getGridNodebyPanelID(panelid, nodes[i].subGrid.engine.nodes)
                  if(node != null)
                    return node;
                }
          }
          return null;          
        },
        getChildPanels: function(parent, panel, parentkey){             
          let panellist =[];
          UI.Log("get the panel tree node:", panel, parent)
          
          if(panel.hasOwnProperty('iscontainer') && panel.iscontainer == true){
            if(panel.hasOwnProperty('panels')){
              let key = parentkey + "/{'id':'" + panel.id + "'}";
              let data = {
                id: panel.id,
                parent: parent,
                text: panel.name,              
                a_attr: {"node-type": "panel", "panelid":panel.id, "data-key":key} ,
                icon: "fa-solid fa-border-all", 
                state: { opened: true },
              }
              panellist.push(data); 

              for(var i=0;i<panel.panels.length;i++){
                let childpanels = LayoutEditor.getChildPanels(panel.id, panel.panels[i], key + "/panels");
                panellist = panellist.concat(childpanels);
              }
            }
          }else if(panel.hasOwnProperty('view')){
            let datakey = parentkey + "/{'id':'" + panel.id + "'}";
            let data = {
              id: panel.id,
              parent: parent,
              text: panel.name,
              type: "panel",
              a_attr: {"node-type": "panel", "panelid":panel.id,"data-key":datakey} ,
              icon: "fa-solid fa-square",
              state: { opened: true },
            }
            panellist.push(data); 
            if(panel.view.hasOwnProperty('id')){

            }
            let viewID = panel.view.id? panel.view.id: UI.generateUUID();
            let viewkey = datakey + "/view";
            data = {
              id: viewID,
              parent: panel.id,
              text: panel.view.name ? panel.view.name : "-----",
              a_attr: {"node-type": "view", "panelid":panel.id, "data-key":viewkey, viewtype: "view"} ,
              icon: "fa fa-solid fa-flag",
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
                  a_attr: {"node-type": "action", "panelid":panel.id, "data-key": viewkey + "/actions/" + key} ,
                  icon: "fa-solid fa-square-caret-right",
                  state: { opened: true },
                }
                panellist.push(data); 
              }
            }
            let panelviewnodeid = UI.generateUUID()
            let panelviewkey = datakey + "/panelviews";
            data = {
              id: panelviewnodeid,
              parent: panel.id,
              text: "Panel Views",
              a_attr: {"node-type": "panelviewcontainer","panelid":panel.id, "data-key":panelviewkey} ,
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
                  a_attr: {"node-type": "view", "panelid":panel.id, "data-key":panelviewkey + "/{'name':'" + panelview.name + "'}", viewtype: "panelview"} ,
                  icon: "fa-regular fa-flag",
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
                      a_attr: {"node-type": "action", "panelid":panel.id, "data-key":panelviewkey + "/{'name':'" + panelview.name + "'}/actions/"+key} ,
                      icon: "fa-solid fa-square-caret-right",
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