package wsproxy_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"cdr.dev/slog/sloggers/slogtest"
	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/enterprise/wsproxy"
	"github.com/coder/coder/v2/enterprise/wsproxy/wsproxysdk"
	"github.com/coder/coder/v2/testutil"
	"github.com/coder/quartz"
)

func TestCryptoKeyCache(t *testing.T) {
	t.Parallel()

	t.Run("Latest", func(t *testing.T) {
		t.Parallel()

		var (
			ctx    = testutil.Context(t, testutil.WaitShort)
			logger = slogtest.Make(t, nil)
			clock  = quartz.NewMock(t)
		)
	})
}

type fakeCoderd struct {
	server *httptest.Server
	keys   []wsproxysdk.CryptoKey
	called int
	url    *url.URL
}

func newFakeCoderd(t *testing.T, keys []wsproxysdk.CryptoKey) *fakeCoderd {
	t.Helper()

	c := &fakeCoderd{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/workspaceproxies/me/crypto-keys", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(wsproxysdk.CryptoKeysResponse{
			CryptoKeys: keys,
		})
		require.NoError(t, err)
		c.called++
	})

	c.server = httptest.NewServer(mux)
	t.Cleanup(c.server.Close)

	var err error
	c.url, err = url.Parse(c.server.URL)
	require.NoError(t, err)

	return c
}

func withClock(clock quartz.Clock) func(*wsproxy.CryptoKeyCache) {
	return func(cache *wsproxy.CryptoKeyCache) {
		cache.Clock = clock
	}
}
