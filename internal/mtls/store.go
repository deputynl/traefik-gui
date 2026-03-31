package mtls

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ClientEntry is metadata about an issued client certificate.
type ClientEntry struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Issued   time.Time `json:"issued"`
	Expires  time.Time `json:"expires"`
	Password string    `json:"password"`
}

// Store manages the mTLS CA and client certificate files on disk.
type Store struct {
	dir string // e.g. /etc/traefik/mtls
}

func NewStore(traefikConfigDir string) *Store {
	return &Store{dir: filepath.Join(traefikConfigDir, "mtls")}
}

func (s *Store) Dir() string        { return s.dir }
func (s *Store) CAKeyPath() string  { return filepath.Join(s.dir, "ca.key") }
func (s *Store) CACertPath() string { return filepath.Join(s.dir, "ca.crt") }
func (s *Store) clientDir() string  { return filepath.Join(s.dir, "clients") }
func (s *Store) indexPath() string  { return filepath.Join(s.clientDir(), "index.json") }

func (s *Store) clientCertPath(id string) string {
	return filepath.Join(s.clientDir(), id+".crt")
}
func (s *Store) clientKeyPath(id string) string {
	return filepath.Join(s.clientDir(), id+".key")
}
func (s *Store) clientP12Path(id string) string {
	return filepath.Join(s.clientDir(), id+".p12")
}

// CAExists returns true if both ca.crt and ca.key are present.
func (s *Store) CAExists() bool {
	_, err1 := os.Stat(s.CACertPath())
	_, err2 := os.Stat(s.CAKeyPath())
	return err1 == nil && err2 == nil
}

// Clients returns all issued client certificates from the index.
func (s *Store) Clients() ([]ClientEntry, error) {
	data, err := os.ReadFile(s.indexPath())
	if os.IsNotExist(err) {
		return []ClientEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []ClientEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store) saveIndex(entries []ClientEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.indexPath(), data, 0600)
}

// AddClient appends a client entry to the index.
func (s *Store) AddClient(entry ClientEntry) error {
	entries, err := s.Clients()
	if err != nil {
		return err
	}
	entries = append(entries, entry)
	return s.saveIndex(entries)
}

// RemoveClient removes a client from the index and deletes its cert files.
func (s *Store) RemoveClient(id string) error {
	entries, err := s.Clients()
	if err != nil {
		return err
	}
	filtered := entries[:0]
	for _, e := range entries {
		if e.ID != id {
			filtered = append(filtered, e)
		}
	}
	if err := s.saveIndex(filtered); err != nil {
		return err
	}
	_ = os.Remove(s.clientCertPath(id))
	_ = os.Remove(s.clientKeyPath(id))
	_ = os.Remove(s.clientP12Path(id))
	return nil
}

// RemoveAllClients deletes all issued client certificate files and clears the index.
func (s *Store) RemoveAllClients() error {
	entries, err := s.Clients()
	if err != nil {
		return err
	}
	for _, e := range entries {
		_ = os.Remove(s.clientCertPath(e.ID))
		_ = os.Remove(s.clientKeyPath(e.ID))
		_ = os.Remove(s.clientP12Path(e.ID))
	}
	return s.saveIndex([]ClientEntry{})
}

// DeleteCA removes ca.crt and ca.key from the store.
func (s *Store) DeleteCA() error {
	_ = os.Remove(s.CACertPath())
	_ = os.Remove(s.CAKeyPath())
	return nil
}

func (s *Store) ensureDirs() error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return err
	}
	// Ensure existing directories have the correct permissions.
	if err := os.Chmod(s.dir, 0755); err != nil {
		return err
	}
	return os.MkdirAll(s.clientDir(), 0700)
}
