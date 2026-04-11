package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

type tempFile struct {
	f *os.File
}

func (tf tempFile) Name() string {
	return tf.f.Name()
}

func (tf *tempFile) Close() {
	tf.f.Close()
	os.Remove(tf.f.Name())
}

func TestGbToByte(t *testing.T) {
	const gb = 2
	const expectedBytes = 2 * 1024 * 1024 * 1024
	res := GbToBytes(gb)
	if res != expectedBytes {
		t.Log("Failed to convert gb bytes to bytes")
		t.Logf("Expected %d, result %d", expectedBytes, res)
		t.FailNow()
	}
}

func TestExtendString(t *testing.T) {
	const org = "original string"
	orgLen := len(org)
	extendBy := 3
	expectedLen := orgLen + extendBy
	extendRune := 'c'
	res, err := ExtendString(org, extendBy, extendRune)
	if err != nil {
		t.Fatal(err)
	}

	resLen := len(res)
	if resLen != expectedLen {
		t.Log("Failed to extend string")
		t.Logf("expected len : %d, res len : %d", expectedLen, resLen)
		t.FailNow()
	}

	if res[:orgLen] != org {
		t.Log("something bad happend to the result string (it has changed the original string characters)")
		t.Logf("expected : %s, has : %s", org, res[:orgLen])
		t.FailNow()
	}

	expectedResOfChars := slices.Repeat([]rune{extendRune}, extendBy)

	if res[orgLen:] != string(expectedResOfChars) {
		t.Log("The extended part of the string is different than expected")
		t.Logf("Expected %s, has %s", string(expectedResOfChars), res[orgLen:])
		t.FailNow()
	}

	// test receiving error as expected
	_, err = ExtendString(org, -3, ' ')
	if err == nil {
		t.Fatal("Expected to receive an error (negetive length), but did not receive one")
	}
}

const tmpDir = "/tmp"

func createTemp() *tempFile {
	tmpFile, err := os.CreateTemp(tmpDir, "")
	if err != nil {
		panic(err)
	}
	return &tempFile{
		f: tmpFile,
	}
}

