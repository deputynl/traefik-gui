package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"traefik-gui/internal/mtls"
)

func (s *Server) mtlsStore() *mtls.Store {
	return mtls.NewStore(filepath.Dir(s.cfg.TraefikConfigPath))
}

// GET /api/mtls — returns CA status and client list.
func (s *Server) handleMTLSStatus(w http.ResponseWriter, r *http.Request) {
	store := s.mtlsStore()

	var caExpires *string
	if store.CAExists() {
		caCert, _, err := store.LoadCA()
		if err == nil {
			t := caCert.NotAfter.Format("2006-01-02")
			caExpires = &t
		}
	}

	clients, err := store.Clients()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"caExists":  store.CAExists(),
		"caExpires": caExpires,
		"clients":   clients,
		"applied":   s.mtlsApplied(),
	})
}

// POST /api/mtls/ca — generate (or regenerate) the CA.
func (s *Server) handleMTLSGenerateCA(w http.ResponseWriter, r *http.Request) {
	store := s.mtlsStore()
	if err := store.RemoveAllClients(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := store.GenerateCA(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-ca", "generated new mTLS CA")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/mtls/ca — delete the CA and all client certificates.
func (s *Server) handleMTLSDeleteCA(w http.ResponseWriter, r *http.Request) {
	store := s.mtlsStore()
	if err := store.RemoveAllClients(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := store.DeleteCA(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-ca", "deleted mTLS CA and all client certificates")
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GET /api/mtls/ca/download — download ca.crt.
func (s *Server) handleMTLSDownloadCA(w http.ResponseWriter, r *http.Request) {
	store := s.mtlsStore()
	data, err := os.ReadFile(store.CACertPath())
	if err != nil {
		http.Error(w, "CA not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", `attachment; filename="ca.crt"`)
	w.Write(data)
}

// POST /api/mtls/clients — issue a new client certificate.
// Body: {"name": "...", "password": "..."}
func (s *Server) handleMTLSIssueClient(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if strings.TrimSpace(body.Password) == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	store := s.mtlsStore()
	if !store.CAExists() {
		writeError(w, http.StatusConflict, "CA not generated yet")
		return
	}

	entry, err := store.IssueClient(body.Name, body.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-issue", "issued client cert: "+body.Name)
	writeJSON(w, http.StatusOK, entry)
}

// GET /api/mtls/clients/{id}/download — download ZIP for a client cert.
func (s *Server) handleMTLSDownloadClient(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store := s.mtlsStore()

	clients, err := store.Clients()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var name string
	for _, c := range clients {
		if c.ID == id {
			name = c.Name
			break
		}
	}
	if name == "" {
		http.Error(w, "client not found", http.StatusNotFound)
		return
	}

	filename := fmt.Sprintf("mtls-%s.zip", id)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if err := store.WriteClientZip(id, w); err != nil {
		// Headers already sent, nothing we can do.
		return
	}
}

// DELETE /api/mtls/clients/{id} — revoke a client certificate.
func (s *Server) handleMTLSRevokeClient(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store := s.mtlsStore()

	clients, _ := store.Clients()
	var name string
	for _, c := range clients {
		if c.ID == id {
			name = c.Name
			break
		}
	}

	if err := store.RemoveClient(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-revoke", "revoked client cert: "+name)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /api/mtls/apply — write mtls.yml to the dynamic config directory.
func (s *Server) handleMTLSApply(w http.ResponseWriter, r *http.Request) {
	store := s.mtlsStore()
	if !store.CAExists() {
		writeError(w, http.StatusConflict, "CA not generated yet")
		return
	}

	dynamicDir := s.paths.DynamicDir
	if dynamicDir == "" {
		writeError(w, http.StatusConflict, "dynamic config directory not configured")
		return
	}

	content := fmt.Sprintf(`tls:
  options:
    mtls:
      clientAuth:
        caFiles:
          - %s
        clientAuthType: RequireAndVerifyClientCert
`, store.CACertPath())

	path := filepath.Join(dynamicDir, "mtls.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-apply", "wrote "+path)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// mtlsApplied returns true if mtls.yml exists in the dynamic config directory.
func (s *Server) mtlsApplied() bool {
	if s.paths.DynamicDir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(s.paths.DynamicDir, "mtls.yml"))
	return err == nil
}
