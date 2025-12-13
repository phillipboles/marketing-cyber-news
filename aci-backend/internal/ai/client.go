package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the Anthropic Claude SDK client
type Client struct {
	client anthropic.Client
	model  anthropic.Model
}

// Config holds configuration for the AI client
type Config struct {
	APIKey string
	Model  string
}

// NewClient creates a new AI client instance
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	modelName := cfg.Model
	if modelName == "" {
		modelName = "claude-3-haiku-20240307" // Default to Haiku for cost efficiency
	}

	client := anthropic.NewClient(
		option.WithAPIKey(cfg.APIKey),
	)

	return &Client{
		client: client,
		model:  anthropic.Model(modelName),
	}, nil
}

// Complete sends a message to Claude and returns the response
func (c *Client) Complete(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	if systemPrompt == "" {
		return "", fmt.Errorf("system prompt is required")
	}

	if userMessage == "" {
		return "", fmt.Errorf("user message is required")
	}

	// Build system parameter
	system := []anthropic.TextBlockParam{
		{Text: systemPrompt},
	}

	// Build messages
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userMessage)),
	}

	// Call the API
	response, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: int64(4096),
		System:    system,
		Messages:  messages,
	})

	if err != nil {
		return "", fmt.Errorf("claude api call failed: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("empty response from claude")
	}

	// Extract text from the first content block
	contentBlock := response.Content[0]
	if contentBlock.Type == "text" {
		textBlock := contentBlock.AsText()
		return textBlock.Text, nil
	}

	return "", fmt.Errorf("unexpected content type in response: %s", contentBlock.Type)
}

// CompleteWithJSON sends a message and parses JSON response
func (c *Client) CompleteWithJSON(ctx context.Context, systemPrompt, userMessage string, result interface{}) error {
	if result == nil {
		return fmt.Errorf("result pointer is required")
	}

	response, err := c.Complete(ctx, systemPrompt, userMessage)
	if err != nil {
		return fmt.Errorf("completion failed: %w", err)
	}

	if err := json.Unmarshal([]byte(response), result); err != nil {
		return fmt.Errorf("failed to parse json response: %w", err)
	}

	return nil
}
