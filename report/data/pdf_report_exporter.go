package data

import (
	"fmt"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/models"
	"github.com/signintech/gopdf"
)

type pdfReportExporter struct {
	png *pngReportExporter
}

func NewPdfReportExporter(png *pngReportExporter) *pdfReportExporter {
	return &pdfReportExporter{png: png}
}

func (pre *pdfReportExporter) Export(url string) ([]byte, *models.PrintOptions, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.data.Export"}
	png, opts, err := pre.png.Export(url)
	if err != nil {
		iLog.Error(fmt.Sprintf("failed to export pdf, image screenshot has a failure for url: %s with error: %v", url, err))
		return nil, nil, err
	}

	pdf := gopdf.GoPdf{}
	defer pdf.Close()

	rect := gopdf.Rect{
		W: opts.PageWidth,
		H: opts.PageHeight,
	}
	rect.UnitsToPoints(gopdf.UnitPT)

	pdf.Start(gopdf.Config{PageSize: rect})

	pdf.AddPage()
	image, _ := gopdf.ImageHolderByBytes(png)
	if err := pdf.ImageByHolder(image, 0, 0, nil); err != nil {
		iLog.Error(fmt.Sprintf("failed to embed image into pdf for url : %s with error: %v", url, err))
		return nil, nil, err
	}

	return pdf.GetBytesPdf(), opts, nil
}
