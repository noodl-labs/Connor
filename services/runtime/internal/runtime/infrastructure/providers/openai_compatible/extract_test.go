package openai_compatible

import "testing"

func TestAssistantContent_ok(t *testing.T) {
	body := []byte(`{
		"choices":[{"message":{"role":"assistant","content":"{\"ok\":true}"}}]
	}`)
	got, err := AssistantContent(body)
	if err != nil || got != `{"ok":true}` {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestAssistantContent_emptyChoices(t *testing.T) {
	_, err := AssistantContent([]byte(`{"choices":[]}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssistantContent_invalidJSON(t *testing.T) {
	_, err := AssistantContent([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCompletion_toolCallsAndUsage(t *testing.T) {
	body := []byte(`{
		"choices":[{
			"message":{
				"role":"assistant",
				"content":"",
				"tool_calls":[
					{"id":"1","type":"function","function":{"name":"search","arguments":"{}"}},
					{"id":"2","type":"function","function":{"name":"book","arguments":"{}"}}
				]
			}
		}],
		"usage":{"prompt_tokens":100,"completion_tokens":20,"total_tokens":120}
	}`)
	got, err := ParseCompletion(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.ToolCalls) != 2 {
		t.Fatalf("tool_calls: %+v", got.ToolCalls)
	}
	if got.ToolCalls[0].Name != "search" || got.ToolCalls[1].Name != "book" {
		t.Fatalf("names: %+v", got.ToolCalls)
	}
	if got.PromptTokens != 100 || got.CompletionTokens != 20 {
		t.Fatalf("tokens: prompt=%d completion=%d", got.PromptTokens, got.CompletionTokens)
	}
}

func TestParseCompletion_skipsEmptyToolName(t *testing.T) {
	body := []byte(`{
		"choices":[{
			"message":{
				"role":"assistant",
				"content":"hi",
				"tool_calls":[
					{"id":"1","type":"function","function":{"name":"","arguments":"{}"}},
					{"id":"2","type":"function","function":{"name":"search","arguments":"{}"}}
				]
			}
		}]
	}`)
	got, err := ParseCompletion(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.ToolCalls) != 1 || got.ToolCalls[0].Name != "search" {
		t.Fatalf("got %+v", got.ToolCalls)
	}
}
