package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/CmdrVasquess/watched"
	"github.com/CmdrVasquess/watched/edeh/edehnet"
)

type dump struct {
	*bufio.Writer
}

func (d dump) OnJournalEvent(e watched.JounalEvent) error {
	d.Write(e.Event)
	fmt.Fprintln(d)
	return d.Flush()
}

func (d dump) OnStatusEvent(e watched.StatusEvent) error {
	d.Write(e.Event)
	fmt.Fprintln(d)
	return d.Flush()
}

func (d dump) Close() error { return nil }

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <listen address>\n", os.Args[0])
		os.Exit(1)
	}
	toStdout := dump{bufio.NewWriter(os.Stdout)}
	nrcv := edehnet.Receiver{os.Args[1]}
	fmt.Fprintln(os.Stderr, nrcv.Run(toStdout))
}
