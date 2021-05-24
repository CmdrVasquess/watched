package main

import (
	"flag"
	"log"
	"os"

	"git.fractalqb.de/fractalqb/gomk"
	"git.fractalqb.de/fractalqb/gomk/task"
)

type target = string

const (
	TOOLS target = "tools"
	GEN   target = "gen"
	TEST  target = "test"
	BUILD target = "build"
	DEPS  target = "deps"
)

var (
	tasks    = make(gomk.Tasks)
	buildCmd = []string{"build", "--trimpath"}
	update   = false
)

func init() {
	tasks.Def(TOOLS, func(dir *gomk.WDir) {
		task.GetStringer(dir.Build(), update)
		task.GetVersioner(dir.Build(), update)
	})

	tasks.Def(GEN, func(dir *gomk.WDir) {
		gomk.Exec(dir, "go", "generate", "./...")
	}, TOOLS)

	tasks.Def(TEST, func(dir *gomk.WDir) {
		gomk.Exec(dir, "go", "test", "./...")
	})

	tasks.Def(BUILD, func(dir *gomk.WDir) {
		gomk.Exec(dir, "go", buildCmd...)
		gomk.Step(dir.Cd("edeh"), "build edeh", func(dir *gomk.WDir) {
			gomk.Exec(dir, "go", buildCmd...)
			gomk.Step(dir.Cd("plugin"), "build plugins", func(dir *gomk.WDir) {
				gomk.Exec(dir.Cd("echo"), "go", "build", "--trimpath")
				gomk.Exec(dir.Cd("speak"), "go", "build", "--trimpath")
				gomk.Exec(dir.Cd("screenshot"), "go", "build", "--trimpath")
			})
		})
	}, GEN)

	tasks.Def(DEPS, func(dir *gomk.WDir) {
		task.DepsGraph(dir.Build(), update)
	}, TEST)
}

func main() {
	fCDir := flag.String("C", "", "change working dir")
	flag.BoolVar(&update, "update", update, "Check tools for updates")
	flag.Parse()
	if *fCDir != "" {
		if err := os.Chdir(*fCDir); err != nil {
			log.Fatal(err)
		}
	}
	build, _ := gomk.NewBuild("", os.Environ())
	log.Printf("project root: %s\n", build.PrjRoot)
	if len(flag.Args()) == 0 {
		tasks.Run(BUILD, build.WDir())
		tasks.Run(TEST, build.WDir())
	} else {
		for _, task := range flag.Args() {
			tasks.Run(task, build.WDir())
		}
	}
}
