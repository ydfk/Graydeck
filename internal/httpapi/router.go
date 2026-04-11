package httpapi

import (
	"encoding/json"
	"net/http"

	"mihomo-manager/internal/model"
	"mihomo-manager/internal/store"
)

type Router struct {
	store *store.MemoryStore
	mux   *http.ServeMux
}

func NewRouter(memStore *store.MemoryStore) http.Handler {
	router := &Router{
		store: memStore,
		mux:   http.NewServeMux(),
	}

	router.registerRoutes()
	return router.withCORS(router.mux)
}

func (r *Router) registerRoutes() {
	r.mux.HandleFunc("/api/system/status", r.handleSystemStatus)
	r.mux.HandleFunc("/api/subscriptions", r.handleSubscriptions)
	r.mux.HandleFunc("/api/zashboard/mode", r.handleZashboardMode)
	r.mux.HandleFunc("/api/healthz", r.handleHealth)
}

func (r *Router) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (r *Router) handleSystemStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, r.store.GetSystemStatus())
}

func (r *Router) handleSubscriptions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string][]model.Subscription{
		"items": r.store.GetSubscriptions(),
	})
}

func (r *Router) handleZashboardMode(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, r.store.GetZashboardMode())
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
