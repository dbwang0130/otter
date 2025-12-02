package agent

import (
	_ "embed"
	"fmt"
	"log"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

type AgentConfig struct {
	Model model.LLM
}

type Option func(*AgentConfig)

func WithModel(m model.LLM) Option {
	return func(cfg *AgentConfig) {
		cfg.Model = m
	}
}

//go:embed prompts/calendar_agent.md
var calendarInstruction string

func SetupAgent(opts ...Option) (*agent.Agent, error) {
	cfg := &AgentConfig{}

	// 应用所有选项
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Model == nil {
		return nil, fmt.Errorf("model 是必需的，请使用 WithModel() 设置")
	}

	instruction := calendarInstruction
	if instruction == "" {
		instruction = "You are a calendar agent that can help you manage your calendar and schedule your events."
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "calendar_agent",
		Model:       cfg.Model,
		Description: "A calendar agent that can help you manage your calendar and schedule your events.",
		Instruction: instruction,
	})

	if err != nil {
		log.Fatalf("Failed to create calendar agent: %v", err)
	}

	return &a, nil
}
