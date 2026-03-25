package api

import (
	"encoding/json"
	"net/http"
)

// POST /auth/login — validates credentials and sets a session cookie.
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if !s.auth.ValidateCredentials(creds.Username, creds.Password) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	s.auth.SetCookie(w, creds.Username)
	writeJSON(w, http.StatusOK, map[string]string{"user": creds.Username})
}

// POST /auth/logout — clears the session cookie.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.auth.ClearCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// GET /auth/check — returns 200 + user if authenticated, 401 otherwise.
// Also refreshes the cookie (sliding 7-day window).
func (s *Server) handleAuthCheck(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth.FromRequest(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	s.auth.SetCookie(w, user)
	writeJSON(w, http.StatusOK, map[string]string{"user": user})
}
