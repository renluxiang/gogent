package gogent

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

type ITool interface {
	SetAgentAgent(agent IAgent)
	GetName() string
	GetDescription() string
	GetNamespace() string
	Run(args ...any) (any, error)
	Close()
}

type IAgent interface {
	GetName() string
	Chat(msg string, session string) string
}

type ILogger interface {
	Info(args ...interface{})
}

const RetryPromptPrefix = "The previous response format was incorrect. Please try again and follow the format below."

const resultFormat = `<response>
	<think>
		Your thoughts
	</think>
	<action>
		call:packageName.funcName(args1,args2)
	</action>
	<final-answer>
		If you confirm that you have sufficient information and no tools need to be called, please write the final answer here.
	</final-answer>
</response>
`

type BaseAgent struct {
	name             string
	tools            []ITool
	responsibilities []string
	systemPrompt     string
	prompt           string
}

var dLog ILogger = &defaultLogger{}

type defaultLogger struct{}

func (d *defaultLogger) Info(args ...interface{}) {
	log.Println(args...)
}

func (b *BaseAgent) SetName(name string) {
	b.name = name
}

func (b *GenericAgent) WithPrompt(prompt string) *GenericAgent {
	b.prompt = prompt
	return b
}

func (b *BaseAgent) GetPrompt() string {
	return b.prompt
}

func (b *GenericAgent) WithSystemPrompt(prompt string) *GenericAgent {
	b.systemPrompt = prompt
	return b
}

func (b *GenericAgent) GetSystemPrompt() string {
	return b.systemPrompt
}
func (b *GenericAgent) GetName() string {
	return b.name
}

func (b *GenericAgent) GetTools() []ITool {
	return b.tools
}

func (g *GenericAgent) WithTools(tools []ITool) *GenericAgent {
	g.tools = tools
	return g
}

type GenericAgent struct {
	BaseAgent
	b      IBrain
	log    ILogger
	memory IMemory
}

func (g *GenericAgent) GetLogger() ILogger {
	return g.log
}

func NewGenericAgent() *GenericAgent {
	agent := &GenericAgent{}
	return agent
}

func (g *GenericAgent) WithLogger(logger ILogger) *GenericAgent {
	g.log = logger
	return g
}

func (g *GenericAgent) WithBrain(b IBrain) *GenericAgent {
	g.b = b
	return g
}

func (g *GenericAgent) WithName(name string) *GenericAgent {
	g.SetName(name)
	return g
}

func (g *GenericAgent) Start() {
	if g.b == nil {
		g.b = NewOpenAiBrain()
	}
	if g.memory == nil {
		g.memory = NewSimpleMemory(100, g.GetSystemPrompt())
	}
	if g.log == nil {
		g.log = dLog
	}
	g.GetLogger().Info("Agent ", g.GetName(), " starting...")
	toolPrompt := "You have the following tools available:\n"
	for i, tool := range g.tools {
		toolPrompt += fmt.Sprintf("%d: namespace: %s name: %s description: %s \n", i, tool.GetNamespace(), tool.GetName(), tool.GetDescription())
	}
	toolPrompt += `Tools can be used in combination. Use line breaks to separate tool calls and add the prefix "call:", e.g., call:packageName.funcName(args1,args2). string args need "", The tool will return the execution result to you. Note that even if there are many parameters, each tool call must be written on a single line without line breaks.
Your response should follow the format below. If you believe you have completed the task, write the final answer:
`
	toolPrompt += resultFormat

	initialMsg := fmt.Sprintf("Your name is: %s\n %s \n %s ", g.GetName(), g.GetSystemPrompt(), toolPrompt)
	g.GetLogger().Info("Agent ", g.GetName(), " initial message: ", initialMsg)
	g.memory.SetSystem(initialMsg)
	for _, tool := range g.tools {
		tool.SetAgentAgent(g)
	}
}

func (g *GenericAgent) Chat(msg string, session string) string {
	msg = "now:" + time.Now().Format(time.RFC3339) + "\n" + g.GetPrompt() + "\nmsg：" + msg
	ret := thinkStepByStep(g, msg, session)
	return ret
}

func StringMsgToOpenAiMsg(msg string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: msg,
	}
}

func (g *GenericAgent) think(msg, session string) string {
	msgs := g.memory.GetWithNewMsg(StringMsgToOpenAiMsg(msg), session)
	resp, err := g.b.ChatCompletion(msgs)
	if err != nil {
		return err.Error()
	}
	g.memory.Add(resp, session)
	return resp.Content
}

