package main

import (
	"runtime"

	"github.com/robinliucn/filebrowser/v2/cmd"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cmd.Execute()
}
