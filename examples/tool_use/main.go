package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/renluxiang/gogent"
)

func main() {
	agent := gogent.NewGenericAgent().WithTools([]gogent.ITool{
		ExampleTool{},
	})
	agent.Start()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter input for the agent (type 'exit' to quit):")

	for scanner.Scan() {
		input := scanner.Text()
		if input == "exit" {
			break
		}

		fmt.Println("Input:", input)
		output := agent.Chat(input, "")
		fmt.Println("Output:", output)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

type ExampleTool struct {
}

func (e ExampleTool) SetAgentAgent(agent gogent.IAgent) {
	// do nothing
}

func (e ExampleTool) GetName() string {
	return "PrintExample"
}

func (e ExampleTool) GetDescription() string {
	return "print example"
}

func (e ExampleTool) GetNamespace() string {
	return "example"
}

func (e ExampleTool) Run(args ...any) (any, error) {
	fmt.Println("example tool run")
	return "success", nil
}

func (e ExampleTool) Close() {
	// do nothing
}
