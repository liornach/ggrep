package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
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

type filter struct {
	textOnly bool
}

func isBinaryFile(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, fmt.Errorf("could not open %s, error : %v", filename, err)
	}

	firstBytes := make([]byte, 512)
	readBytes, err := file.Read(firstBytes)
	if err != nil {
		return false, fmt.Errorf("could not read first bytes, error : %v", err)
	}

	contentType := http.DetectContentType(firstBytes[:readBytes])
	if !strings.Contains(contentType, "text") && !strings.Contains(contentType, "json") {
		return false, nil
	}

	return true, nil
}

func (f filter) shouldFilter(filename string) (bool, error) {
	if f.textOnly {
		res, err := isBinaryFile(filename)
		if err != nil {
			return false, err
		}
		if res == true {
			return true, nil
		}
	}

	return false, nil

}

func FindAllOccurencesRec(dirpath, str string, f filter) ([]OccurenceInLine, error) {
	res := []OccurenceInLine{}

	// type WalkDirFunc func(path string, d DirEntry, err error) error
	err := filepath.WalkDir(dirpath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		filter, err := f.shouldFilter(path)
		if err != nil {
			return err
		}

		if filter {
			return nil
		}

		thisFileRes, err := FindAllOccurences(path, str)
		if err != nil {
			return err
		}

		res = append(res, thisFileRes...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func print(format string, a ...any) {
	fmt.Printf(format, a...)
}

func fatal(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
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

type args struct {
	recursive          bool
	searchPath         string
	strToSearch        string
	includeBinaryFiles bool
}

func parseArgs() (args, error) {
	recursive := flag.Bool("r", false, "search the string recursivly")
	includeBinaryFiles := flag.Bool("include-binary", false, "whether to include binary files in the search")
	flag.Parse()
	remainingArgs := flag.Args()
	if len(remainingArgs) < 1 {
		return args{}, fmt.Errorf("must specify string to search")
	}

	if len(remainingArgs) < 2 {
		return args{}, fmt.Errorf("search path must be specified")
	}

	strToSearch := remainingArgs[0]
	searchPath := remainingArgs[1]

	return args{
		recursive:          *recursive,
		searchPath:         searchPath,
		strToSearch:        strToSearch,
		includeBinaryFiles: *includeBinaryFiles,
	}, nil
}

func CreateFilterFromArgs(a args) (filter, error) {
	return filter{
		textOnly: !a.includeBinaryFiles,
	}, nil
}

func main() {
	args, err := parseArgs()
	if err != nil {
		fatal(err.Error())
	}

	str := args.strToSearch
	searchPath := args.searchPath

	var result []OccurenceInLine
	if !args.recursive {
		result, err = FindAllOccurences(searchPath, str)
		if err != nil {
			fatal("failed to search in file, err :\n%s", err.Error())
		}
	} else {
		f, err := CreateFilterFromArgs(args)
		if err != nil {
			fatal(err.Error())
		}
		result, err = FindAllOccurencesRec(searchPath, str, f)
		if err != nil {
			fatal("failed to search in path, err :\n%s", err.Error())
		}
	}

	for _, res := range result {
		line, err := ReadLine(searchPath, int(res.lineNum))
		if err != nil {
			fatal("%s", err.Error())
		}
		print("line %d : %s\n", res.lineNum, paintSubstrRed(line, str))
	}
}
