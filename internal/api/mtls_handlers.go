package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"traefik-gui/internal/mtls"
	"traefik-gui/internal/traefik"
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

	publicServices, err := store.LoadPublicServices()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Read certResolver name from static config so TCP router can use ACME.
	certResolver := ""
	if s.paths.StaticConfigFound {
		if cfg, err := traefik.Load(s.cfg.TraefikConfigPath); err == nil {
			for name := range cfg.CertResolvers {
				certResolver = name
				break
			}
		}
	}

	content := buildMTLSYAML(store.CACertPath(), publicServices, certResolver)

	path := filepath.Join(dynamicDir, "mtls.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-apply", "wrote "+path)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// buildMTLSYAML generates the mtls.yml content using TCP routers for reliable mTLS enforcement.
//
// Architecture:
//   - TCP routers operate at the TLS handshake level (before HTTP), so HostSNI("*")
//     with tls.options: mtls reliably enforces client certificate requirements for every
//     connection on the websecuremtls entrypoint — regardless of the HTTP path or host.
//   - Public exception routers use specific HostSNI("host") rules. A specific HostSNI
//     match has higher priority than HostSNI("*"), so these hosts bypass mTLS.
//   - After TLS termination, TCP traffic (now plaintext HTTP) is forwarded to
//     websecure-internal (127.0.0.1:8444) — a plain HTTP entrypoint. Docker containers
//     must register their routers on websecure-internal to be reachable via mTLS.
//   - Note: path-based exceptions are not supported at the TCP level. Only host-based
//     exceptions can be configured here.
func buildMTLSYAML(caPath string, publicServices []mtls.PublicService, certResolver string) string {
	var sb strings.Builder

	certResolverLine := ""
	if certResolver != "" {
		certResolverLine = fmt.Sprintf("        certResolver: %s\n", certResolver)
	}

	sb.WriteString("tcp:\n  routers:\n")

	// Public exception routers — specific HostSNI rules win over HostSNI(`*`).
	// tls.options: public skips the client certificate requirement for these hosts.
	for _, svc := range publicServices {
		fmt.Fprintf(&sb, "    public-%s:\n", svc.ID)
		fmt.Fprintf(&sb, "      entryPoints:\n        - websecuremtls\n")
		fmt.Fprintf(&sb, "      rule: \"HostSNI(`%s`)\"\n", svc.Host)
		fmt.Fprintf(&sb, "      tls:\n        options: public\n")
		fmt.Fprintf(&sb, "%s", certResolverLine)
		fmt.Fprintf(&sb, "      service: forward-to-websecure-internal\n")
	}

	// Catch-all mTLS router — HostSNI(`*`) matches every connection not claimed
	// by a more specific rule. Enforces mutual TLS at the handshake level.
	sb.WriteString("    mtls-catchall:\n")
	sb.WriteString("      entryPoints:\n        - websecuremtls\n")
	sb.WriteString("      rule: \"HostSNI(`*`)\"\n")
	sb.WriteString("      tls:\n        options: mtls\n")
	fmt.Fprintf(&sb, "%s", certResolverLine)
	sb.WriteString("      service: forward-to-websecure-internal\n")

	// TCP forwarding service — sends plaintext HTTP to the internal entrypoint.
	// This port is bound to 127.0.0.1 only and is not publicly accessible.
	sb.WriteString("  services:\n")
	sb.WriteString("    forward-to-websecure-internal:\n")
	sb.WriteString("      loadBalancer:\n")
	sb.WriteString("        servers:\n")
	sb.WriteString("          - address: \"127.0.0.1:8444\"\n")

	// TLS options.
	// mtls: enforces mutual TLS (client certificate required).
	// public: standard TLS only, no client certificate — used for public exceptions.
	fmt.Fprintf(&sb, "tls:\n  options:\n")
	fmt.Fprintf(&sb, "    mtls:\n      clientAuth:\n")
	fmt.Fprintf(&sb, "        caFiles:\n          - %s\n", caPath)
	fmt.Fprintf(&sb, "        clientAuthType: RequireAndVerifyClientCert\n")
	fmt.Fprintf(&sb, "    public:\n")
	fmt.Fprintf(&sb, "      minVersion: VersionTLS12\n")

	return sb.String()
}

