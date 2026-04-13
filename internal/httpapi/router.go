package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"mihomo-manager/internal/manager"
	"mihomo-manager/internal/model"
)

type Router struct {
	service *manager.Service
	mux     *http.ServeMux
}

func NewRouter(service *manager.Service) http.Handler {
	router := &Router{
		service: service,
		mux:     http.NewServeMux(),
	}

	router.registerRoutes()
	return router.withCORS(router.mux)
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/api/healthz", r.handleHealth)
	r.mux.HandleFunc("/api/system/status", r.handleSystemStatus)
	r.mux.HandleFunc("/api/system/refresh", r.handleSystemRefresh)
	r.mux.HandleFunc("/api/system/start", r.handleSystemStart)
	r.mux.HandleFunc("/api/system/restart", r.handleSystemRestart)
	r.mux.HandleFunc("/api/system/stop", r.handleSystemStop)
	r.mux.HandleFunc("/api/system/core/update", r.handleSystemCoreUpdate)
	r.mux.HandleFunc("/api/zashboard/update", r.handleZashboardUpdate)
	r.mux.HandleFunc("/api/subscriptions", r.handleSubscriptions)
	r.mux.HandleFunc("/api/subscriptions/create", r.handleSubscriptionCreate)
	r.mux.HandleFunc("/api/subscriptions/update", r.handleSubscriptionUpdate)
	r.mux.HandleFunc("/api/subscriptions/activate", r.handleSubscriptionActivate)
	r.mux.HandleFunc("/api/subscriptions/preview", r.handleSubscriptionPreview)
	r.mux.HandleFunc("/api/logs", r.handleLogs)
	r.mux.Handle("/api/clash/", r.buildClashProxy())
	r.mux.Handle("/api/clash", r.buildClashProxy())
	r.mux.Handle("/zashboard-ui/", r.buildZashboardHandler())
	r.mux.Handle("/zashboard-ui", r.buildZashboardHandler())
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleSystemStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, r.service.Status())
}

func (r *Router) handleSystemRefresh(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.service.RefreshAll(req.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, r.service.Status())
}

func (r *Router) handleSystemStart(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status, err := r.service.StartRuntime(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleSystemRestart(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status, err := r.service.RestartRuntime(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleSystemStop(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, r.service.StopRuntime())
}

func (r *Router) handleSystemCoreUpdate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status, err := r.service.UpdateCore(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleZashboardUpdate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	status, err := r.service.UpdateZashboard(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleSubscriptions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]model.Subscription{
		"items": r.service.Subscriptions(),
	})
}

func (r *Router) handleSubscriptionCreate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Name         string `json:"name"`
		URL          string `json:"url"`
		SyncInterval string `json:"syncInterval"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := r.service.CreateSubscription(req.Context(), payload.Name, payload.URL, payload.SyncInterval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleSubscriptionUpdate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		URL          string `json:"url"`
		SyncInterval string `json:"syncInterval"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := r.service.UpdateSubscription(req.Context(), payload.ID, payload.Name, payload.URL, payload.SyncInterval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleSubscriptionActivate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	item, err := r.service.ActivateSubscription(req.Context(), payload.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (r *Router) handleSubscriptionPreview(w http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	preview, err := r.service.PreviewSubscription(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, preview)
}

func (r *Router) handleLogs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]model.LogEntry{
		"items": r.service.Logs(),
	})
}

func (r *Router) buildZashboardHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/zashboard-ui" {
			http.Redirect(w, req, "/zashboard-ui/", http.StatusTemporaryRedirect)
			return
		}

		root := r.service.ZashboardRoot()
		indexPath := filepath.Join(root, "index.html")
		if _, err := os.Stat(indexPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, "zashboard not ready", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.StripPrefix("/zashboard-ui/", http.FileServer(http.Dir(root))).ServeHTTP(w, req)
	})
}

func (r *Router) buildClashProxy() http.Handler {
	target, err := url.Parse(r.service.ControllerURL())
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		})
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	director := proxy.Director
	proxy.Director = func(req *http.Request) {
		director(req)

		trimmedPath := strings.TrimPrefix(req.URL.Path, "/api/clash")
		if trimmedPath == "" {
			trimmedPath = "/"
		}

		req.URL.Path = trimmedPath
		req.URL.RawPath = trimmedPath
		req.Host = target.Host

		if secret := r.service.RuntimeSecret(); secret != "" {
			req.Header.Set("Authorization", "Bearer "+secret)
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, proxyErr error) {
		http.Error(w, proxyErr.Error(), http.StatusBadGateway)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/clash" {
			http.Redirect(w, req, "/api/clash/", http.StatusTemporaryRedirect)
			return
		}

		proxy.ServeHTTP(w, req)
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (r *Router) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
	})
}
