package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/business"
)

func ReportListHandler(tmplSrv business.TemplateService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.handlers"}

		templates, err := tmplSrv.ListTemplates()
		if err != nil {
			iLog.ErrorLog(err)
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"list": templates,
		})
	}
}
