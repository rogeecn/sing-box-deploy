package templates

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rogeecn/sing-box-deploy/internal/spec"
	"github.com/rogeecn/sing-box-deploy/tmpl"
)

type Data struct {
	Domain      string
	Email       string
	Inbounds    map[string]spec.InboundSpec
	TLSKeyPath  string
	TLSCertPath string
}

var (
	inboundTemplates = map[string]*template.Template{}
	caddyTemplate    *template.Template
)

func init() {
	if err := loadInboundTemplates(); err != nil {
		panic(err)
	}
	if err := loadCaddyTemplate(); err != nil {
		panic(err)
	}
}

func loadInboundTemplates() error {
	files, err := fs.Glob(tmpl.Files, "sing-box/inbounds/*.tmpl")
	if err != nil {
		return err
	}
	for _, file := range files {
		name := filepath.Base(file)
		key := strings.TrimSuffix(strings.TrimSuffix(name, filepath.Ext(name)), ".json")
		tpl, err := template.New(name).ParseFS(tmpl.Files, file)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", file, err)
		}
		inboundTemplates[key] = tpl
	}
	return nil
}

func loadCaddyTemplate() error {
	tpl, err := template.ParseFS(tmpl.Files, "caddy/site.caddy.tmpl")
	if err != nil {
		return err
	}
	caddyTemplate = tpl
	return nil
}

// RenderInbounds renders template files for the provided data set.
func RenderInbounds(data Data) (map[string][]byte, error) {
	outputs := make(map[string][]byte, len(data.Inbounds))
	for key := range data.Inbounds {
		tpl, ok := inboundTemplates[key]
		if !ok {
			return nil, fmt.Errorf("no template for key %s", key)
		}
		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return nil, fmt.Errorf("execute template %s: %w", key, err)
		}
		outputs[key] = bytes.TrimSpace(buf.Bytes())
	}
	return outputs, nil
}

// RenderCaddy renders the Caddyfile template.
func RenderCaddy(data Data) ([]byte, error) {
	var buf bytes.Buffer
	if err := caddyTemplate.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
