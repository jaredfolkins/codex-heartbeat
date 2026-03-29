package main

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed templates/*.md
var embeddedTemplates embed.FS

func embeddedTemplate(name string) string {
	data, err := embeddedTemplates.ReadFile("templates/" + name)
	if err != nil {
		panic(fmt.Sprintf("missing embedded template %q: %v", name, err))
	}
	return strings.TrimSpace(string(data))
}
