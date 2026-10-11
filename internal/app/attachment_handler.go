package app

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var attachmentIDRe = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// AttachmentHandler serves staged or stored attachment bytes to the WebView
// at GET /attachments/{id}. Only a well-formed uuid is looked up; everything
// else is 404, so the route can never be used to probe paths or files.
func (s *Service) AttachmentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id, ok := strings.CutPrefix(r.URL.Path, "/attachments/")
		if !ok || !attachmentIDRe.MatchString(id) {
			http.NotFound(w, r)
			return
		}
		mime, data, found := s.Attachment(id)
		if !found {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", mime)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "private, max-age=3600")
		h.Set("Content-Disposition", "inline")
		h.Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})
}
