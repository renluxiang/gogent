
# GoGent - 轻量级Go语言智能代理框架

GoGent是一个用Go语言编写的轻量级、灵活的智能代理框架，它简化了AI代理的创建和管理过程。它提供了与LLM API（如OpenAI）的简单集成，并支持工具增强，非常适合构建具有高级功能的对话代理。

## 特性

- 🧠 简单集成OpenAI等LLM API
- 🔧 工具调用能力，API简洁易用
- 💾 内置会话历史记忆管理
- 🔄 自动会话管理和清理
- 🛠️ 可扩展架构，方便添加自定义功能

## 安装

```bash
go get github.com/renluxiang/gogent
```

## 快速开始

### 基础代理

```go
package main

import (
    "github.com/renluxiang/gogent"
    "fmt"
)

func main() {
    // 创建一个新代理
    agent := gogent.NewGenericAgent().
        WithName("我的助手").
        WithSystemPrompt("你是一个有帮助的助手。")
    
    // 启动代理
    agent.Start()
    
    // 与代理聊天
    response := agent.Chat("你好，你是谁？", "用户1")
    fmt.Println(response)
}
```

### 带工具的代理

```go
package main

import (
    "github.com/renluxiang/gogent"
    "fmt"
)

func main() {
    // 创建一个计算器工具
    calcTool := MyCalculatorTool{}
    
    // 创建并配置带有工具的代理
    agent := gogent.NewGenericAgent().
        WithName("数学机器人").
        WithSystemPrompt("你是一个数学助手。").
        WithTools([]gogent.ITool{calcTool})
    
    // 启动代理
    agent.Start()
    
    // 与代理聊天
    response := agent.Chat("你能计算15 * 7吗？", "用户1")
    fmt.Println(response)
}

// 工具实现示例
type MyCalculatorTool struct{}

func (t MyCalculatorTool) SetAgentAgent(agent gogent.IAgent) {}
func (t MyCalculatorTool) GetName() string { return "Calculate" }
func (t MyCalculatorTool) GetDescription() string { return "执行基本的数学运算" }
func (t MyCalculatorTool) GetNamespace() string { return "math" }
func (t MyCalculatorTool) Close() {}

func (t MyCalculatorTool) Run(args ...any) (any, error) {
    // 工具实现代码
    // ...
    return "105", nil
}
```

## 环境变量

GoGent使用以下环境变量：

- `OPENAI_API_KEY` - 您的OpenAI API密钥（必需）
- `OPENAI_API_BASE_URL` - 自定义API端点URL（可选）
- `OPENAI_MODEL` - 要使用的模型，如未指定默认为"gpt-4o"（可选）

## 贡献

欢迎贡献！请随时提交拉取请求。

## 许可证

MIT