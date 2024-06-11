package business

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/data"
	"github.com/mdaxf/iac/report/models"
)

type templateService struct {
	engine data.TemplateEngine
	repo   data.TemplateRepository
}

func (ts *templateService) ListTemplates() ([]map[string]interface{}, error) {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.ListTemplates"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.ListTemplates", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.ListTemplates with error: %s", err))
			return
		}
	}()

	templates, err := ts.repo.ListAll()
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to list all templates with error: %v", err))
		return nil, err
	}
	return templates, nil
}

func (ts *templateService) GetTemplate(reportName string) ([]byte, error) {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.GetTemplate"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.GetTemplate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.GetTemplate with error: %s", err))
			return
		}
	}()

	content, err := ts.repo.LoadTemplate(reportName)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to get single template with id: %s with error:%v", reportName, err))
		return nil, err
	}

	return content, nil
}

func (ts *templateService) RenderTemplate(reportName string, data interface{}) ([]byte, error) {

	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.RenderTemplate"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.RenderTemplate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.RenderTemplate with error: %s", err))
			return
		}
	}()

	ctx := &models.TemplateContext{
		Values: data,
	}

	tmpl, err := ts.repo.LoadTemplate(reportName)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to get single template with id: %s with error: %v", reportName, err))
		return nil, err
	}

	html, err := ts.engine.Render(tmpl, ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to render template with id: %s and data: %+v with error: %v", reportName, ctx, err))
		return nil, err
	}

	return html, nil
}

func NewTemplateService(engine data.TemplateEngine, repo data.TemplateRepository) *templateService {
	return &templateService{engine: engine, repo: repo}
}
