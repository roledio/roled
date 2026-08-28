package mail

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed templates
var templateFS embed.FS

func LoadTemplate(path string) (*template.Template, error) {
	return template.ParseFS(templateFS, path)
}

func ParseTemplate(tpl *template.Template, data any) (string, error) {
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func LoadAndParseTemplate(path string, data any) (string, error) {
	tpl, err := LoadTemplate(path)
	if err != nil {
		return "", err
	}
	return ParseTemplate(tpl, data)
}