func TestCountLines(t *testing.T) {

	tmpFile := createTemp()
	defer tmpFile.Close()

	const threeLinesContent string = "this is somefile content\nit has three lines\nthis is the third line"
	n, err := tmpFile.f.WriteString(threeLinesContent)
	if err != nil {
		t.Fatal(err)
	}

	if n != len(threeLinesContent) {
		t.Fatalf("did not wrote %d bytes as expeted, wrote %d", len(threeLinesContent), n)
	}

	res, err := CountLines(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if res != 3 {
		t.Fatalf("Expected 3 lines but received %d", res)
	}
}

func writeAndIncrementCount(f *os.File, b []byte, c *int, t *testing.T) {
	_, err := f.Write(b)
	if err != nil {
		t.Fatal(err)
	}

	(*c) += len(b)
}

func TestCountChars(t *testing.T) {
	tmpFile := createTemp()
	defer tmpFile.Close()

	firstLine := "this is first line"
	secondLine := "this is second"
	thirdLine := "third"
	newLine := []byte{'\n'}

	expCount := 0
	writeAndIncrementCount(tmpFile.f, []byte(firstLine), &expCount, t)
	writeAndIncrementCount(tmpFile.f, newLine, &expCount, t)
	writeAndIncrementCount(tmpFile.f, []byte(secondLine), &expCount, t)
	writeAndIncrementCount(tmpFile.f, newLine, &expCount, t)
	writeAndIncrementCount(tmpFile.f, []byte(thirdLine), &expCount, t)
	if expCount < len(firstLine) {
		panic("testing function is bad")
	}

	resCount, err := CountChars(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if resCount != expCount {
		t.Log("result count is different than expected")
		t.Fatalf("expected : %d, got : %d", expCount, resCount)
	}

	newLineCount, err := CountChar(tmpFile.Name(), '\n')
	if err != nil {
		t.Fatal(err)
	}

	if newLineCount != 2 {
		t.Fatal("did not count new lines correctly")
	}
}

func TestAppendEmptyLines(t *testing.T) {
	tf := createTemp()
	defer tf.Close()

	if err := AppendEmptyLinesPaths(tf.Name(), 1); err != nil {
		panic(err)
	}

	res, err := CountLines(tf.Name())
	if err != nil {
		panic(err)
	}

	if res != 1 {
		t.Fatalf("expected 1 lines, got %d", res)
	}

	res, err = CountChars(tf.Name())
	if res != 1 {
		t.Fatalf("expected 1 chars, got %d", res)
	}

	notEmpty := "this is not an empty line"
	if err = WriteLine(tf.Name(), notEmpty); err != nil {
		panic(err)
	}

	res, err = CountLines(tf.Name())
	if err != nil {
		panic(err)
	}

	if res != 2 {
		t.Fatalf("expected 2 lines, got %d", res)
	}

	if err = AppendEmptyLinesPaths(tf.Name(), 5); err != nil {
		panic(err)
	}

	res, err = CountLines(tf.Name())
	if err != nil {
		panic(err)
	}
	if res != 7 {
		t.Fatalf("expected 7 lines, has %d", res)
	}
	res, err = CountChars(tf.Name())
	if err != nil {
		panic(err)
	}
	exp := len(notEmpty) + 7
	if exp != res {
		t.Fatalf("expected %d chars, has %d", exp, res)
	}
}

func TestFillFile(t *testing.T) {
	tf := createTemp()
	defer tf.Close()

	n := 500
	perLine := 10
	if n%perLine != 0 {
		panic("bad arguments, must divide")
	}
	err := FillFile(tf.Name(), 'c', n, perLine)
	if err != nil {
		t.Fatal(err)
	}

	expLines := n / perLine
	resLines, err := CountLines(tf.Name())
	if err != nil {
		t.Fatal(err)
	}

	if resLines != expLines {
		t.Log("bad resulted lines")
		t.Fatalf("expected %d, got %d", expLines, resLines)
	}

	count, err := CountChars(tf.Name())
	if err != nil {
		t.Fatal(err)
	}
	if count < n {
		t.Fatalf("bad count %d", count)
	}

	count, err = CountChar(tf.Name(), 'c')
	if err != nil {
		t.Fatal(err)
	}
	if count != n {
		t.Log("bad char count")
		t.Fatalf("expected %d, result %d", n, count)
	}
}

func TestFileSize(t *testing.T) {
	tmp := createTemp()
	defer tmp.Close()

	expSize := 50
	perLine := expSize
	FillFile(tmp.Name(), 'c', expSize, perLine)

	res, err := FileSize(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if res != int64(expSize) {
		t.Log("bad file size")
		t.Fatalf("expected %d, got %d", expSize, res)
	}
}

func TestGenerateHugeFile(t *testing.T) {
	tmp := createTemp()
	defer tmp.Close()

	mb := 1
	c := 'c'
	perLine := 120

	err := GeneraetHugeFileMb(mb, c, perLine, tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	resSize, err := FileSize(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}

	if resSize < int64(MbToBytes(1)) {
		t.Log("generate file that is too small")
		t.Fatalf("expected at least %d, got %d", MbToBytes(2), resSize)
	}
}

func TestInsertToString(t *testing.T) {
	// test regular insertion
	const org = "This is the original string"
	const insert = "(whoops, I'm new here)"
	subOrg := "the"
	expected := "This is " + insert + "the original string"
	index := strings.Index(org, subOrg)
	res, err := InsertToString(org, insert, index)
	if err != nil {
		t.Fatal(err)
	}

	if res != expected {
		t.Log("bad string insertion")
		t.Fatalf("expected %s, got %s", expected, res)
	}

	// test insert to begining
	expected = insert + org
	res, err = InsertToString(org, insert, 0)
	if err != nil {
		t.Fatal(err)
	}

	if res != expected {
		t.Log("bad string insertion (begining of string)")
		t.Fatalf("expected %s, got %s", expected, res)
	}

	// test insert to end
	expected = org + insert
	res, err = InsertToString(org, insert, len(org))
	if err != nil {
		t.Fatal(err)
	}

	if res != expected {
		t.Log("bad string insertion (end of string)")
		t.Fatalf("expected \"%s\", got \"%s\"", expected, res)
	}
}

func TestInsertToEmptyString(t *testing.T) {
	org := ""
	exp := "exp"
	res, err := InsertToString(org, exp, 0)
	if err != nil {
		t.Fatal(err)
	}

	if res != exp {
		t.Fatal(expGot(exp, res))
	}

}

func expGot(exp, res any) string {
	return fmt.Sprintf("expected '%v', got '%v'", exp, res)
}

func TestOverrideString(t *testing.T) {
	const org = "This is the original string"

	// test override from begining
	override := "Hola"
	exp := "Hola is the original string"
	res, err := OverrideString(org, override, 0)
	if err != nil {
		t.Fatal(err)
	}

	if res != exp {
		t.Log("bad string override")
		t.Fatal(expGot(exp, res))
	}

	// test override from middle
	exp = "This is the Holainal string"
	idx := strings.Index(org, "original")
	res, err = OverrideString(org, override, idx)
	if err != nil {
		t.Fatal(err)
	}

	if res != exp {
		t.Log("bad string override (middle)")
		t.Fatal(expGot(exp, res))
	}

	// test override from end
	exp = "This is the original stHola"
	idx = strings.Index(org, "string") + 2
	res, err = OverrideString(org, override, idx)
	if err != nil {
		t.Fatal(err)
	}

	if res != exp {
		t.Log("bad string override (end)")
		t.Fatal(expGot(exp, res))
	}

	// test override post end
	exp = "This is the original stringHola"
	idx = len(org)
	res, err = OverrideString(org, override, idx)
	if err != nil {
		t.Fatal(err)
	}

	if res != exp {
		t.Log("bad string override (post-end)")
		t.Fatal(expGot(exp, res))
	}
}

func TestReadAt(t *testing.T) {
	tmp := createTemp()
	defer tmp.Close()
	const testText = "First line1\nSecond line2\nThird line3\n"

	initTmp := func() {
		Truncate(tmp.Name())
		tmp.f.WriteString(testText)
	}

	// test reading the first word of the first line
	initTmp()
	exp := "First"
	res, err := ReadAt(tmp.Name(), 1, 1, len(exp))
	if err != nil {
		t.Fatal(err)
	}
	if res != exp {
		t.Log("bad read at")
		t.Fatal(expGot(exp, res))
	}

	// test reading the first word of the second line
	exp = "Second"
	res, err = ReadAt(tmp.Name(), 2, 1, len(exp))
	if err != nil {
		t.Fatal(err)
	}
	if res != exp {
		t.Log("bad read at")
		t.Fatal(expGot(exp, res))
	}

	// test reading the second word of the second line
	exp = "line2"
	res, err = ReadAt(tmp.Name(), 2, 8, len(exp))
	if err != nil {
		t.Fatal(err)
	}
	if res != exp {
		t.Log("bad read at")
		t.Fatal(expGot(exp, res))
	}

	// test reading past the last word of the second line
	exp = "e2"
	res, err = ReadAt(tmp.Name(), 2, 11, len(exp)+5) // len > exp
	if err != nil {
		t.Fatal(err)
	}
	if res != exp {
		t.Log("bad read at")
		t.Fatal(expGot(exp, res))
	}
}

func TestInsertToFile(t *testing.T) {
	const first = "This is the 1st line"
	const second = "This is 2nd"
	const third = "This is 3rd"
	lines := []string{first, second, third}
	var testText strings.Builder
	for _, l := range lines {
		testText.WriteString(l + "\n")
	}

	tmp := createTemp()
	defer tmp.Close()

	initTmpFile := func() {
		if err := Truncate(tmp.Name()); err != nil {
			t.Fatal(err)
		}

		tmp.f.Seek(0, 0)

		if _, err := tmp.f.WriteString(testText.String()); err != nil {
			t.Fatal(err)
		}
	}

	test := func(str string, line, col int, override bool) {
		onErr := func(err error) {
			panic(err)
		}
		initTmpFile()
		err := InsertToFile(tmp.Name(), str, line, col, override)
		if err != nil {
			onErr(err)
		}
		lineIdx := line - 1
		orgLine := lines[lineIdx]

		var exp string
		if override {
			exp, err = OverrideString(orgLine, str, col-1)
		} else {
			exp, err = InsertToString(orgLine, str, col-1)
		}

		if err != nil {
			onErr(err)
		}

		res, err := ReadLine(tmp.Name(), line)
		if err != nil {
			onErr(err)
		}

		if res != exp {
			t.Log("bad insert to file")
			onErr(errors.New(expGot(exp, res)))
		}

		// make sure the rest of the lines are ok
		for i, exp := range lines {
			if i == lineIdx {
				continue
			}

			lineNum := i + 1
			res, err := ReadLine(tmp.Name(), lineNum)
			if err != nil {
				onErr(err)
			}
			if exp != res {
				errStr := "bad rest of file\n"
				onErr(errors.New(errStr + expGot(exp, res)))
			}
		}
	}

	testNotOverride := func(str string, line, col int) {
		test(str, line, col, false)
	}

	testOverride := func(str string, line, col int) {
		test(str, line, col, true)
	}

	// test insert at the begining of a line
	testNotOverride("Hi! ", 1, 1)
	testNotOverride("Hola!", 2, 1)
	testNotOverride("HEYY", 3, 1)
	testOverride("Hi! ", 1, 1)
	testOverride("Hola!", 2, 1)
	testOverride("ADIOS MOCHACUS!", 3, 1)
	// to the middle
	testNotOverride("Hi! ", 1, 5)
	testNotOverride("Hola!", 2, 4)
	testNotOverride("HEYY", 3, 3)
	testOverride("Hi! ", 1, 7)
	testOverride("Hola!", 2, 3)
	testOverride("ADIOS MOCHACUS!", 3, 2)
	// at the end
	testNotOverride("Hi! ", 1, len(lines[0]))
	testNotOverride("Ho!", 2, len(lines[1]))
	testNotOverride("HEYY", 3, len(lines[2]))
	testOverride("Hi! ", 1, len(lines[0]))
	testOverride("Hola!", 2, len(lines[1]))
	testOverride("ADIOS MOCHACUS!", 3, len(lines[2]))
	// long strings
	testNotOverride("Hiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii!", 1, 5)
	testNotOverride("Holaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!", 2, 4)
	testNotOverride("HEYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY", 3, 3)
	testOverride("Hiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii! ", 1, 7)
	testOverride("Holaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!", 2, 3)
	testOverride("ADIOS MOCHACUSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS!", 3, 2)
	// empty strings
	testNotOverride("", 1, 1)
	testNotOverride("", 2, 1)
	testNotOverride("", 3, 1)
	testOverride("", 1, 1)
	testOverride("", 2, 1)
	testOverride("", 3, 1)
}

func TestInsertToFileBeyondItsLineNumber(t *testing.T) {
	tf := createTemp()
	defer tf.Close()

	if err := WriteLine(tf.Name(), "some line"); err != nil {
		t.Fatal(err)
	}

	if err := InsertToFile(tf.Name(), "some string", 3, 3, false); err != nil {
		t.Fatal(err)
	}

	res, err := ReadAt(tf.Name(), 3, 3, len("some string"))
	if err != nil {
		t.Fatal(err)
	}

	if res != "some string" {
		t.Fatal("expected different string")
	}
}

func TestAppendNLines(t *testing.T) {
	lines := []string{
		"first", "second", "third",
	}

	dst := createTemp()
	defer dst.Close()
	src := initTmpWithLines(lines)
	defer src.Close()
	srcpath := src.Name()
	dstpath := dst.Name()

	testZeroSrcn := func(nLines, srcn int) {
		if nLines > len(lines) {
			panic("bad testing")
		}

		prevDstLines, err := CountLines(dstpath)
		if err != nil {
			t.Fatal(err)
		}

		srcLines, err := CountLines(src.Name())
		if err != nil {
			t.Fatal(err)
		}

		var isExpectedEof bool = srcLines < srcn+nLines
		linesAvailableForCopy := max(len(lines)-(srcn-1), 0)
		expCopiedLine := min(linesAvailableForCopy, nLines)
		linesCopied, err := AppendFromTo(srcpath, dstpath, nLines, srcn)
		if err != nil {
			if IsEof(err) && isExpectedEof {
				// all good
			} else {
				t.Fatal(fmt.Errorf("error during tests - nLines:%d,srcn:%d :\n%w", nLines, srcn, err))
			}
		}

		if linesCopied != expCopiedLine {
			t.Fatalf("expected to copy %d lines, copied :%d, nLines:%d, srcn:%d", expCopiedLine, linesCopied, nLines, srcn)
		}

		curDstLines, err := CountLines(dstpath)
		if err != nil {
			t.Fatal(err)
		}

		if curDstLines != prevDstLines+expCopiedLine {
			panic(fmt.Errorf("dst file expected to have %d line (has %d)", nLines, linesCopied))
		}

		for i := 1; i < expCopiedLine; i++ {
			dstLine, err := ReadLine(dstpath, i+prevDstLines)
			if err != nil {
				t.Fatal(err)
			}

			srcLine, err := ReadLine(srcpath, srcn+i-1)
			if err != nil {
				t.Fatal(err)
			}

			if dstLine != srcLine {
				t.Fatalf("line is different than expected at index %d, has \"%s\", expected \"%s\", nLines:%d, srcn;%d", i, dstLine, srcLine, nLines, srcn)
			}
		}
	}

	testZeroSrcn(0, 1)
	testZeroSrcn(0, 2)
	testZeroSrcn(0, 3)
	testZeroSrcn(1, 1)
	testZeroSrcn(1, 2)
	testZeroSrcn(1, 3)
	testZeroSrcn(2, 1)
	testZeroSrcn(2, 2)
	testZeroSrcn(2, 3)
	testZeroSrcn(2, 4)
	testZeroSrcn(3, 1)
	testZeroSrcn(3, 2)
	testZeroSrcn(3, 3)
	testZeroSrcn(3, 4)
	testZeroSrcn(3, 5)
}

func TestInsertToFileMultipleWithOneArgumentOnEmptyFile(t *testing.T) {
	emptyFile := createTemp()
	defer emptyFile.Close()
	str := "Some Unique String"

	line := 1
	col := 1
	override := false
	arg := InsertToFileArg{
		line:     line,
		col:      col,
		override: override,
		str:      str,
	}

	argArray := []InsertToFileArg{arg}
	if err := InsertToFileMultiple(emptyFile.Name(), argArray); err != nil {
		t.Fatal(err)
	}

	res, err := ReadAt(emptyFile.Name(), line, col, len(str))
	if err != nil {
		t.Fatal(err)
	}

	if res != str {
		t.Fatalf("expected identical strings (has %s, expected %s)", res, str)
	}

	line = 9
	col = 15
	arg = InsertToFileArg{
		line:     line,
		col:      col,
		override: override,
		str:      str,
	}

	if err := os.Truncate(emptyFile.Name(), 0); err != nil {
		panic(err)
	}

	argArray = []InsertToFileArg{arg}
	if err := InsertToFileMultiple(emptyFile.Name(), argArray); err != nil {
		t.Fatal(err)
	}

	res, err = ReadAt(emptyFile.Name(), line, col, len(str))
	if err != nil {
		t.Fatal(err)
	}

	if res != str {
		t.Fatalf("expected identical strings (has %s, expected %s)", res, str)
	}

	// override
	str = "override the previous one"
	override = true
	arg = InsertToFileArg{
		line:     line,
		col:      col,
		override: override,
		str:      str,
	}

	if err := os.Truncate(emptyFile.Name(), 0); err != nil {
		t.Fatal(err)
	}

	argArray = []InsertToFileArg{arg}
	if err := InsertToFileMultiple(emptyFile.Name(), argArray); err != nil {
		t.Fatal(err)
	}

	res, err = ReadLine(emptyFile.Name(), line)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) <= 0 {
		t.Fatalf("line len is %d", len(res))
	}

	res, err = ReadAt(emptyFile.Name(), line, col, len(str))
	if err != nil {
		t.Fatal(err)
	}

	if res != str {
		t.Fatalf("expected identical strings (has %s, expected %s)", res, str)
	}

	newArg := func() InsertToFileArg {
		return InsertToFileArg{
			line:     line,
			col:      col,
			override: override,
			str:      str,
		}
	}

	newArgArray := func() []InsertToFileArg {
		return []InsertToFileArg{newArg()}
	}
	// look for errors
	line = 0
	argArray = newArgArray()
	if err = InsertToFileMultiple(emptyFile.Name(), argArray); err == nil {
		t.Fatal("expected some error since line number is 0")
	}

	line = 99
	col = 0
	argArr := newArgArray()
	if err = InsertToFileMultiple(emptyFile.Name(), argArr); err == nil {
		t.Fatal("expected some error since col number is 0")
	}
}

func TestInsertToFileMultipleWithMultipleArgumentsOnEmptyFile(t *testing.T) {
	f := createTemp()
	defer f.Close()

	str1 := "str1111111111111111111111111111"
	line1 := 5
	col1 := 6
	override1 := false
	arg1 := InsertToFileArg{
		line:     line1,
		col:      col1,
		override: override1,
		str:      str1,
	}

	str2 := "str22222222222222222222222"
	line2 := 19
	col2 := 33
	override2 := false
	arg2 := InsertToFileArg{
		line:     line2,
		col:      col2,
		override: override2,
		str:      str2,
	}

	str3 := "str3333333333333333333333333"
	line3 := 994
	col3 := 15
	override3 := true
	arg3 := InsertToFileArg{
		line:     line3,
		col:      col3,
		override: override3,
		str:      str3,
	}

	argArray := []InsertToFileArg{arg1, arg2, arg3}
	if err := InsertToFileMultiple(f.Name(), argArray); err != nil {
		t.Fatal(err)
	}

	// arg1
	res, err := ReadAt(f.Name(), line1, col1, len(str1))
	if err != nil {
		t.Fatal(err)
	}

	if res != str1 {
		t.Fatalf("expected %s, has %s", str1, res)
	}

	// arg2
	res, err = ReadLine(f.Name(), line2)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) <= 0 {
		t.Fatal("line is empty")
	}

	res, err = ReadAt(f.Name(), line2, col2, len(str2))
	if err != nil {
		t.Fatal(err)
	}

	if res != str2 {
		t.Fatalf("expected %s, has %s", str2, res)
	}

	// arg3
	res, err = ReadLine(f.Name(), line3)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) <= 0 {
		t.Fatal("line is empty")
	}
	res, err = ReadAt(f.Name(), line3, col3, len(str3))
	if err != nil {
		t.Fatal(err)
	}
	if res != str3 {
		t.Fatalf("expected %s, has %s", str3, res)
	}

	// look for errors
	argArray[2].col = -1
	if err = InsertToFileMultiple(f.Name(), argArray); err == nil {
		t.Fatal("expected error since last argument col is -1")
	}
}

func TestInsertToFileMultipleOneArgNonEmptyFile(t *testing.T) {
	// initialize file
	lines := []string{}
	const linesNum = 999
	for i := 1; i <= linesNum; i++ {
		lines = append(lines, fmt.Sprintf("this is line %d", i))
	}
	tf := initTmpWithLines(lines)
	if cnt, err := CountLines(tf.Name()); err != nil || cnt != linesNum {
		panic("something bad happened")
	}

	line := 1
	col := 1
	override := false
	str := "Hello"

	newArg := func() InsertToFileArg {
		return InsertToFileArg{
			line:     line,
			col:      col,
			override: override,
			str:      str,
		}
	}

	newArgArr := func() []InsertToFileArg {
		return []InsertToFileArg{newArg()}
	}

	callInsert := func() error {
		return InsertToFileMultiple(tf.Name(), newArgArr())
	}

	if err := callInsert(); err != nil {
		t.Fatal(err)
	}

	srcPath := tf.Name()

	test := func() {
		var origLine string
		if len(lines) < line {
			origLine = ""
		} else {
			origLine = lines[line-1]
		}

		getExpLine := func() string {
			var ret string
			var err error
			if override {
				ret, err = OverrideString(origLine, str, col-1)
			} else {
				ret, err = InsertToString(origLine, str, col-1)
			}

			if err != nil {
				panic(err)
			}

			return ret
		}

		expLine := getExpLine()
		resLine, err := ReadLine(srcPath, line)
		if err != nil {
			t.Fatal(err)
		}

		fmtArg := fmt.Sprintf("%+v", newArg())

		if expLine != resLine {
			t.Fatalf("exp: %s, res: %s, args : %v", expLine, resLine, fmtArg)
		}

		if len(lines) < line {
			for len(lines) < line {
				lines = append(lines, origLine)
			}
		}

		lines[line-1] = resLine // file lines have changed

		linesCnt, err := CountLines(srcPath)
		if err != nil {
			t.Fatal(err)
		}

		if linesCnt != len(lines) {
			t.Fatalf("something happened to the lines count, args : %v", fmtArg)
		}
	}
	test()

	callInsertNoErrors := func() {
		if err := callInsert(); err != nil {
			t.Fatal(err)
		}
	}

	callTestNoErrors := func() {
		callInsertNoErrors()
		test()
	}

	line = 9
	callTestNoErrors()

	str = "same line different string"
	callTestNoErrors()

	col = 3
	callTestNoErrors()

	str = "same col"
	callTestNoErrors()

	col = 2
	callTestNoErrors()

	override = true
	callTestNoErrors()

	override = false
	callTestNoErrors()

	override = true
	str = "different str"
	callTestNoErrors()

	line = 1001
	callTestNoErrors()

	// now expect some errors
	col = -1
	if err := callInsert(); err == nil {
		t.Fatal("expected some error since col is -1s")
	}

	col = 229
	callTestNoErrors()

	line = -99
	if err := callInsert(); err == nil {
		t.Fatal("expected some error since line is -99")
	}

}

func TestInsertToFileMultipleManyArgsNonEmptyFile(t *testing.T) {
	expLines := []string{}
	const numOfLines = 1000
	for i := range numOfLines {
		expLines = append(expLines, fmt.Sprintf("This is line number %d", i))
	}
	tmp := initTmpWithLines(expLines)

	newArg := func(line, col int, str string, override bool) InsertToFileArg {
		return InsertToFileArg{
			line:     line,
			col:      col,
			override: override,
			str:      str,
		}
	}

	argArrayWithErrors := []InsertToFileArg{
		newArg(-1, 5, "should cause an error (line)", true),
		newArg(2, 0, "should cause an error (col)", true),
		newArg(1000, 80, "Should be inserted at line 1000", true),
		newArg(999, 1, "This is line number 999 (but with some text)", true),
		newArg(1, 1, "Holla", true),
		newArg(1, 30, "yo", false),
		newArg(500, 9, "500 500 500 500", false),
		newArg(500, 11, "500-11 500-11 500-11 500-11", false),
		newArg(3000, 9, "3000 3000 3000 3000", true),
	}

	if err := InsertToFileMultiple(tmp.Name(), argArrayWithErrors); err == nil {
		t.Fatal("expected errors to occur")
	}

	validArgArray := []InsertToFileArg{}
	for _, a := range argArrayWithErrors {
		if a.col <= 0 || a.line <= 0 {
			continue
		}
		validArgArray = append(validArgArray, a)
	}

	if len(validArgArray) < 5 {
		panic("not what you wanted")
	}

	updateExpLines := func() {
		clone := slices.Clone(validArgArray)
		sortByLineAndCol(clone)

		linesVisited := []int{}
		for i, curArg := range validArgArray {
			l := curArg.line
			if slices.Contains(linesVisited, l) {
				continue
			}

			sameLineArgs := []InsertToFileArg{curArg}
			for _, anotherArg := range validArgArray[i+1:] {
				if anotherArg.line != l {
					break
				}

				sameLineArgs = append(sameLineArgs, anotherArg)
			}

			for len(expLines) < l {
				expLines = append(expLines, "")
			}

			for _, a := range sameLineArgs {
				var newExpLine string
				var err error
				orgLine := expLines[l-1]
				if a.override {
					newExpLine, err = OverrideString(orgLine, a.str, a.col-1)
				} else {
					newExpLine, err = InsertToString(orgLine, a.str, a.col-1)
				}

				if err != nil {
					t.Fatalf("err : %s, arg: %+v", err.Error(), a)
				}

				expLines[l-1] = newExpLine
			}

			linesVisited = append(linesVisited, l)
		}

	}

	test := func(arg InsertToFileArg) {
		expLine := expLines[arg.line-1]
		resLine, err := ReadLine(tmp.Name(), arg.line)
		if err != nil {
			t.Fatalf("error occured: %s, arg %+v", err.Error(), arg)
		}

		if resLine != expLines[arg.line-1] {
			t.Fatalf("exp : %s, res %s, arg %+v", expLine, resLine, arg)
		}
	}

	if err := InsertToFileMultiple(tmp.Name(), validArgArray); err != nil {
		t.Fatal(err)
	}

	updateExpLines()
	for _, a := range validArgArray {
		test(a)
	}
}
