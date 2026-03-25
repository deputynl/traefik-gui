package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"traefik-gui/internal/accesslog"
)

// handleAccessLogRecent returns the last N parsed lines from the access log.
func (s *Server) handleAccessLogRecent(w http.ResponseWriter, r *http.Request) {
	path := s.paths.AccessLogPath
	if path == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"entries":   []any{},
			"available": false,
			"reason":    "no accessLog.filePath configured in traefik.yml — Traefik is logging to stdout",
		})
		return
	}

	f, err := os.Open(path)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"entries":   []any{},
			"available": false,
			"reason":    "access log file not found: " + path,
		})
		return
	}
	defer f.Close()

	// Read all lines, keep last 300.
	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 300 {
			lines = lines[1:]
		}
	}

	entries := make([]accesslog.Entry, 0, len(lines))
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

// handleAccessLogStream streams new access log lines via SSE.
func (s *Server) handleAccessLogStream(w http.ResponseWriter, r *http.Request) {
	path := s.paths.AccessLogPath
	if path == "" {
		http.Error(w, "no accessLog.filePath configured", http.StatusServiceUnavailable)
		return
	}

	f, err := os.Open(path)
	if err != nil {
		http.Error(w, "cannot open access log: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer f.Close()

	// Seek to end so we only stream new entries.
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		http.Error(w, "seek error: "+err.Error(), http.StatusInternalServerError)
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
	w.Header().Set("X-Accel-Buffering", "no") // prevent nginx buffering

	// Send a keepalive comment every 15 s to prevent proxy timeouts.
	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()

		case <-poll.C:
			sent := false
			for scanner.Scan() {
				line := scanner.Text()
				e, ok := accesslog.ParseLine(line)
				if !ok {
					continue
				}
				data, err := json.Marshal(e)
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
