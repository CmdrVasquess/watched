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
)

func init() {
	tasks.Def(TOOLS, func(dir *gomk.WDir) {
		task.GetStringer(dir.Build())
		task.GetVersioner(dir.Build())
	})

	tasks.Def(GEN, func(dir *gomk.WDir) {
		dir.Exec("go", "generate", "./...")
	}, TOOLS)

	tasks.Def(TEST, func(dir *gomk.WDir) {
		dir.Exec("go", "test", "./...")
	})

	tasks.Def(BUILD, func(dir *gomk.WDir) {
		dir.Exec("go", buildCmd...)
		dir.Cd("edeh").Do("build edeh", func(dir *gomk.WDir) {
			dir.Exec("go", buildCmd...)
			dir.Cd("plugin").Do("build plugins", func(dir *gomk.WDir) {
				dir.Cd("echo").Exec("go", "build", "--trimpath")
				dir.Cd("speak").Exec("go", "build", "--trimpath")
				dir.Cd("screenshot").Exec("go", "build", "--trimpath")
			})
		})
	}, GEN)

	tasks.Def(DEPS, func(dir *gomk.WDir) {
		task.DepsGraph(dir.Build())
	}, TEST)
}

func main() {
	fCDir := flag.String("C", "", "change working dir")
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
