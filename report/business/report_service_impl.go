package business

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/data"
)

type reportService struct {
	pdf             data.ReportExporter
	png             data.ReportExporter
	templateService TemplateService
	baseUrl         string
}

func (rs *reportService) ExportReportHtml(reportId string, data interface{}) ([]byte, error) {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.ExportReportHtml"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.ExportReportHtml", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.ExportReportHtml with error: %s", err))
			return
		}
	}()

	html, err := rs.templateService.RenderTemplate(reportId, data)
	if err != nil {
		iLog.Error(fmt.Sprintf("render template with err %v", err))
		return nil, err
	}

	b64Html := base64.StdEncoding.EncodeToString(html)

	return []byte(b64Html), nil
}

func (rs *reportService) ExportReportPdf(reportId string, data interface{}) ([]byte, error) {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.ExportReportPdf"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.ExportReportPdf", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.ExportReportPdf with error: %s", err))
			return
		}
	}()

	url, err := rs.buildUrl(reportId, data)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to build url for pdf export on reportId: %s and data: %+v with err: %v", reportId, data, err))
		return nil, err
	}

	pdf, _, err := rs.pdf.Export(url)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to export the report with pdf with err: %v", err))
		return nil, err
	}

	b64Pdf := base64.StdEncoding.EncodeToString(pdf)

	return []byte(b64Pdf), nil
}

func (rs *reportService) ExportReportPng(reportId string, data interface{}) ([]byte, error) {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.ExportReportPng"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.ExportReportPng", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.ExportReportPdf with error: %s", err))
			return
		}
	}()

	url, err := rs.buildUrl(reportId, data)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to build url for png export on reportId: %s and data: %+v with err", reportId, data, err))
		return nil, err
	}

	png, _, err := rs.png.Export(url)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to export on reportId: %s and data: %+v with err", reportId, data, err))
		return nil, err
	}

	b64Png := base64.StdEncoding.EncodeToString(png)

	return []byte(b64Png), nil
}

func (rs *reportService) PrintReport(reportName string, data interface{}, printerName string) error {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.PrintReport"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.PrintReport", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.PrintReport with error: %s", err))
			return
		}
	}()

	reportData, err := rs.ExportReportPdf(reportName, data)
	if err != nil {
		iLog.ErrorLog(err)
		return nil
	}

	return PrintByteData(reportData, printerName)
}
func PrintByteData(data []byte, printerName string) error {
	startTime := time.Now()
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.PrintByteData"}
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("Report.business.PrintByteData", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to Report.business.PrintByteData with error: %s", err))
			return
		}
	}()
	//	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.Print"}

	cmd := exec.Command("lp", "-d", printerName)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.Copy(stdin, bytes.NewReader(data))
	}()

	output, err := cmd.CombinedOutput()
	if err != nil {
		iLog.Error(fmt.Sprintf("Command Output: %s fail", string(output)))
		return err
	}

	return nil
}

func (rs *reportService) buildUrl(reportName string, data interface{}) (string, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.business.buildUrl"}
	jsonData, err := json.Marshal(data)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to marshal %+v to json with error: %v", data, err))
		return "", err
	}

	b64Data := base64.StdEncoding.EncodeToString(jsonData)
	if b64Data == "e30=" {
		b64Data = ""
	}

	return fmt.Sprintf("http://%s/reports/render/%s?d=%s", rs.baseUrl, reportName, b64Data), nil
}

func NewReportService(pdf data.ReportExporter, png data.ReportExporter, templateService TemplateService, baseUrl string) *reportService {
	return &reportService{pdf: pdf, png: png, templateService: templateService, baseUrl: baseUrl}
}
