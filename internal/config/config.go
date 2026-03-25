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
}

// Load reads environment variables and returns the app config.
func Load() *AppConfig {
	return &AppConfig{
		Port:              envOr("TRAEFIK_GUI_PORT", DefaultPort),
		TraefikConfigPath: envOr("TRAEFIK_CONFIG_PATH", DefaultConfigPath),
		TraefikAPIURL:     envOr("TRAEFIK_API_URL", DefaultAPIURL),
		GUIUser:           envOr("TRAEFIK_GUI_USER", DefaultUser),
		GUIPassword:       envOr("TRAEFIK_GUI_PASSWORD", DefaultPassword),
	}
}

// ResolvedPaths contains the file/directory paths discovered from traefik.yml
// (or the defaults when no config file is found).
type ResolvedPaths struct {
	StaticConfig string `json:"staticConfig"`
	DynamicDir   string `json:"dynamicDir"`
	AcmePath     string `json:"acmePath"`
	// Whether the static config file was actually found on disk.
	StaticConfigFound bool `json:"staticConfigFound"`
}

// Resolve determines the dynamic dir and acme path from the static config path.
// It uses defaults if the file does not exist or the relevant fields are empty.
// The traefik parser fills in DynamicDir and AcmePath; this function only
// populates the base fields so callers always get a valid struct.
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

	return rp
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