// GET /api/mtls/public — list all public service exceptions.
func (s *Server) handleListPublicServices(w http.ResponseWriter, r *http.Request) {
	store := s.mtlsStore()
	services, err := store.LoadPublicServices()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, services)
}

// POST /api/mtls/public — add a public service exception.
// Body: {"host": "...", "path": "...", "description": "..."}
func (s *Server) handleAddPublicService(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Host        string `json:"host"`
		Path        string `json:"path"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	body.Host = strings.TrimSpace(body.Host)
	body.Path = strings.TrimSpace(body.Path)
	body.Description = strings.TrimSpace(body.Description)
	if body.Host == "" {
		writeError(w, http.StatusBadRequest, "host is required")
		return
	}
	if body.Path != "" && !strings.HasPrefix(body.Path, "/") {
		writeError(w, http.StatusBadRequest, "path must start with /")
		return
	}

	store := s.mtlsStore()
	services, err := store.LoadPublicServices()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	svc := mtls.PublicService{
		ID:          sanitizeRouterID(body.Host + body.Path),
		Host:        body.Host,
		Path:        body.Path,
		Description: body.Description,
	}
	for _, existing := range services {
		if existing.ID == svc.ID {
			writeError(w, http.StatusConflict, "a public exception with this host/path already exists")
			return
		}
	}

	services = append(services, svc)
	if err := store.SavePublicServices(services); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-public-add", "added public exception: "+body.Host+body.Path)
	writeJSON(w, http.StatusOK, svc)
}

// PUT /api/mtls/public/{id} — update a public service exception.
// Body: {"host": "...", "path": "...", "description": "..."}
func (s *Server) handleUpdatePublicService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Host        string `json:"host"`
		Path        string `json:"path"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	body.Host = strings.TrimSpace(body.Host)
	body.Path = strings.TrimSpace(body.Path)
	body.Description = strings.TrimSpace(body.Description)
	if body.Host == "" {
		writeError(w, http.StatusBadRequest, "host is required")
		return
	}
	if body.Path != "" && !strings.HasPrefix(body.Path, "/") {
		writeError(w, http.StatusBadRequest, "path must start with /")
		return
	}

	store := s.mtlsStore()
	services, err := store.LoadPublicServices()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	found := false
	for i, svc := range services {
		if svc.ID == id {
			services[i].Host = body.Host
			services[i].Path = body.Path
			services[i].Description = body.Description
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "public service not found")
		return
	}

	if err := store.SavePublicServices(services); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-public-update", "updated public exception: "+id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// DELETE /api/mtls/public/{id} — delete a public service exception.
func (s *Server) handleDeletePublicService(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	store := s.mtlsStore()
	services, err := store.LoadPublicServices()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filtered := make([]mtls.PublicService, 0, len(services))
	found := false
	for _, svc := range services {
		if svc.ID == id {
			found = true
		} else {
			filtered = append(filtered, svc)
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "public service not found")
		return
	}

	if err := store.SavePublicServices(filtered); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.audit.Log(s.userFromRequest(r), "mtls-public-delete", "deleted public exception: "+id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// sanitizeRouterID converts a raw string (host+path) to a valid Traefik router name.
func sanitizeRouterID(raw string) string {
	result := make([]byte, 0, len(raw))
	for _, c := range strings.ToLower(raw) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result = append(result, byte(c))
		} else {
			result = append(result, '-')
		}
	}
	return strings.Trim(string(result), "-")
}

// mtlsApplied returns true if mtls.yml exists in the dynamic config directory.
func (s *Server) mtlsApplied() bool {
	if s.paths.DynamicDir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(s.paths.DynamicDir, "mtls.yml"))
	return err == nil
}
