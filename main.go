package main

import (
	"os"

	"github.com/seapvnk/qokl/application"
)

func main() {
	// Init server
	baseDir := "./"
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}

	addr := ":8080"
	if len(os.Args) > 2 {
		addr = os.Args[2]
	}

	// initialize application
	app := application.New(baseDir, addr)
	app.InitMemory()
	defer app.CloseMemory()

	// run server
	app.Run()
}
