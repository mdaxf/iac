class UIChart extends HTMLElement{
    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.canvas = document.createElement('canvas');
        this.shadowRoot.appendChild(this.canvas);
    }

    render() {

      this.chart = new Chart(this.canvas, {
            type: this.type,
            data: this.data,
            options: this.options
      });
    }

    disconnectedCallback() {
        this.chart.destroy();
    }

    setChartData(data) {
        this.chart.data = data;
        this.chart.update();
    }

    refresh() {
        this.chart.update();
    }
}
customElements.define('ui-line-chart', class extends UIChart {
    connectedCallback() {
      this.type = 'line';
      this.data = {
        labels: ['January', 'February', 'March', 'April', 'May'],
        datasets: [{
          label: 'Line Chart',
          data: [10, 25, 32, 48, 55],
          borderColor: 'blue',
          backgroundColor: 'transparent'
        }]
      };
      
      this.options = {};
      this.render();
      
    }
});

  // Bar Chart Web Component
customElements.define('ui-bar-chart', class extends UIChart {
    connectedCallback() {
      this.type = 'bar';
      /*this.data = {
        labels: ['A', 'B', 'C', 'D', 'E'],
        datasets: [{
          label: 'Bar Chart',
          data: [15, 28, 42, 30, 50],
          backgroundColor: 'green'
        },
        {
          label: 'let Chart',
          data: [32, 38, 22, 50, 20],
          backgroundColor: 'red'
        }
      ]
      }; */
      this.data ={};
      this.options = {};
      this.render();
    }
    refresh() {
      this.chart.update();
    }
});

  // Pie Chart Web Component
customElements.define('ui-pie-chart', class extends UIChart {
    connectedCallback() {
      this.type = 'pie';
      this.data = {
        labels: ['Red', 'Blue', 'Yellow', 'Green'],
        datasets: [{
          data: [20, 25, 15, 40],
          backgroundColor: ['red', 'blue', 'yellow', 'green']
        }]
      };
      this.options = {};
      this.render();
    }
});

  // Scatter Chart Web Component
customElements.define('ui-scatter-chart', class extends HTMLElement {
    connectedCallback() {
      this.type = 'scatter';
      this.data = {
        datasets: [{
          label: 'Scatter Chart',
          data: [{ x: 10, y: 15 }, { x: 20, y: 30 }, { x: 30, y: 45 }],
          borderColor: 'purple',
          backgroundColor: 'transparent'
        }]
      };
      this.options = {
        scales: {
          x: { type: 'linear', position: 'bottom' },
          y: { type: 'linear', position: 'left' }
        }
      };
      this.render();
    }
 
});

  // Scatter Chart Web Component
  customElements.define('ui-doughnut-chart', class extends HTMLElement {
    connectedCallback() {
      this.type = 'doughnut';
      this.data = {
        labels: [
          'Red',
          'Blue',
          'Yellow'
        ],
        datasets: [{
          label: 'My First Dataset',
          data: [300, 50, 100],
          backgroundColor: [
            'rgb(255, 99, 132)',
            'rgb(54, 162, 235)',
            'rgb(255, 205, 86)'
          ],
          hoverOffset: 4
        }]
      };;
      this.render();
    }
 
});