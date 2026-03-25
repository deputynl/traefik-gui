package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultConfigPath  = "/etc/traefik/traefik.yml"
	DefaultDynamicDir  = "/etc/traefik/dynamic"
	DefaultAcmePath    = "/etc/traefik/acme.json"
	DefaultAPIURL      = "http://localhost:8080"
	DefaultPort        = "8888"
	DefaultUser        = "admin"
	DefaultPassword    = "admin"
)

// AppConfig holds all runtime configuration for traefik-gui.
type AppConfig struct {
	Port              string
	TraefikConfigPath string
	TraefikAPIURL     string
	GUIUser           string
	GUIPassword       string
	// Optional overrides — useful when acme.json / access log live at a
	// different path inside this container than in the Traefik container.
	AcmePathOverride      string
	AccessLogPathOverride string
}

// Load reads environment variables and returns the app config.
func Load() *AppConfig {
	return &AppConfig{
		Port:                  envOr("TRAEFIK_GUI_PORT", DefaultPort),
		TraefikConfigPath:     envOr("TRAEFIK_CONFIG_PATH", DefaultConfigPath),
		TraefikAPIURL:         envOr("TRAEFIK_API_URL", DefaultAPIURL),
		GUIUser:               envOr("TRAEFIK_GUI_USER", DefaultUser),
		GUIPassword:           envOr("TRAEFIK_GUI_PASSWORD", DefaultPassword),
		AcmePathOverride:      os.Getenv("TRAEFIK_ACME_PATH"),
		AccessLogPathOverride: os.Getenv("TRAEFIK_ACCESS_LOG_PATH"),
	}
}

// ResolvedPaths contains the file/directory paths discovered from traefik.yml
// (or the defaults when no config file is found).
type ResolvedPaths struct {
	StaticConfig      string `json:"staticConfig"`
	DynamicDir        string `json:"dynamicDir"`
	AcmePath          string `json:"acmePath"`
	AccessLogPath     string `json:"accessLogPath"`
	StaticConfigFound bool   `json:"staticConfigFound"`
	DynamicDirFound   bool   `json:"dynamicDirFound"`
	AcmePathFound     bool   `json:"acmePathFound"`
	AccessLogFound    bool   `json:"accessLogFound"`
}

// Resolve determines the dynamic dir and acme path from the static config path.
// It uses defaults if the file does not exist or the relevant fields are empty.
func Resolve(staticConfigPath string) *ResolvedPaths {
	rp := &ResolvedPaths{
		StaticConfig: staticConfigPath,
		DynamicDir:   DefaultDynamicDir,
		AcmePath:     DefaultAcmePath,
	}

	if _, err := os.Stat(staticConfigPath); err == nil {
		rp.StaticConfigFound = true
		// Derive defaults relative to the config file's directory.
		dir := filepath.Dir(staticConfigPath)
		rp.DynamicDir = filepath.Join(dir, "dynamic")
		rp.AcmePath = filepath.Join(dir, "acme.json")
	}

	rp.RefreshFoundFlags()
	return rp
}

// RefreshFoundFlags re-checks whether DynamicDir, AcmePath and AccessLogPath exist on disk.
// Call this after overriding the paths from traefik.yml.
func (rp *ResolvedPaths) RefreshFoundFlags() {
	if info, err := os.Stat(rp.DynamicDir); err == nil && info.IsDir() {
		rp.DynamicDirFound = true
	} else {
		rp.DynamicDirFound = false
	}
	if _, err := os.Stat(rp.AcmePath); err == nil {
		rp.AcmePathFound = true
	} else {
		rp.AcmePathFound = false
	}
	if rp.AccessLogPath != "" {
		if _, err := os.Stat(rp.AccessLogPath); err == nil {
			rp.AccessLogFound = true
		} else {
			rp.AccessLogFound = false
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
