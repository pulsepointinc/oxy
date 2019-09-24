package roundrobin

import (
	"net/http"
	"net/url"
)

// StickySession is a mixin for load balancers that implements layer 7 (http cookie) session affinity
type StickySession struct {
	cookieName string
	options    CookieOptions
}

// CookieOptions has all the options one would like to set on the affinity cookie
type CookieOptions struct {
	HTTPOnly       bool
	Secure         bool
	ChromeSameSite bool
}

// NewStickySession creates a new StickySession
func NewStickySession(cookieName string) *StickySession {
	return &StickySession{cookieName: cookieName}
}

// NewStickySessionWithOptions creates a new StickySession whilst allowing for options to
// shape its affinity cookie such as "httpOnly" or "secure"
func NewStickySessionWithOptions(cookieName string, options CookieOptions) *StickySession {
	return &StickySession{cookieName: cookieName, options: options}
}

// GetBackend returns the backend URL stored in the sticky cookie, iff the backend is still in the valid list of servers.
func (s *StickySession) GetBackend(req *http.Request, servers []*url.URL) (*url.URL, bool, error) {
	cookie, err := req.Cookie(s.cookieName)
	switch err {
	case nil:
	case http.ErrNoCookie:
		return nil, false, nil
	default:
		return nil, false, err
	}

	serverURL, err := url.Parse(cookie.Value)
	if err != nil {
		return nil, false, err
	}

	if s.isBackendAlive(serverURL, servers) {
		return serverURL, true, nil
	}
	return nil, false, nil
}

// StickBackend creates and sets the cookie
func (s *StickySession) StickBackend(backend *url.URL, req *http.Request, w *http.ResponseWriter) {
	opt := s.options
	//if chrome rewriting is enabled and user agent is chrome, set SameSite = True, Secure = True if browser is chrome
	if opt.ChromeSameSite && IsUANewChrome(req.Header.Get("User-Agent")) {
		cookie := &http.Cookie{Name: s.cookieName, Value: backend.String(), Path: "/", HttpOnly: opt.HTTPOnly, Secure: true, SameSite: http.SameSiteNoneMode}
		http.SetCookie(*w, cookie)
	} else {
		cookie := &http.Cookie{Name: s.cookieName, Value: backend.String(), Path: "/", HttpOnly: opt.HTTPOnly, Secure: opt.Secure}
		http.SetCookie(*w, cookie)
	}
}

func (s *StickySession) isBackendAlive(needle *url.URL, haystack []*url.URL) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, serverURL := range haystack {
		if sameURL(needle, serverURL) {
			return true
		}
	}
	return false
}
