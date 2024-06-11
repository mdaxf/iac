package cmd

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/report/business"
	"github.com/mdaxf/iac/report/cmd/handlers"
	"github.com/mdaxf/iac/report/data"

	"github.com/mdaxf/iac/logger"
)

type Server struct {
	config          Config
	templateService business.TemplateService
	reportService   business.ReportService
	templateEngine  data.TemplateEngine
}

func NewServer(config Config, templateService business.TemplateService, engine data.TemplateEngine, reportService business.ReportService) *Server {
	return &Server{config: config, templateService: templateService, reportService: reportService, templateEngine: engine}
}

func (s *Server) setupServer(r *gin.Engine) *gin.Engine {
	//r := gin.Default()
	r.GET("/reports/health", handlers.Health())
	r.GET("/reports/list", handlers.ReportListHandler(s.templateService))
	r.GET("/reports/render/*reportName", handlers.ReportRenderHandlder(s.templateService))
	r.GET("/reports/preview/*reportName", handlers.ReportPreviewHandler(s.templateService, s.templateEngine, s.reportService))
	r.POST("/reports/export/html/*reportName", handlers.ReportExportHandler(s.reportService, "html"))
	r.POST("/reports/export/png/*reportName", handlers.ReportExportHandler(s.reportService, "png"))
	r.POST("/reports/export/pdf/*reportName", handlers.ReportExportHandler(s.reportService, "pdf"))
	r.POST("/reports/export/print/*printerName/*reportName", handlers.ReportExportHandler(s.reportService, "print"))
	return r
}

func SetupReportServer(r *gin.Engine) *gin.Engine {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.Server"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.cmd.Server", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.cmd.Server with error: %s", err))
			return
		}
	}()
	config, err := ParseConfig()
	if err != nil {
		iLog.ErrorLog(err)
		return nil
	}
	module := NewModule(config)

	return module.Server.setupServer(r)

}
