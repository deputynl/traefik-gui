package api

import (
	"net/http"

	"traefik-gui/internal/docker"
)

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
