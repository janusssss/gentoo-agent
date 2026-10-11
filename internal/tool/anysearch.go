package tool

import (
	"context"
	"time"

	"github.com/zendev-sh/goai/mcp"
)

type anysearch struct {
	client
}

func (a *anysearch) initialize() error {
	a.name = "anysearch"
	a.description = "Unified real-time search engine skill for AI agents." +
		" Supports general web search, vertical domain search, parallel batch search, and full-page content extraction."
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	transport := mcp.NewHTTPTransport("https://api.anysearch.com/mcp")
	a.Client = mcp.NewClient(a.name, "", mcp.WithTransport(transport))
	if err := a.Client.Connect(ctx); err != nil {
		return err
	}

	return nil

}
