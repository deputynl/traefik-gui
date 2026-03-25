package api

import (
	"net/http"

	"traefik-gui/internal/traefik"
)

func (s *Server) handleGetCerts(w http.ResponseWriter, r *http.Request) {
	if !s.paths.AcmePathFound {
		writeJSON(w, http.StatusOK, map[string]any{"certs": []any{}, "available": false})
		return
	}

	certs, err := traefik.LoadAcmeCerts(s.paths.AcmePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"certs": certs, "available": true})
}
