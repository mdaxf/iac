package data

import (
	"testing"

	"github.com/mdaxf/iac/report/models"
	"github.com/stretchr/testify/assert"
)

func TestTemplateContext_Sequence(t *testing.T) {
	template := `{{- range $i, $item := .Sequence 3 }}
a
{{- end }}`

	te := NewGolangTemplateEngine()
	rendered, err := te.Render([]byte(template), &models.TemplateContext{})

	assert.Nil(t, err)
	assert.NotNil(t, rendered)
	assert.EqualValues(t, "\na\na\na", string(rendered))
}

func TestTemplateContext_Sequence2(t *testing.T) {
	template := `{{- range $i, $item := .Sequence 10 }}
{{- if lt $i (len $.Values.array) }}
{{ index $.Values.array $i }}
{{- else }}
0
{{- end }}

{{- end }}`

	te := NewGolangTemplateEngine()
	rendered, err := te.Render([]byte(template), &models.TemplateContext{Values: map[string]interface{}{
		"array": []int{4, 5, 6},
	}})

	assert.Nil(t, err)
	assert.NotNil(t, rendered)
	assert.EqualValues(t, "\n4\n5\n6\n0\n0\n0\n0\n0\n0\n0", string(rendered))
}
