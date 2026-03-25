package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"traefik-gui/internal/traefik"
)

// handleDynamic dispatches GET (list) and POST (create) on /api/dynamic.
func (s *Server) handleDynamic(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.dynamicList(w, r)
	case http.MethodPost:
		s.dynamicCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDynamicFile dispatches GET / PUT / DELETE on /api/dynamic/{file}.
func (s *Server) handleDynamicFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.dynamicGet(w, r)
	case http.MethodPut:
		s.dynamicSave(w, r)
	case http.MethodDelete:
		s.dynamicDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /api/dynamic — return FileSummary list.
func (s *Server) dynamicList(w http.ResponseWriter, r *http.Request) {
	summaries, err := traefik.ListDynamic(s.paths.DynamicDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

// GET /api/dynamic/{file} — return raw YAML + parsed struct.
func (s *Server) dynamicGet(w http.ResponseWriter, r *http.Request) {
	name, ok := s.validFile(w, r)
	if !ok {
		return
	}
	path := filepath.Join(s.paths.DynamicDir, name)

	raw, err := traefik.ReadRaw(path)
	if err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "file not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	parsed, _ := traefik.LoadDynamic(path)
	writeJSON(w, http.StatusOK, map[string]any{
		"name":   name,
		"raw":    raw,
		"parsed": parsed,
	})
}

// PUT /api/dynamic/{file} — save raw YAML content.
func (s *Server) dynamicSave(w http.ResponseWriter, r *http.Request) {
	name, ok := s.validFile(w, r)
	if !ok {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 512*1024))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read body")
		return
	}

	path := filepath.Join(s.paths.DynamicDir, name)
	if err := traefik.WriteRaw(path, string(body)); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// DELETE /api/dynamic/{file} — remove the file.
func (s *Server) dynamicDelete(w http.ResponseWriter, r *http.Request) {
	name, ok := s.validFile(w, r)
	if !ok {
		return
	}

	path := filepath.Join(s.paths.DynamicDir, name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			writeError(w, http.StatusNotFound, "file not found")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// POST /api/dynamic — create a new service file from a ServiceSpec.
func (s *Server) dynamicCreate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read body")
		return
	}

	var spec traefik.ServiceSpec
	if err := json.Unmarshal(body, &spec); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if spec.Name == "" || spec.Hostname == "" || spec.BackendURL == "" {
		writeError(w, http.StatusBadRequest, "name, hostname and backendUrl are required")
		return
	}
	if !isValidName(spec.Name) {
		writeError(w, http.StatusBadRequest, "name must contain only letters, digits, hyphens and underscores")
		return
	}

	filename := spec.Name + ".yml"
	path := filepath.Join(s.paths.DynamicDir, filename)

	if _, err := os.Stat(path); err == nil {
		writeError(w, http.StatusConflict, "file already exists: "+filename)
		return
	}

	cfg := traefik.GenerateServiceConfig(spec)
	if err := traefik.SaveDynamic(path, cfg); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"name": filename, "status": "created"})
}

// validFile extracts and validates the {file} path parameter.
func (s *Server) validFile(w http.ResponseWriter, r *http.Request) (string, bool) {
	name := r.PathValue("file")
	if name == "" || strings.Contains(name, "/") || strings.Contains(name, "..") {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return "", false
	}
	// Only allow .yml and .yml.bak files.
	if !strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yml.bak") {
		writeError(w, http.StatusBadRequest, "only .yml and .yml.bak files are allowed")
		return "", false
	}
	return name, true
}

// isValidName checks that a service name is safe for use as a filename.
func isValidName(name string) bool {
	if name == "" || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
