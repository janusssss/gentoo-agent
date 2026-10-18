package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"honi/config"
	"honi/internal/tool"
	"log/slog"
	"slices"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const minoName = "mimo"

const (
	userRole      = "user"
	assistantRole = "assistant"
	systemRole    = "system"
	developerRole = "developer"

	inputTextType  = "input_text"
	inputImageType = "input_image"
	inputAudioType = "input_audio"
	inputVideType  = "input_video"
)

// ===== OpenAI Responses API：请求体 =====
type functionCallOutput struct {
	CallId string `json:"call_id"`
	Output string `json:"output"`
	Id     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Tyep   string `json:"type"`
}
type inputConstraint interface {
	InputText | InputImage | InputAudio | InputVideo | responsesOutputMessage
}
type Input[T inputConstraint] struct {
	Type    string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
	Role    string `json:"role"` // user|assistant|system|developer
	Content T      `json:"content"`
}
type InputBase struct {
	Type string `json:"type"` // input_text|input_image|input_audio|input_video
}
type InputText struct {
	Text string `json:"text,omitempty"`
	InputBase
}
type InputImage struct {
	ImageUrl string `json:"image_url,omitempty"`
	InputBase
}
type InputAudio struct {
	AudioUrl string `json:"audio_url,omitempty"`
	InputBase
}
type InputVideo struct {
	AudioUrl        string `json:"video_url,omitempty"`
	FPS             int    `json:"fps,omitempty"`
	MediaResolution string `json:"media_resolution,omitempty"`
	InputBase
}

type responsesRequest struct {
	apiKey          string               `json:"-"`
	baseUrl         string               `json:"-"`
	Model           string               `json:"model"`
	Input           Inputs               `json:"input"`
	Instructions    string               `json:"instructions,omitempty"` // 系统/开发者指令
	MaxOutputTokens *int                 `json:"max_output_tokens,omitempty"`
	Stream          bool                 `json:"stream,omitempty"`
	Reasoning       *responsesReasoning  `json:"reasoning,omitempty"`
	Temperature     *float64             `json:"temperature,omitempty"`
	TopP            *float64             `json:"top_p,omitempty"`
	Text            *responsesTextConfig `json:"text,omitempty"`
	ToolChoice      string               `json:"tool_choice,omitempty"` // 仅支持 "auto"
	Tools           []responsesTool      `json:"tools,omitempty"`
}
type Inputs []json.RawMessage

func (i *Inputs) Add[T InputText | InputImage | InputAudio | InputVideo](input Input[T]) {
	buff, _ := json.Marshal(input)
	*i = append(*i, json.RawMessage(buff))
}

type responsesReasoning struct {
	Effort string `json:"effort"` // none|minimal|low|medium|high|xhigh|max|ultra
}

type responsesTextConfig struct {
	Format responsesTextFormat `json:"format"`
}

type responsesTextFormat struct {
	Type string `json:"type"` // text | json_object
}

type responsesTool struct {
	Type        string          `json:"type"` // function
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema 原样透传
	Strict      bool            `json:"strict"`
}

// responsesInput 是 string | array 的联合类型：
// 纯文本直接放 Text，多模态/多轮放 Items。
type responsesInput struct {
	Text  string
	Items []responsesInputItem
}

func (r responsesInput) MarshalJSON() ([]byte, error) {
	if len(r.Items) == 0 && len(r.Text) == 0 {
		return nil, errors.New("input is empty")
	}
	if r.Items != nil {
		return json.Marshal(r.Items)
	}
	return json.Marshal(r.Text)
}

func (r *responsesInput) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		r.Text = s
		return nil
	}
	return json.Unmarshal(data, &r.Items)
}

const (
	responseTypeMessage            = "message"
	responseTypeFunctionCall       = "function_call"
	responseTypeFunctionCallOutput = "function_call_output"
	responseTypeReasoning          = "function_reasoning"
)

// responsesInputItem 是 message | function_call | function_call_output | reasoning 的联合类型。
type responsesInputItem struct {
	Type      string                  `json:"type,omitempty"` // message|function_call|function_call_output|reasoning
	Role      string                  `json:"role,omitempty"` // message 时：user|assistant
	Content   []responsesInputContent `json:"content,omitempty"`
	ID        string                  `json:"id,omitempty"` // function_call / reasoning 的标识
	CallID    string                  `json:"call_id,omitempty"`
	Name      string                  `json:"name,omitempty"`      // function_call 的函数名
	Arguments string                  `json:"arguments,omitempty"` // function_call 的参数 JSON 串
	Output    string                  `json:"output,omitempty"`    // function_call_output 的输出串
}

// responsesInputContent 是 input_text | input_image | input_audio | input_video 的联合类型。
type responsesInputContent struct {
	Type   string `json:"type"`
	Text   string `json:"text,omitempty"`   // input_text
	URL    string `json:"url,omitempty"`    // 图像/音频/视频的地址
	Detail string `json:"detail,omitempty"` // 图像细节档位
}

