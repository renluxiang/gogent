package main

import (
	"fmt"
	"github.com/renluxiang/gogent"
)

func main() {
	agent := gogent.NewGenericAgent().WithTools([]gogent.ITool{
		ExampleTool{},
	})
	agent.Start()
	agent.Chat("hello, can you print example?", "")
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
