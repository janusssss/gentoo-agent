package tool

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/zendev-sh/goai"
)

var GetMCPResource = sync.OnceValue(func() *MCPResource {
	m := &MCPResource{
		clients: []mcpClient{
			&anysearch{},
			&fileSystem{},
		},
	}
	m.initTool()
	return m
})

type MCPResource struct {
	toolMap map[string][]goai.Tool
	active  []goai.Tool
	clients []mcpClient
	mu      *sync.Mutex
}

type mcpClient interface {
	tools() ([]goai.Tool, error)
	initialize() error
	getName() string
	getDescription() string
}

func (m *MCPResource) fetchTool(clients ...mcpClient) error {
	wg := sync.WaitGroup{}
	wg.Add(len(clients))
	for _, c := range clients {
		go func() {
			defer wg.Done()
			if err := c.initialize(); err != nil {
				slog.Error("connect to map server failed", "err", err)
				return
			}
			fmt.Println("name:", c.getName())

			tools, err := c.tools()
			if err != nil {
				slog.Error("get tool from mcp resources failed", "err", err)
				return
			}

			m.mu.Lock()
			m.toolMap[c.getName()] = tools
			m.mu.Unlock()

		}()
	}

	wg.Wait()
	return nil
}

func (m *MCPResource) toolSetSchema() string {
	var schema strings.Builder
	for _, c := range m.clients {
		fmt.Println(c.getDescription())
		fmt.Fprintln(&schema, c.getName(), ":", c.getDescription())
	}
	return schema.String()
}

func (m *MCPResource) initTool() ([]goai.Tool, error) {
	m.fetchTool()
	tool := goai.NewTool(
		"加载工具资源",
		"加载资源，开放如下工具资源以供使用:\n"+m.toolSetSchema(),
		func(ctx context.Context, in struct {
			Name string `json:"name" jsonschema:"description=需用开启工具资源name"`
		}) (string, error) {
			tools := m.toolMap[in.Name]
			if len(tools) == 0 {
				return "", fmt.Errorf("找不到%v资源", in.Name)
			}
			m.active = append(m.active, tools...)
			return in.Name + "资源已启用", nil
		})

	m.active = append(m.active, tool)
	return m.active, nil
}

func (m *MCPResource) ActionTools() []goai.Tool {
	return m.active
}
