package tool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGetName(t *testing.T) {
	var c mcpClient = &anysearch{}
	err := c.initialize()
	assert.NoError(t, err)
	require.Equal(t, "anysearch", c.getName())
}
