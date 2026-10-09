package ai

// ChatEvent is the single structured event forwarded to the frontend.
// One variant with type field, per docs 05 section 6.
type ChatEvent struct {
	Type           string       `json:"type"` // started|delta|completed|cancelled|failed
	RequestID      string       `json:"requestId"`
	ConversationID string       `json:"conversationId"`
	MessageID      string       `json:"messageId"`
	Delta          string       `json:"delta,omitempty"`
	Error          *PublicError `json:"error,omitempty"`
}

// PublicError is safe to show in UI. Never carries secrets.
type PublicError struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	Retryable     bool   `json:"retryable"`
	StatusCode    int    `json:"statusCode,omitempty"`
	RetryAfterSec int    `json:"retryAfterSec,omitempty"`
}

func (e *PublicError) Error() string { return e.Code + ": " + e.Message }
