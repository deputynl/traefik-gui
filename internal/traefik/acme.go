package traefik

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// AcmeCertInfo is a parsed certificate entry from acme.json.
type AcmeCertInfo struct {
	Resolver string    `json:"resolver"`
	Domain   string    `json:"domain"`
	SANs     []string  `json:"sans"`
	Expiry   time.Time `json:"expiry"`
	DaysLeft int       `json:"daysLeft"`
}

type acmeFile map[string]acmeResolver

type acmeResolver struct {
	Certificates []acmeCertEntry `json:"Certificates"`
}

type acmeCertEntry struct {
	Domain struct {
		Main string   `json:"main"`
		SANs []string `json:"sans"`
	} `json:"domain"`
	Certificate string `json:"certificate"`
}

// LoadAcmeCerts parses acme.json and returns cert info for all stored certs.
func LoadAcmeCerts(path string) ([]AcmeCertInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var af acmeFile
	if err := json.Unmarshal(data, &af); err != nil {
		return nil, fmt.Errorf("parsing acme.json: %w", err)
	}

	now := time.Now()
	var certs []AcmeCertInfo
	for resolverName, resolver := range af {
		for _, entry := range resolver.Certificates {
			info := AcmeCertInfo{
				Resolver: resolverName,
				Domain:   entry.Domain.Main,
				SANs:     entry.Domain.SANs,
			}

			// The certificate field is base64-encoded PEM.
			certPEM, err := base64.StdEncoding.DecodeString(entry.Certificate)
			if err != nil {
				certPEM, _ = base64.RawStdEncoding.DecodeString(entry.Certificate)
			}
			block, _ := pem.Decode(certPEM)
			if block != nil {
				if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
					info.Expiry = cert.NotAfter
					info.DaysLeft = int(cert.NotAfter.Sub(now).Hours() / 24)
				}
			}

			certs = append(certs, info)
		}
	}
	return certs, nil
}