// ===== OpenAI Responses API：非流式出参 =====
const (
	outputTypeMessage            = "message"
	outputTypeFunctionCall       = "function_call"
	outputTypeFunctionCallOutput = "function_call_output"
	outputTypeReasoning          = "reasoning"
)

type outputItem json.RawMessage

func (o *outputItem) MarshalJSON() ([]byte, error) {
	return json.Marshal(json.RawMessage(*o))
}

func (o *outputItem) UnmarshalJSON(data []byte) error {
	*o = append((*o)[:0], data...)
	return nil
}
func (o outputItem) Type() (string, error) {
	var base responsesOutputItemBase
	if err := json.Unmarshal([]byte(o), &base); err != nil {
		slog.Error("outputItem MarshalJSON json Unmarshal failed", "err", err)
		return "", err
	}

	switch base.Type {
	case outputTypeMessage:
		return outputTypeMessage, nil
	case outputTypeFunctionCall:
		return outputTypeFunctionCall, nil
	case outputTypeReasoning:
		return outputTypeReasoning, nil
	default:
		return "", errors.New("outputItem Type can not find type")
	}
}

func (o outputItem) Marshal[T outputConstraint](t T) (T, error) {
	if err := json.Unmarshal(o, &t); err != nil {
		slog.Error("outputItem Marshal json Unmarshal failed", "err", err)
		return t, err
	}
	return t, nil
}

type responsesResponse struct {
	ID                string               `json:"id"`
	Object            string               `json:"object"` // response
	CreatedAt         int64                `json:"created_at"`
	Status            string               `json:"status"` // completed|in_progress|incomplete
	Error             *responsesError      `json:"error"`
	IncompleteDetails *responsesIncomplete `json:"incomplete_details"`
	Model             string               `json:"model"`
	Output            []outputItem         `json:"output"`
	OutputText        string               `json:"output_text"` // 便捷聚合字段，服务端直接返回
	Usage             responsesUsage       `json:"usage"`
}

type outputConstraint interface {
	responsesOutputMessage |
		*responsesOutputFunctionCall
}

type responsesError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type responsesIncomplete struct {
	Reason string `json:"reason"` // max_output_tokens | content_filter
}

// responsesOutputItem 是 message | function_call | function_call_output | reasoning 的联合类型。
type responsesOutputItem struct {
	Type      string                   `json:"type"` // message|function_call|function_call_output|reasoning
	ID        string                   `json:"id,omitempty"`
	Status    string                   `json:"status,omitempty"` // in_progress|completed
	Role      string                   `json:"role,omitempty"`   // message 时恒为 assistant
	Content   []responsesOutputContent `json:"content,omitempty"`
	CallID    string                   `json:"call_id,omitempty"`
	Name      string                   `json:"name,omitempty"`      // function_call 的函数名
	Arguments string                   `json:"arguments,omitempty"` // function_call 的参数 JSON 串
	Output    string                   `json:"output,omitempty"`    // function_call_output 的输出串
	Summary   []responsesOutputContent `json:"summary,omitempty"`   // reasoning 的摘要
}
type responsesOutputItemBase struct {
	ID   string `json:"id"`
	Type string `json:"type"` // message|function_call|function_call_output|reasoning
}

type responsesOutputMessage struct {
	responsesOutputItemBase
	Role    string                   `json:"role,omitempty"`   // message 时恒为 assistant
	Status  string                   `json:"status,omitempty"` // in_progress|completed
	Content []responsesOutputContent `json:"content,omitempty"`
}

