package accesslog

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Entry is a parsed access log line.
type Entry struct {
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Host       string    `json:"host"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	DurationMS float64   `json:"durationMs"`
	RouterName string    `json:"routerName"`
	ClientIP   string    `json:"clientIp"`
	Raw        string    `json:"raw"`
}

// ParseLine tries JSON format first, then falls back to CLF.
// Returns (entry, true) on success; (nil, false) if the line is blank or unparseable.
func ParseLine(line string) (*Entry, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, false
	}

	if strings.HasPrefix(line, "{") {
		if e := parseJSON(line); e != nil {
			return e, true
		}
	}

	if e := parseCLF(line); e != nil {
		return e, true
	}

	return nil, false
}

// ── JSON format ────────────────────────────────────────────────────────────

type jsonEntry struct {
	Time             string  `json:"time"`
	RequestMethod    string  `json:"RequestMethod"`
	RequestPath      string  `json:"RequestPath"`
	RequestHost      string  `json:"RequestHost"`
	RequestAddr      string  `json:"RequestAddr"`
	DownstreamStatus int     `json:"DownstreamStatus"`
	Duration         int64   `json:"Duration"` // nanoseconds
	RouterName       string  `json:"RouterName"`
	ClientHost       string  `json:"ClientHost"`
}

func parseJSON(line string) *Entry {
	var j jsonEntry
	if err := json.Unmarshal([]byte(line), &j); err != nil {
		return nil
	}

	t, _ := time.Parse(time.RFC3339Nano, j.Time)
	if t.IsZero() {
		t = time.Now()
	}

	host := j.RequestHost
	if host == "" {
		host = j.RequestAddr
	}

	return &Entry{
		Time:       t,
		Method:     j.RequestMethod,
		Host:       host,
		Path:       j.RequestPath,
		Status:     j.DownstreamStatus,
		DurationMS: float64(j.Duration) / 1e6,
		RouterName: j.RouterName,
		ClientIP:   j.ClientHost,
		Raw:        line,
	}
}

// ── CLF format ─────────────────────────────────────────────────────────────
// 10.0.0.1 - user [25/Mar/2026:12:00:00 +0000] "GET /path HTTP/2.0" 200 1234 "-" "-" 1 "router@file" "http://backend" 12ms

var clfRe = regexp.MustCompile(
	`^(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"(\w+)\s+([^\s"]+)[^"]*"\s+(\d+)\s+\S+\s+"[^"]*"\s+"[^"]*"\s+\S+\s+"([^"]*)"\s+\S+\s+([\d.]+)ms`,
)

var clfTime = "02/Jan/2006:15:04:05 -0700"

func parseCLF(line string) *Entry {
	m := clfRe.FindStringSubmatch(line)
	if m == nil {
		return nil
	}

	t, _ := time.Parse(clfTime, m[2])
	status, _ := strconv.Atoi(m[5])
	dur, _ := strconv.ParseFloat(m[7], 64)

	return &Entry{
		Time:       t,
		Method:     m[3],
		Path:       m[4],
		Status:     status,
		DurationMS: dur,
		RouterName: m[6],
		ClientIP:   m[1],
		Raw:        line,
	}
}
