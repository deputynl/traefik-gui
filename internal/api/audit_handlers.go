package api

import "net/http"

func (s *Server) handleGetAudit(w http.ResponseWriter, r *http.Request) {
	entries := s.audit.Recent(200)
	writeJSON(w, http.StatusOK, map[string]any{"entries": entries})
}
