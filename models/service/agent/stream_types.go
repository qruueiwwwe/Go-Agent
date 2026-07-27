package agent

// StreamChunkType 流式 chunk 类型
type StreamChunkType string

const (
	ChunkSession     StreamChunkType = "session"      // 会话信息（首帧）
	ChunkThought     StreamChunkType = "thought"      // 思考片段
	ChunkAnswer      StreamChunkType = "answer"       // 答案片段
	ChunkAnswerReset StreamChunkType = "answer_reset" // 通知前端清空 answer 累积（用于工具调用重放）
	ChunkToolCall    StreamChunkType = "tool_call"    // 工具调用
	ChunkToolResult  StreamChunkType = "tool_result"  // 工具执行结果
	ChunkTitle       StreamChunkType = "title"        // 异步生成的会话标题
	ChunkDone        StreamChunkType = "done"         // 流结束
	ChunkError       StreamChunkType = "error"        // 出错
)

// StreamChunk 流式输出的单个片段
type StreamChunk struct {
	Type      StreamChunkType `json:"type"`
	Content   string          `json:"content,omitempty"`
	SessionID string          `json:"session_id,omitempty"`
	Title     string          `json:"title,omitempty"`
	Tool      string          `json:"tool,omitempty"`
	ToolInput string          `json:"tool_input,omitempty"`
}
