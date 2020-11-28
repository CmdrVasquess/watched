package main

import (
	"flag"

	"github.com/CmdrVasquess/watched/edeh/plugin"
)

var speaker Speaker

func main() {
	flag.BoolVar(&speaker.Verbose, "v", false, "Print messages to stdout")
	flag.StringVar(&speaker.Exe, "tts", "espeak-ng", "Set TTS CLI executable")
	flag.Parse()
	plugin.RunRecv(&speaker, nil)
}
