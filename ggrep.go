package main	

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Occurences struct {
	lineText   string
	lineNumber uint
}

func findAllOccurences(filePath, str string) ([]Occurences, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var ret []Occurences

	scanner := bufio.NewScanner(f)
	lineNumber := uint(1)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, str) {
			ret = append(ret, Occurences{
				lineText:   line,
				lineNumber: lineNumber,
			})
		}

		lineNumber++
	}

	return ret, nil
}

func print(format string, a ...any) {
	fmt.Printf(format, a...)
}

func fatal(format string, a ...any) {
	fmt.Printf(format, a...)
	os.Exit(-1)
}

var red = color.New(color.FgRed).SprintFunc()

func paintRed(str string) string {
	return red(str)
}

func paintSubstrRed(str, substr string) string {
	redSubstr := paintRed(substr)
	return strings.ReplaceAll(str, substr, redSubstr)
}

func main() {
	if len(os.Args) < 3 {
		fatal("Usage : ggrep <str> <file>")
	}

	str := os.Args[1]
	filePath := os.Args[2]

	results, err := findAllOccurences(filePath, str)
	if err != nil {
		fatal("failed to search in file, err :\n%s", err.Error())
	}

	for _, res := range results {
		print("line %d : %s\n", res.lineNumber, paintSubstrRed(res.lineText, str))
	}
}
