package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/seapvnk/qokl/core"
)

type Server struct {
	baseDir string
	Router  chi.Router
}

func New(baseDir string) *Server {
	server := &Server{
		baseDir: baseDir,
		Router:  chi.NewRouter(),
	}

	// query endpoint
	server.Router.Post("/", queryHandler)

	// discover api routes
	server.Router.Route("/api", func(r chi.Router) {
		_ = server.setupApi(r)
	})

	// discover channels
	server.Router.Route("/channels", func(r chi.Router) {
		_ = server.setupChannels(r)
	})

	return server
}

func (server *Server) Start(addr string) {
	err := http.ListenAndServe(addr, server.Router)
	if err != nil {
		log.Fatal(err)
	}
}

// server base routes handlers
func queryHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, errReadingBody := io.ReadAll(r.Body)
	if errReadingBody != nil {
		http.Error(w, "unable to read body", http.StatusBadRequest)
		return
	}

	vm := core.NewVM()
	result, err := vm.ExecuteString(string(bodyBytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if result.Error == nil {
		response, _ := core.SexpToGo(result.Value)
		json.NewEncoder(w).Encode(response)
	} else {
		response := ErrorResponse{Error: result.Error.Error()}
		json.NewEncoder(w).Encode(response)
	}
}
