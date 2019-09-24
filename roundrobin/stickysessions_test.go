package roundrobin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/testutils"
)

func TestBasic(t *testing.T) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	sticky := NewStickySession("test")
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.UpsertServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	client := http.DefaultClient

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		require.NoError(t, err)
		assert.Equal(t, "a", string(body))
	}
}

func TestStickCookie(t *testing.T) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	sticky := NewStickySession("test")
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.UpsertServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	resp, err := http.Get(proxy.URL)
	require.NoError(t, err)

	cookie := resp.Cookies()[0]
	assert.Equal(t, "test", cookie.Name)
	assert.Equal(t, a.URL, cookie.Value)
}

func TestStickCookieWithOptions(t *testing.T) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	options := CookieOptions{HTTPOnly: true, Secure: true}
	sticky := NewStickySessionWithOptions("test", options)
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.UpsertServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	resp, err := http.Get(proxy.URL)
	require.NoError(t, err)

	cookie := resp.Cookies()[0]
	assert.Equal(t, "test", cookie.Name)
	assert.Equal(t, a.URL, cookie.Value)
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
}

func TestStickCookieWithChromeSameSite(t *testing.T) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	options := CookieOptions{ChromeSameSite: true}
	sticky := NewStickySessionWithOptions("test", options)
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.UpsertServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	resp, err := http.Get(proxy.URL)
	require.NoError(t, err)

	cookie := resp.Cookies()[0]
	assert.Equal(t, "test", cookie.Name)
	assert.Equal(t, a.URL, cookie.Value)
	assert.False(t, cookie.Secure)
	assert.False(t, cookie.HttpOnly)
	// assert samesite is not set
	assert.Equal(t, http.SameSite(0), cookie.SameSite)

	//lets try again with a chrome 78 user agent (and no stickiness cookie)
	req, _ := http.NewRequest(http.MethodGet, proxy.URL, nil)
	req.Header.Set("User-Agent", "Chrome/78.0")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	cookie = resp.Cookies()[0]
	assert.Equal(t, "test", cookie.Name)
	assert.Equal(t, b.URL, cookie.Value)
	// assert that secure got overriden
	assert.True(t, cookie.Secure)
	assert.False(t, cookie.HttpOnly)
	// assert samesite got set to none
	assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite)

	//try with nil user agent header
	req, _ = http.NewRequest(http.MethodGet, proxy.URL, nil)
	req.Header.Set("User-Agent", "")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	cookie = resp.Cookies()[0]
	assert.Equal(t, "test", cookie.Name)
	assert.Equal(t, a.URL, cookie.Value)
	assert.False(t, cookie.Secure)
	assert.False(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSite(0), cookie.SameSite)
}

func TestRemoveRespondingServer(t *testing.T) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	sticky := NewStickySession("test")
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.UpsertServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	client := http.DefaultClient

	for i := 0; i < 10; i++ {
		req, errReq := http.NewRequest(http.MethodGet, proxy.URL, nil)
		require.NoError(t, errReq)

		req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})

		resp, errReq := client.Do(req)
		require.NoError(t, errReq)
		defer resp.Body.Close()

		body, errReq := ioutil.ReadAll(resp.Body)
		require.NoError(t, errReq)

		assert.Equal(t, "a", string(body))
	}

	err = lb.RemoveServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)

	// Now, use the organic cookie response in our next requests.
	req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})
	resp, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, "test", resp.Cookies()[0].Name)
	assert.Equal(t, b.URL, resp.Cookies()[0].Value)

	for i := 0; i < 10; i++ {
		req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		require.NoError(t, err)
		assert.Equal(t, "b", string(body))
	}
}

func TestRemoveAllServers(t *testing.T) {
	a := testutils.NewResponder("a")
	b := testutils.NewResponder("b")

	defer a.Close()
	defer b.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	sticky := NewStickySession("test")
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.UpsertServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	client := http.DefaultClient

	for i := 0; i < 10; i++ {
		req, errReq := http.NewRequest(http.MethodGet, proxy.URL, nil)
		require.NoError(t, errReq)
		req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})

		resp, errReq := client.Do(req)
		require.NoError(t, errReq)
		defer resp.Body.Close()

		body, errReq := ioutil.ReadAll(resp.Body)
		require.NoError(t, errReq)

		assert.Equal(t, "a", string(body))
	}

	err = lb.RemoveServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)
	err = lb.RemoveServer(testutils.ParseURI(b.URL))
	require.NoError(t, err)

	// Now, use the organic cookie response in our next requests.
	req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "test", Value: a.URL})
	resp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestBadCookieVal(t *testing.T) {
	a := testutils.NewResponder("a")

	defer a.Close()

	fwd, err := forward.New()
	require.NoError(t, err)

	sticky := NewStickySession("test")
	require.NotNil(t, sticky)

	lb, err := New(fwd, EnableStickySession(sticky))
	require.NoError(t, err)

	err = lb.UpsertServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
	defer proxy.Close()

	client := http.DefaultClient

	req, err := http.NewRequest(http.MethodGet, proxy.URL, nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "test", Value: "This is a patently invalid url!  You can't parse it!  :-)"})

	resp, err := client.Do(req)
	require.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "a", string(body))

	// Now, cycle off the good server to cause an error
	err = lb.RemoveServer(testutils.ParseURI(a.URL))
	require.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}
