package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

type OccurenceInLine struct {
	lineNum  int
	len      int
	indexes  []int
	lineText string
}

// func FindAllIndexesStr(org, substr string) ([]int, error) {
// 	if substr == "" {
// 		return nil, errors.New("substr must not be empty")
// 	}

// 	return findAllIndexesStrRec(org, substr, 0), nil
// }

// func findAllIndexesStrRec(org, substr string, baseIdx int) []int {
// 	if org == "" {
// 		return nil
// 	}

// 	idx := strings.Index(org, substr)
// 	if idx == -1 {
// 		return nil
// 	}
// 	idx += baseIdx

// 	if len(org) < len(substr)+idx+1 { // no more room to advance
// 		return []int{idx}
// 	}

// 	nextIdx := len(substr) + idx
// 	res := findAllIndexesStrRec(org[nextIdx:], substr, nextIdx)
// 	if res == nil {
// 		return []int{idx}
// 	}

// 	return append([]int{idx}, res...)
// }

func FindAllIndexes(str, substr string) []int {
	if substr == "" {
		return nil
	}

	i := 0
	indexes := []int{}

	for {
		idx := strings.Index(str[i:], substr)
		if idx == -1 {
			break
		}

		indexes = append(indexes, idx+i)
		i += idx + 1
	}

	if len(indexes) == 0 {
		return nil
	}

	return indexes
}

func FindAllOccurences(filePath, str string) ([]OccurenceInLine, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	var occurences []OccurenceInLine

	scanner := bufio.NewScanner(f)
	lineNumber := 0
	for scanner.Scan() {

		line := scanner.Text()
		lineNumber++
		indexes := FindAllIndexes(line, str)
		if len(indexes) == 0 {
			continue
		}

		res := OccurenceInLine{
			lineNum:  lineNumber,
			len:      len(str),
			indexes:  indexes,
			lineText: line,
		}

		occurences = append(occurences, res)
	}

	return occurences, nil
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

	results, err := FindAllOccurences(filePath, str)
	if err != nil {
		fatal("failed to search in file, err :\n%s", err.Error())
	}

	for _, res := range results {
		line, err := ReadLine(filePath, int(res.lineNum))
		if err != nil {
			fatal("%s", err.Error())
		}
		print("line %d : %s\n", res.lineNum, paintSubstrRed(line, str))
	}
}
