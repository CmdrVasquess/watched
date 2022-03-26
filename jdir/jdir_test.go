package jdir

// func TestSplitEmpty(t *testing.T) {
// 	rd := str.NewReader("")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	scnCount := 0
// 	for scn.Scan() {
// 		scnCount++
// 	}
// 	if scnCount > 0 {
// 		t.Errorf("scanned too many lines: %d, expected 0", scnCount)
// 	}
// }

// func TestSplit1LineNoEndl(t *testing.T) {
// 	rd := str.NewReader("foo bar baz")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	scnCount := 0
// 	for scn.Scan() {
// 		scnCount++
// 	}
// 	if scnCount > 0 {
// 		t.Errorf("scanned too many lines: %d, expected 0", scnCount)
// 	}
// }

// func ExampleSplit1LineNL() {
// 	rd := str.NewReader("foo bar baz\n")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	for scn.Scan() {
// 		fmt.Printf("SCN:[%s]\n", scn.Text())
// 	}
// 	// Output:
// 	// SCN:[foo bar baz]
// }

// func ExampleSplit1LineCR() {
// 	rd := str.NewReader("foo bar baz\r")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	for scn.Scan() {
// 		fmt.Printf("SCN:[%s]\n", scn.Text())
// 	}
// 	// Output:
// 	// SCN:[foo bar baz]
// }

// func ExampleSplit1LineCRNL() {
// 	rd := str.NewReader("foo bar baz\r\n")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	for scn.Scan() {
// 		fmt.Printf("SCN:[%s]\n", scn.Text())
// 	}
// 	// Output:
// 	// SCN:[foo bar baz]
// }

// func ExampleSplit3LineNoCRNL() {
// 	rd := str.NewReader("foo\r\nbar\r\nbaz")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	for scn.Scan() {
// 		fmt.Printf("SCN:[%s]\n", scn.Text())
// 	}
// 	// Output:
// 	// SCN:[foo]
// 	// SCN:[bar]
// }

// func ExampleSplit3LineCRNL() {
// 	rd := str.NewReader("foo\r\nbar\r\nbaz\n")
// 	scn := bufio.NewScanner(rd)
// 	scn.Split(splitLogLines)
// 	for scn.Scan() {
// 		fmt.Printf("SCN:[%s]\n", scn.Text())
// 	}
// 	// Output:
// 	// SCN:[foo]
// 	// SCN:[bar]
// 	// SCN:[baz]
// }

// func TestJournalBulk(t *testing.T) {
// 	os.RemoveAll(t.Name())
// 	os.Mkdir(t.Name(), 0700)
// 	count := 0
// 	if testing.Verbose() {
// 		internal.LogCfg.SetLevel(c4hgol.Trace)
// 	}
// 	jd := JournalDir{
// 		Dir:      t.Name(),
// 		PerJLine: func(_ []byte) { count++ },
// 		Stop:     MakeStopChan(),
// 	}
// 	go jd.Watch("")
// 	time.Sleep(1 * time.Second)
// 	const jFile = "Journal.012345678901.01.log"
// 	pack.CopyFile(filepath.Join(t.Name(), jFile), jFile, nil)
// 	time.Sleep(2 * time.Second)
// 	jd.Stop <- watched.Stop
// 	<-jd.Stop
// 	if count != 190 {
// 		t.Errorf("expected 190 events, got %d", count)
// 	}
// }
