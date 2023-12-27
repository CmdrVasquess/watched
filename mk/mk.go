package main

import (
	"flag"
	"os"

	"git.fractalqb.de/fractalqb/gomk"
	"git.fractalqb.de/fractalqb/gomk/mktask"
)

var (
	goBuild = gomk.CommandDef{
		Name: "go",
		Args: []string{"build", "--trimpath", "-ldflags", "-s -w"}, // What about -a
	}
	update = false
	onErr  = gomk.LogMust

	plugins = []string{"echo", "screenshot", "speak"}
)

func main() {
	flag.BoolVar(&update, "u", false, "Get tool update")
	flag.Parse()

	prj := gomk.NewProject(onErr, &gomk.Config{Env: os.Environ()})

	tStringer := mktask.NewGetStringer(onErr, prj, update)
	tVersioner := mktask.NewGetVersioner(onErr, prj, update)

	tGoGen := gomk.NewCmdTask(onErr, prj, "generate", "go", "generate", "./...").
		DependOn(tStringer.Name(), tVersioner.Name())

	cmds := gomk.NewNopTask(onErr, prj, "cmds")

	task := gomk.NewCmdDefTask(onErr, prj, "edeh", goBuild).
		WorkDir("edeh").
		DependOn(tGoGen.Name())
	cmds.DependOn(task.Name())

	for _, p := range plugins {
		task = gomk.NewCmdDefTask(onErr, prj, p, goBuild).
			WorkDir("edeh/plugin", p).
			DependOn(tGoGen.Name())
		cmds.DependOn(task.Name())
	}

	task = gomk.NewCmdDefTask(onErr, prj, "jreplay", goBuild).
		WorkDir("jreplay")
	cmds.DependOn(task.Name())

	if len(flag.Args()) == 0 {
		gomk.Build(prj, "cmds")
	} else {
		for _, target := range flag.Args() {
			gomk.Build(prj, target)
		}
	}
}
