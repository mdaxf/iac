package models

type TemplateContext struct {
	Values interface{}
}

func (ctx *TemplateContext) Sequence(size int) []int {
	var out []int
	for i := 0; i < size; i++ {
		out = append(out, i+1)
	}
	return out
}
