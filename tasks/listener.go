package tasks

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/seapvnk/qokl/core"
)

type Listener struct {
	baseDir string
}

func New(baseDir string) *Listener {
	listener := &Listener{
		baseDir: baseDir,
	}
	return listener
}

func (listener *Listener) Run() {
	// listener main loop
	for {
		tasksPath := filepath.Join(listener.baseDir, tasksDir)
		filepath.Walk(tasksPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			rel, _ := filepath.Rel(tasksPath, path)
			queue := strings.ToLower(rel)
			queue = strings.Replace(queue, ".lisp", "", 1)

			var msg []byte
			msg, err = core.StoreDequeue(queue)
			if err == nil {
				go handleTask(path, msg)
			}

			return nil
		})
	}
}

func handleTask(queuePath string, msg []byte) {
	vm := core.NewVM()
	vm.AddVariables(map[string]any{
		"msg": msg,
	})

	_, err := vm.Execute(queuePath)
	if err != nil {
		log.Println(err.Error())
	}
}