// thinkStepByStep 处理问题，支持多次反思和工具调用
func thinkStepByStep(g *GenericAgent, msg, session string) (ret string) {
	defer func() {
		if err := recover(); err != nil {
			g.GetLogger().Info("Recovered from panic in thinkStepByStep:", err, debug.Stack()[0:1024])
			ret = "Sorry, an error occurred. Please try again."
		}
	}()

	resp := g.think(msg, session)

	const maxTries = 100
	const retryDelay = 1 * time.Second

	for tries := 0; tries < maxTries; tries++ {

		think, action, finalAnswer := g.extractSections(resp)
		if think != "" {
			g.GetLogger().Info("Think: ", think)
		}
		if action != "" {
			g.GetLogger().Info("Action: ", action)
		}
		if finalAnswer != "" {
			g.GetLogger().Info("Final answer: ", finalAnswer)
		}

		// Check if response format is correct
		if think == "" && action == "" && finalAnswer == "" {
			prompt := RetryPromptPrefix
			prompt += resultFormat

			// Add sleep before retry 0,1,2,3,3
			pauseTime := tries
			if pauseTime > 3 {
				pauseTime = 3
			}
			time.Sleep(time.Duration(pauseTime) * retryDelay)

			resp = g.think(prompt, session)
			continue
		}

		if action != "" {
			lines := strings.Split(action, "\n")
			result := "你的工具使用结果：\n"
			results := make([]string, len(lines))
			wg := sync.WaitGroup{}

			for i, line := range lines {
				wg.Add(1)
				go func(i int, line string) {
					defer wg.Done()
					results[i] = processCallLine(g, line)
				}(i, line)
			}
			wg.Wait()
			for _, res := range results {
				if res != "" {
					result += res + "\n"
				}
			}

			resp = g.think(result, session)
			continue
		}

		if think != "" {
			ret = think
		}
		if finalAnswer != "" {
			ret = finalAnswer
		}
		if ret != "" {
			return ret
		}
	}
	return "Unable to process your request."
}

func processCallLine(g *GenericAgent, line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "call:") {
		// Parse tool call
		callExpr, err := parse(line)
		outExpr := strings.TrimPrefix(line, "call:")
		if err != nil {
			return "Tool call format error, parsing failed, " + outExpr + ", Error: " + err.Error()
		}
		toolName := callExpr.PackageName + "." + callExpr.FunctionName
		var callArgs []any
		for _, arg := range callExpr.Arguments {
			callArgs = append(callArgs, arg)
		}
		for _, tool := range g.GetTools() {
			if tool.GetNamespace()+"."+tool.GetName() == toolName {
				res, err := tool.Run(callArgs...)
				if err != nil {
					return outExpr + " Call error: " + err.Error()

				} else {
					if resStr, ok := res.(string); ok {
						return outExpr + " Call success: " + resStr
					} else {
						jsonRes, err := json.Marshal(res)
						if err != nil {
							return fmt.Sprintf("%s Call success, but result cannot be converted to String: %s", outExpr, err.Error())
						} else {
							return outExpr + " Call success: " + string(jsonRes)
						}
					}
				}
			}
		}
		return outExpr + " Not found, please check if the package name is provided or if the tool exists"
	}
	return ""
}

// extractSections splits the response into think, action, and finalAnswer sections
func (g *GenericAgent) extractSections(resp string) (think, action, finalAnswer string) {
	// 处理没有<response>标签的情况
	if !strings.Contains(resp, "<response>") {
		think = extractTag(resp, "think")
		action = extractTag(resp, "action")
		finalAnswer = extractTag(resp, "final-answer")
	} else {
		type Response struct {
			XMLName     xml.Name `xml:"response"`
			Think       string   `xml:"think"`
			Action      string   `xml:"action"`
			FinalAnswer string   `xml:"final-answer"`
		}
		if strings.Contains(resp, "<response>") {
			resp = resp[strings.Index(resp, "<response>"):]
		}
		if strings.Contains(resp, "</response>") {
			resp = resp[:strings.Index(resp, "</response>")+11]
		}

		xmlBytes := []byte(resp)

		// 解析XML数据到结构体
		var response Response
		err := xml.Unmarshal(xmlBytes, &response)
		if err != nil {
			g.GetLogger().Info("解析XML出错:", err, "响应:", resp)
			return "", "", ""
		}

		think = response.Think
		action = response.Action
		finalAnswer = response.FinalAnswer
	}

	// 去除空白字符
	think = strings.TrimSpace(think)
	action = strings.TrimSpace(action)
	finalAnswer = strings.TrimSpace(finalAnswer)
	return think, action, finalAnswer
}

func extractTag(content, tag string) string {
	startTag := "<" + tag + ">"
	endTag := "</" + tag + ">"
	startIndex := strings.Index(content, startTag)
	endIndex := strings.Index(content, endTag)
	if startIndex != -1 && endIndex != -1 && startIndex < endIndex {
		return content[startIndex+len(startTag) : endIndex]
	}
	return ""
}
