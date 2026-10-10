package api

import (
	"encoding/json"
	"honi/internal/agent"
	"time"
	"uuid"

	"github.com/gofiber/fiber/v3"
)

func wechatBot(c fiber.Ctx) error {
	req := &chatRequest{}
	if err := json.Unmarshal(c.Request().Body(), req); err != nil {
		return err
	}
	answer, err := agent.GetAgent().Ask(*req.Messages[len(req.Messages)].Content)
	if err != nil {
		return err
	}
	resp := chatResponse{
		ID:      "chatcmpl-" + uuid.New().String(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []chatChoice{{
			Index:        0,
			Message:      chatMessage{Role: "assistant", Content: &answer},
			FinishReason: "stop",
		}},
	}

	return c.Status(200).JSON(resp)
}

// ===== OpenAI 兼容：入参 =====

type chatRequest struct {
	Model            string        `json:"model"`
	Messages         []chatMessage `json:"messages"`
	Tools            []chatTool    `json:"tools,omitempty"`
	ToolChoice       any           `json:"tool_choice,omitempty"` // "auto"|"none"|"required"|object
	Temperature      *float64      `json:"temperature,omitempty"`
	TopP             *float64      `json:"top_p,omitempty"`
	MaxTokens        *int          `json:"max_tokens,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
	Stop             []string      `json:"stop,omitempty"`
	ResponseFormat   *respFormat   `json:"response_format,omitempty"`
	PresencePenalty  *float64      `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64      `json:"frequency_penalty,omitempty"`
	User             string        `json:"user,omitempty"`
}

type chatMessage struct {
	Role       string     `json:"role"` // system|user|assistant|tool
	Content    *string    `json:"content"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"` // role=tool 时必填
	Name       string     `json:"name,omitempty"`
}

type chatTool struct {
	Type     string       `json:"type"` // 固定 "function"
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"` // JSON Schema 原样透传
}

type toolCall struct {
	Index    *int         `json:"index,omitempty"` // 流式增量用
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function toolCallFunc `json:"function"`
}

type toolCallFunc struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments"` // 字符串，不是对象
}

type respFormat struct {
	Type string `json:"type"` // text | json_object
}

// ===== OpenAI 兼容：非流式出参 =====

type chatResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"` // chat.completion
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   tokenUsage   `json:"usage"`
}

type chatChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"` // stop|length|tool_calls|content_filter
}

type tokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ===== OpenAI 兼容：流式 chunk =====

type chatChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"` // chat.completion.chunk
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []chunkChoice `json:"choices"`
	Usage   *tokenUsage   `json:"usage,omitempty"`
}

type chunkChoice struct {
	Index        int        `json:"index"`
	Delta        chunkDelta `json:"delta"`
	FinishReason *string    `json:"finish_reason"`
}

type chunkDelta struct {
	Role      string     `json:"role,omitempty"`
	Content   *string    `json:"content,omitempty"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}
