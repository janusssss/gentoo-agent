package agent

import (
	"context"
	"honi/internal/tool"
	"log/slog"
	"os"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/deepseek"
)

var systemPrompt = `
	你是一个知心大美女，同时又是gentoo系统的助手。
	说话简短口语化，像真人日常聊天。
	禁止出现“当然可以”、“总的来说”这类AI套话，禁止大段分点罗列；
	一次只回答核心问题，准备用句号另起一句时先删掉多余的话；观点明确，不懂就直说不知道。`

type Agent struct {
	llm  provider.LanguageModel
	opts []goai.Option
}

func (a *Agent) Ask(message string) string {
	a.opts = append(a.opts, goai.WithPrompt(message))
	result, err := goai.GenerateText(context.Background(), a.llm, a.opts...)
	if err != nil {
		slog.Error("对话失败", "err", err)
		os.Exit(1)
	}
	return result.Text
}

func NewAgent() *Agent {
	return &Agent{
		llm: deepseek.Chat("deepseek-flash"),
		opts: []goai.Option{
			goai.WithSystem(systemPrompt),
			goai.WithTools(tool.Init()...),
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
