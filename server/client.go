package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/seapvnk/qokl/core"
	"github.com/seapvnk/qokl/parser"
)

func (server *Server) setupClient(r chi.Router) error {
	clientPath := filepath.Join(server.baseDir, clientDir)
	return filepath.Walk(clientPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, _ := filepath.Rel(clientPath, path)
		parts := strings.Split(rel, string(filepath.Separator))
		ext := strings.ToLower(filepath.Ext(path))

		route := buildRoutePath(parts[:len(parts)-1])

		var routeParts []string
		if len(parts) > 2 {
			routeParts = parts[1 : len(parts)-1]
		} else {
			routeParts = []string{}
		}

		handler := injectRouteVars(wrapClientHandler(path, ext), routeParts)

		switch ext {
		case ".html":
			r.Get(route, handler)
		case ".lisp":
			r.Get(route, handler)
			r.Post(route, handler)
		default:
			log.Printf("Skipping unsupported client file: %s", path)
		}

		return nil
	})
}

func wrapClientHandler(path string, ext string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch ext {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeFile(w, r, path)

		case ".lisp":
			// route vars
			vars := GetRouteVars(r.Context())

			// headers
			headers := map[string]string{}
			for k, v := range r.Header {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}

			formData := map[string]interface{}{}
			if err := r.ParseForm(); err == nil {
				for k, v := range r.Form {
					if len(v) == 1 {
						formData[k] = v[0]
					} else {
						formData[k] = v
					}
				}
			}

			// VM with variables
			vm := core.NewVM().UseCommunicationModule().UseStoreModule().UseClientModule()
			vm.AddVariables(map[string]any{
				"method":  r.Method,
				"params":  vars,
				"headers": headers,
				"form":    formData,
			})

			result, err := vm.Execute(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if result.Error != nil {
				http.Error(w, result.Error.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			val, _ := parser.SexpToGo(result.Value)
			io.WriteString(w, fmt.Sprintf("%v", val))
		}
	}
}
