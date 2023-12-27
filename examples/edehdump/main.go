package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/edeh/edehnet"
)

type dump struct {
	*bufio.Writer
}

func (d dump) OnJournalEvent(e watched.JounalEvent) error {
	fmt.Fprintf(d.Writer, "%s:%d\t", e.File, e.EventNo)
	d.Write(e.Event)
	fmt.Fprintln(d)
	return d.Flush()
}

func (d dump) OnStatusEvent(e watched.StatusEvent) error {
	fmt.Fprintf(d.Writer, "%s: ", e.Type.String())
	d.Write(e.Event)
	fmt.Fprintln(d)
	return d.Flush()
}

func (d dump) Close() error { return nil }

func flags() {
	fLog := flag.String("log", "", "Set log level")
	flag.Parse()
	if *fLog != "" {
		qblog.DefaultConfig.ParseFlag(*fLog)
	}
}

func main() {
	flags()
	if flag.NArg() != 1 {
		fmt.Printf("Usage: %s <listen address>\n", os.Args[0])
		os.Exit(1)
	}
	toStdout := dump{bufio.NewWriter(os.Stdout)}
	nrcv := edehnet.Receiver{Listen: flag.Arg(0)}
	for {
		nrcv.Run(toStdout)
	}
}
