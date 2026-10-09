package tool

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/mcp"
)

func Init() []goai.Tool {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)
	transport := mcp.NewStdioTransport("mcp-filesystem-server", []string{"/home/janus/"})
	client := mcp.NewClient("mcp-filesystem-server", "", mcp.WithTransport(transport))
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	//	defer client.Close()
	// Collect all tools with pagination.
	toolsResult, err := client.ListTools(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	allMCPTools := toolsResult.Tools
	for toolsResult.NextCursor != "" {
		toolsResult, err = client.ListTools(ctx, &mcp.ListParams{Cursor: toolsResult.NextCursor})
		if err != nil {
			log.Fatal(err)
		}
		allMCPTools = append(allMCPTools, toolsResult.Tools...)
	}

	fmt.Println(allMCPTools[0])
	// Convert MCP tools to GoAI tools for the LLM.
	return mcp.ConvertTools(client, allMCPTools)
}
