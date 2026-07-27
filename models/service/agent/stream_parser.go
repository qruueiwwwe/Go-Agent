package agent

import (
	"strings"
)

// TagStreamParser 基于【思考】/【回答】标签的流式解析器
// 状态：
//
//	StateInitial   → 未遇到任何标签
//	StateInThought → 已进入【思考】
//	StateInAnswer  → 已进入【回答】
type ParserState int

const (
	StateInitial ParserState = iota
	StateInThought
	StateInAnswer
)

const (
	tagThoughtOpen = "【思考】"
	tagAnswerOpen  = "【回答】"
	// 最长可能的标签前缀，用于缓冲判定
	maxTagLen = 12
)

// TagStreamParser 逐字扫描输出，检测标签切换状态并推 chunk
// 兼容模型未输出标签的情况：初始状态下所有内容都作为 answer 推送
type TagStreamParser struct {
	state ParserState
	buf   strings.Builder // 缓存可能是标签的字节
	// 是否见到过 thought 标签（用于判断"未输出标签"的兼容情况）
	sawAnyTag bool
}

// NewTagStreamParser 创建标签解析器
func NewTagStreamParser() *TagStreamParser {
	return &TagStreamParser{state: StateInitial}
}

// Feed 接收一个 token（可能是多字符），产出 chunk 到 out
func (p *TagStreamParser) Feed(token string, out chan<- StreamChunk) {
	if token == "" {
		return
	}
	p.buf.WriteString(token)
	p.drain(out, false)
}

// Flush 结束时冲刷缓冲区
func (p *TagStreamParser) Flush(out chan<- StreamChunk) {
	p.drain(out, true)
}

// drain 尝试消费缓冲区
// 策略：
//   - 检查缓冲区是否以 【 开头，若是则尝试匹配【思考】或【回答】
//     · 完整匹配 → 切换状态并去掉标签
//     · 前缀匹配但不完整 → 若非 final，等待更多数据
//     · 不匹配 → 把缓冲区当作普通文本输出
//   - 其它情况：找到下一个 【 之前的部分整体输出，剩余留在 buf
func (p *TagStreamParser) drain(out chan<- StreamChunk, final bool) {
	for p.buf.Len() > 0 {
		s := p.buf.String()

		// 查找第一个 【
		idx := strings.Index(s, "【")
		if idx == -1 {
			// 没有标签，全部输出
			p.emit(s, out)
			p.buf.Reset()
			return
		}

		// 输出【之前的普通内容
		if idx > 0 {
			p.emit(s[:idx], out)
			p.buf.Reset()
			p.buf.WriteString(s[idx:])
			continue
		}

		// idx == 0：缓冲区从 【 开始，尝试匹配已知标签
		remaining := p.buf.String()
		if strings.HasPrefix(remaining, tagThoughtOpen) {
			p.state = StateInThought
			p.sawAnyTag = true
			p.buf.Reset()
			p.buf.WriteString(remaining[len(tagThoughtOpen):])
			continue
		}
		if strings.HasPrefix(remaining, tagAnswerOpen) {
			p.state = StateInAnswer
			p.sawAnyTag = true
			p.buf.Reset()
			p.buf.WriteString(remaining[len(tagAnswerOpen):])
			continue
		}

		// 前缀可能是标签的一部分，等待更多字节
		if !final && len(remaining) < maxTagLen {
			return
		}

		// final 或超过最大标签长度 → 当作普通文本输出第一个字符
		// 保守做法：输出【本身，然后继续
		firstRune, size := firstRuneOf(remaining)
		p.emit(string(firstRune), out)
		p.buf.Reset()
		p.buf.WriteString(remaining[size:])
	}
}

// emit 根据当前状态推送对应类型的 chunk
func (p *TagStreamParser) emit(text string, out chan<- StreamChunk) {
	if text == "" {
		return
	}
	switch p.state {
	case StateInThought:
		out <- StreamChunk{Type: ChunkThought, Content: text}
	case StateInAnswer:
		out <- StreamChunk{Type: ChunkAnswer, Content: text}
	default:
		// Initial 状态：模型未输出标签 → 全部作为 answer（兼容普通输出）
		out <- StreamChunk{Type: ChunkAnswer, Content: text}
	}
}

// firstRuneOf 返回字符串的首个 rune 及其字节长度
func firstRuneOf(s string) (rune, int) {
	for _, r := range s {
		return r, len(string(r))
	}
	return 0, 0
}

// NativeReasoningEmit 供付费模型直接推送 thought/answer chunk（不经过标签解析）
// content 来自 delta.reasoning_content 时推 thought，delta.content 时推 answer
func NativeReasoningEmit(reasoning, content string, out chan<- StreamChunk) {
	if reasoning != "" {
		out <- StreamChunk{Type: ChunkThought, Content: reasoning}
	}
	if content != "" {
		out <- StreamChunk{Type: ChunkAnswer, Content: content}
	}
}
