package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"strings"
)

// StreamChunk is one parsed text delta from an SSE frame.
type StreamChunk struct {
	Text string
	Done bool
}

type chatStreamPayload struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// ScanSSE reads an OpenAI-compatible SSE stream incrementally.
// It calls onChunk for each text delta. [DONE] or EOF ends cleanly.
// Malformed JSON in one event is skipped, not fatal.
func ScanSSE(ctx context.Context, r *bufio.Reader, onChunk func(StreamChunk) error) error {
	var dataLines []string
	flush := func() error {
		if len(dataLines) == 0 {
			return nil
		}
		raw := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		if raw == "[DONE]" {
			return onChunk(StreamChunk{Done: true})
		}
		var p chatStreamPayload
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return nil
		}
		if len(p.Choices) == 0 {
			return nil
		}
		if t := p.Choices[0].Delta.Content; t != "" {
			if err := onChunk(StreamChunk{Text: t}); err != nil {
				return err
			}
		}
		return nil
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		line, err := r.ReadString('\n')
		if len(line) > 0 {
			// Cap runaway lines: trim, bufio already bounds memory per read.
			if len(line) > 64*1024 {
				dataLines = dataLines[:0]
				continue
			}
			t := strings.TrimRight(line, "\r\n")
			switch {
			case t == "":
				if err := flush(); err != nil {
					return err
				}
			case strings.HasPrefix(t, ":"):
				// comment / keepalive, ignore
			case strings.HasPrefix(t, "data:"):
				v := strings.TrimSpace(strings.TrimPrefix(t, "data:"))
				dataLines = append(dataLines, v)
			default:
				// unknown field, ignore per SSE rules
			}
		}
		if err != nil {
			_ = flush()
			return nil
		}
	}
}
