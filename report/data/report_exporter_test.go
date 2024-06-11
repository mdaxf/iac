package data

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestPngReportExporter_Export_PrintableNotFound(t *testing.T) {
	ts := NewPngReportExporter(1*time.Second, 2000, 1920)
	png, _, err := ts.Export(fmt.Sprintf("http://localhost:%d/subfolder/report2.html", testPort))
	assert.NotNil(t, err)
	assert.Nil(t, png)
	assert.EqualValues(t, "context deadline exceeded", err.Error())
}

func TestPngReportExporter_Export_PrintableFound(t *testing.T) {
	ts := NewPngReportExporter(2*time.Second, 2000, 1920)
	png, _, err := ts.Export(fmt.Sprintf("http://localhost:%d/report1.html", testPort))
	assert.Nil(t, err)
	assert.NotEmpty(t, png)

	expected, _ := ioutil.ReadFile("testdata/report1.png")
	assert.EqualValues(t, expected, png)
}

func TestPdfReportExporter_Export_PrintableNotFound(t *testing.T) {
	ts := NewPdfReportExporter(NewPngReportExporter(1*time.Second, 2000, 1920))
	png, _, err := ts.Export(fmt.Sprintf("http://localhost:%d/subfolder/report2.html", testPort))
	assert.NotNil(t, err)
	assert.Nil(t, png)
	assert.Contains(t, err.Error(), "Caused by: context deadline exceeded")
}

func TestPdfReportExporter_Export_PrintableFound(t *testing.T) {
	ts := NewPdfReportExporter(NewPngReportExporter(1*time.Second, 2000, 1920))
	pdf, _, err := ts.Export(fmt.Sprintf("http://localhost:%d/report1.html", testPort))
	assert.Nil(t, err)
	assert.NotEmpty(t, pdf)

	expected, _ := ioutil.ReadFile("testdata/report1.pdf")
	assert.EqualValues(t, expected, pdf)
}

const testPort = 9878

func TestMain(m *testing.M) {
	testServeFiles()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/report1.html", testPort))
	panicIf(err, err != nil)
	panicIf("failed to get 200 ok", resp.StatusCode != 200)

	body, err := ioutil.ReadAll(resp.Body)
	panicIf(err, err != nil)
	panicIf("nil body", body == nil)

	expected, err := ioutil.ReadFile("testdata/report1.html")
	panicIf(err, err != nil)
	panicIf("report server not equal to the one in file system", !assert.ObjectsAreEqual(expected, body))

	os.Exit(m.Run())
}

func testServeFiles() {
	fs := http.FileServer(http.Dir("testdata"))
	http.Handle("/", fs)
	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", testPort), nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}

func panicIf(value interface{}, condition bool) {
	if condition {
		panic(value)
	}
}
