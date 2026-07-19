package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"trongcon-api/internal/config"

	goopenai "github.com/sashabaranov/go-openai"
)

type Client struct {
	sdk   *goopenai.Client
	model string
	timeout time.Duration
}

func NewClient(cfg config.OpenAIConfig) *Client {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil
	}
	apiCfg := goopenai.DefaultConfig(cfg.APIKey)
	if cfg.BaseURL != "" {
		apiCfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	}
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Client{
		sdk:     goopenai.NewClientWithConfig(apiCfg),
		model:   model,
		timeout: timeout,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.sdk != nil
}

func (c *Client) Model() string {
	return c.model
}

// ChatJSON asks the model for a JSON object response (no tools).
func (c *Client) ChatJSON(ctx context.Context, system, user string) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.sdk.CreateChatCompletion(callCtx, goopenai.ChatCompletionRequest{
		Model: c.model,
		Messages: []goopenai.ChatCompletionMessage{
			{Role: goopenai.ChatMessageRoleSystem, Content: system},
			{Role: goopenai.ChatMessageRoleUser, Content: user},
		},
		ResponseFormat: &goopenai.ChatCompletionResponseFormat{
			Type: goopenai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Temperature: 0.4,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty openai response")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

// ChatWithTools runs a tool-calling loop (maxRounds). onTool handles each call.
func (c *Client) ChatWithTools(
	ctx context.Context,
	messages []goopenai.ChatCompletionMessage,
	tools []goopenai.Tool,
	onTool func(name, args string) (string, error),
	maxRounds int,
) (string, error) {
	if maxRounds <= 0 {
		maxRounds = 3
	}
	msgs := append([]goopenai.ChatCompletionMessage{}, messages...)

	for round := 0; round < maxRounds; round++ {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		resp, err := c.sdk.CreateChatCompletion(callCtx, goopenai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    msgs,
			Tools:       tools,
			Temperature: 0.5,
		})
		cancel()
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("empty openai response")
		}
		msg := resp.Choices[0].Message
		if len(msg.ToolCalls) == 0 {
			return strings.TrimSpace(msg.Content), nil
		}
		msgs = append(msgs, msg)
		for _, tc := range msg.ToolCalls {
			result, toolErr := onTool(tc.Function.Name, tc.Function.Arguments)
			if toolErr != nil {
				result = fmt.Sprintf(`{"error":%q}`, toolErr.Error())
			}
			msgs = append(msgs, goopenai.ChatCompletionMessage{
				Role:       goopenai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
			})
		}
	}

	// Final reply without tools
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.sdk.CreateChatCompletion(callCtx, goopenai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    msgs,
		Temperature: 0.5,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty openai response")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func DecodeJSON[T any](raw string, out *T) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty json")
	}
	// Strip markdown fences if any
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}
	return json.Unmarshal([]byte(raw), out)
}
