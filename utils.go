package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

func OpenForWrite(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
}

func OpenForRead(filePath string) (*os.File, error) {
	return os.Open(filePath)
}

func CountLines(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}

	return count, scanner.Err()
}

func FillFile(filePath string, c rune, n, perLine int) error {
	f, err := OpenForWrite(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	chunkSize := perLine + 1
	chunk := slices.Repeat([]byte{byte(c)}, chunkSize)
	chunk[chunkSize-1] = '\n'

	for n > 0 {
		toWrite := min(n, chunkSize)

		if _, err := f.Write(chunk[:toWrite]); err != nil {
			return err
		}

		n -= perLine
	}

	return nil
}

func ExtendString(str string, n int, c rune) (string, error) {
	if n < 0 {
		return "", fmt.Errorf("n must be greated than zero")
	}

	newLen := len(str) + n
	for len(str) < newLen {
		str += string(c)
	}

	return str, nil
}

func GeneraetHugeFileGb(gb int, c rune, perLine int, filePath string) error {
	return FillFile(filePath, c, GbToBytes(gb), perLine)
}

func GeneraetHugeFileMb(gb int, c rune, perLine int, filePath string) error {
	return FillFile(filePath, c, MbToBytes(gb), perLine)
}

func GbToBytes(gb int) int {
	return MbToBytes(gb * 1024)
}

func MbToBytes(mb int) int {
	return mb * 1024 * 1024
}

func AppendEmptyLines(dst io.Writer, lines int) error {
	newLines := slices.Repeat([]byte{'\n'}, lines)
	_, err := dst.Write(newLines)
	if err != nil {
		return fmt.Errorf("failed to write to file:\n%w", err)
	}
	return nil
}

func AppendEmptyLinesPaths(filePath string, lines int) error {
	f, err := OpenForWrite(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file:\n%w", err)
	}
	defer f.Close()

	newLines := slices.Repeat([]byte{'\n'}, lines)
	_, err = f.WriteString(string(newLines))
	if err != nil {
		return fmt.Errorf("failed to write to file:\n%w", err)
	}
	return nil
}

func FileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return -1, err
	}

	return info.Size(), nil
}

func CountChars(filePath string) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0

	reader := bufio.NewReader(f)
	for {
		_, _, err := reader.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}
		count++
	}

	return count, err
}

func CountChar(filePath string, c rune) (int, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0

	reader := bufio.NewReader(f)
	for {
		readVal, _, err := reader.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, err
		}

		if readVal == c {
			count++
		}
	}

	return count, nil
}

func WriteLine(filePath, line string) error {
	f, err := OpenForWrite(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for writing:\n%w", err)
	}

	line = line + "\n"
	_, err = f.WriteString(line)
	if err != nil {
		return fmt.Errorf("failed to write to file\n%w", err)
	}

	return nil
}

func idxOutOfStrErr(str string, idx int) error {
	if len(str) < idx {
		return fmt.Errorf("index is out of string bounds")
	}

	return nil
}

func InsertToString(org, str string, idx int) (string, error) {
	reqLen := idx
	if len(org) < reqLen {
		extOrg, err := ExtendString(org, reqLen-len(org), ' ')
		if err != nil {
			return "", fmt.Errorf("error extending string(at InsertToString):\n%w", err)
		}
		org = extOrg
	}

	prefix := org[:idx]
	postFix := org[idx:]
	return prefix + str + postFix, nil
}

func OverrideString(org, str string, idx int) (string, error) {
	if err := idxOutOfStrErr(org, idx); err != nil {
		reqLen := idx + 1
		needToExtendBy := reqLen - len(org)
		org, err = ExtendString(org, needToExtendBy, ' ')
		if err != nil {
			return "", err
		}
	}

	prefix := org[:idx]
	postFix := ""
	postIdx := idx + len(str)
	if postIdx+1 < len(org) {
		postFix = org[postIdx:]
	}

	return prefix + str + postFix, nil
}

type scannerReader struct {
	s       *bufio.Scanner
	curLine int
}

func NewScannerReader(r io.Reader) scannerReader {
	return scannerReader{
		s:       bufio.NewScanner(r),
		curLine: 0,
	}
}

func (s *scannerReader) SkipLines(n int) error {
	for range n {
		if _, err := s.Read(); err != nil {
			return fmt.Errorf("error at SkipLines:\n%w", err)
		}
	}

	return nil
}

func (s *scannerReader) Read() (string, error) {
	if !s.s.Scan() {
		if s.s.Err() == nil {
			return "", io.EOF
		}

		return "", s.s.Err()
	}

	s.curLine++
	return s.s.Text(), nil
}

