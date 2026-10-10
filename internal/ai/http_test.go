package ai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSafeClientDropsAuthOnCrossHost proves a redirect to a different host
// does not receive the bearer token (docs 08 security checklist).
func TestSafeClientDropsAuthOnCrossHost(t *testing.T) {
	var leaked string
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		leaked = r.Header.Get("Authorization")
		io.WriteString(w, "ok")
	}))
	defer other.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	defer origin.Close()

	req, _ := http.NewRequest(http.MethodPost, origin.URL, nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp, err := SafeClient().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	if leaked != "" {
		t.Fatalf("cross-host redirect leaked Authorization: %q", leaked)
	}
}

// TestSafeClientKeepsAuthOnSameHost proves normal same-host redirects still
// carry the token, so providers that canonicalise paths keep working.
func TestSafeClientKeepsAuthOnSameHost(t *testing.T) {
	var got string
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/final" {
			got = r.Header.Get("Authorization")
			io.WriteString(w, "ok")
			return
		}
		http.Redirect(w, r, srv.URL+"/final", http.StatusFound)
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/start", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	resp, err := SafeClient().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	if got != "Bearer secret-token" {
		t.Fatalf("same-host redirect dropped auth: %q", got)
	}
}
