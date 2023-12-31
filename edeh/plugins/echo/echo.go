package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	fTStart time.Duration
	fTMsg   time.Duration
	fTQuit  time.Duration
)

func main() {
	flag.DurationVar(&fTStart, "ts", 0, "startup delay")
	flag.DurationVar(&fTMsg, "tm", 0, "message delay")
	flag.DurationVar(&fTQuit, "tq", 0, "shutdown on quit delay")
	flag.Parse()
	log.Printf("echo: startup with delay %s", fTStart)
	time.Sleep(fTStart)
	wd, _ := os.Getwd()
	log.Println("echo: plugin running in", wd)
	scn := bufio.NewScanner(os.Stdin)
	for scn.Scan() {
		fmt.Printf("Echo:[%s]\n", scn.Text())
		if fTMsg > 0 {
			log.Printf("echo: message delay %s…", fTMsg)
			time.Sleep(fTMsg)
			log.Println("echo: message delay done")
		}
	}
	log.Printf("echo: input closed, shutting down with delay %s…", fTQuit)
	time.Sleep(fTQuit)
	log.Println("echo: bye!")
}
