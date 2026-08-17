package agent

import (
	"strings"
	"testing"

	"github.com/ollama/ollama/api"
)

// newAgentServiceForTest 构造仅关注 trimHistory 的最小 AgentService。
func newAgentServiceForTest(sysPrompt string, maxChars, maxMsgs int) *AgentService {
	return &AgentService{
		systemPrompt: sysPrompt,
		maxChars:     maxChars,
		maxMsgs:      maxMsgs,
	}
}

func TestTrimHistory_Empty(t *testing.T) {
	s := newAgentServiceForTest("SYS", 1000, 8)
	got := s.trimHistory(nil, "hi")
	if got != nil {
		t.Errorf("expected nil for empty history, got %d msgs", len(got))
	}
}

func TestTrimHistory_WithinBudget(t *testing.T) {
	s := newAgentServiceForTest("SYS", 1000, 8)
	history := []api.Message{
		{Role: "user", Content: "u1"},
		{Role: "assistant", Content: "a1"},
		{Role: "user", Content: "u2"},
		{Role: "assistant", Content: "a2"},
	}
	got := s.trimHistory(history, "curr")
	if len(got) != 4 {
		t.Fatalf("expected 4 msgs, got %d", len(got))
	}
	if got[0].Content != "u1" || got[3].Content != "a2" {
		t.Errorf("order broken: %+v", got)
	}
}

func TestTrimHistory_TrimTail(t *testing.T) {
	// 预算刚够 2 条历史（每条 100 字符）+ userMessage(10)
	s := newAgentServiceForTest("SYS", 210, 8)
	long := strings.Repeat("x", 100)
	history := []api.Message{
		{Role: "user", Content: long},
		{Role: "assistant", Content: long},
		{Role: "user", Content: long},
		{Role: "assistant", Content: long},
	}
	got := s.trimHistory(history, "curr(10ch)")
	if len(got) != 2 {
		t.Fatalf("expected 2 msgs kept, got %d", len(got))
	}
	// 最近 2 条应保留（顺序为升序）
	if got[0].Role != "user" || got[1].Role != "assistant" {
		t.Errorf("expected last user+assistant, got %+v", got)
	}
}

func TestTrimHistory_FallbackWhenUserMessageOverBudget(t *testing.T) {
	// userMessage 直接超预算，仍应兜底保留 1 条最近历史
	s := newAgentServiceForTest("SYS", 50, 8)
	history := []api.Message{
		{Role: "user", Content: "u1"},
		{Role: "assistant", Content: "a1"},
	}
	got := s.trimHistory(history, strings.Repeat("y", 100))
	if len(got) != 1 {
		t.Fatalf("expected 1 fallback msg, got %d", len(got))
	}
	if got[0].Content != "a1" {
		t.Errorf("expected last message 'a1', got %q", got[0].Content)
	}
}

func TestTrimHistory_FallbackWhenFirstHistoryTooLarge(t *testing.T) {
	// 每条历史单条即超预算，兜底至少保留 1 条
	s := newAgentServiceForTest("SYS", 30, 8)
	history := []api.Message{
		{Role: "user", Content: strings.Repeat("a", 200)},
		{Role: "assistant", Content: strings.Repeat("b", 200)},
	}
	got := s.trimHistory(history, "hi")
	if len(got) != 1 {
		t.Fatalf("expected 1 fallback msg, got %d", len(got))
	}
}

func TestTrimHistory_MaxMsgsCap(t *testing.T) {
	s := newAgentServiceForTest("SYS", 100000, 3) // 预算大但条数上限 3
	history := []api.Message{
		{Role: "user", Content: "u1"},
		{Role: "assistant", Content: "a1"},
		{Role: "user", Content: "u2"},
		{Role: "assistant", Content: "a2"},
		{Role: "user", Content: "u3"},
	}
	got := s.trimHistory(history, "curr")
	if len(got) != 3 {
		t.Fatalf("expected 3 msgs (maxMsgs cap), got %d", len(got))
	}
	if got[2].Content != "u3" {
		t.Errorf("expected latest at tail, got %+v", got)
	}
}

func TestTrimHistory_SkipUnknownRole(t *testing.T) {
	s := newAgentServiceForTest("SYS", 1000, 8)
	history := []api.Message{
		{Role: "system", Content: "should be skipped"},
		{Role: "user", Content: "u1"},
		{Role: "tool", Content: "should be skipped"},
		{Role: "assistant", Content: "a1"},
	}
	got := s.trimHistory(history, "curr")
	if len(got) != 2 {
		t.Fatalf("expected 2 msgs (only user/assistant), got %d", len(got))
	}
}

func TestGetContextBudget(t *testing.T) {
	cases := []struct {
		mode     string
		wantMax  int
		wantMsgs int
	}{
		{"", 64000, 40},
		{"2k", 2000, 8},
		{"4k", 4000, 12},
		{"8k", 8000, 16},
		{"16k", 16000, 20},
		{"32k", 32000, 30},
		{"64k", 64000, 40},
		{"unknown", 64000, 40},
		{"  2K  ", 2000, 8}, // 大小写 + 空白
	}
	for _, c := range cases {
		gotMax, gotMsgs := getContextBudget(c.mode)
		if gotMax != c.wantMax || gotMsgs != c.wantMsgs {
			t.Errorf("getContextBudget(%q) = (%d, %d), want (%d, %d)",
				c.mode, gotMax, gotMsgs, c.wantMax, c.wantMsgs)
		}
	}
}
