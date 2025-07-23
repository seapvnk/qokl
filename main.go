package main

import (
	"os"

	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/server"
	"github.com/seapvnk/qokl/storage"
	"github.com/seapvnk/qokl/tasks"
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

	// Setup kv store and entity database
	core.OpenStore()
	storage.OpenDB(baseDir)
	defer core.CloseStore()
	defer storage.CloseDB()

	// run server
	srv := server.New(baseDir)
	go srv.Start(addr)

	// process tasks
	listener := tasks.New(baseDir)
	listener.Run()
}
