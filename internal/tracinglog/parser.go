package tracinglog

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

// Line is a single parsed Traefik general log entry.
type Line struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"msg"`
	Raw     string `json:"raw"`
}

// kvRe matches key=value and key="quoted value" pairs in Traefik text log lines.
var kvRe = regexp.MustCompile(`(\w+)=(?:"([^"]*)"|(\S+))`)

// ParseLine parses a Traefik general log line.
// Handles both JSON format and the default key=value text format.
// Lines that match neither are returned with empty Level/Message and the raw text.
func ParseLine(raw string) Line {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Line{}
	}

	// Try JSON format.
	var obj struct {
		Time  string `json:"time"`
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal([]byte(raw), &obj); err == nil && obj.Level != "" {
		return Line{
			Time:    obj.Time,
			Level:   strings.ToUpper(obj.Level),
			Message: obj.Msg,
			Raw:     raw,
		}
	}

	// Try key=value text format: time="..." level=info msg="..."
	var timeStr, level, msg string
	for _, m := range kvRe.FindAllStringSubmatch(raw, -1) {
		val := m[2]
		if val == "" {
			val = m[3]
		}
		switch m[1] {
		case "time":
			timeStr = val
		case "level":
			level = strings.ToUpper(val)
		case "msg":
			msg = val
		}
	}
	if level != "" {
		return Line{Time: timeStr, Level: level, Message: msg, Raw: raw}
	}

	// Unparseable — return raw with current timestamp.
	return Line{
		Time: time.Now().UTC().Format(time.RFC3339),
		Raw:  raw,
	}
}
