package agent

import (
	"bytes"
	"fmt"
	"honi/config"
	"honi/internal/tool"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/provider"
	"github.com/zendev-sh/goai/provider/openai"
)

var GetAgent = sync.OnceValue(func() *Agent {
	llm := config.GetConfig().LLM
	if err := os.Setenv("OPENAI_API_KEY", llm[0].ApiKey); err != nil {
		panic(err)
	}
	if err := os.Setenv("OPENAI_BASE_URL", llm[0].BaseUrl); err != nil {
		panic(err)
	}
	return NewAgent()
})

type Agent struct {
	llm  provider.LanguageModel
	opts []goai.Option
}

func (a *Agent) Ask(message string) (string, error) {
	mimo, err := newMimo()
	if err != nil {
		slog.Error("Agent Ask newMimo failed", "err", err)
		return "", err
	}
	answer, err := mimo.Chat(message)
	if err != nil {
		slog.Error("Agent Ask Chat failed", "err", err)
		return "", err
	}
	return answer, nil
}

type dumpTransport struct{ base http.RoundTripper }

func (d dumpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(body))
	slog.Debug("request LLM", "body:", string(body))
	resp, err := d.base.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	body, _ = io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(body))
	slog.Debug("resp", "status", resp.Status, "body", string(body))

	return resp, err
}
func NewAgent() *Agent {
	conf := config.GetConfig()

	return &Agent{
		llm: openai.Chat(config.GetConfig().LLM[0].Model, openai.WithHTTPClient(&http.Client{
			Transport: &dumpTransport{base: http.DefaultTransport},
		})),
		opts: []goai.Option{
			goai.WithSystem(conf.SystemPrompt),
			//goai.WithTools(tool.ControlPc()),
			//	goai.WithTools(tool.GetMCPResource().ActionTools()...),
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

func (a *Agent) optionByMessage(prompt string) []goai.Option {
	return []goai.Option{
		goai.WithSystem(config.GetConfig().SystemPrompt),
		//goai.WithTools(tool.ControlPc()),
		goai.WithTools(tool.GetMCPResource().ActionTools()...),
		goai.WithMaxSteps(10),
		goai.WithPrompt(prompt),
		goai.WithOnResponse(func(ri goai.ResponseInfo) {
			fmt.Printf("response:%+v\n\n", ri)
		}),
		goai.WithOnBeforeStep(func(bsi goai.BeforeStepInfo) goai.BeforeStepResult {
			fmt.Printf("step before:%+v\n\n", bsi)
			return goai.BeforeStepResult{}
		}),
	}
}
