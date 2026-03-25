package api

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"traefik-gui/internal/auth"
	"traefik-gui/internal/config"
	"traefik-gui/internal/traefik"
)

// Server is the HTTP server for traefik-gui.
type Server struct {
	cfg   *config.AppConfig
	paths *config.ResolvedPaths
	mux   *http.ServeMux
	webFS fs.FS
	auth  *auth.Manager
}

// New creates and configures a new Server.
func New(cfg *config.AppConfig, webFiles fs.FS) *Server {
	distFS, err := fs.Sub(webFiles, "web/dist")
	if err != nil {
		log.Fatalf("could not sub web/dist: %v", err)
	}

	s := &Server{
		cfg:  cfg,
		auth: auth.NewManager(cfg.GUIUser, cfg.GUIPassword),
	}
	s.webFS = distFS
	s.refreshPaths()
	s.mux = http.NewServeMux()
	s.registerRoutes()
	return s
}

// Start begins listening on the configured port.
func (s *Server) Start() error {
	addr := ":" + s.cfg.Port
	log.Printf("traefik-gui listening on http://0.0.0.0%s", addr)
	log.Printf("  static config : %s (found=%v)", s.cfg.TraefikConfigPath, s.paths.StaticConfigFound)
	log.Printf("  dynamic dir   : %s", s.paths.DynamicDir)
	log.Printf("  acme.json     : %s", s.paths.AcmePath)
	log.Printf("  traefik api   : %s", s.cfg.TraefikAPIURL)
	log.Printf("  auth user     : %s", s.cfg.GUIUser)
	// Wrap the whole mux with the auth gate.
	return http.ListenAndServe(addr, s.authGate(s.mux))
}

// authGate protects all /api/* routes with session auth.
// /auth/* and static assets are always public.
func (s *Server) authGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			// Static files and /auth/* — no auth required.
			next.ServeHTTP(w, r)
			return
		}
		user, ok := s.auth.FromRequest(r)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"error":"unauthorized"}`)
			return
		}
		// Slide the cookie window on every authenticated API call.
		s.auth.SetCookie(w, user)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) refreshPaths() {
	rp := config.Resolve(s.cfg.TraefikConfigPath)

	if rp.StaticConfigFound {
		cfg, err := traefik.Load(s.cfg.TraefikConfigPath)
		if err != nil {
			log.Printf("warn: parsing traefik config for path resolution: %v", err)
		} else {
			rp.DynamicDir, rp.AcmePath = traefik.ResolvePaths(s.cfg.TraefikConfigPath, cfg)
		}
	}

	s.paths = rp
}

func (s *Server) registerRoutes() {
	// Auth routes (no session required).
	s.mux.HandleFunc("POST /auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /auth/logout", s.handleLogout)
	s.mux.HandleFunc("GET /auth/check", s.handleAuthCheck)

	// API routes (protected by authGate).
	s.mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.handleGetConfig(w, r)
		case http.MethodPut:
			s.handleSaveConfig(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/dynamic", s.handleDynamic)
	s.mux.HandleFunc("/api/dynamic/{file}", s.handleDynamicFile)
	s.mux.HandleFunc("/api/traefik/", s.handleTraefikProxy)

	// SPA fallback — serve index.html for all unknown paths.
	fileServer := http.FileServer(http.FS(s.webFS))
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(s.webFS, path); err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fileServer.ServeHTTP(w, r2)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
