package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/business"
)

func ReportRenderHandlder(templateService business.TemplateService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.handlers.ReportRenderHandlder"}
		reportName := strings.TrimPrefix(ctx.Param("reportName"), "/")
		data, err := extractData(ctx)
		if err != nil {
			iLog.ErrorLog(err)
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		html, err := templateService.RenderTemplate(reportName, data)
		if err != nil {
			iLog.Error(err.Error())
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Data(http.StatusOK, "text/html; charset=utf-8", html)
	}
}
