package gatekeeper_test

import (
	"github.com/Thooms/gatekeeper"
	"github.com/Thooms/gatekeeper/backend/inmemory"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func getDummyHandler() (http.Handler, func() bool) {
	reachedHandler := false
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reachedHandler = true
		io.WriteString(w, "OK")
	}), func() bool { return reachedHandler }
}

/*
req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
*/

func TestNominalCaseWithDefaults(t *testing.T) {
	gkBackend := inmemory.New()
	gkBackend.Set("someAPIKey", 10)
	gk := gatekeeper.FromKeeper(gkBackend)

	req := httptest.NewRequest("GET", "http://acme.acme", nil)
	req.Header.Add("APIKEY", "someAPIKey")
	w := httptest.NewRecorder()

	dummyHandler, hasReachedHandler := getDummyHandler()
	gk.Wrap(dummyHandler).ServeHTTP(w, req)
	require.True(t, hasReachedHandler()) // we passed through
	result := w.Result()
	require.Equal(t, "10", result.Header.Get("X-API-Limit"))    // API limits are good
	require.Equal(t, "9", result.Header.Get("X-API-Remaining")) // API limits are good
}

func TestNotAllowed(t *testing.T) {
	gkBackend := inmemory.New()
	gkBackend.Set("someAPIKey", 0)
	gk := gatekeeper.FromKeeper(gkBackend)

	req := httptest.NewRequest("GET", "http://acme.acme", nil)
	req.Header.Add("APIKEY", "someAPIKey")
	w := httptest.NewRecorder()

	dummyHandler, hasReachedHandler := getDummyHandler()
	gk.Wrap(dummyHandler).ServeHTTP(w, req)
	require.False(t, hasReachedHandler())
	body, _ := ioutil.ReadAll(w.Result().Body)
	require.Equal(t, "limit reached\n", string(body))
	require.Equal(t, http.StatusTooManyRequests, w.Result().StatusCode)
}

func TestUnknownKey(t *testing.T) {
	gkBackend := inmemory.New()
	gk := gatekeeper.FromKeeper(gkBackend)

	req := httptest.NewRequest("GET", "http://acme.acme", nil)
	req.Header.Add("APIKEY", "unkonwnAPIKey")
	w := httptest.NewRecorder()

	dummyHandler, hasReachedHandler := getDummyHandler()
	gk.Wrap(dummyHandler).ServeHTTP(w, req)
	require.False(t, hasReachedHandler())
	body, _ := ioutil.ReadAll(w.Result().Body)
	require.Equal(t, "unknown API key\n", string(body))
	require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestMissingKey(t *testing.T) {
	gkBackend := inmemory.New()
	gk := gatekeeper.FromKeeper(gkBackend)

	req := httptest.NewRequest("GET", "http://acme.acme", nil)
	w := httptest.NewRecorder()

	dummyHandler, hasReachedHandler := getDummyHandler()
	gk.Wrap(dummyHandler).ServeHTTP(w, req)
	require.False(t, hasReachedHandler())
	body, _ := ioutil.ReadAll(w.Result().Body)
	require.Equal(t, "missing API key\n", string(body))
	require.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}
