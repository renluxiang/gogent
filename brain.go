package gogent

import (
	"context"
	openai "github.com/sashabaranov/go-openai"
	"os"
)

type IBrain interface {
	ChatCompletion(msgs []openai.ChatCompletionMessage) (msg openai.ChatCompletionMessage, err error)
}

type OpenAiBrain struct {
	c     *openai.Client
	model string
}

func NewOpenAiBrain() IBrain {
	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		panic("OPENAI_API_KEY is not set")
	}
	cfg := openai.DefaultConfig(token)
	baseUrl := os.Getenv("OPENAI_API_BASE_URL")
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = openai.GPT4o
	}
	if baseUrl != "" {
		cfg.BaseURL = baseUrl
	}
	return &OpenAiBrain{
		c:     openai.NewClientWithConfig(cfg),
		model: model,
	}
}

func (oab *OpenAiBrain) ChatCompletion(msgs []openai.ChatCompletionMessage) (msg openai.ChatCompletionMessage, err error) {
	resp, err := oab.c.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    oab.model,
		Messages: msgs,
	})
	if err != nil {
		return openai.ChatCompletionMessage{}, err
	}
	return resp.Choices[0].Message, nil
}
