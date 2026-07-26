// Package auth handles the web UI and SMTP authentication
package auth

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/tg123/go-htpasswd"
	"golang.org/x/crypto/bcrypt"
)

var (
	// UICredentials passwords
	UICredentials *htpasswd.File
	// SendAPICredentials passwords
	SendAPICredentials *htpasswd.File
	// SMTPCredentials passwords
	SMTPCredentials *htpasswd.File
	// POP3Credentials passwords
	POP3Credentials *htpasswd.File
	uiAuthMu        sync.RWMutex
)

// SetUIAuth will set Basic Auth credentials required for the UI & API
func SetUIAuth(s string) error {
	credentials := credentialsFromString(s)
	if len(credentials) == 0 {
		uiAuthMu.Lock()
		UICredentials = nil
		uiAuthMu.Unlock()
		return nil
	}

	r := strings.NewReader(strings.Join(credentials, "\n"))
	parsed, err := htpasswd.NewFromReader(r, htpasswd.DefaultSystems, nil)
	if err != nil {
		return err
	}

	uiAuthMu.Lock()
	UICredentials = parsed
	uiAuthMu.Unlock()
	return nil
}

// MatchUI verifies UI credentials without racing credential reloads.
func MatchUI(username, password string) bool {
	uiAuthMu.RLock()
	defer uiAuthMu.RUnlock()
	return UICredentials != nil && UICredentials.Match(username, password)
}

// UIAuthEnabled reports whether UI credentials are configured.
func UIAuthEnabled() bool {
	uiAuthMu.RLock()
	defer uiAuthMu.RUnlock()
	return UICredentials != nil
}

// RegisterUIUser persists a bcrypt credential and reloads the UI auth store.
func RegisterUIUser(path, username, password string) error {
	if path == "" {
		return fmt.Errorf("registration requires a UI auth file")
	}
	if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{2,31}$`).MatchString(username) {
		return fmt.Errorf("username must be 3-32 characters and use only letters, numbers, dot, dash or underscore")
	}
	if len(password) < 12 || len(password) > 128 {
		return fmt.Errorf("password must contain between 12 and 128 characters")
	}

	uiAuthMu.Lock()
	defer uiAuthMu.Unlock()

	current, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read credential store: %w", err)
	}
	for _, line := range strings.Split(string(current), "\n") {
		name, _, found := strings.Cut(strings.TrimSpace(line), ":")
		if found && strings.EqualFold(name, username) {
			return fmt.Errorf("this username is already registered")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	updated := strings.TrimSpace(string(current))
	if updated != "" {
		updated += "\n"
	}
	updated += username + ":" + string(hash) + "\n"

	parsed, err := htpasswd.NewFromReader(strings.NewReader(updated), htpasswd.DefaultSystems, nil)
	if err != nil {
		return fmt.Errorf("parse credential store: %w", err)
	}
	if err := os.WriteFile(path, []byte(updated), 0o600); err != nil {
		return fmt.Errorf("write credential store: %w", err)
	}
	UICredentials = parsed
	return nil
}

// SetSendAPIAuth will set Send API credentials
func SetSendAPIAuth(s string) error {
	var err error

	credentials := credentialsFromString(s)
	if len(credentials) == 0 {
		return nil
	}

	r := strings.NewReader(strings.Join(credentials, "\n"))

	SendAPICredentials, err = htpasswd.NewFromReader(r, htpasswd.DefaultSystems, nil)
	if err != nil {
		return err
	}

	return nil
}

// SetSMTPAuth will set SMTP credentials
func SetSMTPAuth(s string) error {
	var err error

	credentials := credentialsFromString(s)
	if len(credentials) == 0 {
		return nil
	}

	r := strings.NewReader(strings.Join(credentials, "\n"))

	SMTPCredentials, err = htpasswd.NewFromReader(r, htpasswd.DefaultSystems, nil)
	if err != nil {
		return err
	}

	return nil
}

// SetPOP3Auth will set POP3 server credentials
func SetPOP3Auth(s string) error {
	var err error

	credentials := credentialsFromString(s)
	if len(credentials) == 0 {
		return nil
	}

	r := strings.NewReader(strings.Join(credentials, "\n"))

	POP3Credentials, err = htpasswd.NewFromReader(r, htpasswd.DefaultSystems, nil)
	if err != nil {
		return err
	}

	return nil
}

func credentialsFromString(s string) []string {
	// split string by any whitespace character
	re := regexp.MustCompile(`\s+`)

	words := re.Split(s, -1)
	credentials := []string{}
	for _, w := range words {
		if w != "" {
			credentials = append(credentials, w)
		}
	}

	return credentials
}
