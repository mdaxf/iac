class UIPlotlyChart extends HTMLElement{
    constructor() {
        super();
      //  this.attachShadow({ mode: 'open' });

        //let link = document.createElement('style');
        //link.setAttribute('id', 'plotly.js-style-global');       
        //this.appendChild(link);

        
    }

    render() {
        this.chart = Plotly.newPlot(this.element, this.data, this.layout,{displayModeBar: false});
    }

    disconnectedCallback() {
    //    this.chart.destroy();

    }

    setChartData(data) {
        this.chart.data = data;
        this.chart.update();
    }

}

customElements.define('ui-plotly-guage', class extends UIPlotlyChart {
    connectedCallback() {
      this.element = document.createElement('div');
      this.appendChild(this.element);

      this.type = 'indicator';
      this.shape = this.getAttribute('shape');
      this.width = this.getAttribute('width');
      this.height = this.getAttribute('height');
    
      this.minValue = this.getAttribute('min-value');
      this.maxValue = this.getAttribute('max-value');
      this.referenceValue = this.getAttribute('reference-value');
      this.value = this.getAttribute('value');
      this.title = this.getAttribute('title');
      this.unit = this.getAttribute('unit');
      this.mode = this.getAttribute('mode');
    
   
      let guage ={ axis: { visible: false, range: [this.minValue? this.minValue:0, this.maxValue?this.maxValue:100] } }
      if (this.shape && this.shape !="" && this.shape !="guage")
        guage.shape = this.shape
        
    //  console.log(this.mode,this.title,this.unit,this.value,this.referenceValue,this.minValue,this.maxValue);

      let data = {
        type: this.type,
        value: this.value? this.value: 0,
        delta: { reference: this.referenceValue? this.referenceValue: 80 },
        domain: { row: 0, column: 0 }
      }

      if (this.shape && this.shape !="" )
        data.guage = guage

      this.data = [data];
      let indicators =[]
      let indicator = {}
      if(this.title && this.title != null && this.title !=undefined &&  this.title !=""){
        indicator.title = {text: this.title}
        if(this.titlefontSize && this.titlefontSize !="")
            indicator.title["fontSize"] = this.titlefontSize
      }
      indicator.mode = this.mode? this.mode: "number+delta+gauge";

      if (this.referenceValue && this.referenceValue !="")
        indicator.delta = { reference: this.referenceValue? this.referenceValue: 80  }
        
      indicators.push(indicator)
      this.layout = {
        width: this.width? this.width: 150,
        height: this.height? this.height: 100,
        margin:  { t: 25, b: 25, l: 25, r: 25 },
        grid: { rows: 0, columns: 0, pattern: "independent" },
        template: {
            data: {
                indicator: indicators
            }
        }        
      };
      UI.Log(this.data, this.layout)
      this.render();
      
    }
});