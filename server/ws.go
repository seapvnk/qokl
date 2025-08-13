package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/seapvnk/qokl/core"
	"github.com/olahol/melody"
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

		last := parts[len(parts)-1]
		parts[len(parts)-1] = strings.TrimSuffix(last, ".lisp")

		route := buildRoutePath(parts)

		r.Get(route, func(w http.ResponseWriter, r *http.Request) {
			core.WS.HandleRequest(w, r)
		})

		initHandlers(core.WS, path)

		return nil
	})
}

func initHandlers(m *melody.Melody, defaultPath string) {
	m.HandleConnect(func(s *melody.Session) {
		connID := core.ConnID(uuid.NewString())
		s.Set("conn_id", connID)
		core.RegisterConn(connID, s)

		headers := map[string]string{}
		for k, v := range s.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		input := map[string]any{
			"conn_id": connID,
			"headers": headers,
			"params":  extractParams(s.Request),
			"init":    false,
			"msg":     "",
		}

		vm := core.NewVM().UseCommunicationModule().UseStoreModule()
		vm.AddVariables(input)
		vm.Execute(defaultPath)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		m.BroadcastFilter(msg, func(q *melody.Session) bool {
			return q.Request.URL.Path == s.Request.URL.Path
		})

		sessions, _ := m.Sessions()

		for _, q := range sessions {
			if q.Request.URL.Path == s.Request.URL.Path {
				connIDValQ, _ := q.Get("conn_id")
				paramsQ := extractParams(q.Request)
				inputQ := map[string]any{
					"conn_id": connIDValQ,
					"headers": nil,
					"init":    true,
					"params":  paramsQ,
					"msg":     string(msg),
				}

				vm := core.NewVM().UseCommunicationModule().UseStoreModule()
				vm.AddVariables(inputQ)
				vm.Execute(defaultPath)
			}
		}
	})

	m.HandleDisconnect(func(s *melody.Session) {
		connIDVal, ok := s.Get("conn_id")
		if ok {
			core.UnregisterConn(connIDVal.(core.ConnID))
		}
	})
}

func extractParams(r *http.Request) map[string]string {
	ctx := chi.RouteContext(r.Context())
	params := map[string]string{}
	for i, key := range ctx.URLParams.Keys {
		params[key] = ctx.URLParams.Values[i]
	}
	return params
}

