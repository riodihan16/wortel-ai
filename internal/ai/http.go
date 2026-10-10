package ai

import (
	"net/http"
)

// SafeClient returns an HTTP client whose redirect policy drops the
// Authorization header when a redirect leaves the original host. Go's default
// policy copies every header onto the next request, so an HTTP 3xx to another
// host would leak the bearer token. Same-host redirects keep the header
// (normal for trailing-slash and path canonicalisation on the provider side).
//
// Timeout is left to the caller.
func SafeClient() *http.Client {
	return &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		// via[0] is the first request in the chain: it carries the
		// Authorization header we must protect.
		if len(via) == 0 {
			return nil
		}
		origin := via[0].URL
		if origin == nil || req.URL == nil {
			req.Header.Del("Authorization")
			return nil
		}
		if origin.Scheme != req.URL.Scheme || origin.Host != req.URL.Host {
			req.Header.Del("Authorization")
		}
		return nil
	}}
}