type responsesOutputFunctionCall struct {
	responsesOutputItemBase
	Arguments string `json:"arguments,omitempty"`
	CallId    string `json:"call_id"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Status    string `json:"status,omitempty"` // in_progress|completed
}
type responsesOutputReasoning struct {
	responsesOutputItemBase
	Status  string `json:"status,omitempty"` // in_progress|completed
	Content struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
}

// responsesOutputContent 是 output_text | reasoning_text 的联合类型。
type responsesOutputContent struct {
	Type        string `json:"type"` // output_text | reasoning_text
	Text        string `json:"text"`
	Annotations []any  `json:"annotations,omitempty"`
}

type responsesUsage struct {
	InputTokens         int                         `json:"input_tokens"`
	InputTokensDetails  responsesInputTokensDetail  `json:"input_tokens_details"`
	OutputTokens        int                         `json:"output_tokens"`
	OutputTokensDetails responsesOutputTokensDetail `json:"output_tokens_details"`
	TotalTokens         int                         `json:"total_tokens"`
}

type responsesInputTokensDetail struct {
	CachedTokens int `json:"cached_tokens"`
}

type responsesOutputTokensDetail struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

// ===== OpenAI Responses API：流式事件 =====
// 所有 SSE 事件共用一个外壳，按 Type 区分，字段按事件取用：
//   response.created / response.in_progress / response.completed / response.incomplete -> Response
//   response.output_item.added / response.output_item.done                            -> Item, OutputIndex
//   response.content_part.added / response.content_part.done                          -> Part, ItemID, ContentIndex
//   response.output_text.delta / response.output_text.done                            -> Delta / Text
//   response.function_call_arguments.delta / response.function_call_arguments.done     -> Delta / Arguments, Name
//   response.reasoning_text.delta / response.reasoning_text.done                       -> Delta / Text
//   response.custom_tool_call_input.delta / response.custom_tool_call_input.done       -> Delta / Input

type responsesEvent struct {
	Type           string                  `json:"type"`
	SequenceNumber int64                   `json:"sequence_number"`
	Response       *responsesResponse      `json:"response,omitempty"`
	Item           *responsesOutputItem    `json:"item,omitempty"`
	Part           *responsesOutputContent `json:"part,omitempty"`
	ItemID         string                  `json:"item_id,omitempty"`
	OutputIndex    *int                    `json:"output_index,omitempty"`
	ContentIndex   *int                    `json:"content_index,omitempty"`
	Delta          string                  `json:"delta,omitempty"`
	Text           string                  `json:"text,omitempty"`      // output_text.done / reasoning_text.done
	Arguments      string                  `json:"arguments,omitempty"` // function_call_arguments.done
	Name           string                  `json:"name,omitempty"`      // function_call_arguments.done
	Input          string                  `json:"input,omitempty"`     // custom_tool_call_input.done
}

func newMimo() (mino *responsesRequest, err error) {
	conf := config.GetConfig()
	index := slices.IndexFunc(conf.LLM, func(llm config.LLM) bool {
		return llm.Provider == minoName
	})
	if index < 0 {
		slog.Error("provider is not config", "provide", conf.Provider)
		return nil, errors.New("can not find provede info")
	}
	input := Input[InputText]{
		Role: systemRole,
		Content: InputText{
			Text: conf.SystemPrompt,
			Type: inputTextType,
		},
	}
	r := &responsesRequest{
		Model:   conf.LLM[index].Model,
		apiKey:  conf.LLM[index].ApiKey,
		baseUrl: conf.LLM[index].BaseUrl,
	}
	r.Input.Add(input)
	r.fetchTools()
	return r, nil
}

var sessionTmp *mcp.ClientSession

func (r *responsesRequest) fetchTools() error {
	anysearch := &tool.Anysearch{}
	session, err := anysearch.Connect()
	if err != nil {
		return err
	}
	toolResult, err := session.ListTools(context.Background(), nil)
	if err != nil {
		return err
	}
	for _, tool := range toolResult.Tools {
		input, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return err
		}
		rtool := responsesTool{
			Type:        "function",
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  input,
		}
		r.Tools = append(r.Tools, rtool)
	}

	sessionTmp = session
	return nil
}

func (r *responsesRequest) Chat(message string) (answer string, err error) {
	if len(message) != 0 {
		text := Input[InputText]{
			Role: userRole,
			Content: InputText{
				Text: message,
				Type: inputTextType,
			},
		}
		r.Input.Add(text)
	}
	body, err := json.Marshal(r)
	if err != nil {
		slog.Error("responses input MarshalJSON failed", "err", err)
		return "", err
	}
	reps, err := openAIResponseApi(r.apiKey, r.baseUrl, body)
	if err != nil {
		slog.Error("reques openAIResponseApi failed", "err", err)
		return "", err
	}

	return r.requesMoreForOutput(*reps)
}

func (r *responsesRequest) requesMoreForOutput(reps responsesResponse) (string, error) {
	for _, output := range reps.Output {
		r.Input = append(r.Input, json.RawMessage(output))
	}
	if reps.OutputText != "" {
		return reps.OutputText, nil
	}

	fmt.Println(reps.OutputText)
	for _, item := range reps.Output {
		outputType, err := item.Type()
		if err != nil {
			return "", err
		}
		switch outputType {
		case outputTypeMessage:
			return reps.OutputText, nil
		case outputTypeReasoning:
			continue
		case outputTypeFunctionCall:
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			functionCall, err := item.Marshal(&responsesOutputFunctionCall{})
			if err != nil {
				return "", err
			}
			args := map[string]any{}
			if err := json.Unmarshal([]byte(functionCall.Arguments), &args); err != nil {
				// 解析失败说明 LLM 吐了脏东西，考虑重试或修复
				slog.Error("outputTypeFunctionCall Unmarshal failed", "err", err)
				return "", err
			}
			callResult, err := sessionTmp.CallTool(ctx, &mcp.CallToolParams{
				Name:         functionCall.Name,
				Arguments:    args,
				RequestState: functionCall.Status,
			})
			if err != nil {
				return "", err
			}
			for _, content := range callResult.Content {
				val, err := content.MarshalJSON()
				if err != nil {
					slog.Error("responsesRequest requesMoreForOutput MarshalJSON failed", "err", err)
					return "", nil
				}
				fmt.Println("content:", string(val))
			}
			content, err := callResult.Content[0].MarshalJSON()
			if err != nil {
				return "", err
			}
			item := functionCallOutput{
				CallId: functionCall.CallId,
				Status: callResult.RequestState,
				Output: string(content),
				Tyep:   "function_call_output",
			}
			buff, err := json.Marshal(item)
			if err != nil {
				return "", nil
			}
			fmt.Println("iput:", string(buff))
			r.Input = append(r.Input, buff)
		}
	}

	return r.Chat("")
}
