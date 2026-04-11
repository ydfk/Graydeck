package app

import (
	"fmt"
	"net/http"

	"mihomo-manager/internal/httpapi"
	"mihomo-manager/internal/store"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	memStore := store.NewMemoryStore(cfg.ZashboardMode)
	handler := httpapi.NewRouter(memStore)

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.ListenAddress,
			Handler: handler,
		},
	}, nil
}

func (s *Server) ListenAndServe() error {
	fmt.Printf("managerd listening on %s\n", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}
