package main

import (
	"fmt"
	"honi/internal/agent"
	"os"
)

func main() {
	answer := agent.NewAgent().Ask(os.Args[1])
	fmt.Printf("%+v", answer)
}