// return the last read line
func (s scannerReader) Line() int {
	return s.curLine
}

func Truncate(filePath string) error {
	if err := os.Truncate(filePath, 0); err != nil {
		return err
	}

	return nil
}

func ReadLine(filePath string, line int) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// scanner := NewScannerReader(f)
	// var curLineText string
	// for scanner.Line() < line {
	// 	curLineText, err = scanner.Read()
	// 	if err != nil {
	// 		return "", fmt.Errorf("error at ReadLine:\n%w", err)
	// 	}
	// }

	s := bufio.NewScanner(f)
	linesRead := 0
	for linesRead < line-1 {
		if ok := s.Scan(); !ok {
			if s.Err() == nil {
				return "", fmt.Errorf("not enough lines (stopped at %d)", linesRead)
			}

			return "", s.Err()
		}
		linesRead++
	}

	if ok := s.Scan(); !ok {
		if s.Err() == nil {
			return "", fmt.Errorf("not enough lines (stopped at %d)", linesRead)
		}

		return "", s.Err()
	}

	lineText := s.Text()
	return lineText, nil
}

func ReadAt(filePath string, line, col, n int) (string, error) {
	lineText, err := ReadLine(filePath, line)
	if err != nil {
		return "", err
	}

	startIdx := col - 1
	endIdx := min(len(lineText), startIdx+n)
	return lineText[startIdx:endIdx], nil
}

func copyFileContent(src, dst string) error {
	srcf, err := OpenForRead(src)
	if err != nil {
		return err
	}
	defer srcf.Close()

	dstf, err := OpenForWrite(dst)
	if err != nil {
		return err
	}
	defer dstf.Close()

	if err := Truncate(dst); err != nil {
		return err
	}

	_, err = io.Copy(dstf, srcf)
	return err
}

type InsertToFileArg struct {
	line, col int
	override  bool
	str       string
}

func IsEof(e error) bool {
	return errors.Is(e, io.EOF)
}

// AppendFromTo copy dstn lines from the file at srcpath to the file at dstpath,
// starting the line srcn. The copied lines are appended to dstpath.
// The function returned the actual copied lines. In case srcpath
// has less than dstn + srcn lines, the function copy what is can.
func AppendFromTo(srcpath, dstpath string, dstn, srcn int) (int, error) {
	if dstn <= 0 {
		return 0, nil
	}

	src, err := OpenForRead(srcpath)
	if err != nil {
		return -1, fmt.Errorf("error open file for read in AppendNLines")
	}
	defer src.Close()

	dst, err := OpenForWrite(dstpath)
	if err != nil {
		return -1, fmt.Errorf("error open file for write in AppendNLines")
	}
	defer dst.Close()

	// skip to srcn line
	sr := NewScannerReader(src)
	if err := sr.SkipLines(srcn - 1); err != nil {
		return 0, fmt.Errorf("error at AppenNLines:\n%w", err)
	}

	linesCopied := 0
	dw := bufio.NewWriter(dst)
	defer dw.Flush()

	for linesCopied < dstn {
		line, err := sr.Read()
		if err != nil {
			return linesCopied, err
		}

		if _, err := dw.WriteString(line + "\n"); err != nil {
			return linesCopied, err
		}

		linesCopied++
	}

	return linesCopied, nil
}

func sortByLineAndCol(inserts []InsertToFileArg) {
	slices.SortStableFunc(inserts, func(l InsertToFileArg, r InsertToFileArg) int {
		if l.line == r.line {
			return l.col - r.col
		}

		return l.line - r.line
	})
}

