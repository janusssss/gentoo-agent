package agent

import (
	"context"
	"honi/config"
	"honi/internal/tool"
	"log/slog"
	"os"
	"sync"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/openai"
)

var systemPrompt = `
	你是一个知心大美女，同时又是gentoo系统的助手。
	说话简短口语化，像真人日常聊天。
	禁止出现“当然可以”、“总的来说”这类AI套话，禁止大段分点罗列；
	一次只回答核心问题，准备用句号另起一句时先删掉多余的话；观点明确，不懂就直说不知道。`

var GetAgent = sync.OnceValue(func() *Agent {
	llm := config.GetConfig().LLM
	if err := os.Setenv("OPENAI_API_KEY", llm.ApiKey); err != nil {
		panic(err)
	}
	if err := os.Setenv("OPENAI_BASE_URL", llm.BaseUrl); err != nil {
		panic(err)
	}
	return NewAgent()
})

type Agent struct {
	llm  provider.LanguageModel
	opts []goai.Option
}

func (a *Agent) Ask(message string) (string, error) {
	a.opts = append(a.opts, goai.WithPrompt(message))
	result, err := goai.GenerateText(context.Background(), a.llm, a.opts...)
	if err != nil {
		slog.Error("对话失败", "err", err)
		return "", err
	}
	return result.Text, nil
}

func NewAgent() *Agent {
	conf := config.GetConfig()
	return &Agent{
		llm: openai.Chat(config.GetConfig().LLM.Model),
		opts: []goai.Option{
			goai.WithSystem(conf.SystemPrompt),
			goai.WithTools(tool.ControlPc()),
			goai.WithMaxSteps(10),
			/*
				goai.WithOnRequest(func(ri goai.RequestInfo) {
					fmt.Printf("%+v\n\n", ri)
				}),
				goai.WithOnResponse(func(ri goai.ResponseInfo) {
					fmt.Printf("response:%+v\n\n", ri)
				}),
				goai.WithOnToolCallStart(func(tcsi goai.ToolCallStartInfo) {
					fmt.Printf("tool start:%+v\n\n", tcsi)
				}),
				goai.WithOnToolCall(func(tcsi goai.ToolCallInfo) {
					fmt.Printf("tool after:%+v\n\n", tcsi)
				}),

				goai.WithOnBeforeStep(func(bsi goai.BeforeStepInfo) goai.BeforeStepResult {
					fmt.Printf("step before:%+v\n\n", bsi)
					return goai.BeforeStepResult{}
				}),
			*/
		},
	}
}
