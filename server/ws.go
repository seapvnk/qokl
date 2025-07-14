package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (server *Server) setupChannels(r chi.Router) error {
	wsPath := filepath.Join(server.baseDir, wsDir)
	return filepath.Walk(wsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		rel, _ := filepath.Rel(wsDir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		route := buildRoutePath(parts)

		r.Get(route, wrapWSHandler(path))
		return nil
	})
}

func wrapWSHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("websocket upgrade error:", err)
			return
		}
		defer conn.Close()

		conn.WriteMessage(websocket.TextMessage, []byte("WebSocket for "+path))
	}
}

