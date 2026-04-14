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
	"time"

	"mihomo-manager/internal/manager"
	"mihomo-manager/internal/model"
)

type Router struct {
	service  *manager.Service
	mux      *http.ServeMux
	sessions *sessionStore
}

func NewRouter(service *manager.Service) http.Handler {
	router := &Router{
		service:  service,
		mux:      http.NewServeMux(),
		sessions: newSessionStore(7 * 24 * time.Hour),
	}

	router.registerRoutes()
	return router.withCORS(router.mux)
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/api/healthz", r.handleHealth)
	r.mux.HandleFunc("/api/auth/status", r.handleAuthStatus)
	r.mux.HandleFunc("/api/auth/login", r.handleAuthLogin)
	r.mux.HandleFunc("/api/auth/logout", r.handleAuthLogout)
	r.mux.HandleFunc("/api/system/status", r.requireAuthAPI(r.handleSystemStatus))
	r.mux.HandleFunc("/api/system/config/current", r.requireAuthAPI(r.handleSystemCurrentConfig))
	r.mux.HandleFunc("/api/system/config/default", r.requireAuthAPI(r.handleSystemDefaultConfig))
	r.mux.HandleFunc("/api/system/refresh", r.requireAuthAPI(r.handleSystemRefresh))
	r.mux.HandleFunc("/api/system/start", r.requireAuthAPI(r.handleSystemStart))
	r.mux.HandleFunc("/api/system/restart", r.requireAuthAPI(r.handleSystemRestart))
	r.mux.HandleFunc("/api/system/stop", r.requireAuthAPI(r.handleSystemStop))
	r.mux.HandleFunc("/api/system/core/update", r.requireAuthAPI(r.handleSystemCoreUpdate))
	r.mux.HandleFunc("/api/system/core/upload", r.requireAuthAPI(r.handleSystemCoreUpload))
	r.mux.HandleFunc("/api/zashboard/update", r.requireAuthAPI(r.handleZashboardUpdate))
	r.mux.HandleFunc("/api/zashboard/upload", r.requireAuthAPI(r.handleZashboardUpload))
	r.mux.HandleFunc("/api/subscriptions", r.requireAuthAPI(r.handleSubscriptions))
	r.mux.HandleFunc("/api/subscriptions/create", r.requireAuthAPI(r.handleSubscriptionCreate))
	r.mux.HandleFunc("/api/subscriptions/update", r.requireAuthAPI(r.handleSubscriptionUpdate))
	r.mux.HandleFunc("/api/subscriptions/sync", r.requireAuthAPI(r.handleSubscriptionSync))
	r.mux.HandleFunc("/api/subscriptions/activate", r.requireAuthAPI(r.handleSubscriptionActivate))
	r.mux.HandleFunc("/api/subscriptions/preview", r.requireAuthAPI(r.handleSubscriptionPreview))
	r.mux.HandleFunc("/api/logs", r.requireAuthAPI(r.handleLogs))
	r.mux.Handle("/api/clash/", r.buildClashProxy())
	r.mux.Handle("/api/clash", r.buildClashProxy())
	r.mux.Handle("/zashboard-ui/", r.buildZashboardHandler())
	r.mux.Handle("/zashboard-ui", r.buildZashboardHandler())
	r.mux.Handle("/", r.buildWebHandler())
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleSystemStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, r.service.Status())
}

func (r *Router) handleSystemCurrentConfig(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	preview, err := r.service.RuntimeConfigPreview()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "运行中配置不存在", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, preview)
}

func (r *Router) handleSystemDefaultConfig(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, r.service.DefaultConfig())
	case http.MethodPost:
		var payload model.DefaultRuntimeConfig
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		status, err := r.service.UpdateDefaultConfig(req.Context(), payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, http.StatusOK, status)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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

func (r *Router) handleSubscriptionSync(w http.ResponseWriter, req *http.Request) {
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

	item, err := r.service.SyncSubscription(req.Context(), payload.ID)
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
		if !r.isAuthenticated(req) {
			http.Error(w, "未登录", http.StatusUnauthorized)
			return
		}

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
		if !r.isAuthenticated(req) {
			http.Error(w, "未登录", http.StatusUnauthorized)
			return
		}

		if req.URL.Path == "/api/clash" {
			http.Redirect(w, req, "/api/clash/", http.StatusTemporaryRedirect)
			return
		}

		proxy.ServeHTTP(w, req)
	})
}

func (r *Router) buildWebHandler() http.Handler {
	webRoot := strings.TrimSpace(r.service.WebRoot())
	if webRoot == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "web ui not configured", http.StatusNotFound)
		})
	}

	fileServer := http.FileServer(http.Dir(webRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/api/") || strings.HasPrefix(req.URL.Path, "/zashboard-ui") {
			http.NotFound(w, req)
			return
		}

		trimmedPath := strings.TrimPrefix(req.URL.Path, "/")
		if trimmedPath == "" {
			trimmedPath = "index.html"
		}

		if !r.isAuthenticated(req) {
			if r.shouldServePublicAsset(webRoot, trimmedPath) {
				fileServer.ServeHTTP(w, req)
				return
			}

			if req.URL.Path == "/login" || req.URL.Path == "/login/" {
				r.serveWebIndex(w, req, webRoot)
				return
			}

			http.Redirect(w, req, "/login", http.StatusTemporaryRedirect)
			return
		}

		if req.URL.Path == "/login" || req.URL.Path == "/login/" {
			http.Redirect(w, req, "/", http.StatusTemporaryRedirect)
			return
		}

		if info, statErr := os.Stat(filepath.Join(webRoot, filepath.Clean(trimmedPath))); statErr == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, req)
			return
		}

		if strings.Contains(filepath.Base(trimmedPath), ".") {
			http.NotFound(w, req)
			return
		}

		r.serveWebIndex(w, req, webRoot)
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func (r *Router) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if origin := req.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func (r *Router) shouldServePublicAsset(webRoot, trimmedPath string) bool {
	if trimmedPath == "" || trimmedPath == "index.html" {
		return false
	}

	info, err := os.Stat(filepath.Join(webRoot, filepath.Clean(trimmedPath)))
	if err != nil || info.IsDir() {
		return false
	}

	return strings.Contains(filepath.Base(trimmedPath), ".")
}

func (r *Router) serveWebIndex(w http.ResponseWriter, req *http.Request, webRoot string) {
	indexPath := filepath.Join(webRoot, "index.html")
	if _, statErr := os.Stat(indexPath); statErr != nil {
		http.Error(w, "web ui not ready", http.StatusServiceUnavailable)
		return
	}

	http.ServeFile(w, req, indexPath)
}
