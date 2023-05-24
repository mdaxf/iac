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
        cellHeight: 50,
        cellWidth: 50,
        initialize: function (){
             console.log(CustomGridStack)
            let width = window.innerWidth || document.documentElement.clientWidth || document.body.clientWidth;
            let height = window.innerHeight || document.documentElement.clientHeight || document.body.clientHeight;
            width = width - 50;
            height = height - 100;
            LayoutEditor.subOptions = {
              cellHeight: LayoutEditor.cellHeight, // should be 50 - top/bottom
              cellWidth: LayoutEditor.cellWidth, // should be 50 - left/right
              column: 'auto', // size to match container. make sure to include gridstack-extra.min.css
              acceptWidgets: true, // will accept .grid-stack-item by default
              margin: 1,
              subGridDynamic: true, 
            };
            LayoutEditor.Options = { // main grid options
                cellHeight: LayoutEditor.cellHeight, // should be 50 - top/bottom
                cellWidth: LayoutEditor.cellWidth, // should be 50 - left/right
                margin: 1,
                minRow: 2, // don't collapse when empty
                disableOneColumnMode: true,
                acceptWidgets: true,
                subGridOpts: LayoutEditor.subOptions,
                subGridDynamic: true,
                id: 'main_layout_panel',
                children: []
            };
            
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
        addPanelContainer: function (){
          LayoutEditor.Options = LayoutEditor.grid.save(true, false);
          let node={
            x:0,
            y:0,
            w:3,
            h:3,
            content: 'The panel can include other panels',
            id: 'sub_grid'+ (LayoutEditor.Options.length+1),
            name: 'sub_grid'+ (LayoutEditor.Options.length+1),
            width: 3,
            height: 3,
            subGridOpts: {children: [], id:'sub_grid'+ (LayoutEditor.Options.length+1), class: 'sub_grid', ...LayoutEditor.subOptions}
          }
          
          LayoutEditor.Options.push(node); 
          LayoutEditor.grid.removeAll();
          LayoutEditor.load(LayoutEditor.Options,false);    
        /*  let subgrid = LayoutEditor.grid.makeSubGrid(node);
          LayoutEditor.addSelectEvent(subgrid)      */  
        },

        addPanel: function (){
            let count = $('.grid-stack-item').length + 1;
            let content = '<div class="layout_panel_operations" style="display:inline-block"><div>panel'+count+'" ></div></div>'            
             let cell = LayoutEditor.grid.addWidget({x:0, y:100, content:content, w:3, h:3, id:'panel'+count, name:'panel'+count,width:3,height:3,view:'',class:'layout_panel'});
            LayoutEditor.addSelectEvent(cell)
        },
        render:function() {
          LayoutEditor.Options = LayoutEditor.grid.save(true, false);
          LayoutEditor.grid.removeAll();
          LayoutEditor.load(LayoutEditor.Options,false); 
          LayoutEditor.attachContextEvents();
        },
        save:function(content = true, full = true) {
            options = LayoutEditor.grid.save(content, full);
            console.log(options);
        // console.log(JSON.stringify(options));
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

                      case 'Properties':
                        LayoutEditor.ShowProperties($triggerElement);
                        break;
                      
                    }

                  }, 
                  items:{
                    'Properties':{
                      name: 'Properties',
                      icon: 'fa-cog',
                      disabled: false
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
        ShowProperties: function(selectedElement){
            console.log('ShowProperties',selectedElement) 
            let subgrid = selectedElement.hasClass('grid-stack-sub-grid')   
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
            new UI.FormControl(container, 'h1',{innerHTML: 'Properties'});
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
            attrs={
                for: 'width',
                innerHTML: 'Width'
            }
            new UI.FormControl(container, 'label',attrs);
            attrs={
                id: 'width',
                type: 'text',
                value: grid.w * LayoutEditor.cellWidth || '',

                style: 'width: 100%;'
            }
            new UI.FormControl(container, 'input',attrs);
            attrs={
                for: 'height',
                innerHTML: 'Height',
                style: 'width: 100%;'
            }
            new UI.FormControl(container, 'label',attrs);
            attrs={
                id: 'height',
                type: 'text',
                value: grid.h * LayoutEditor.cellHeight || '',
                style: 'width: 100%;'
            }
            new UI.FormControl(container, 'input',attrs);
            attrs={
                for: 'style',
                innerHTML: 'inline Style'
            }
            new UI.FormControl(container, 'label',attrs);
            attrs={
                id: 'style',
                type: 'text',
                value: grid.inlinestyle || '',
                style: 'width: 100%;'
            }
            new UI.FormControl(container, 'textarea',attrs);

            if(!subgrid){
              attrs={
                innerHTML: 'Link View',
                
              }
              new UI.FormControl(container, 'h3',attrs);
              attrs={
                  id: 'view_name',
                  value:  (grid.view ?  grid.view.name : ''),
              }
              new UI.FormControl(container, 'input',attrs);
              attrs={
                  id: 'view_config',
                  value: (grid.view ? grid.view.config : ''),
              }
              new UI.FormControl(container, 'input',attrs);
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
                let height = $('#height').val();
                let style = $('#style').val();
                let oritention = $('#oritention').val();
                console.log(name, width, height, style,oritention)
                grid.name = name;
               //grid.width = width;
                grid.w = width / LayoutEditor.cellWidth;
               // grid.height = height;
                grid.h = height / LayoutEditor.cellHeight;
                grid.inlinestyle = style;
                grid.oritention = oritention;
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
        }
        
}