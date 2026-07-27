package agent

import (
	"strings"
	"testing"
)

func collect(chunks <-chan StreamChunk) (thought, answer string) {
	var tb, ab strings.Builder
	for c := range chunks {
		switch c.Type {
		case ChunkThought:
			tb.WriteString(c.Content)
		case ChunkAnswer:
			ab.WriteString(c.Content)
		}
	}
	return tb.String(), ab.String()
}

func runParser(tokens []string) (thought, answer string) {
	out := make(chan StreamChunk, 64)
	p := NewTagStreamParser()
	for _, t := range tokens {
		p.Feed(t, out)
	}
	p.Flush(out)
	close(out)
	return collect(out)
}

func TestTagStreamParser_FullTags(t *testing.T) {
	th, an := runParser([]string{"【思考】", "我在", "想事情", "【回答】", "你好", "世界"})
	if th != "我在想事情" || an != "你好世界" {
		t.Fatalf("unexpected thought=%q answer=%q", th, an)
	}
}

func TestTagStreamParser_TagSplitAcrossChunks(t *testing.T) {
	th, an := runParser([]string{"【思", "考】想一想", "【回", "答】答案"})
	if th != "想一想" || an != "答案" {
		t.Fatalf("unexpected thought=%q answer=%q", th, an)
	}
}

func TestTagStreamParser_NoTags(t *testing.T) {
	// 模型未输出标签 → 全部作为 answer
	th, an := runParser([]string{"hello", " ", "world"})
	if th != "" || an != "hello world" {
		t.Fatalf("unexpected thought=%q answer=%q", th, an)
	}
}

func TestTagStreamParser_OnlyThought(t *testing.T) {
	th, an := runParser([]string{"【思考】只有思考没有回答"})
	if th != "只有思考没有回答" || an != "" {
		t.Fatalf("unexpected thought=%q answer=%q", th, an)
	}
}

func TestTagStreamParser_TextBeforeTag(t *testing.T) {
	th, an := runParser([]string{"引导语", "【思考】t", "【回答】a"})
	// 引导语在 initial 状态 → answer
	if th != "t" || an != "引导语a" {
		t.Fatalf("unexpected thought=%q answer=%q", th, an)
	}
}
