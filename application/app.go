package application

import (
	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/server"
	"github.com/seapvnk/qokl/storage"
	"github.com/seapvnk/qokl/tasks"
)

type Application struct {
	baseDir string
	addr    string
	vm      *core.VM
	server  *server.Server
	tasks   *tasks.Listener
}

func New(baseDir, addr string) *Application {
	vm := core.NewVM()
	app := &Application{
		baseDir: baseDir,
		addr:    addr,
		vm:      vm,
	}

	core.InitWS()
	app.server = server.New(baseDir)
	app.tasks = tasks.New(baseDir)

	return app
}

func (app *Application) Run() {
	go app.server.Start(app.addr)
	app.tasks.Run()
}

func (app *Application) GetHTTPHandler() *server.Server {
	return app.server
}

func (app *Application) GetTaskListener() *tasks.Listener {
	return app.tasks
}

func (app *Application) InitMemory() {
	core.OpenStore()
	storage.OpenDB(app.baseDir)
}

func (app *Application) CloseMemory() {
	core.CloseStore()
	storage.CloseDB()
}
