package main

import (
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

func writeAndIncrementCount(f *os.File, b []byte, c *uint64, t *testing.T) {
	_, err := f.Write(b)
	if err != nil {
		t.Fatal(err)
	}

	(*c) += uint64(len(b))
}

func TestCountChars(t *testing.T) {
	tmpFile := createTemp()
	defer tmpFile.Close()

	firstLine := "this is first line"
	secondLine := "this is second"
	thirdLine := "third"
	newLine := []byte{'\n'}

	expCount := uint64(0)
	writeAndIncrementCount(tmpFile.f, []byte(firstLine), &expCount, t)
	writeAndIncrementCount(tmpFile.f, newLine, &expCount, t)
	writeAndIncrementCount(tmpFile.f, []byte(secondLine), &expCount, t)
	writeAndIncrementCount(tmpFile.f, newLine, &expCount, t)
	writeAndIncrementCount(tmpFile.f, []byte(thirdLine), &expCount, t)
	if expCount < uint64(len(firstLine)) {
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

	if err := AppendEmptyLines(tf.Name(), 1); err != nil {
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

	if err = AppendEmptyLines(tf.Name(), 5); err != nil {
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
	if uint64(exp) != res {
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

	if resLines != uint64(expLines) {
		t.Log("bad resulted lines")
		t.Fatalf("expected %d, got %d", expLines, resLines)
	}

	count, err := CountChars(tf.Name())
	if err != nil {
		t.Fatal(err)
	}
	if count < uint64(n) {
		t.Fatalf("bad count %d", count)
	}

	count, err = CountChar(tf.Name(), 'c')
	if err != nil {
		t.Fatal(err)
	}
	if count != uint64(n) {
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
		t.Fatalf("expected %s, got %s", expected, res)
	}
}

func expGot(exp, res any) string {
	return fmt.Sprintf("expected %v, got %v", exp, res)
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

	// test error
	idx = len(org) + 1
	_, err = OverrideString(org, override, idx)
	if err == nil {
		t.Fatal("expected some error but got none")
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
	const first = "This is the 1st line\n"
	const second = "This is 2nd\n"
	const third = "This is 3rd\n"
	const testText = first + second + third
	tmp := createTemp()
	defer tmp.Close()

	initTmpFile := func() {
		if err := Truncate(tmp.Name()); err != nil {
			t.Fatal(err)
		}

		tmp.f.Seek(0, 0)

		if _, err := tmp.f.WriteString(testText); err != nil {
			t.Fatal(err)
		}
	}

	// test insert and NOT override at the begining
	initTmpFile()
	str := "Hi! "
	err := InsertToFile(tmp.Name(), str, 1, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ReadAt(tmp.Name(), 1, 1, 8)
	if err != nil {
		t.Fatal(err)
	}
	exp := str + "This"
	if res != exp {
		t.Log("bad insert to file")
		t.Fatal(expGot(exp, res))
	}

	// test insert and override at the begining
	initTmpFile()
	err = InsertToFile(tmp.Name(), str, 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	res, err = ReadAt(tmp.Name(), 1, 1, 8)
	if err != nil {
		t.Fatal(err)
	}
	exp = str + " is "
	if res != exp {
		t.Log("bad insert to file (override)")
		t.Fatal(expGot(exp, res))
	}
	exp = "This is 2nd"
	res, err = ReadAt(tmp.Name(), 2, 1, len(exp)) // make sure the rest of the file is fine
	if err != nil {
		t.Fatal(err) // must debug this one
	}
	if res != exp {
		t.Log("bad isnert to file (2nd line validation)")
		t.Fatal(expGot(exp, res))
	}

	// test insert and NOT override to begining of the 2nd line
	initTmpFile()
	err = InsertToFile(tmp.Name(), str, 2, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	exp = str + second[:5]
	res, err = ReadAt(tmp.Name(), 2, 1, len(str)+5)
	if err != nil {
		t.Fatal(err)
	}
	if res != exp {
		t.Log("bad insert to file (2nd line insert and NOT override)")
		t.Fatal(expGot(exp, res))
	}

	// test insert and override to the begining of the 2nd line
	initTmpFile()
	err = InsertToFile(tmp.Name(), str, 2, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(str)+3 > len(second) {
		panic("bad test parameters")
	}
	exp = str + second[len(str):len(str)+3]
	res, err = ReadAt(tmp.Name(), 2, 1, len(str)+3)
	if err != nil {
		t.Fatal(err)
	}
	if res != exp {
		t.Log("bad insert to file (2nd line insert and override)")
		t.Fatal(expGot(exp, res))
	}
}
