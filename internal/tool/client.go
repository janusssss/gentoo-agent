package tool

import (
	"context"
	"time"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/mcp"
)

type client struct {
	*mcp.Client
	name        string
	Description string
}

func (c client) tools() ([]goai.Tool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	toolsResult, err := c.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}
	allMCPTools := toolsResult.Tools
	for toolsResult.NextCursor != "" {
		toolsResult, err = c.ListTools(ctx, &mcp.ListParams{Cursor: toolsResult.NextCursor})
		if err != nil {
			return nil, err
		}
		allMCPTools = append(allMCPTools, toolsResult.Tools...)
	}

	return mcp.ConvertTools(c.Client, allMCPTools), nil
}

func (c *client) getName() string {
	return c.name
}

func (c *client) getDescription() string {
	return c.Description
}
