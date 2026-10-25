package agent

import (
	"fmt"
	"testing"
)

func TestAsk(t *testing.T) {
	agent := GetAgent()
	answer, err := agent.Ask("你是谁")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(answer)
}
