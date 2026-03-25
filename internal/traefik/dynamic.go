package traefik

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var hostRuleRe = regexp.MustCompile(`Host\(\x60([^\x60]+)\x60\)`)

// ListDynamic returns a summary card for every file in dir.
func ListDynamic(dir string) ([]FileSummary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileSummary{}, nil
		}
		return nil, err
	}

	var out []FileSummary
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		active := strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yml.bak")

		sum := FileSummary{Name: name, Active: active}
		if active {
			if cfg, err := LoadDynamic(filepath.Join(dir, name)); err == nil {
				sum = summarise(name, active, cfg)
			}
		}
		out = append(out, sum)
	}
	return out, nil
}

func summarise(name string, active bool, cfg *DynamicConfig) FileSummary {
	s := FileSummary{Name: name, Active: active}
	if cfg.HTTP == nil {
		return s
	}
	h := cfg.HTTP
	s.RouterCount = len(h.Routers)
	s.ServiceCount = len(h.Services)
	s.MiddlewareCount = len(h.Middlewares)

	for _, r := range h.Routers {
		if m := hostRuleRe.FindStringSubmatch(r.Rule); len(m) > 1 {
			s.Hostnames = append(s.Hostnames, m[1])
		}
		if r.TLS != nil && r.TLS.CertResolver != "" {
			s.CertResolver = r.TLS.CertResolver
		}
	}
	for _, svc := range h.Services {
		if svc.LoadBalancer != nil {
			for _, srv := range svc.LoadBalancer.Servers {
				if srv.URL != "" {
					s.Backends = append(s.Backends, srv.URL)
				}
			}
		}
	}
	for _, st := range h.ServersTransports {
		if st.InsecureSkipVerify {
			s.InsecureSkipVerify = true
		}
	}
	return s
}

// LoadDynamic reads and parses a dynamic config file.
func LoadDynamic(path string) (*DynamicConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var cfg DynamicConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

// ReadRaw returns the raw file content as a string.
func ReadRaw(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

// WriteRaw validates the content is valid YAML then writes it to path.
func WriteRaw(path, content string) error {
	var check interface{}
	if err := yaml.Unmarshal([]byte(content), &check); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// SaveDynamic marshals cfg as YAML and writes it to path.
func SaveDynamic(path string, cfg *DynamicConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// GenerateServiceConfig builds a DynamicConfig from a ServiceSpec wizard input.
func GenerateServiceConfig(spec ServiceSpec) *DynamicConfig {
	t := true
	svcName := spec.Name + "-svc"

	if len(spec.EntryPoints) == 0 {
		spec.EntryPoints = []string{"websecure"}
	}

	lb := &LoadBalancer{
		Servers:        []BackendServer{{URL: spec.BackendURL}},
		PassHostHeader: &t,
	}
	if spec.InsecureBackend {
		lb.ServersTransport = "insecure-skip"
	}

	cfg := &DynamicConfig{
		HTTP: &HTTPDynamicConfig{
			Routers: map[string]DynRouter{
				spec.Name: {
					Rule:        fmt.Sprintf("Host(`%s`)", spec.Hostname),
					EntryPoints: spec.EntryPoints,
					Service:     svcName,
					TLS:         &DynTLS{CertResolver: spec.CertResolver},
				},
			},
			Services: map[string]DynService{
				svcName: {LoadBalancer: lb},
			},
		},
	}

	if spec.InsecureBackend {
		cfg.HTTP.ServersTransports = map[string]ServersTransport{
			"insecure-skip": {InsecureSkipVerify: true},
		}
	}

	return cfg
}
