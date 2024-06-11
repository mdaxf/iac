package data

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/report/models"
)

type pngReportExporter struct {
	timeout        time.Duration
	viewportHeight int
	viewportWidth  int
}

func NewPngReportExporter(timeout time.Duration, vpHeight, vpWidth int) *pngReportExporter {
	return &pngReportExporter{timeout: timeout, viewportHeight: vpHeight, viewportWidth: vpWidth}
}

func (pre *pngReportExporter) Export(url string) ([]byte, *models.PrintOptions, error) {
	iLog := logger.Log{ModuleName: "Report", User: "System", ControllerName: "Report.cmd.data.NewPngReportExporter"}

	ctx, cancel := createContext(pre.timeout, pre.viewportHeight, pre.viewportWidth)
	defer cancel()

	var res []byte
	var options models.PrintOptions
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Evaluate(extractPrintOptions, &options),
		chromedp.WaitVisible(`#printable`, chromedp.ByID),
		chromedp.Screenshot("#printable", &res, chromedp.NodeVisible, chromedp.ByID),
	)
	if err != nil {
		iLog.ErrorLog(err)
		return nil, nil, err
	}

	return res, &options, nil
}

func createContext(timeout time.Duration, vph int, vpw int) (context.Context, context.CancelFunc) {
	baseCtx, cancelTimeout := context.WithTimeout(context.Background(), timeout)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.WindowSize(vpw, vph),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(baseCtx, opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	cancelFuncs := func() {
		cancelTimeout()
		cancelAlloc()
		cancelCtx()
	}

	return ctx, cancelFuncs
}

const extractPrintOptions = `function extractStyles() {
	var styles = getComputedStyle(document.body);

	return {
		"page_height": document.body.offsetHeight,
		"page_width": document.body.offsetWidth,
		"orientation": styles['orientation'] || "portrait",
	}
}
extractStyles();`
