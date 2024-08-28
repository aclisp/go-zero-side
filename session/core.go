package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/zeromicro/go-zero/core/lang"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

var (
	sessionContextKey contextKey
	sessionStore      *redisStore
	sessionConfig     SessionConfig
)

type contextKey struct{}

func Setup(c SessionConfig, store *redis.Redis) {
	if len(c.SessionSecret) != 32 {
		logx.Must(fmt.Errorf("expect a session secret of 32 bytes"))
	}

	sessionStore = newRedisStore(store, c)
	sessionConfig = c
}

func isDev(mode string) bool {
	switch mode {
	case service.DevMode, service.TestMode, service.RtMode:
		return true
	default:
		return false
	}
}

func Middleware(serviceConfMode string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get a session. Get() always returns a session, even if empty.
			session, err := sessionStore.Get(r, sessionConfig.SessionCookieName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Save it before we write to the response/return from the handler.
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Overwrite session values, for debugging purpose only
			injectedAuthentication := false
			if isDev(serviceConfMode) {
				q := r.URL.Query()
				if uid := q.Get("uid"); uid != "" {
					if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
						session.Values[UserID] = id
						session.Values[Username] = "devuser"
						session.Values[UserType] = nil
						session.Values[Authenticated] = 1
						injectedAuthentication = true
						if userType := q.Get("ut"); userType != "" {
							session.Values[UserType] = userType
							session.Values[Username] = "devop"
						}
					}
				}
			}
			session.Values[UserAgent] = r.UserAgent()
			session.Values[UserIPAddr] = GetUserAddr(r)

			// Log session values
			logFieldNames := []string{UserID, Authenticated, Username, UserType, Created, UserAgent, UserIPAddr}
			logFields := make([]logx.LogField, 0, len(logFieldNames)+3)
			for _, fieldName := range logFieldNames {
				logFields = append(logFields, logx.Field(fieldName, session.Values[fieldName]))
			}
			logFields = append(logFields, logx.Field("session_id", session.ID))
			logFields = append(logFields, logx.Field("is_new_session", session.IsNew))
			logFields = append(logFields, logx.Field("path", r.URL.Path))
			logc.Infow(r.Context(), "[Session]", logFields...)

			next(w, r.WithContext(context.WithValue(r.Context(), sessionContextKey, session)))

			// Update session values. Put these after `next` to prevent from changing by handlers.
			// Actually inside `next` we are reading them as `LastUpdated` and `LastPath`
			if session.IsNew {
				session.Values[Created] = time.Now().Format(time.RFC3339)
			}
			session.Values[Updated] = time.Now().Format(time.RFC3339)
			session.Values[Path] = r.URL.Path
			// Use a relatively short age for unauthenticated session, to save capacity of redis storage
			if (Session{session}).GetInt(Authenticated) == 0 {
				session.Options.MaxAge = sessionConfig.SessionStorageUnauthenticatedTTL
			}
			if injectedAuthentication {
				session.Options.MaxAge = sessionConfig.SessionStorageInjectedAuthenticationTTL
			}

			err = sessionStore.save(session)
			if err != nil {
				logx.Errorf("Can not write session.Values to redis: %v", err)
			}
		}
	}
}

type Session struct {
	s *sessions.Session
}

func From(ctx context.Context) Session {
	session := ctx.Value(sessionContextKey).(*sessions.Session)
	return Session{session}
}

func (s Session) Get(key string) any {
	return s.s.Values[key]
}

func (s Session) Set(key string, value any) {
	s.s.Values[key] = value
}

func (s Session) GetInt(key string) int64 {
	value := s.Get(key)
	switch v := value.(type) {
	case nil:
		return 0
	case bool:
		if v {
			return 1
		} else {
			return 0
		}
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return int64(v)
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	logx.Errorf("Session Get(%q) does not have an integer value, the value type is %T", key, value)
	return 0
}

func (s Session) GetStr(key string) string {
	value := s.Get(key)
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		return lang.Repr(v)
	}
}

func (s Session) Del(key string) {
	s.s.Values[key] = nil
}

// Clear is usually used on user logout. It deletes the session both at server and at client.
func (s Session) Clear(r *http.Request, w http.ResponseWriter) {
	s.s.Options.MaxAge = -1
	_ = s.s.Store().Save(r, w, s.s)
}

func (s Session) ID() string {
	return s.s.ID
}

func (s Session) Authenticated() bool {
	return s.GetInt(Authenticated) != 0
}

func (s Session) CreatedAt() time.Time {
	t, _ := time.Parse(time.RFC3339, s.GetStr(Created))
	return t
}

func (s Session) UpdatedAt() time.Time {
	t, _ := time.Parse(time.RFC3339, s.GetStr(Updated))
	return t
}

// Forcely delete a session if we know its ID. Note that it only deletes the session at server side
// while client could initiate a session with a same ID, which he remembered in the past.
// ForceDelete is usually used to forbid a session programmly, maybe upon the user's password change.
func ForceDelete(sessionID string) {
	sessionStore.erase(&sessions.Session{ID: sessionID})
}
