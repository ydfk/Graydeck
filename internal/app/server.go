package app

import (
	"fmt"
	"net/http"

	"mihomo-manager/internal/httpapi"
	"mihomo-manager/internal/manager"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	service, err := manager.New(manager.Config{
		DataDir:          cfg.DataDir,
		CoreTargetOS:     cfg.CoreTargetOS,
		CoreTargetArch:   cfg.CoreTargetArch,
		ControllerAddr:   cfg.ControllerAddr,
		RuntimeMixedPort: cfg.RuntimeMixedPort,
		RuntimeSecret:    cfg.RuntimeSecret,
		BaseConfigPath:   cfg.BaseConfigPath,
	})
	if err != nil {
		return nil, err
	}

	handler := httpapi.NewRouter(service)

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
