package business

type TemplateService interface {
	ListTemplates() ([]map[string]interface{}, error)
	GetTemplate(reportId string) ([]byte, error)
	RenderTemplate(reportId string, data interface{}) ([]byte, error)
}

type ReportService interface {
	ExportReportHtml(reportName string, data interface{}) ([]byte, error)
	ExportReportPdf(reportName string, data interface{}) ([]byte, error)
	ExportReportPng(reportName string, data interface{}) ([]byte, error)
	PrintReport(reportName string, data interface{}, printerName string) error
}
