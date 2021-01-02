package main

import (
	"flag"
	"log"

	"github.com/CmdrVasquess/watched/edeh/plugin"
	"github.com/CmdrVasquess/watched/examples/speak"
)

var speaker speak.Speaker

func main() {
	flag.BoolVar(&speaker.Verbose, "v", false, "Print messages to stdout")
	flag.StringVar(&speaker.Exe, "tts", "espeak-ng", "Set TTS CLI executable")
	flag.Parse()
	log.Println("start edeh plugin 'speak'")
	defer speaker.Close()
	plugin.RunRecv(&speaker, nil)
}
