package data

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestFilesystemTemplateRepo_ListAll(t *testing.T) {
	ts := NewFilesystemTemplateRepo("testdata")

	templates, err := ts.ListAll()

	assert.Nil(t, err)
	assert.NotEmpty(t, templates)
	assert.EqualValues(t, 2, len(templates))
	assert.Contains(t, templates, "report1")
	assert.Contains(t, templates, "subfolder/report2")
}

func TestFilesystemTemplateRepo_LoadTemplate(t *testing.T) {
	ts := NewFilesystemTemplateRepo("testdata")

	report1, err := ts.LoadTemplate("report1")
	assert.Nil(t, err)
	assert.NotEmpty(t, report1)
	report1Exp, _ := ioutil.ReadFile("./testdata/report1.html")
	assert.EqualValues(t, report1Exp, report1)

	report2, err := ts.LoadTemplate("subfolder/report2")
	assert.Nil(t, err)
	assert.NotEmpty(t, report2)
	report2Exp, _ := ioutil.ReadFile("./testdata/subfolder/report2.html")
	assert.EqualValues(t, report2Exp, report2)
}
