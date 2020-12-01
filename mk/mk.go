package main

import (
	"log"
	"os/exec"

	"git.fractalqb.de/fractalqb/gomk"
)

func main() {
	build, _ := gomk.NewBuild("")
	log.Printf("project root: %s\n", build.PrjRoot)
	gomk.Try(func() {
		build.WDir().Exec("go", "test", "./...")
		build.WDir().Cd("edeh").Do("build edeh", func(dir *gomk.WDir) {
			dir.Exec("go", "build")
			dir.Cd("plugin").Do("build plugins", func(dir *gomk.WDir) {
				dir.Cd("echo").Exec("go", "build")
				dir.Cd("speak").Exec("go", "build")
			})
		})
		build.WDir().ExecPipe(
			exec.Command("go", "mod", "graph"),
			exec.Command("modgraphviz"),
			exec.Command("dot", "-Tsvg", "-o", "depgraph.svg"),
		)
	})
}
