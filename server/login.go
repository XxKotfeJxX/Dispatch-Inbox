package server

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/axllent/mailpit/config"
	"github.com/axllent/mailpit/internal/auth"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.UIAuthEnabled() {
		http.Redirect(w, r, config.Webroot, http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
		if err := r.ParseForm(); err != nil {
			renderLogin(w, "Invalid login request.")
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		if username == "" || !auth.MatchUI(username, r.FormValue("password")) {
			w.WriteHeader(http.StatusUnauthorized)
			renderLogin(w, "Incorrect username or password.")
			return
		}

		token, expires, err := auth.IssueUISession(username)
		if err != nil {
			http.Error(w, "Could not create session.", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     auth.SessionCookieName,
			Value:    token,
			Path:     config.Webroot,
			Expires:  expires,
			MaxAge:   int(time.Until(expires).Seconds()),
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteStrictMode,
		})
		http.Redirect(w, r, config.Webroot, http.StatusSeeOther)
		return
	}

	if _, ok := auth.ValidateUISession(r); ok {
		http.Redirect(w, r, config.Webroot, http.StatusSeeOther)
		return
	}
	renderLogin(w, "")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if !config.UIRegistration || config.UIAuthFile == "" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
		if err := r.ParseForm(); err != nil {
			renderRegister(w, "Invalid registration request.")
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")
		if password != r.FormValue("password_confirm") {
			w.WriteHeader(http.StatusBadRequest)
			renderRegister(w, "Passwords do not match.")
			return
		}
		if err := auth.RegisterUIUser(config.UIAuthFile, username, password); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			renderRegister(w, err.Error())
			return
		}

		token, expires, err := auth.IssueUISession(username)
		if err != nil {
			http.Error(w, "Could not create session.", http.StatusInternalServerError)
			return
		}
		setSessionCookie(w, r, token, expires)
		http.Redirect(w, r, config.Webroot, http.StatusSeeOther)
		return
	}

	renderRegister(w, "")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	auth.RevokeUISession(r)
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Path:     config.Webroot,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, config.Webroot+"login", http.StatusSeeOther)
}

func renderLogin(w http.ResponseWriter, errorMessage string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; form-action 'self'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	data := struct {
		Action       string
		RegisterURL  string
		Registration bool
		Error        string
	}{
		Action:       config.Webroot + "login",
		RegisterURL:  config.Webroot + "register",
		Registration: config.UIRegistration && config.UIAuthFile != "",
		Error:        errorMessage,
	}
	_ = template.Must(template.New("login").Parse(loginPage)).Execute(w, data)
}

func renderRegister(w http.ResponseWriter, errorMessage string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; form-action 'self'")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	data := struct {
		Action   string
		LoginURL string
		Error    string
	}{Action: config.Webroot + "register", LoginURL: config.Webroot + "login", Error: errorMessage}
	_ = template.Must(template.New("register").Parse(registerPage)).Execute(w, data)
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    token,
		Path:     config.Webroot,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})
}

const loginPage = `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<meta name="robots" content="noindex,nofollow">
	<title>Sign in · Dispatch Inbox</title>
	<style>
		*{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:#050a12;color:#eef6ff;font:16px system-ui,sans-serif}
		body:before{content:"";position:fixed;inset:0;background:radial-gradient(circle at 70% 10%,rgba(0,119,255,.2),transparent 42rem);pointer-events:none}
		main{position:relative;width:min(92vw,430px);padding:32px;border:1px solid #1d2a3b;border-radius:24px;background:rgba(10,17,29,.94);box-shadow:0 24px 80px rgba(0,0,0,.35)}
		header{display:flex;align-items:center;gap:14px;margin-bottom:28px}img{width:52px;height:52px}h1{font-size:22px;margin:0}header span{display:block;color:#16d9f5;font-size:13px;margin-top:3px}
		label{display:block;color:#a9bad0;font-size:13px;font-weight:650;margin:16px 0 7px}input{width:100%;padding:13px 14px;border:1px solid #26364b;border-radius:12px;background:#070e19;color:#fff;font:inherit;outline:none}input:focus{border-color:#16d9f5;box-shadow:0 0 0 3px rgba(22,217,245,.12)}
		button{width:100%;margin-top:22px;padding:13px;border:0;border-radius:12px;background:#16d9f5;color:#031018;font:700 15px system-ui;cursor:pointer}button:hover{background:#46e4fa}
		.error{padding:11px 13px;border:1px solid rgba(255,69,103,.4);border-radius:10px;background:rgba(255,69,103,.08);color:#ff8298;font-size:14px}
		.switch{margin-top:20px;text-align:center;color:#8fa2b8;font-size:13px}.switch a{color:#16d9f5;text-decoration:none}
		footer{margin-top:22px;color:#6f8199;font-size:12px;text-align:center}
	</style>
</head>
<body><main>
	<header><img src="dispatch-inbox.svg" alt=""><div><h1>Dispatch Inbox</h1><span>Your notifications, organized.</span></div></header>
	{{if .Error}}<div class="error">{{.Error}}</div>{{end}}
	<form method="post" action="{{.Action}}">
		<label for="username">Username</label><input id="username" name="username" autocomplete="username" required autofocus>
		<label for="password">Password</label><input id="password" name="password" type="password" autocomplete="current-password" required>
		<button type="submit">Sign in</button>
	</form>
	{{if .Registration}}<div class="switch">New to Dispatch Inbox? <a href="{{.RegisterURL}}">Create account</a></div>{{end}}
	<footer>Part of the Dispatch notification ecosystem</footer>
</main></body></html>`

