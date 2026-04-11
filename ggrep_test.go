package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
)

const firstLine = "First : This is the 1st line"
const secondLine = "Second : This is the 2nd line"
const thirdLine = "Third : This is the 3rd line"

func initTmpFileWithThreeLines() *tempFile {
	tmp := createTemp()
	tmpPath := tmp.Name()
	if err := WriteLine(tmpPath, firstLine); err != nil {
		panic(err)
	}

	if err := WriteLine(tmpPath, secondLine); err != nil {
		panic(err)
	}

	if err := WriteLine(tmpPath, thirdLine); err != nil {
		panic(err)
	}

	return tmp
}

func initTmpWithLines(lines []string) *tempFile {
	tmp := createTemp()
	tmpPath := tmp.Name()

	for i, _ := range lines {
		WriteLine(tmpPath, lines[i])
	}

	return tmp
}

func TestFindAllIndexes(t *testing.T) {
	str := "One, Two, Three, Four, One"
	substr := "One"

	res := FindAllIndexes(str, substr)
	if len(res) != 2 {
		t.Fatal("expected two results")
	}

	if res[0] != 0 {
		t.Fatal("expected first result index to be 0")
	}

	if res[1] != 23 {
		t.Fatal("expected second result index to be 23")
	}

	if res = FindAllIndexes(str, ""); res != nil {
		t.Fatal("expected nil")
	}

	res = FindAllIndexes(str, "e")
	if len(res) != 4 {
		t.Fatal("expected 4 indexes")
	}

	firstEidx := 2
	if res[0] != firstEidx {
		t.Fatal("expected first e index to be 2")
	}

	secondEidx := strings.Index(str, "Three") + 3
	if res[1] != secondEidx {
		t.Fatalf("expected second e index to be %d", secondEidx)
	}

	thirdEidx := secondEidx + 1
	if res[2] != thirdEidx {
		t.Fatalf("expected third e index to be %d", thirdEidx)
	}

	fourthEidx := len(str) - 1
	if res[3] != fourthEidx {
		t.Fatalf("expected fourth e index to be %d", fourthEidx)
	}
}

type differentIndexesError struct {
	left, right []int
}

func (e differentIndexesError) Error() string {
	err := "Indexes are different, Left : ["
	for i, l := range e.left {
		err += fmt.Sprintf(" %d", l)
		if i == len(e.left)-1 {
			break
		}
		err += ","
	}
	err += " ]"

	err += ", Right : ["
	for i, r := range e.right {
		err += fmt.Sprintf(" %d", r)
		if i == len(e.right)-1 {
			break
		}
		err += ","
	}
	err += " ]"
	return err
}

func compareOccurence(left, right OccurenceInLine) error {
	if !slices.Equal(left.indexes, right.indexes) {
		return differentIndexesError{
			left:  left.indexes,
			right: right.indexes,
		}
	}

	if left.len != right.len {
		return errors.New("occurences len are not the same")
	}

	if left.lineNum != right.lineNum {
		return errors.New("line numes are not the same")
	}

	if left.lineText != right.lineText {
		return errors.New("line texts are not the same")
	}

	return nil
}

