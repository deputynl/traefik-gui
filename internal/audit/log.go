package audit

import (
	"bytes"
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry is a single audit log record.
type Entry struct {
	Time   time.Time `json:"time"`
	User   string    `json:"user"`
	Action string    `json:"action"`
	Detail string    `json:"detail"`
}

// Logger appends JSON-lines entries to a log file.
type Logger struct {
	path string
	mu   sync.Mutex
}

func NewLogger(path string) *Logger {
	return &Logger{path: path}
}

func (l *Logger) Log(user, action, detail string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{
		Time:   time.Now().UTC(),
		User:   user,
		Action: action,
		Detail: detail,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(append(line, '\n')) //nolint:errcheck
}

// Recent returns up to n entries, newest first.
func (l *Logger) Recent(n int) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil
	}

	rawLines := bytes.Split(bytes.TrimRight(data, "\n"), []byte("\n"))
	if len(rawLines) > n {
		rawLines = rawLines[len(rawLines)-n:]
	}

	entries := make([]Entry, 0, len(rawLines))
	for _, line := range rawLines {
		if len(line) == 0 {
			continue
		}
		var e Entry
		if err := json.Unmarshal(line, &e); err == nil {
			entries = append(entries, e)
		}
	}

	// Reverse to newest-first.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
}
