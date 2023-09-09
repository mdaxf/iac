/*
 a web component to display a tree using jstree
*/
customElements.define('ui-tree', class extends HTMLElement {
    constructor() {
        super();

        // Create a shadow DOM
        this.shadow = this.attachShadow({ mode: 'open' });
        let link = document.createElement('link');
        link.setAttribute('rel', 'stylesheet');
        link.setAttribute('href', 'styles/jstree/jstree.css');
        this.shadowRoot.appendChild(link);
        this.container =null;
    }

    refresh() {
        this.treeInstance.refresh();
    }

    initialize() {
        if(this.container != null)
            this.shadowRoot.removeChild(this.container);
        this.container = document.createElement('div');
        this.container.setAttribute('id', 'jstree');
        this.shadowRoot.appendChild(this.container);
        this.data = [{
            text: "Root",
            id: "root",
            parent: "#",
            state: { opened: true },
            icon: "fa fa-newspaper",          
          }];
       $(function() {
            $("#jstree").jstree({
            'core': {
              'data': this.data
            }
            });		
          });  
    //    this.treeInstance = $(this.container).jstree(true);  
    }
    connectedCallback() {
        this.initialize();
    }

    disconnectedCallback() {
    }
    
    setData(data) {
        this.data = data;
        $(function() {
            $("#jstree").jstree({
            'core': {
              'data': this.data
            }
            });		
          });
        this.treeInstance = $(this.container).jstree(true);
    //    this.treeInstance.refresh();
    }
    addNode(parent, node) {
        this.treeInstance.create_node(parent, node, "last");
        this.treeInstance.refresh();
    }

    deleteNode(node) {
        this.treeInstance.delete_node(node);
        this.treeInstance.refresh();
    }

    getNode(node) {
        return this.treeInstance.get_node(node);
    }
});