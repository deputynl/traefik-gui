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
	// Core
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	Host       string    `json:"host"`
	Path       string    `json:"path"`
	Protocol   string    `json:"protocol"`
	Scheme     string    `json:"scheme"`
	Status     int       `json:"status"`
	DurationMS float64   `json:"durationMs"`
	ClientIP   string    `json:"clientIp"`
	// Routing
	RouterName     string `json:"routerName"`
	ServiceName    string `json:"serviceName"`
	ServiceAddr    string `json:"serviceAddr"`
	EntryPoint     string `json:"entryPoint"`
	// Upstream
	OriginStatus     int     `json:"originStatus"`
	OriginDurationMS float64 `json:"originDurationMs"`
	RetryAttempts    int     `json:"retryAttempts"`
	// Sizes
	ResponseSize int64 `json:"responseSize"`
	RequestSize  int64 `json:"requestSize"`
	// TLS
	TLSVersion string `json:"tlsVersion"`
	TLSCipher  string `json:"tlsCipher"`
	// Raw line for display
	Raw string `json:"raw"`
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
	Time                 string `json:"time"`
	RequestMethod        string `json:"RequestMethod"`
	RequestPath          string `json:"RequestPath"`
	RequestHost          string `json:"RequestHost"`
	RequestAddr          string `json:"RequestAddr"`
	RequestProtocol      string `json:"RequestProtocol"`
	RequestScheme        string `json:"RequestScheme"`
	RequestContentSize   int64  `json:"RequestContentSize"`
	DownstreamStatus     int    `json:"DownstreamStatus"`
	DownstreamContentSize int64 `json:"DownstreamContentSize"`
	Duration             int64  `json:"Duration"` // nanoseconds
	OriginStatus         int    `json:"OriginStatus"`
	OriginDuration       int64  `json:"OriginDuration"` // nanoseconds
	RouterName           string `json:"RouterName"`
	ServiceName          string `json:"ServiceName"`
	ServiceAddr          string `json:"ServiceAddr"`
	EntryPointName       string `json:"entryPointName"`
	ClientHost           string `json:"ClientHost"`
	RetryAttempts        int    `json:"RetryAttempts"`
	TLSVersion           string `json:"TLSVersion"`
	TLSCipher            string `json:"TLSCipher"`
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

	var originDurationMS float64
	if j.OriginDuration > 0 {
		originDurationMS = float64(j.OriginDuration) / 1e6
	}

	return &Entry{
		Time:              t,
		Method:            j.RequestMethod,
		Host:              host,
		Path:              j.RequestPath,
		Protocol:          j.RequestProtocol,
		Scheme:            j.RequestScheme,
		Status:            j.DownstreamStatus,
		DurationMS:        float64(j.Duration) / 1e6,
		ClientIP:          j.ClientHost,
		RouterName:        j.RouterName,
		ServiceName:       j.ServiceName,
		ServiceAddr:       j.ServiceAddr,
		EntryPoint:        j.EntryPointName,
		OriginStatus:      j.OriginStatus,
		OriginDurationMS:  originDurationMS,
		RetryAttempts:     j.RetryAttempts,
		ResponseSize:      j.DownstreamContentSize,
		RequestSize:       j.RequestContentSize,
		TLSVersion:        j.TLSVersion,
		TLSCipher:         j.TLSCipher,
		Raw:               line,
	}
}

// ── CLF format ─────────────────────────────────────────────────────────────
// 10.0.0.1 - user [25/Mar/2026:12:00:00 +0000] "GET /path HTTP/2.0" 200 1234 "-" "-" 1 "router@file" "http://backend" 12ms

var clfRe = regexp.MustCompile(
	`^(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"(\w+)\s+([^\s"]+)\s+([^"]+)"\s+(\d+)\s+(\S+)\s+"[^"]*"\s+"[^"]*"\s+\S+\s+"([^"]*)"\s+\S+\s+([\d.]+)ms`,
)

const clfTime = "02/Jan/2006:15:04:05 -0700"

func parseCLF(line string) *Entry {
	m := clfRe.FindStringSubmatch(line)
	if m == nil {
		return nil
	}

	t, _ := time.Parse(clfTime, m[2])
	status, _ := strconv.Atoi(m[6])
	dur, _ := strconv.ParseFloat(m[9], 64)

	var respSize int64
	if m[7] != "-" {
		respSize, _ = strconv.ParseInt(m[7], 10, 64)
	}

	return &Entry{
		Time:         t,
		Method:       m[3],
		Path:         m[4],
		Protocol:     strings.TrimSpace(m[5]),
		Status:       status,
		ResponseSize: respSize,
		DurationMS:   dur,
		RouterName:   m[8],
		ClientIP:     m[1],
		Raw:          line,
	}
}
