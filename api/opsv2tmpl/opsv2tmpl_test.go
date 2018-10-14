package opsv2tmpl

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"testing"
)

func TestTpl(t *testing.T) {
	// funcMap := template.FuncMap{
	// 	"add": Addition,
	// }
	templateText := `Output: {{Mod .a .b | printf "%.f"}}
Output2: {{ if .c}}{{.c}}{{end}}
Output3: {{range .d}} {{printf "%s," .}}{{end}}
`
	tmpl, err := template.New("titleTest").Funcs(FuncMap).Parse(templateText)
	if err != nil {
		log.Fatalf("parsing: %s", err)
	}
	dd := make(map[string]interface{})
	dd["a"] = "23"
	dd["b"] = "10"
	dd["c"] = "test str"
	dd["d"] = []string{"111", "222", "333"}
	tmpl.Option("missingkey=error")
	err = tmpl.Execute(os.Stdout, dd)
	if err != nil {
		log.Fatalf("execution: %s", err)
	}
	tmp := "appd"
	fmt.Printf("URL: %s\n", tmp[:3])
}
