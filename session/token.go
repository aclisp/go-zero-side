package session

import (
	"errors"
	"net/http"
	"net/textproto"
	"strings"
)

var ErrNoToken = errors.New("http: named token not present")

// Token returns the named token provided in the request or [ErrNoToken] if not found.
// If multiple tokens match the given name, only one token will be returned.
func Token(r *http.Request, name string) (*http.Cookie, error) {
	if name == "" {
		return nil, ErrNoToken
	}
	for _, c := range readCookiesByHeaderName(r.Header, "Session-Token", name) {
		return c, nil
	}
	for _, c := range readCookiesByHeaderName(r.Header, "Cookie", name) {
		return c, nil
	}
	return nil, ErrNoToken
}

// SetToken adds a Set-Session-Token header to the provided [ResponseWriter]'s headers.
// The provided cookie must have a valid Name. Invalid cookies may be silently dropped.
func SetToken(w http.ResponseWriter, cookie *http.Cookie) {
	if v := cookie.String(); v != "" {
		w.Header().Add("Set-Session-Token", v)
		w.Header().Add("Set-Cookie", v)
	}
}

func readCookiesByHeaderName(h http.Header, header, filter string) []*http.Cookie {
	lines := h[header]
	if len(lines) == 0 {
		return []*http.Cookie{}
	}

	cookies := make([]*http.Cookie, 0, len(lines)+strings.Count(lines[0], ";"))
	for _, line := range lines {
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
			if filter != "" && filter != name {
				continue
			}
			val, ok := parseCookieValue(val, true)
			if !ok {
				continue
			}
			cookies = append(cookies, &http.Cookie{Name: name, Value: val})
		}
	}
	return cookies
}

func parseCookieValue(raw string, allowDoubleQuote bool) (string, bool) {
	// Strip the quotes, if present.
	if allowDoubleQuote && len(raw) > 1 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	for i := 0; i < len(raw); i++ {
		if !validCookieValueByte(raw[i]) {
			return "", false
		}
	}
	return raw, true
}

func validCookieValueByte(b byte) bool {
	return 0x20 <= b && b < 0x7f && b != '"' && b != ';' && b != '\\'
}

func parseSameSite(sameSite string) http.SameSite {
	switch sameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
