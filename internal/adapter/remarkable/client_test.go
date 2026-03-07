package remarkable

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token/json/2/device/new", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("fake-device-token-jwt"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	token, err := client.RegisterDevice("abcd1234")
	require.NoError(t, err)
	assert.Equal(t, "fake-device-token-jwt", token)
}

func TestRefreshUserToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/token/json/2/user/new", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer fake-device-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("fake-user-token-jwt"))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	token, err := client.RefreshUserToken("fake-device-token")
	require.NoError(t, err)
	assert.Equal(t, "fake-user-token-jwt", token)
}

func TestRegisterDeviceFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.RegisterDevice("badcode1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
