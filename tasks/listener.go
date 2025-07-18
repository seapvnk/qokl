package tasks

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/seapvnk/qokl/core"
)

type Listener struct {
	baseDir string
	closed  chan struct{}
}

func New(baseDir string) *Listener {
	return &Listener{
		baseDir: baseDir,
		closed:  make(chan struct{}),
	}
}

func (listener *Listener) Close() {
	close(listener.closed)
}

func (listener *Listener) Run() {
	tasksPath := filepath.Join(listener.baseDir, tasksDir)

	for {
		select {
		case <-listener.closed:
			return
		default:
			_ = filepath.Walk(tasksPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}

				rel, _ := filepath.Rel(tasksPath, path)
				queue := strings.TrimSuffix(strings.ToLower(rel), ".lisp")

				msg, err := core.StoreDequeue(queue)
				if err == nil && len(msg) > 0 {
					go handleTask(path, msg)
				}

				return nil
			})

			time.Sleep(10 * time.Millisecond)
		}
	}
}

func handleTask(queuePath string, msg []byte) {
	vm := core.NewVM().UseStoreModule()
	vm.AddVariables(map[string]any{
		"msg": msg,
	})

	_, err := vm.Execute(queuePath)
	if err != nil {
		log.Printf("[task - %s] error: %s\n", queuePath, err.Error())
	}
}
