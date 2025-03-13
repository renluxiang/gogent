package gogent

import (
	openai "github.com/sashabaranov/go-openai"

	"sync"
	"time"
)

type IMemory interface {
	GetWithNewMsg(msg openai.ChatCompletionMessage, sessionId string) []openai.ChatCompletionMessage
	Add(msg openai.ChatCompletionMessage, sessionId string)
	SetSystem(msg string)
}

type simpleMemory struct {
	MaxHistory        int
	Buffer            map[string][]openai.ChatCompletionMessage
	BufferLastUseTime map[string]int64
	sync.Mutex
}

func (memory *simpleMemory) getSessionBuffers(sessionId string) []openai.ChatCompletionMessage {
	memory.BufferLastUseTime[sessionId] = time.Now().Unix()
	buffers, ok := memory.Buffer[sessionId]
	if !ok {
		// Create a new buffer for this session with the system message
		buffers = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: memory.Buffer[""][0].Content, // Assuming the system message is always the first in the "" session
			},
		}
		// Store the new buffer in the memory
		memory.Buffer[sessionId] = buffers
	}
	return buffers
}

func (memory *simpleMemory) reduceBuffer(sessionId string) {
	buffers, ok := memory.Buffer[sessionId]
	if !ok {
		return
	}
	if memory.MaxHistory > 0 && len(buffers) > memory.MaxHistory {
		finalBuffers := make([]openai.ChatCompletionMessage, 0, len(buffers))
		delLen := len(buffers) - memory.MaxHistory
		finalBuffers = append(buffers[:1], buffers[delLen+1:]...)
		memory.Buffer[sessionId] = finalBuffers
	}
}

func (memory *simpleMemory) GetWithNewMsg(msg openai.ChatCompletionMessage, sessionId string) []openai.ChatCompletionMessage {
	memory.Lock()
	defer memory.Unlock()

	buffers := memory.getSessionBuffers(sessionId)
	buffers = append(buffers, msg)
	memory.Buffer[sessionId] = buffers
	memory.reduceBuffer(sessionId)
	return memory.Buffer[sessionId]
}

func (memory *simpleMemory) Add(msg openai.ChatCompletionMessage, sessionId string) {
	memory.Lock()
	defer memory.Unlock()
	memory.Buffer[sessionId] = append(memory.getSessionBuffers(sessionId), msg)
	memory.reduceBuffer(sessionId)
}

// 修改构造函数
func NewSimpleMemory(maxHistory int, system string) IMemory {
	ret := &simpleMemory{
		MaxHistory: maxHistory,
		Buffer: map[string][]openai.ChatCompletionMessage{
			"": {
				openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: system,
				},
			},
		},
		BufferLastUseTime: make(map[string]int64),
	}
	go ret.Start()
	return ret
}

func (memory *simpleMemory) Start() {
	for {
		time.Sleep(time.Minute * 15)
		memory.Lock()
		var expiredSessionIds []string
		for sessionId, lastUseTime := range memory.BufferLastUseTime {
			if time.Now().Unix()-lastUseTime > 60*15 && sessionId != "" {
				expiredSessionIds = append(expiredSessionIds, sessionId)
			}
		}

		for _, sessionId := range expiredSessionIds {
			delete(memory.Buffer, sessionId)
			delete(memory.BufferLastUseTime, sessionId)
		}
		memory.Unlock()
	}
}

func (memory *simpleMemory) SetSystem(msg string) {
	memory.Lock()
	defer memory.Unlock()
	memory.Buffer[""][0] = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: msg,
	}
}

func (memory *simpleMemory) Clear(sessionId string) {
	memory.Lock()
	defer memory.Unlock()

	if sessionId == "" {
		// Clear all sessions except the system prompt in default session
		systemPrompt := memory.Buffer[""][0]
		memory.Buffer = map[string][]openai.ChatCompletionMessage{
			"": {systemPrompt},
		}
	} else {
		// Clear specific session while preserving the system message
		if _, exists := memory.Buffer[sessionId]; exists {
			memory.Buffer[sessionId] = memory.getSessionBuffers(sessionId)[:1]
		}
	}
}
