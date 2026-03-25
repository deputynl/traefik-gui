package traefik

import "fmt"

// ValidationWarning is a single semantic warning about a static config.
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateStaticConfig performs semantic checks and returns any warnings.
// The config is still saved even when warnings are present.
func ValidateStaticConfig(cfg *StaticConfig) []ValidationWarning {
	if cfg == nil {
		return nil
	}

	var warns []ValidationWarning

	epNames := make(map[string]bool, len(cfg.EntryPoints))
	for name := range cfg.EntryPoints {
		epNames[name] = true
	}

	// Check redirect targets exist.
	for name, ep := range cfg.EntryPoints {
		if ep.HTTP != nil && ep.HTTP.Redirections != nil && ep.HTTP.Redirections.EntryPoint != nil {
			to := ep.HTTP.Redirections.EntryPoint.To
			if to != "" && !epNames[to] {
				warns = append(warns, ValidationWarning{
					Field:   fmt.Sprintf("entryPoints.%s.http.redirections.entryPoint.to", name),
					Message: fmt.Sprintf("target entry point %q does not exist", to),
				})
			}
		}
	}

	// Check ACME HTTP challenge entry points exist.
	for resolverName, resolver := range cfg.CertResolvers {
		if resolver.ACME == nil {
			continue
		}
		if resolver.ACME.HTTPChallenge != nil {
			ep := resolver.ACME.HTTPChallenge.EntryPoint
			if ep != "" && !epNames[ep] {
				warns = append(warns, ValidationWarning{
					Field:   fmt.Sprintf("certificatesResolvers.%s.acme.httpChallenge.entryPoint", resolverName),
					Message: fmt.Sprintf("entry point %q does not exist", ep),
				})
			}
		}
		if resolver.ACME.Storage == "" {
			warns = append(warns, ValidationWarning{
				Field:   fmt.Sprintf("certificatesResolvers.%s.acme.storage", resolverName),
				Message: "no storage path configured for ACME certificates",
			})
		}
	}

	return warns
}
