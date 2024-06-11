package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/report/business"

	"net/http"
	"strings"

	"github.com/mdaxf/iac/logger"
)

func ReportExportHandler(reportService business.ReportService, kind string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.handlers"}

		reportName := strings.TrimPrefix(ctx.Param("reportName"), "/")
		var body interface{}
		if err := ctx.BindJSON(&body); err != nil {
			iLog.Error(fmt.Sprintf("%v", err))
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		var export []byte
		var err error
		switch kind {
		case "html":
			export, err = reportService.ExportReportHtml(reportName, body)
		case "pdf":
			export, err = reportService.ExportReportPdf(reportName, body)
		case "png":
			export, err = reportService.ExportReportPng(reportName, body)
		case "print":
			printerName := strings.TrimPrefix(ctx.Param("printerName"), "/")
			if printerName == "" {
				iLog.Error("There is no printer supplied.")
				ctx.String(http.StatusInternalServerError, "There is no printer supplied.")
				return
			}
			err = reportService.PrintReport(reportName, body, printerName)
		}

		if err != nil {
			iLog.ErrorLog(err)
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, map[string]interface{}{
			"reportName": reportName,
			"data":       string(export),
			"type":       kind,
		})
	}
}
