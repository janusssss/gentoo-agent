package tool

import (
	"context"
	"time"

	"github.com/zendev-sh/goai/mcp"
)

type fileSystem struct {
	client
}

func (f *fileSystem) initialize() error {
	f.name = "file system server"
	f.Description = "provides secure access to the local filesystem via the Model Context Protocol"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	transport := mcp.NewStdioTransport("mcp-filesystem-server", []string{"/home/janus/"})
	f.Client = mcp.NewClient("mcp-filesystem-server", "", mcp.WithTransport(transport))
	if err := f.client.Connect(ctx); err != nil {
		return err
	}

	return nil
}
