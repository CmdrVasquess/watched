package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/CmdrVasquess/watched/edeh/plugin"
	"github.com/CmdrVasquess/watched/examples/speak"
)

var speaker speak.Speaker

func readCfg(name string) {
	log.Printf("read config: %s", name)
	rd, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer rd.Close()
	dec := json.NewDecoder(rd)
	if err = dec.Decode(&speaker); err != nil {
		log.Fatal(err)
	}
}

func main() {
	fVerb := flag.Bool("v", false, "Verbose output")
	flag.Parse()
	log.Println("start edeh plugin 'speak'")
	for _, arg := range flag.Args() {
		readCfg(arg)
	}
	speaker.Verbose = speaker.Verbose || *fVerb
	defer speaker.Close()
	plugin.RunRecv(&speaker, nil)
}
