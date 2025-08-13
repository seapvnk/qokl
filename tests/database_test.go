package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/zygomys/v9/zygo"
	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/storage"
)

// Checks if insert can be performed
func TestDBSpecs(t *testing.T) {
	core.InitWS()
	filepath.Walk("./specs", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		dbPath := storage.OpenDB("./")
		defer os.RemoveAll(dbPath)
		vm := core.NewVM()

		var sexpr *core.ZygResult

		sexpr, err = vm.Execute(path)
		if err != nil {
			t.Errorf("syntax error on %s", path)
		}

		sexp := sexpr.Value
		sexpBool, ok := sexp.(*zygo.SexpBool)
		if !ok {
			t.Errorf("assertion failed on file: %s, receiving %T, error: %s", path, sexp, sexpr.Error.Error())
		}

		if !sexpBool.Val {
			t.Errorf("assertion failed on returning value: %s", path)
		}

		return nil
	})
}
