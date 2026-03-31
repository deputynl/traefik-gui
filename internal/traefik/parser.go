package traefik

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a traefik static config file.
// Returns the parsed config and any error encountered.
func Load(path string) (*StaticConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var cfg StaticConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	return &cfg, nil
}

// Save writes a StaticConfig back to disk as YAML.
func Save(path string, cfg *StaticConfig) error {
	// Ensure JSON format whenever a log block is present.
	if cfg.Log != nil {
		cfg.Log.Format = "json"
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// ResolvePaths extracts the dynamic files directory and acme.json path from a
// parsed static config. Falls back to paths relative to the config file's
// directory when the config does not specify them.
func ResolvePaths(staticConfigPath string, cfg *StaticConfig) (dynamicDir, acmePath string) {
	configDir := filepath.Dir(staticConfigPath)

	dynamicDir = filepath.Join(configDir, "dynamic")
	acmePath = filepath.Join(configDir, "acme.json")

	if cfg == nil {
		return
	}

	// Override with values from the file provider.
	if cfg.Providers != nil && cfg.Providers.File != nil {
		fp := cfg.Providers.File
		if fp.Directory != "" {
			dynamicDir = fp.Directory
		} else if fp.Filename != "" {
			// Single-file mode: use that file's directory.
			dynamicDir = filepath.Dir(fp.Filename)
		}
	}

	// Override with ACME storage path from the first resolver that defines one.
	for _, resolver := range cfg.CertResolvers {
		if resolver.ACME != nil && resolver.ACME.Storage != "" {
			acmePath = resolver.ACME.Storage
			break
		}
	}

	return
}
