package main

import (
	"fmt"
	"os"

	"github.com/TGScheme/TLExtractorBot/internal/telegram/scheme"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: tlvalidate <scheme.tl> [scheme.tl ...]")
		os.Exit(2)
	}
	var files []scheme.TLFile
	for _, path := range os.Args[1:] {
		text, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		files = append(files, scheme.TLFile{Name: path, Text: string(text)})
	}
	problems := scheme.ValidateFiles(files)
	fatal := scheme.FatalProblems(problems)
	for _, problem := range problems {
		level := "warning"
		if problem.Fatal {
			level = "error"
		}
		fmt.Printf("%s:%d: %s: %s\n", problem.File, problem.Line, level, problem.Reason)
	}
	fmt.Printf("%d problems, %d of them fatal\n", len(problems), len(fatal))
	if len(fatal) > 0 {
		os.Exit(1)
	}
}
