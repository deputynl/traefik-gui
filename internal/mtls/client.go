package mtls

import (
	"archive/zip"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"regexp"
	"strings"
	"time"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(name string) string {
	s := slugRe.ReplaceAllString(strings.ToLower(name), "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "client"
	}
	return s
}

// IssueClient generates a new client certificate signed by the CA and returns
// the ClientEntry metadata. The cert and key PEM files are written to the store.
func (s *Store) IssueClient(name string) (ClientEntry, error) {
	if err := s.ensureDirs(); err != nil {
		return ClientEntry{}, err
	}

	caCert, caKey, err := s.LoadCA()
	if err != nil {
		return ClientEntry{}, fmt.Errorf("loading CA: %w", err)
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return ClientEntry{}, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return ClientEntry{}, err
	}

	now := time.Now()
	expires := now.AddDate(2, 0, 0)

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    now.Add(-1 * time.Minute),
		NotAfter:     expires,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return ClientEntry{}, err
	}

	// Build a unique ID: slug + serial suffix to avoid collisions.
	id := toSlug(name) + "-" + serial.Text(16)[len(serial.Text(16))-6:]

	// Write cert PEM.
	certFile, err := os.OpenFile(s.clientCertPath(id), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return ClientEntry{}, err
	}
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certFile.Close()

	// Write key PEM.
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return ClientEntry{}, err
	}
	keyFile, err := os.OpenFile(s.clientKeyPath(id), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return ClientEntry{}, err
	}
	pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	keyFile.Close()

	entry := ClientEntry{ID: id, Name: name, Issued: now, Expires: expires}
	if err := s.AddClient(entry); err != nil {
		return ClientEntry{}, err
	}
	return entry, nil
}

// WriteClientZip writes a ZIP archive containing ca.crt, client.crt, client.key
// and a README to the provided writer.
func (s *Store) WriteClientZip(id string, w io.Writer) error {
	caCertPEM, err := os.ReadFile(s.CACertPath())
	if err != nil {
		return fmt.Errorf("reading CA cert: %w", err)
	}
	clientCertPEM, err := os.ReadFile(s.clientCertPath(id))
	if err != nil {
		return fmt.Errorf("reading client cert: %w", err)
	}
	clientKeyPEM, err := os.ReadFile(s.clientKeyPath(id))
	if err != nil {
		return fmt.Errorf("reading client key: %w", err)
	}

	readme := `mTLS Client Certificate
========================

Files in this archive:
  ca.crt      - Certificate Authority (trust this in your browser/OS)
  client.crt  - Your client certificate
  client.key  - Your client private key (keep this secret)

Installation instructions:

macOS:
  1. Double-click ca.crt → add to Keychain → mark as "Always Trust"
  2. Combine into PKCS#12: openssl pkcs12 -export -in client.crt -inkey client.key -out client.p12
  3. Double-click client.p12 to import into Keychain

Windows:
  1. Double-click ca.crt → Install Certificate → Local Machine → Trusted Root CAs
  2. Run: certutil -importpfx client.p12  (after creating it with the openssl command above)

Firefox:
  Settings → Privacy & Security → Certificates → View Certificates
  → Authorities tab → Import ca.crt
  → Your Certificates tab → Import client.p12

Linux (Chrome/Chromium):
  Settings → Privacy and security → Security → Manage certificates
  Import ca.crt as Authority, client.p12 as Your certificates
`

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, f := range []struct {
		name    string
		content []byte
	}{
		{"ca.crt", caCertPEM},
		{"client.crt", clientCertPEM},
		{"client.key", clientKeyPEM},
		{"README.txt", []byte(readme)},
	} {
		fw, err := zw.Create(f.name)
		if err != nil {
			return err
		}
		if _, err := fw.Write(f.content); err != nil {
			return err
		}
	}
	return nil
}
