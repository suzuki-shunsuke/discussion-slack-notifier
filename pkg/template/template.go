package template

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func Parse(s string) (*template.Template, error) {
	return template.New("_").Funcs(sprig.TxtFuncMap()).Parse(s) //nolint:wrapcheck
}

func Execute(tpl *template.Template, param interface{}) (string, error) {
	buf := &bytes.Buffer{}
	if err := tpl.Execute(buf, param); err != nil {
		return "", fmt.Errorf("render a template: %w", err)
	}
	return buf.String(), nil
}
