package data

import "github.com/mdaxf/iac/report/models"

type TemplateEngine interface {
	Render(templateContent []byte, ctx *models.TemplateContext) ([]byte, error)
}

type TemplateRepository interface {
	ListAll() ([]map[string]interface{}, error)
	LoadTemplate(reportName string) ([]byte, error)
}

type ReportExporter interface {
	Export(url string) ([]byte, *models.PrintOptions, error)
}
