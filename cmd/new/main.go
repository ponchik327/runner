package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates
var templateFS embed.FS

// templateData — данные, передаваемые в каждый шаблон при генерации.
type templateData struct {
	Module string // имя Go-модуля воркспейса, напр. "contest"
	P      int    // номер задачи
}

func main() {
	problems := flag.Int("p", 1, "количество задач")
	dir := flag.String("dir", "contest", "имя создаваемой директории")
	mod := flag.String("mod", "contest", "имя Go-модуля в go.mod воркспейса")
	flag.Parse()

	if err := os.MkdirAll(*dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// render читает шаблон из embed.FS, применяет data и записывает результат в dst.
	render := func(tmplName, dst string, data any) {
		content, err := templateFS.ReadFile("templates/" + tmplName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		tmpl := template.Must(template.New(tmplName).Parse(string(content)))
		f, err := os.Create(dst)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		if err := tmpl.Execute(f, data); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	root := templateData{Module: *mod}

	render("go.mod.tmpl", filepath.Join(*dir, "go.mod"), root)
	render("Makefile.tmpl", filepath.Join(*dir, "Makefile"), root)

	for p := 1; p <= *problems; p++ {
		pd := templateData{Module: *mod, P: p}
		pdir := filepath.Join(*dir, fmt.Sprintf("p%d", p))
		os.MkdirAll(filepath.Join(pdir, "testdata"), 0o755)

		render("main.go.tmpl", filepath.Join(pdir, "main.go"), pd)
		render("main_test.go.tmpl", filepath.Join(pdir, "main_test.go"), pd)
		render("stress_test.go.tmpl", filepath.Join(pdir, "stress_test.go"), pd)

		// .gitkeep нужен чтобы пустая директория попала в git
		os.WriteFile(filepath.Join(pdir, "testdata", ".gitkeep"), nil, 0o644)
	}

	fmt.Printf("Воркспейс создан в ./%s/\n", *dir)
	fmt.Printf("Следующие шаги:\n  cd %s\n  go mod tidy\n  make test P=1\n", *dir)
}