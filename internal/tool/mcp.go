package tool

import "encoding/json"

// ===== MCP Tool 定义（tools/list 返回项） =====

type mcpTool struct {
	Name         string             `json:"name"`                   // 工具唯一标识
	Title        string             `json:"title,omitempty"`        // 供人阅读的展示名
	Description  string             `json:"description,omitempty"`  // 供模型理解的功能描述
	InputSchema  mcpSchema          `json:"inputSchema"`            // JSON Schema，即入参定义
	OutputSchema *mcpSchema         `json:"outputSchema,omitempty"` // 可选，结构化输出的校验 schema
	Annotations  mcpToolAnnotations `json:"annotations,omitempty"`  // 行为提示，不可全信
}

// mcpSchema 就是标准 JSON Schema，这里原样透传。
type mcpSchema = json.RawMessage

type mcpToolAnnotations struct {
	Title           string `json:"title,omitempty"`
	ReadOnlyHint    bool   `json:"readOnlyHint,omitempty"`    // 只读，不产生副作用
	DestructiveHint bool   `json:"destructiveHint,omitempty"` // 可能产生破坏性副作用
	IdempotentHint  bool   `json:"idempotentHint,omitempty"`  // 重复调用结果一致
	OpenWorldHint   bool   `json:"openWorldHint,omitempty"`   // 与外部开放世界交互
}

// ===== MCP Tool 调用（tools/call） =====

type mcpToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"` // 需符合 inputSchema
}

type mcpToolResult struct {
	Content           []mcpContentBlock `json:"content"`                     // 非结构化内容块
	StructuredContent json.RawMessage   `json:"structuredContent,omitempty"` // 结构化输出，需符合 outputSchema
	IsError           bool              `json:"isError,omitempty"`           // true 表示执行失败，非协议错误
}

// mcpContentBlock 是 text | image | audio | resource_link | resource 的联合类型。
type mcpContentBlock struct {
	Type        string          `json:"type"`
	Text        string          `json:"text,omitempty"`        // text
	Data        string          `json:"data,omitempty"`        // image / audio：base64
	MimeType    string          `json:"mimeType,omitempty"`    // image / audio / resource_link
	URI         string          `json:"uri,omitempty"`         // resource_link / resource
	Name        string          `json:"name,omitempty"`        // resource_link
	Description string          `json:"description,omitempty"` // resource_link
	Resource    *mcpResource    `json:"resource,omitempty"`    // resource：内嵌资源
	Annotations *mcpAnnotations `json:"annotations,omitempty"` // 受众、优先级等元数据
}

type mcpResource struct {
	URI         string          `json:"uri"`
	MimeType    string          `json:"mimeType,omitempty"`
	Text        string          `json:"text,omitempty"`
	Blob        string          `json:"blob,omitempty"`
	Annotations *mcpAnnotations `json:"annotations,omitempty"`
}

type mcpAnnotations struct {
	Audience     []string `json:"audience,omitempty"` // user | assistant
	Priority     float64  `json:"priority,omitempty"`
	LastModified string   `json:"lastModified,omitempty"`
}

// ===== tools/list 响应 =====

type mcpListToolsResult struct {
	Tools      []mcpTool `json:"tools"`
	NextCursor string    `json:"nextCursor,omitempty"` // 非空表示有下一页
}
