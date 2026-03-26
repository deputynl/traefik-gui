package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"traefik-gui/internal/accesslog"
	"traefik-gui/internal/docker"
)

// handleAccessLogRecent returns the last N parsed lines from the access log.
func (s *Server) handleAccessLogRecent(w http.ResponseWriter, r *http.Request) {
	// Docker-container mode.
	if s.cfg.TraefikContainerName != "" {
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
		return
	}

	// File mode.
	path := s.paths.AccessLogPath
	if path == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"entries":   []any{},
			"available": false,
			"reason":    "no accessLog.filePath configured in traefik.yml and TRAEFIK_CONTAINER_NAME not set",
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
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // prevent nginx buffering

	// Docker-container mode.
	if s.cfg.TraefikContainerName != "" {
		s.streamDockerLogs(w, r, flusher)
		return
	}

	// File mode.
	path := s.paths.AccessLogPath
	if path == "" {
		http.Error(w, "no accessLog.filePath configured and TRAEFIK_CONTAINER_NAME not set", http.StatusServiceUnavailable)
		return
	}
	s.streamFileLogs(w, r, flusher, path)
}

func sendSSEEntry(w http.ResponseWriter, flusher http.Flusher, line string) {
	e, ok := accesslog.ParseLine(line)
	if !ok {
		return
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// streamDockerLogs reads from the container's stdout/stderr via the Docker socket.
func (s *Server) streamDockerLogs(w http.ResponseWriter, r *http.Request, flusher http.Flusher) {
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
			sendSSEEntry(w, flusher, line)
		case <-errCh:
			return
		}
	}
}

// streamFileLogs tails a log file and sends new lines as SSE events.
func (s *Server) streamFileLogs(w http.ResponseWriter, r *http.Request, flusher http.Flusher, path string) {
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

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	poll := time.NewTicker(500 * time.Millisecond)
	defer poll.Stop()

	// We tail by reading raw bytes. bufio.Scanner cannot be reused after EOF,
	// so we manage a partial-line remainder ourselves.
	buf := make([]byte, 64*1024)
	var remainder []byte

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()

		case <-poll.C:
			n, err := f.Read(buf)
			if n > 0 {
				chunk := append(remainder, buf[:n]...)
				lines := bytes.Split(chunk, []byte("\n"))
				// The last element is an incomplete line (or empty); save it.
				remainder = lines[len(lines)-1]
				lines = lines[:len(lines)-1]

				sent := false
				for _, raw := range lines {
					e, ok := accesslog.ParseLine(string(raw))
					if !ok {
						continue
					}
					data, merr := json.Marshal(e)
					if merr != nil {
						continue
					}
					fmt.Fprintf(w, "data: %s\n\n", data)
					sent = true
				}
				if sent {
					flusher.Flush()
				}
			}
			// io.EOF just means no new data yet — keep polling.
			if err != nil && err != io.EOF {
				return
			}
		}
	}
}
