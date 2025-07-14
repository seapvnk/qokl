package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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

	server.Router.Route("/api", func(r chi.Router) {
		_ = server.setupApi(r)
	})

	server.Router.Route("/channels", func(r chi.Router) {
		_ = server.setupChannels(r)
	})

	return server
}

func (server *Server) Start(addr string) error {
	return http.ListenAndServe(addr, server.Router)
}