func compareOccurences(left, right []OccurenceInLine) error {
	if len(left) != len(right) {
		return errors.New("left and right are not at the same length")
	}

	for i := range left {
		err := compareOccurence(left[i], right[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func TestFindAllOccurencesNotExist(t *testing.T) {
	tmp := initTmpFileWithThreeLines()
	defer tmp.Close()

	// test finding no occurences
	results, err := FindAllOccurences(tmp.Name(), "NOT_EXIST")
	if err != nil {
		t.Fatal(err)
	}

	if results != nil {
		t.Fatal("expected result to be nil, but it's not")
	}
}

func TestFindSingleOccurencesFirstLineBegining(t *testing.T) {
	tmp := initTmpFileWithThreeLines()
	defer tmp.Close()

	str := "First"
	results, err := FindAllOccurences(tmp.Name(), str)
	if err != nil {
		t.Fatal(err)
	}

	exp := []OccurenceInLine{
		{
			lineNum:  1,
			len:      len(str),
			indexes:  []int{0},
			lineText: firstLine,
		},
	}

	if err := compareOccurences(exp, results); err != nil {
		t.Fatal(err)
	}
}

func TestFindSingleOccurencesFirstLineMiddle(t *testing.T) {
	tf := initTmpFileWithThreeLines()
	defer tf.Close()

	str := "1st"
	results, err := FindAllOccurences(tf.Name(), str)
	if err != nil {
		t.Fatal(err)
	}

	exp := []OccurenceInLine{
		{
			lineNum:  1,
			len:      len(str),
			indexes:  []int{strings.Index(firstLine, str)},
			lineText: firstLine,
		},
	}

	if err := compareOccurences(exp, results); err != nil {
		t.Fatal(err)
	}
}

func TestFindSingleOccurenceFirstLineEnd(t *testing.T) {
	lines := []string{
		"First line is best",
		"Second is less",
		"Third is worst",
	}

	tf := initTmpWithLines(lines)
	defer tf.Close()

	str := "best"
	res, err := FindAllOccurences(tf.Name(), str)
	if err != nil {
		t.Fatal(err)
	}

	exp := []OccurenceInLine{
		{
			lineNum:  1,
			len:      len(str),
			indexes:  []int{strings.Index(lines[0], str)},
			lineText: lines[0],
		},
	}

	if err = compareOccurences(exp, res); err != nil {
		t.Fatal(err)
	}
}

func newExpSingleOccurence(lineNum int, orgLine, str string) []OccurenceInLine {
	idx := strings.Index(orgLine, str)
	if idx == -1 {
		panic("bad testing, no str found in line")
	}
	return []OccurenceInLine{
		{
			lineNum:  lineNum,
			len:      len(str),
			indexes:  []int{idx},
			lineText: orgLine,
		},
	}
}

func testSingle(lineNum int, orgLine, str, filePath string) {
	exp := newExpSingleOccurence(lineNum, orgLine, str)
	res, err := FindAllOccurences(filePath, str)
	if err != nil {
		panic(err)
	}

	if err := compareOccurences(res, exp); err != nil {
		panic(err)
	}
}

func TestRestSingleOccurence(t *testing.T) {
	lines := []string{
		"I am the 1st line",
		"And I am the 2nd",
		"I am third",
		"4th Is the actual best",
	}

	tf := initTmpWithLines(lines)
	defer tf.Close()

	// 2nd line
	testSingle(2, lines[1], "And", tf.Name())   // begining
	testSingle(2, lines[1], "the 2", tf.Name()) // middle
	testSingle(2, lines[1], "2nd", tf.Name())   // end

	// 4th line
	testSingle(4, lines[3], "4th", tf.Name())    // begining
	testSingle(4, lines[3], "actual", tf.Name()) // middle
	testSingle(4, lines[3], "best", tf.Name())   // end

}

func TestMultipleOccurencesFirstLine(t *testing.T) {
	lines := []string{
		"1st line is often the best, becuase it's the 1st one. BTW, 1st lines are often the best ones",
		"2nd should be not too bad",
		"4th usually sucks",
		"Luckily the previous one was not really 4th",
		"But the previous one was, so it's the one that's actually sucks",
	}

	tf := initTmpWithLines(lines)
	defer tf.Close()

	firstLine := lines[0]
	str := "1st"
	firstIdx := strings.Index(firstLine, str)
	if firstIdx == -1 {
		panic("bad testing (1)")
	}

	secondIdx := strings.Index(firstLine[firstIdx+1:], str)
	if secondIdx == -1 {
		panic("bad testing (2)")
	}
	secondIdx += firstIdx + 1

	thirdIdx := strings.Index(firstLine[secondIdx+1:], str)
	if thirdIdx == -1 {
		panic("bad testing (3)")
	}
	thirdIdx += secondIdx + 1

	expIdxs := []int{firstIdx, secondIdx, thirdIdx}
	exp := []OccurenceInLine{
		{
			lineNum:  1,
			len:      len(str),
			indexes:  expIdxs,
			lineText: firstLine,
		},
	}

	res, err := FindAllOccurences(tf.Name(), str)
	if err != nil {
		t.Fatal(err)
	}

	if err := compareOccurences(exp, res); err != nil {
		t.Fatal(err)
	}
}

func TestMultipleOccurences(t *testing.T) {
	lines := []string{
		"first line",
		"second line",
		"third line (next one is fourth)",
		"fourth line",
	}

	tf := initTmpWithLines(lines)
	defer tf.Close()

	occrs, err := FindAllOccurences(tf.Name(), "line")
	if err != nil {
		t.Fatal(err)
	}

	if len(occrs) != 4 {
		t.Fatal("expected 4 occurences")
	}

	for i := range occrs {
		if occrs[i].lineNum != i+1 {
			t.Fatalf("different line number than expected, has %d, expected %d", occrs[i].lineNum, i+1)
		}

		if len(occrs[i].indexes) != 1 {
			t.Fatal("expected only 1 index for a line")
		}

		if strings.Index(occrs[i].lineText, "line") != occrs[i].indexes[0] {
			t.Fatal("expected different index")
		}
	}

	occrs, err = FindAllOccurences(tf.Name(), "fourth")
	if err != nil {
		t.Fatal(err)
	}

	if len(occrs) != 2 {
		t.Fatal("expected 2 occurences")
	}

	if occrs[0].lineNum != 3 {
		t.Fatal("expected 1st occurence to be on line 3")
	}

	if occrs[1].lineNum != 4 {
		t.Fatal("expected 2nd occurence to be on line 4")
	}

	if len(occrs[0].indexes) != 1 {
		t.Fatal("expected only one index")
	}

	expIdx := strings.Index(lines[2], "fourth")
	if occrs[0].indexes[0] != expIdx {
		t.Fatal("expected different index here")
	}

	if len(occrs[1].indexes) != 1 {
		t.Fatal("expected one index")
	}

	if occrs[1].indexes[0] != strings.Index(lines[3], "fourth") {
		t.Fatal("expected different index")
	}
}

func BenchmarkHugeFile(b *testing.B) {
	tf := createTemp()
	defer tf.Close()

	b.Log("generating file")
	if err := GeneraetHugeFileMb(10, 'r', 500, tf.Name()); err != nil {
		b.Fatal(err)
	}
	b.Log("counting lines")
	linesCount, err := CountLines(tf.Name())
	if err != nil {
		b.Fatal(err)
	}

	b.Logf("lines : %d", linesCount)
	// every 500 lines insert some string
	for l := 0; l < linesCount; l += 500 {
		if err := InsertToFile(tf.Name(), "test string", l, 50, false); err != nil {
			b.Fatal(err)
		}
		b.Log("write")
	}
	b.Log("init")

	i := 0
	for b.Loop() {
		i++
	}
}
