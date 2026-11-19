package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"id.azz/pow/internal/tools"
	"id.azz/pow/pkg/logger"
)

type Client struct {
	HTTPClient *http.Client
	APIKey     string
	BaseURL    string
	Tools      []Tool
}

func NewClient(baseURL string, apiKey string) (*Client, error) {
	if baseURL == "" {
		return nil, errors.New("baseURL is required")
	}
	if apiKey == "" {
		return nil, errors.New("API key is required")
	}

	gitTool := Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "read_git_history",
			Description: "Read recent commits from a Git repository.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit": map[string]any{
						"type":        "integer",
						"description": "Number of commits to retrieve.",
					},
				},
				"required": []string{"limit"},
			},
		},
	}

	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Tools:      []Tool{gitTool},
	}, nil
}

func (c *Client) CallAIService(ctx context.Context, userPrompt string, tools ...Message) (*AIResponse, error) {
	// prepare ai model, prepare tools
	messages := []Message{
		{
			Role:    "user",
			Content: userPrompt,
		},
	}
	messages = append(messages, tools...)

	aiReq := AIRequest{
		Model:     "gpt-4o-mini",
		Messages:  messages,
		Tools:     c.Tools,
		MaxTokens: 150,
	}

	jsonData, err := json.Marshal(aiReq)
	if err != nil {
		return nil, err
	}

	// prepare http client request
	req, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("User-Agent", "Go-HTTP-Client/1.0")
	logger.Printf("[%s] %s | Req: %s", "POST", c.BaseURL, jsonData)

	response, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Fatalf("[%s] %s | Error: %s", "POST", c.BaseURL, err)
		return nil, err
	}

	// prepare http client response
	resBody, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Fatalf("[%s] %s | Error reading response body: %s", "POST", c.BaseURL, err)
		return nil, err
	}

	var aiResponse AIResponse
	if err := json.Unmarshal(resBody, &aiResponse); err != nil {
		logger.Fatalf("[%s] %s | Error unmarshaling response: %s", "POST", c.BaseURL, err)
		return nil, err
	}
	logger.Printf("[%s] %s | Res: %s", "POST", c.BaseURL, resBody)

	return &aiResponse, nil
}

func (c *Client) ToolCallService(ctx context.Context, choices ResponseChoices) ([]Message, error) {
	var toolResult []Message = nil
	for _, toolCall := range choices.Message.ToolCalls {
		if toolCall.Function.Name == "read_git_history" {
			var gitArgs struct {
				Limit int `json:"limit"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &gitArgs); err != nil {
				logger.Fatalf("[%s] %s | Error unmarshaling tool arguments: %s", "POST", c.BaseURL, err)
				return nil, err
			}
			result, err := tools.ReadGitHistory(ctx, gitArgs)
			if err != nil {
				return nil, err
			}

			toolResult = append(toolResult, Message{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Content:    string(result),
			})
		}
		// else if - insert here for another tools
	}
	return toolResult, nil
}
