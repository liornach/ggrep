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

func CountLines(filePath string) (uint64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	count := uint64(0)
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

func AppendEmptyLines(filePath string, lines int) error {
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

func CountChars(filePath string) (uint64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := uint64(0)

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

func CountChar(filePath string, c rune) (uint64, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := uint64(0)

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
	if err := idxOutOfStrErr(org, idx); err != nil {
		return "", err
	}

	prefix := org[:idx]
	postFix := org[idx:]
	return prefix + str + postFix, nil
}

func OverrideString(org, str string, idx int) (string, error) {
	if err := idxOutOfStrErr(org, idx); err != nil {
		return "", err
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

func (s *scannerReader) Read() (string, bool) {
	if !s.s.Scan() {
		return "", false
	}

	s.curLine++
	return s.s.Text(), true
}

func (s *scannerReader) Err() error {
	return s.s.Err()
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

	scanner := NewScannerReader(f)
	var curLineText string
	var ok bool
	for scanner.Line() < line {
		curLineText, ok = scanner.Read()
		if !ok {
			return "", fmt.Errorf("could not read")
		}
	}

	return curLineText, nil
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

func copyContent(from, to string) error {
	tf, err := OpenForWrite(to)
	if err != nil {
		return err
	}
	defer tf.Close()

	ff, err := OpenForRead(from)
	if err != nil {
		return err
	}
	defer ff.Close()

	s := NewScannerReader(ff)

	if err = Truncate(tf.Name()); err != nil {
		return err
	}

	for {
		lineText, ok := s.Read()
		if !ok {
			if err = s.Err(); err != nil {
				return err
			}

			break
		}
		writeLineToOpenFile(tf, lineText)
	}

	return nil
}

func writeLineToOpenFile(f *os.File, line string) error {
	_, err := f.WriteString(line + "\n")
	return err
}

func InsertToFile(filePath, str string, lineNum, col int, override bool) error {
	lines, err := CountLines(filePath)
	if err != nil {
		return err
	}

	if lines < uint64(lineNum) {
		if err = AppendEmptyLines(filePath, lineNum-int(lines)); err != nil {
			return err
		}
	}

	f, err := OpenForRead(filePath)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}

	scanner := NewScannerReader(f)

	// seek the correct line, copy until then
	line, ok := scanner.Read()
	if !ok {
		return fmt.Errorf("could not read, err : %w", scanner.Err())
	}
	for scanner.Line() < lineNum {
		if err := writeLineToOpenFile(tmpFile, line); err != nil {
			return err
		}

		line, ok = scanner.Read()
		if !ok {
			return fmt.Errorf("could not read, err : %w", scanner.Err())
		}
	}

	// we're at the target line
	// something is bad here
	// requiredLen := col + len(str)
	// if len(line) < requiredLen {
	// 	line, err = ExtendString(line, requiredLen-len(line), ' ')
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	idx := col - 1
	if override {
		line, err = OverrideString(line, str, idx)
	} else {
		line, err = InsertToString(line, str, idx)
	}

	if err != nil {
		return err
	}

	if err := writeLineToOpenFile(tmpFile, line); err != nil {
		return err
	}

	// copy the rest
	for {
		line, ok := scanner.Read()
		if !ok {
			break
		}

		if err := writeLineToOpenFile(tmpFile, line); err != nil {
			return err
		}
	}

	if err = copyContent(tmpFile.Name(), filePath); err != nil {
		return err
	}

	return nil
}
