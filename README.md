# GoGent - Go Agent Framework

# GoGent - A Lightweight Agent Framework in Go

GoGent is a lightweight, flexible agent framework written in Go that simplifies the creation and management of AI agents. It provides easy integration with LLM APIs (like OpenAI) and supports tool augmentation, making it perfect for building conversational agents with advanced capabilities.

## Features

- 🧠 Simple integration with OpenAI and other LLM APIs
- 🔧 Tool-using capabilities with a simple API
- 💾 Built-in memory management for conversation history
- 🔄 Automatic session management and cleanup
- 🛠️ Extensible architecture to add custom features

## Installation

```bash
go get github.com/renluxiang/gogent
```

## Quick Start

### Basic Agent

```go
package main

import (
    "github.com/renluxiang/gogent"
    "fmt"
)

func main() {
    // Create a new agent
    agent := gogent.NewGenericAgent().
        WithName("MyAssistant").
        WithSystemPrompt("You are a helpful assistant.")
    
    // Start the agent
    agent.Start()
    
    // Chat with the agent
    response := agent.Chat("Hello, who are you?", "user1")
    fmt.Println(response)
}
```

### Agent with Tools

```go
package main

import (
    "github.com/renluxiang/gogent"
    "fmt"
)

func main() {
    // Create a calculator tool
    calcTool := MyCalculatorTool{}
    
    // Create and configure an agent with the tool
    agent := gogent.NewGenericAgent().
        WithName("MathBot").
        WithSystemPrompt("You are a mathematics assistant.").
        WithTools([]gogent.ITool{calcTool})
    
    // Start the agent
    agent.Start()
    
    // Chat with the agent
    response := agent.Chat("Can you calculate 15 * 7?", "user1")
    fmt.Println(response)
}

// Tool implementation example
type MyCalculatorTool struct{}

func (t MyCalculatorTool) SetAgentAgent(agent gogent.IAgent) {}
func (t MyCalculatorTool) GetName() string { return "Calculate" }
func (t MyCalculatorTool) GetDescription() string { return "Performs basic math operations" }
func (t MyCalculatorTool) GetNamespace() string { return "math" }
func (t MyCalculatorTool) Close() {}

func (t MyCalculatorTool) Run(args ...any) (any, error) {
    // Tool implementation here
    // ...
    return "105", nil
}
```

## Environment Variables

GoGent uses the following environment variables:

- `OPENAI_API_KEY` - Your OpenAI API key (required)
- `OPENAI_API_BASE_URL` - Custom API endpoint URL (optional)
- `OPENAI_MODEL` - Model to use, defaults to "gpt-4o" if not specified (optional)

## Contributing

Contributions are welcome! Please feel free to submit pull requests.

## License

MIT

