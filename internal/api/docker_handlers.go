package api

import (
	"net/http"

	"traefik-gui/internal/docker"
)

func (s *Server) handleRestartTraefik(w http.ResponseWriter, r *http.Request) {
	name := s.cfg.TraefikContainerName
	if name == "" {
		writeError(w, http.StatusServiceUnavailable, "TRAEFIK_CONTAINER_NAME is not set")
		return
	}
	if err := docker.RestartContainer(name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "restart", "restarted container "+name)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleGetDocker(w http.ResponseWriter, r *http.Request) {
	if !docker.Available() {
		writeJSON(w, http.StatusOK, map[string]any{"containers": []any{}, "available": false})
		return
	}

	containers, err := docker.ListContainers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"containers": containers, "available": true})
}
