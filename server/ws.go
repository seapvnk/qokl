package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/seapvnk/qokl/core"
)

func (server *Server) setupChannels(r chi.Router) error {
	wsPath := filepath.Join(server.baseDir, wsDir)
	return filepath.Walk(wsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if !strings.HasSuffix(path, ".lisp") {
			return nil
		}

		rel, _ := filepath.Rel(wsPath, path)
		parts := strings.Split(rel, string(filepath.Separator))

		// remove .lisp extension from last part
		last := parts[len(parts)-1]
		parts[len(parts)-1] = strings.TrimSuffix(last, ".lisp")

		route := buildRoutePath(parts)

		r.Get(route, wrapWSHandler(path))
		return nil
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func wrapWSHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		connID := core.ConnID(uuid.NewString())
		core.RegisterConn(connID, conn)
		defer func() {
			core.UnregisterConn(connID)
			conn.Close()
		}()

		headers := map[string]string{}
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		input := map[string]any{
			"conn_id": connID,
			"headers": headers,
			"params":  extractParams(r),
			"message": "",
		}

		vm := core.NewVM().UseCommunicationModule()
		vm.AddVariables(input)
		vm.Execute(path)

		// Loop for incoming messages
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			input["message"] = string(msg)
			vm.AddVariables(input)
			vm.Execute(path)
		}
	}
}

func extractParams(r *http.Request) map[string]string {
	ctx := chi.RouteContext(r.Context())
	params := map[string]string{}
	for i, key := range ctx.URLParams.Keys {
		params[key] = ctx.URLParams.Values[i]
	}
	return params
}
