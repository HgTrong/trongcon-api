package mail

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Render executes a Go text/template with the given data map.
func Render(name, tpl string, data map[string]interface{}) (string, error) {
	tpl = strings.TrimSpace(tpl)
	if tpl == "" {
		return "", nil
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	t, err := template.New(name).Option("missingkey=zero").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}
