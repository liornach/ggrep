package main

// func insertStringToFile(f *os.File, line, column int, str string) error {
// 	in, err := os.Open(f.Name())
// 	if err != nil {
// 		return err
// 	}
// 	defer in.Close()

// 	tmp, err := os.CreateTemp("", "tmpfile")
// 	if err != nil {
// 		return err
// 	}
// 	defer tmp.Close()

// 	reader := bufio.NewReader(in)
// 	curLineNum := 0

// 	for {
// 		l, err := reader.ReadString('\n')
// 		if err != nil {
// 			if err == io.EOF { // write empty lines to the file
// 				l = "\n"
// 			}

// 			panic(err)
// 		}

// 		if curLineNum == line {
// 			if column > len(l) {
// 				l, err := extendString(l, column-len(l), ' ')
// 				if err != nil {
// 					panic(err)
// 				}

// 				l = l[:column] + str + l[column:]
// 			}
// 		}

// 		if _, err := tmp.WriteString(l); err != nil {
// 			panic(err)
// 		}

// 		curLineNum++
// 	}

// 	if err := os.Rename(tmp.Name(), f.Name()); err != nil {
// 		panic(err)
// 	}

// 	return nil
// }

// func TestFind(t *testing.T) {
// 	const oneGb = 1
// 	f, err := GeneraetHugeFile(oneGb)
// 	if err != nil {
// 		panic(err)
// 	}

// 	defer f.Close()
// 	defer os.Remove(f.Name())
// 	insertStringToFile(f, 5, 3, "hello")
// }