const registerPage = `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width,initial-scale=1">
	<meta name="robots" content="noindex,nofollow">
	<title>Create account · Dispatch Inbox</title>
	<style>
		*{box-sizing:border-box}body{margin:0;min-height:100vh;display:grid;place-items:center;background:#050a12;color:#eef6ff;font:16px system-ui,sans-serif}
		body:before{content:"";position:fixed;inset:0;background:radial-gradient(circle at 70% 10%,rgba(0,119,255,.2),transparent 42rem);pointer-events:none}
		main{position:relative;width:min(92vw,430px);padding:32px;border:1px solid #1d2a3b;border-radius:24px;background:rgba(10,17,29,.94);box-shadow:0 24px 80px rgba(0,0,0,.35)}
		header{display:flex;align-items:center;gap:14px;margin-bottom:24px}img{width:52px;height:52px}h1{font-size:22px;margin:0}header span{display:block;color:#16d9f5;font-size:13px;margin-top:3px}
		label{display:block;color:#a9bad0;font-size:13px;font-weight:650;margin:14px 0 7px}input{width:100%;padding:13px 14px;border:1px solid #26364b;border-radius:12px;background:#070e19;color:#fff;font:inherit;outline:none}input:focus{border-color:#16d9f5;box-shadow:0 0 0 3px rgba(22,217,245,.12)}
		button{width:100%;margin-top:22px;padding:13px;border:0;border-radius:12px;background:#16d9f5;color:#031018;font:700 15px system-ui;cursor:pointer}button:hover{background:#46e4fa}
		.error{padding:11px 13px;border:1px solid rgba(255,69,103,.4);border-radius:10px;background:rgba(255,69,103,.08);color:#ff8298;font-size:14px}
		.hint{color:#70839a;font-size:12px;margin-top:6px}.switch{margin-top:20px;text-align:center;color:#8fa2b8;font-size:13px}.switch a{color:#16d9f5;text-decoration:none}footer{margin-top:18px;color:#6f8199;font-size:12px;text-align:center}
	</style>
</head>
<body><main>
	<header><img src="dispatch-inbox.svg" alt=""><div><h1>Create your account</h1><span>Dispatch Inbox</span></div></header>
	{{if .Error}}<div class="error">{{.Error}}</div>{{end}}
	<form method="post" action="{{.Action}}">
		<label for="username">Username</label><input id="username" name="username" minlength="3" maxlength="32" pattern="[A-Za-z0-9][A-Za-z0-9._-]{2,31}" autocomplete="username" required autofocus>
		<div class="hint">3–32 characters: letters, numbers, dot, dash or underscore.</div>
		<label for="password">Password</label><input id="password" name="password" type="password" minlength="12" maxlength="128" autocomplete="new-password" required>
		<div class="hint">Use at least 12 characters.</div>
		<label for="password-confirm">Repeat password</label><input id="password-confirm" name="password_confirm" type="password" minlength="12" maxlength="128" autocomplete="new-password" required>
		<button type="submit">Create account</button>
	</form>
	<div class="switch">Already registered? <a href="{{.LoginURL}}">Sign in</a></div>
	<footer>Part of the Dispatch notification ecosystem</footer>
</main></body></html>`
