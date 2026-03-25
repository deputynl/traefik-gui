package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"traefik-gui/internal/config"
	"traefik-gui/internal/traefik"
)

// configResponse is returned by GET /api/config.
type configResponse struct {
	Paths        *config.ResolvedPaths `json:"paths"`
	StaticConfig *traefik.StaticConfig `json:"staticConfig,omitempty"`
	RawConfig    string                `json:"rawConfig,omitempty"`
	TraefikAPI   string                `json:"traefikApiUrl"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	resp := configResponse{
		Paths:      s.paths,
		TraefikAPI: s.cfg.TraefikAPIURL,
	}

	if s.paths.StaticConfigFound {
		cfg, err := traefik.Load(s.cfg.TraefikConfigPath)
		if err != nil {
			log.Printf("warn: could not parse traefik config: %v", err)
		} else {
			resp.StaticConfig = cfg
		}
		raw, err := traefik.ReadRaw(s.cfg.TraefikConfigPath)
		if err != nil {
			log.Printf("warn: could not read raw traefik config: %v", err)
		} else {
			resp.RawConfig = raw
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleSaveConfig handles PUT /api/config.
// Accepts either raw YAML (Content-Type: text/plain) or a JSON StaticConfig.
func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read body")
		return
	}

	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/json") {
		// Form save: JSON → Go struct → YAML on disk.
		var cfg traefik.StaticConfig
		if err := json.Unmarshal(body, &cfg); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
		if err := traefik.Save(s.cfg.TraefikConfigPath, &cfg); err != nil {
			writeError(w, http.StatusInternalServerError, "could not save: "+err.Error())
			return
		}
	} else {
		// YAML tab save: validate and write raw content.
		if err := traefik.WriteRaw(s.cfg.TraefikConfigPath, string(body)); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	s.refreshPaths()
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

// handleTraefikProxy proxies requests to the Traefik API.
func (s *Server) handleTraefikProxy(w http.ResponseWriter, r *http.Request) {
	target := s.cfg.TraefikAPIURL + r.URL.Path[len("/api/traefik"):]
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	resp, err := http.Get(target) //nolint:noctx
	if err != nil {
		writeError(w, http.StatusBadGateway, "traefik api unreachable: "+err.Error())
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body) //nolint:errcheck
}

// handleStatus returns GUI health + whether the Traefik API is reachable.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	traefikOK := false
	resp, err := http.Get(s.cfg.TraefikAPIURL + "/api/overview")
	if err == nil {
		resp.Body.Close()
		traefikOK = resp.StatusCode == http.StatusOK
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"gui":     "ok",
		"traefik": traefikOK,
	})
}

// --- helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
