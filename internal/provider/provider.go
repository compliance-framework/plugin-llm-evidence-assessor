package provider

import (
	"context"
)

type Params struct {
	Provider    string
	Model       string
	MaxTokens   int
	Temperature float64
}

type Payload struct {
	Prompt string
}

type Output struct {
	JSON string
	UsageTokens int
}

type Client interface {
	Generate(ctx context.Context, params Params, in Payload) (Output, error)
}
