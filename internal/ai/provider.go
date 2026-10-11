package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Message is one chat message sent to a provider. Parts, when present, make
// the message multimodal: Content becomes the first text part. Without Parts
// the message marshals as the plain {"role","content":"…"} shape every
// OpenAI-compatible provider understands.
type Message struct {
	Role    string
	Content string
	Parts   []Part
}

// MarshalJSON picks the string or array content shape (see Message).
func (m Message) MarshalJSON() ([]byte, error) {
	if len(m.Parts) == 0 {
		return json.Marshal(struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{m.Role, m.Content})
	}
	parts := make([]Part, 0, len(m.Parts)+1)
	if m.Content != "" {
		parts = append(parts, TextPart(m.Content))
	}
	parts = append(parts, m.Parts...)
	return json.Marshal(struct {
		Role    string `json:"role"`
		Content []Part `json:"content"`
	}{m.Role, parts})
}

// ChatRequest is the internal request contract.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Provider streams text deltas. Implementations must respect ctx cancellation.
type Provider interface {
	StreamChat(ctx context.Context, req ChatRequest, onChunk func(string) error) error
	Name() string
}

// MockProvider emits canned tokens with delay. No network, no secrets.
type MockProvider struct {
	Chunks []string
	Delay  time.Duration
}

func (m *MockProvider) Name() string { return "mock" }

func (m *MockProvider) StreamChat(ctx context.Context, req ChatRequest, onChunk func(string) error) error {
	chunks := m.Chunks
	if len(chunks) == 0 {
		chunks = []string{"Halo! ", "ini ", "mock ", "stream."}
	}
	d := m.Delay
	if d == 0 {
		d = 60 * time.Millisecond
	}
	for _, c := range chunks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}
		if err := onChunk(c); err != nil {
			return err
		}
	}
	return nil
}

// NormalizeBaseURL validates scheme/host and joins the chat completions path.
func NormalizeBaseURL(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid base url")
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("base url must be http or https")
	}
	if u.User != nil {
		return "", fmt.Errorf("base url must not contain credentials")
	}
	p := strings.TrimSuffix(u.Path, "/")
	if !strings.HasSuffix(p, "/chat/completions") {
		p += "/chat/completions"
	}
	u.Path = p
	return u.String(), nil
}

// OpenAICompatible is the baseline HTTP+SSE client. API key passed per call, never stored.
type OpenAICompatible struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func (o *OpenAICompatible) Name() string { return "openai-compatible" }

func (o *OpenAICompatible) StreamChat(ctx context.Context, req ChatRequest, onChunk func(string) error) error {
	endpoint, err := NormalizeBaseURL(o.BaseURL)
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	})
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if o.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+o.APIKey)
	}
	client := o.Client
	if client == nil {
		client = SafeClient()
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return toNetError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return MapHTTPErrorWithHeaders(resp.StatusCode, resp.Header)
	}
	return ScanSSE(ctx, bufio.NewReader(resp.Body), func(c StreamChunk) error {
		if c.Done || c.Text == "" {
			return nil
		}
		return onChunk(c.Text)
	})
}
