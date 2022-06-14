package main

import (
	"flag"
	"os"

	"git.fractalqb.de/fractalqb/gomk"
	"git.fractalqb.de/fractalqb/gomk/mktask"
)

var (
	goBuild = gomk.CmdDef{
		Name: "go",
		Args: []string{"build", "--trimpath", "-ldflags", "-s -w"}, // What about -a
	}
	update = false

	must = gomk.LogMust
)

// func init() {

// 	gomk.NewCmdTask(must, prj, "generate", "go", "generate", "./...").
// 		DependOn(tStringer.Name(), tVersioner.Name())

// 	mktask.Def(TEST, func(dir *gomk.WDir) {
// 		gomk.Exec(dir, "go", "test", "./...")
// 	})

// 	mktask.Def(BUILD, func(dir *gomk.WDir) {
// 		gomk.Exec(dir, "go", buildCmd...)
// 		gomk.Step(dir.Cd("edeh"), "build edeh", func(dir *gomk.WDir) {
// 			gomk.Exec(dir, "go", buildCmd...)
// 			gomk.Step(dir.Cd("plugin"), "build plugins", func(dir *gomk.WDir) {
// 				gomk.Exec(dir.Cd("echo"), "go", "build", "--trimpath")
// 				gomk.Exec(dir.Cd("speak"), "go", "build", "--trimpath")
// 				gomk.Exec(dir.Cd("screenshot"), "go", "build", "--trimpath")
// 			})
// 		})
// 	}, GEN)

// 	mktask.Def(DEPS, func(dir *gomk.Dir) {
// 		task.DepsGraph(dir.Build(), update)
// 	}, TEST)
// }

func main() {
	flag.BoolVar(&update, "u", false, "Get tool update")
	flag.Parse()

	prj := gomk.NewProject(must, &gomk.Config{Env: os.Environ()})

	tStringer := mktask.NewGetStringer(must, prj, update)
	tVersioner := mktask.NewGetVersioner(must, prj, update)

	tGoGen := gomk.NewCmdTask(must, prj, "generate", "go", "generate", "./...").
		DependOn(tStringer.Name(), tVersioner.Name())

	gomk.NewCmdDefTask(must, prj, "edeh", goBuild).WorkDir("edeh").
		DependOn(tGoGen.Name())

	if len(flag.Args()) == 0 {
		gomk.Build(prj, "edeh")
	} else {
		for _, target := range flag.Args() {
			gomk.Build(prj, target)
		}
	}
}
