package tool

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/zendev-sh/goai"
	"github.com/zendev-sh/goai/mcp"
)

var GetTool = sync.OnceValue(func() *Tool {
	tool, err := NewTool()
	if err != nil {
		panic(err)
	}
	return tool
})

type Tool struct {
	all    []goai.Tool
	Active []goai.Tool
}

func NewTool() (*Tool, error) {
	tools, err := newFileSystem()
	if err != nil {
		return nil, err
	}

	tool := &Tool{}
	tool.all = append(tool.all, tools...)

	return tool, nil
}

func Loadtool() (tools []goai.Tool, err error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*60)

	transport := mcp.NewHTTPTransport(
		"https://api.anysearch.com/mcp",
		mcp.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + os.Getenv("ANYSEARCH_API_KEY"),
		}),
	)
	client := mcp.NewClient("anysearch", "", mcp.WithTransport(transport))
	if err := client.Connect(ctx); err != nil {
		return tools, err
	}
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
	return mcp.ConvertTools(client, allMCPTools), nil
}