func InsertToFileMultiple(srcPath string, inserts []InsertToFileArg) error {
	if len(inserts) == 0 {
		return nil
	}

	srcf, err := OpenForRead(srcPath)
	if err != nil {
		return err
	}
	defer srcf.Close()

	tmp, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer tmp.Close()

	sortByLineAndCol(inserts)

	if minLine := inserts[0].line; minLine <= 0 {
		return fmt.Errorf("insert line must be > 0 (got %d)", minLine)
	}

	for i := range inserts {
		if c := inserts[i].col; c <= 0 {
			return fmt.Errorf("insert col must be > 0 (got %d)", c)
		}
	}

	srcLinesNum, err := CountLines(srcPath)
	if err != nil {
		return err
	}

	maxInsertLine := inserts[len(inserts)-1].line
	if srcLinesNum < maxInsertLine {
		if err := AppendEmptyLinesPaths(srcPath, maxInsertLine-srcLinesNum); err != nil {
			return err
		}
		srcLinesNum = maxInsertLine
	}

	insertOrOverride := func(org string, arg InsertToFileArg) (string, error) {
		if arg.override {
			return OverrideString(org, arg.str, arg.col-1)
		}

		return InsertToString(org, arg.str, arg.col-1)
	}

	var nextLineToInsert string
	tmpWriter := bufio.NewWriter(tmp)
	linesWritten := 0
	writeNextLine := func() error {
		if _, err = tmpWriter.WriteString(nextLineToInsert + "\n"); err != nil {
			return err
		}
		linesWritten++
		return nil
	}

	srcScanner := bufio.NewScanner(srcf)
	srcLineScanned := 0
	readLineFromSrc := func() (bool, error) {
		if ok := srcScanner.Scan(); !ok {
			return false, srcScanner.Err()
		}
		srcLineScanned++
		return true, nil
	}

	const initInsertLine = -1
	prevInsertLine := initInsertLine

	for i, arg := range inserts {
		insertLine := inserts[i].line

		if insertLine == prevInsertLine {
			nextLineToInsert, err = insertOrOverride(nextLineToInsert, arg)
			if err != nil {
				return err
			}
			continue
		}

		if prevInsertLine != initInsertLine { // some line is pending to be written
			if err := writeNextLine(); err != nil {
				return err
			}
		}

		for {
			if ok, err := readLineFromSrc(); !ok {
				if err == nil { // eof error
					panic("doesn't make sense, hard logic error")
				}
				return err
			}

			if srcLineScanned == insertLine {
				nextLineToInsert, err = insertOrOverride(srcScanner.Text(), arg)
				if err != nil {
					return err
				}

				break // can not write this line now, maybe the next arg's line is the same
			}

			nextLineToInsert = srcScanner.Text()
			if err := writeNextLine(); err != nil {
				return err
			}
		}

		prevInsertLine = insertLine
	}

	if linesWritten != srcLineScanned { // some line is pending to be written
		if err := writeNextLine(); err != nil {
			return err
		}
	}

	// copy the rest
	for srcLineScanned < srcLinesNum {
		if ok, err := readLineFromSrc(); !ok {
			if err != nil {
				return err
			}

			// else this is an eof error
			panic("eof error should not have happened")
		}

		nextLineToInsert = srcScanner.Text()
		if err := writeNextLine(); err != nil {
			return err
		}
	}

	if err := tmpWriter.Flush(); err != nil {
		return err
	}

	if err := copyFileContent(tmp.Name(), srcf.Name()); err != nil {
		return err
	}

	return nil
}

func InsertToFile(src, str string, lineNum, col int, override bool) error {
	lines, err := CountLines(src) // maybe skip this one and the one after that, somehow lazily do all neccessary work
	if err != nil {
		return fmt.Errorf("error counting lines:\n%w", err)
	}

	if lines < lineNum {
		if err = AppendEmptyLinesPaths(src, lineNum-int(lines)); err != nil {
			return fmt.Errorf("error append empty lines:\n%w", err)
		}
	}

	srcf, err := OpenForRead(src)
	if err != nil {
		return fmt.Errorf("error open file for read:\n%w", err)
	}

	tmpf, err := os.CreateTemp("", "")
	if err != nil {
		return fmt.Errorf("error create temp file:\n%w", err)
	}

	scanner := NewScannerReader(srcf)

	// seek the correct line, copy until then
	line, err := scanner.Read()
	if err != nil {
		return fmt.Errorf("could not read, err : %w", err)
	}

	tw := bufio.NewWriter(tmpf)
	for scanner.Line() < lineNum {
		if _, err := tw.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error write string(first iteration):\n%w", err)
		}

		line, err = scanner.Read()
		if err != nil {
			return fmt.Errorf("could not read, err : %w", err)
		}
	}

	idx := col - 1
	if override {
		line, err = OverrideString(line, str, idx)
		if err != nil {
			return fmt.Errorf("error override string:\n%w", err)
		}
	} else {
		line, err = InsertToString(line, str, idx)
		if err != nil {
			return fmt.Errorf("error insert to string:\n%w", err)
		}
	}

	if _, err := tw.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("error write corrected line:\n%w", err)
	}

	// copy the rest
	for {
		line, err := scanner.Read()
		if err != nil {
			if IsEof(err) {
				break
			}

			return err
		}

		if _, err := tw.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("error write string (2nd iteration):\n%w", err)
		}
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	if err = copyFileContent(tmpf.Name(), srcf.Name()); err != nil {
		return err
	}

	return nil
}
