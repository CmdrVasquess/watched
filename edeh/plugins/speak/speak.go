package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"git.fractalqb.de/fractalqb/ggja"
)

var (
	ttsExe    string
	isJournal = []byte("Journal ")
	msgq      = make(chan string, 32)
)

func speaker() {
	for txt := range msgq {
		cmd := exec.Command(ttsExe, txt)
		cmd.Run()
	}
}

func main() {
	flag.StringVar(&ttsExe, "tts", "espeak-ng", "Set TTS CLI executable")
	flag.Parse()
	go speaker()
	scn := bufio.NewScanner(os.Stdin)
	event := make(map[string]interface{})
	for scn.Scan() {
		if !bytes.HasPrefix(scn.Bytes(), isJournal) {
			continue
		}
		if err := json.Unmarshal(scn.Bytes()[len(isJournal):], &event); err != nil {
			log.Println(err)
			continue
		}
		evt := ggja.Obj{Bare: event}
		switch evt.MStr("event") {
		case "ReceiveText":
			msgq <- fmt.Sprintf("From \"%s\": %s", evt.MStr("From"), evt.MStr("Message"))
		}
	}
	close(msgq)
}
