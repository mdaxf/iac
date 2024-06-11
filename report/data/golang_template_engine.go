package data

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/models"
)

type golangTemplateEngine struct{}

func NewGolangTemplateEngine() *golangTemplateEngine {
	return &golangTemplateEngine{}
}

func (gte *golangTemplateEngine) Render(templateContent []byte, ctx *models.TemplateContext) ([]byte, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.data.Render"}
	eng := template.New("report_template")
	t, err := eng.Parse(string(templateContent))
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to parse template file: %s with data: %+v with error: %v", templateContent, ctx, err))
		return nil, err
	}

	var out bytes.Buffer
	if err := t.Execute(&out, ctx); err != nil {
		iLog.Error(fmt.Sprintf("failed to execute template file: %s with data: %+v with error: %v", templateContent, ctx, err))
		return nil, err
	}

	return out.Bytes(), err
}
