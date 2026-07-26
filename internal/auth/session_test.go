package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUISessionLifecycle(t *testing.T) {
	token, expires, err := IssueUISession("dispatch-user")
	if err != nil {
		t.Fatalf("IssueUISession() error = %v", err)
	}
	if token == "" || !expires.After(time.Now()) {
		t.Fatal("IssueUISession() returned an invalid session")
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: token})
	username, ok := ValidateUISession(request)
	if !ok || username != "dispatch-user" {
		t.Fatalf("ValidateUISession() = %q, %v", username, ok)
	}

	RevokeUISession(request)
	if _, ok := ValidateUISession(request); ok {
		t.Fatal("ValidateUISession() accepted a revoked session")
	}
}

func TestRegisterUIUser(t *testing.T) {
	path := filepath.Join(t.TempDir(), "users.htpasswd")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := RegisterUIUser(path, "new.user", "correct-horse-battery"); err != nil {
		t.Fatalf("RegisterUIUser() error = %v", err)
	}
	if !MatchUI("new.user", "correct-horse-battery") {
		t.Fatal("registered credentials do not authenticate")
	}
	if err := RegisterUIUser(path, "NEW.USER", "another-long-password"); err == nil {
		t.Fatal("RegisterUIUser() accepted a duplicate username")
	}
	if err := RegisterUIUser(path, "x", "another-long-password"); err == nil {
		t.Fatal("RegisterUIUser() accepted an invalid username")
	}
}
