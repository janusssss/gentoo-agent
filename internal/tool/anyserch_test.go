package tool

import (
	"fmt"
	"testing"
)

func TestAnysearchTool(t *testing.T) {
	var c mcpClient = &anysearch{}
	if err := c.initialize(); err != nil {
		t.Fatal(err)
	}
	tools, err := c.tools()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", tools)
}
