package main

import (
	"log"
	"os"

	"github.com/seapvnk/qokl/server"
)

func main() {
	baseDir := "./"
	if len(os.Args) > 1 {
		baseDir = os.Args[1]
	}
	
	addr := ":8080"
	if len(os.Args) > 2 {
		addr = os.Args[2]
	}

	srv := server.New(baseDir)
	if err := srv.Start(addr); err != nil {
		log.Fatal(err)
	}
}

