package main

import (
	"log"
	"os"

	"git.fractalqb.de/fractalqb/gomk"
	"git.fractalqb.de/fractalqb/gomk/task"
)

func main() {
	build, _ := gomk.NewBuild("", os.Environ())
	log.Printf("project root: %s\n", build.PrjRoot)
	gomk.Try(func() {
		task.GetStringer(build)
		task.GetVersioner(build)
		build.WDir().Exec("go", "generate", "./...")
		build.WDir().Exec("go", "test", "./...")
		build.WDir().Cd("edeh").Do("build edeh", func(dir *gomk.WDir) {
			dir.Exec("go", "build", "--trimpath")
			dir.Cd("plugin").Do("build plugins", func(dir *gomk.WDir) {
				dir.Cd("echo").Exec("go", "build", "--trimpath")
				dir.Cd("speak").Exec("go", "build", "--trimpath")
				dir.Cd("screenshot").Exec("go", "build", "--trimpath")
			})
		})
		task.DepsGraph(build)
	})
}
