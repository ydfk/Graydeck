package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type updateRequest struct {
	Source string `json:"source"`
	URL    string `json:"url"`
}

func (r *Router) handleSystemCoreUpdate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload updateRequest
	if req.ContentLength > 0 {
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	source := strings.TrimSpace(payload.Source)
	if source == "" || source == "auto" {
		status, err := r.service.UpdateCore(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		writeJSON(w, http.StatusOK, status)
		return
	}

	if source != "url" {
		http.Error(w, "unsupported update source", http.StatusBadRequest)
		return
	}

	status, err := r.service.InstallCoreFromURL(req.Context(), payload.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleSystemCoreUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tempPath, fileName, cleanup, err := readUploadedTempFile(req, "file", "graydeck-core-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer cleanup()

	status, err := r.service.InstallCoreFromUpload(req.Context(), tempPath, fileName)
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

	var payload updateRequest
	if req.ContentLength > 0 {
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	source := strings.TrimSpace(payload.Source)
	if source == "" || source == "auto" {
		status, err := r.service.UpdateZashboard(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		writeJSON(w, http.StatusOK, status)
		return
	}

	if source != "url" {
		http.Error(w, "unsupported update source", http.StatusBadRequest)
		return
	}

	status, err := r.service.InstallZashboardFromURL(req.Context(), payload.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func (r *Router) handleZashboardUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	tempPath, fileName, cleanup, err := readUploadedTempFile(req, "file", "graydeck-zashboard-*")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer cleanup()

	status, err := r.service.InstallZashboardFromUpload(tempPath, fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func readUploadedTempFile(req *http.Request, fieldName, pattern string) (string, string, func(), error) {
	if err := req.ParseMultipartForm(256 << 20); err != nil {
		return "", "", nil, err
	}

	file, header, err := req.FormFile(fieldName)
	if err != nil {
		return "", "", nil, err
	}
	defer file.Close()

	tempFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", "", nil, err
	}
	tempPath := tempFile.Name()

	if _, err := io.Copy(tempFile, file); err != nil {
		tempFile.Close()
		_ = os.Remove(tempPath)
		return "", "", nil, err
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", "", nil, err
	}

	fileName := safeUploadedFileName(header.Filename, filepath.Base(tempPath))
	return tempPath, fileName, func() { _ = os.Remove(tempPath) }, nil
}

func safeUploadedFileName(fileName, fallback string) string {
	baseName := strings.TrimSpace(filepath.Base(fileName))
	if baseName == "" || baseName == "." {
		return fallback
	}

	return baseName
}
