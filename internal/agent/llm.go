package agent

import (
	"bytes"
	"encoding/json"
	"honi/config"
	"io"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v3"
)

const (
	reponsesAPISuffer = "/responses"
)

type LLM interface {
	Chat(message string) (answer string, err error)
}

type Session struct {
	LLM
}

func NewSession() LLM {
	switch config.GetConfig().Provider {
	case minoName:
		llm, err := newMimo()
		if err != nil {
			slog.Error("new mimo failed", "err", err)
		}
		return llm
	}
	return nil
}

func openAIResponseApi(key, url string, body []byte) (resp *responsesResponse, err error) {
	url += reponsesAPISuffer
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		slog.Error("ceate http request failed", "err", err)
		return nil, err
	}
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("api-key", key)
	//fmt.Println("body:", string(body))
	httpResp, err := http.DefaultClient.Do(request)
	if err != nil {
		slog.Error("http reques post method fialed", "err", err)
		return nil, err
	}

	if httpResp.StatusCode != http.StatusOK {
		slog.Error("responses api reques return not ok", "status", httpResp.Status)
		return nil, err
	}
	buff, err := io.ReadAll(httpResp.Body)
	if err != nil {
		slog.Error("read responses api body fialed", "err", err)
		return nil, err
	}

	if err := json.Unmarshal(buff, &resp); err != nil {
		slog.Error("openAIResponseApi json Unmarshal fialed", "err", err)
		return nil, err
	}

	return
}
