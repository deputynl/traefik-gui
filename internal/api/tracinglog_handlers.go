package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"traefik-gui/internal/docker"
	"traefik-gui/internal/traefik"
	"traefik-gui/internal/tracinglog"
)

// GET /api/tracinglog — returns the last 300 lines from the container's stdout.
func (s *Server) handleTracingLogRecent(w http.ResponseWriter, r *http.Request) {
	if s.cfg.TraefikContainerName == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"lines":     []any{},
			"available": false,
			"reason":    "TRAEFIK_CONTAINER_NAME is not set",
		})
		return
	}

	rawLines, err := docker.ContainerLogLines(s.cfg.TraefikContainerName, 300)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"lines":     []any{},
			"available": false,
			"reason":    "cannot read Docker logs: " + err.Error(),
		})
		return
	}

	lines := make([]tracinglog.Line, 0, len(rawLines))
	for i := len(rawLines) - 1; i >= 0; i-- {
		if l := tracinglog.ParseLine(rawLines[i]); l.Raw != "" {
			lines = append(lines, l)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"lines":     lines,
		"available": true,
	})
}

// GET /api/tracinglog/stream — SSE stream of new log lines from the container.
func (s *Server) handleTracingLogStream(w http.ResponseWriter, r *http.Request) {
	if s.cfg.TraefikContainerName == "" {
		http.Error(w, "TRAEFIK_CONTAINER_NAME is not set", http.StatusServiceUnavailable)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()
	keepalive := time.NewTicker(10 * time.Second)
	defer keepalive.Stop()

	since := time.Now()
	name := s.cfg.TraefikContainerName

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()

		case t := <-poll.C:
			rawLines, err := docker.ContainerLogLinesSince(name, since)
			if err != nil {
				return
			}
			since = t

			sent := false
			for _, raw := range rawLines {
				l := tracinglog.ParseLine(raw)
				if l.Raw == "" {
					continue
				}
				data, err := json.Marshal(l)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				sent = true
			}
			if sent {
				flusher.Flush()
			}
		}
	}
}

// PUT /api/traefik/loglevel — update log.level in static config and restart Traefik.
func (s *Server) handleSetLogLevel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Level string `json:"level"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	level := strings.ToUpper(strings.TrimSpace(body.Level))
	switch level {
	case "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
	default:
		writeError(w, http.StatusBadRequest, "invalid log level: "+level)
		return
	}

	if !s.paths.StaticConfigFound {
		writeError(w, http.StatusConflict, "static config not found")
		return
	}

	cfg, err := traefik.Load(s.cfg.TraefikConfigPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not load static config: "+err.Error())
		return
	}
	if cfg.Log == nil {
		cfg.Log = &traefik.LogConfig{}
	}
	cfg.Log.Level = level

	if err := traefik.Save(s.cfg.TraefikConfigPath, cfg); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save config: "+err.Error())
		return
	}
	s.refreshPaths()

	name := s.cfg.TraefikContainerName
	if name == "" {
		s.audit.Log(s.userFromRequest(r), "log_level", "set log level to "+level)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restarted": false})
		return
	}

	if err := docker.RestartContainer(name); err != nil {
		writeError(w, http.StatusInternalServerError, "config saved but restart failed: "+err.Error())
		return
	}

	s.audit.Log(s.userFromRequest(r), "log_level", "set log level to "+level+", restarted "+name)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "restarted": true})
}
