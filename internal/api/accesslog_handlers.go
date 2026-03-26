package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"traefik-gui/internal/accesslog"
	"traefik-gui/internal/docker"
)

// handleAccessLogRecent returns the last 300 parsed lines from the container's stdout.
func (s *Server) handleAccessLogRecent(w http.ResponseWriter, r *http.Request) {
	if s.cfg.TraefikContainerName == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"entries":   []any{},
			"available": false,
			"reason":    "TRAEFIK_CONTAINER_NAME is not set",
		})
		return
	}

	lines, err := docker.ContainerLogLines(s.cfg.TraefikContainerName, 300)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"entries":   []any{},
			"available": false,
			"reason":    "cannot read Docker logs: " + err.Error(),
		})
		return
	}

	entries := make([]accesslog.Entry, 0)
	for i := len(lines) - 1; i >= 0; i-- {
		if e, ok := accesslog.ParseLine(lines[i]); ok {
			entries = append(entries, *e)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"entries":   entries,
		"available": true,
	})
}

// handleAccessLogStream streams new log lines from the container's stdout via SSE.
func (s *Server) handleAccessLogStream(w http.ResponseWriter, r *http.Request) {
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
	// Flush headers immediately so the browser sees the 200 and fires onopen.
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	lines := make(chan string, 64)
	errCh := make(chan error, 1)

	go func() {
		errCh <- docker.StreamContainerLogs(ctx, s.cfg.TraefikContainerName, lines)
	}()

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		case line, open := <-lines:
			if !open {
				return
			}
			e, ok := accesslog.ParseLine(line)
			if !ok {
				continue
			}
			data, err := json.Marshal(e)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-errCh:
			return
		}
	}
}
