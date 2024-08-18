package session

import (
	"net"
	"net/http"
	"net/textproto"
	"strings"
)

// These are transient session keys
const (
	// Is the session authenticated: 0: no; 1: yes
	Authenticated = "authenticated"
	// The user's ID
	UserID = "user_id"
	// The user's name
	Username = "username"
	// The user's type: admin|<nil>
	UserType = "user_type"
	// The time when the session is created
	Created = "created"
	// The time when the session is last updated
	Updated = "updated"
	// The path where the session is last updated
	Path = "path"
	// User-Agent request header
	UserAgent = "user_agent"
	// The user's IP address
	UserIPAddr = "user_ipaddr"
)

// GetUserAddr from the http request, considering X-Forwarded-For, Forwarded
func GetUserAddr(r *http.Request) string {
	if xFwd := r.Header.Get("X-Forwarded-For"); xFwd != "" {
		client, _, _ := strings.Cut(xFwd, ",")
		return textproto.TrimString(client)
	}
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		client, _, _ := strings.Cut(fwd, ",")
		client = readCookieValue(client, "for")
		return stripPort(client)
	}
	return stripPort(r.RemoteAddr)
}

func stripPort(addr string) string {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

func readCookieValue(line, filter string) string {
	line = textproto.TrimString(line)

	var part string
	for len(line) > 0 { // continue since we have rest
		part, line, _ = strings.Cut(line, ";")
		part = textproto.TrimString(part)
		if part == "" {
			continue
		}
		name, val, _ := strings.Cut(part, "=")
		name = textproto.TrimString(name)
		if !isCookieNameValid(name) {
			continue
		}
		if filter != "" && !strings.EqualFold(filter, name) {
			continue
		}
		val, ok := parseCookieValue(val, true)
		if !ok {
			continue
		}
		return val
	}

	return ""
}
