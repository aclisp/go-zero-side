package session

//lint:file-ignore SA5008 Use gozero config tags
//nolint:staticcheck
type SessionConfig struct {
	SessionSecret           string // used to authenticate session cookies using HMAC
	SessionStorageNamespace string `json:",default=sessions"`
	SessionCookieName       string `json:",default=SID"`
	SessionCookiePath       string `json:",default=/"`
	SessionCookieDomain     string `json:",optional"`
	// The duration in seconds that the session cookie/token is valid,
	// and also how long users stay logged-in to the App.
	SessionCookieTTL      int    `json:",default=600,range=[60:]"`
	SessionCookieSameSite string `json:",default=Lax,options=Strict|Lax|None"`
	SessionCookieSecure   bool   `json:",default=false"`
	// The session storage TTL is derived from its max age plus this grace period.
	SessionStorageGracePeriod               int `json:",default=10,range=[1:60]"`
	SessionStorageUnauthenticatedTTL        int `json:",default=60,range=[0:600]"`
	SessionStorageInjectedAuthenticationTTL int `json:",default=0,range=[0:60]"`
}
