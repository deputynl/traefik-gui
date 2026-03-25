package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	CookieName  = "tg_session"
	SessionDays = 7
)

// Manager handles credential validation and signed session cookies.
// The signing key is derived deterministically from the credentials so sessions
// survive server restarts — changing the password invalidates all existing sessions.
type Manager struct {
	user   string
	pass   string
	secret []byte
}

func NewManager(user, pass string) *Manager {
	h := sha256.Sum256([]byte(user + ":" + pass + ":traefik-gui-v1"))
	return &Manager{user: user, pass: pass, secret: h[:]}
}

// ValidateCredentials checks a username/password pair.
func (m *Manager) ValidateCredentials(user, pass string) bool {
	return user == m.user && pass == m.pass
}

// createToken returns a signed token: "<user>:<expiry_unix>:<hmac_hex>".
func (m *Manager) createToken(user string) string {
	expiry := time.Now().Add(SessionDays * 24 * time.Hour).Unix()
	payload := fmt.Sprintf("%s:%d", user, expiry)
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return payload + ":" + hex.EncodeToString(mac.Sum(nil))
}

// validateToken parses and verifies a token, returning the username on success.
func (m *Manager) validateToken(token string) (string, bool) {
	// Format: user:expiry:sig — but user may contain colons so split from right.
	lastColon := strings.LastIndex(token, ":")
	if lastColon < 0 {
		return "", false
	}
	sig := token[lastColon+1:]
	rest := token[:lastColon]

	secondLast := strings.LastIndex(rest, ":")
	if secondLast < 0 {
		return "", false
	}
	expiryStr := rest[secondLast+1:]
	user := rest[:secondLast]

	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return "", false
	}

	payload := user + ":" + expiryStr
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return "", false
	}
	return user, true
}

// SetCookie writes a fresh 7-day session cookie to the response.
func (m *Manager) SetCookie(w http.ResponseWriter, user string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    m.createToken(user),
		Path:     "/",
		MaxAge:   SessionDays * 24 * 3600,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearCookie expires the session cookie.
func (m *Manager) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

// FromRequest extracts and validates the session from the request cookie.
// Returns the username and true on success.
func (m *Manager) FromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", false
	}
	return m.validateToken(cookie.Value)
}
