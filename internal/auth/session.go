package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"sync"
	"time"
)

const (
	SessionCookieName = "dispatch_inbox_session"
	sessionLifetime   = 12 * time.Hour
)

type session struct {
	username string
	expires  time.Time
}

var uiSessions = struct {
	sync.RWMutex
	values map[string]session
}{values: make(map[string]session)}

// IssueUISession creates an opaque, in-memory browser session after credentials
// have been verified. Sessions intentionally expire on process restart.
func IssueUISession(username string) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().Add(sessionLifetime)

	uiSessions.Lock()
	defer uiSessions.Unlock()
	uiSessions.values[token] = session{username: username, expires: expires}

	return token, expires, nil
}

// ValidateUISession validates a Dispatch Inbox session cookie.
func ValidateUISession(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	uiSessions.RLock()
	entry, ok := uiSessions.values[cookie.Value]
	uiSessions.RUnlock()
	if !ok {
		return "", false
	}
	if time.Now().After(entry.expires) {
		uiSessions.Lock()
		delete(uiSessions.values, cookie.Value)
		uiSessions.Unlock()
		return "", false
	}

	return entry.username, true
}

// RevokeUISession removes a browser session.
func RevokeUISession(r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return
	}
	uiSessions.Lock()
	delete(uiSessions.values, cookie.Value)
	uiSessions.Unlock()
}
