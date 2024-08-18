package session

import (
	_ "embed"
	"encoding/base32"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var _ sessions.Store = (*redisStore)(nil)

var (
	//go:embed hmsetex.lua
	hmsetExLua    string
	hmsetExScript = redis.NewScript(hmsetExLua)
)

// redisStore stores sessions in the redis.
type redisStore struct {
	Codecs     []securecookie.Codec
	Options    *sessions.Options // default configuration
	store      *redis.Redis
	namespace  string
	gracePerid int
}

func newRedisStore(store *redis.Redis, c SessionConfig) *redisStore {
	rs := &redisStore{
		Codecs: securecookie.CodecsFromPairs([]byte(c.SessionSecret)),
		Options: &sessions.Options{
			Path:     c.SessionCookiePath,
			Domain:   c.SessionCookieDomain,
			MaxAge:   c.SessionCookieTTL,
			SameSite: parseSameSite(c.SessionCookieSameSite),
			Secure:   c.SessionCookieSecure,
			HttpOnly: true,
		},
		store:      store,
		namespace:  c.SessionStorageNamespace + ":",
		gracePerid: c.SessionStorageGracePeriod,
	}

	rs.MaxAge(rs.Options.MaxAge)
	return rs
}

// MaxAge sets the maximum age for the store and the underlying cookie
// implementation. Individual sessions can be deleted by setting Options.MaxAge
// = -1 for that session.
func (s *redisStore) MaxAge(age int) {
	s.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

// Get returns a session for the given name after adding it to the registry.
func (s *redisStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
func (s *redisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := Token(r, name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			if err := s.load(session); err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

var base32RawStdEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// Save adds a single session to the response.
//
// If the Options.MaxAge of the session is <= 0 then the session item will be
// deleted from the redis. With this process it enforces the properly
// session cookie handling so no need to trust in the cookie management in the
// web browser.
func (s *redisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		if err := s.erase(session); err != nil {
			return err
		}
		SetToken(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		// Because the ID is used in the filename, encode it to
		// use alphanumeric characters only.
		session.ID = base32RawStdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32))
	}
	// Don't save to the store until the middleware finished.
	// if err := s.save(session); err != nil {
	// 	return err
	// }
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID,
		s.Codecs...)
	if err != nil {
		return err
	}
	SetToken(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// save writes encoded session.Values to redis.
func (s *redisStore) save(session *sessions.Session) error {
	// Deleted if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		return nil
	}
	if len(session.Values) == 0 {
		return nil
	}

	args := make([]any, 0, 1+len(session.Values)*2)
	args = append(args, session.Options.MaxAge+s.gracePerid)
	for k, v := range session.Values {
		if sk, ok := k.(string); ok {
			if sv, err := jsonx.MarshalToString(v); err == nil {
				args = append(args, sk, sv)
			}
		}
	}
	_, err := s.store.ScriptRun(hmsetExScript, []string{s.namespace + session.ID}, args...)
	return err
}

// load reads from redis and decodes its content into session.Values.
func (s *redisStore) load(session *sessions.Session) error {
	fvs, err := s.store.Hgetall(s.namespace + session.ID)
	if err != nil {
		return err
	}
	if len(fvs) == 0 {
		return redis.Nil
	}
	for k, v := range fvs {
		var iv interface{}
		if err := jsonx.UnmarshalFromString(v, &iv); err == nil {
			session.Values[k] = iv
		} else {
			logx.Errorw("Invalid Session Value", logx.Field("SessionID", session.ID), logx.Field("Key", k), logx.Field("Value", v))
		}
	}
	return nil
}

// delete session item
func (s *redisStore) erase(session *sessions.Session) error {
	_, err := s.store.Del(s.namespace + session.ID)
	return err
}
