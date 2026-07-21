package openai_compatible

import (
	"encoding/json"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

// ParsedCompletion is the subset of a chat completion Connor needs (RFC 0002 PR-1).
type ParsedCompletion struct {
	Content          string
	ToolCalls        []entities.ToolCall
	PromptTokens     int
	CompletionTokens int
}

// ParseCompletion decodes a non-stream chat.completions body into content, tools, and usage.
func ParseCompletion(body []byte) (ParsedCompletion, error) {
	var resp chatCompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return ParsedCompletion{}, ErrDecodeResponse
	}
	if len(resp.Choices) == 0 {
		return ParsedCompletion{}, ErrEmptyChoices
	}

	msg := resp.Choices[0].Message
	out := ParsedCompletion{
		Content:          msg.Content,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
	}
	if len(msg.ToolCalls) > 0 {
		out.ToolCalls = make([]entities.ToolCall, 0, len(msg.ToolCalls))
		for _, tc := range msg.ToolCalls {
			name := tc.Function.Name
			if name == "" {
				continue
			}
			out.ToolCalls = append(out.ToolCalls, entities.ToolCall{Name: name})
		}
	}
	return out, nil
}

// AssistantContent returns choices[0].message.content from a non-stream response body.
// Response.Body in ConnorLLM = this string (not the raw HTTP envelope).
func AssistantContent(body []byte) (string, error) {
	parsed, err := ParseCompletion(body)
	if err != nil {
		return "", err
	}
	return parsed.Content, nil
}
