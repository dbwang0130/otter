package agent

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/galilio/otter/internal/agent/tools"
	"github.com/galilio/otter/internal/calendar"
	"github.com/galilio/otter/internal/common/config"
	"github.com/galilio/otter/internal/llm"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

type AgentConfig struct {
	Model           model.LLM
	CalendarService calendar.Service
}

type Option func(*AgentConfig)

func WithModel(m model.LLM) Option {
	return func(cfg *AgentConfig) {
		cfg.Model = m
	}
}

func WithCalendarService(service calendar.Service) Option {
	return func(cfg *AgentConfig) {
		cfg.CalendarService = service
	}
}

func setupAgent(cfg *config.DeepSeekConfig, calendarService calendar.Service) (adkagent.Agent, error) {
	model, err := llm.NewDeepSeekModel(cfg)
	if err != nil {
		slog.Error("Failed to create DeepSeek model", "error", err)
		return nil, err
	}

	ts := []tool.Tool{}
	calendarTools, err := tools.NewCalendarTools(calendarService)
	if err != nil {
		slog.Error("Failed to create calendar tools", "error", err)
		return nil, err
	}
	ts = append(ts, calendarTools...)
	timeTools, err := tools.NewTimeTools()
	if err != nil {
		slog.Error("Failed to create time tools", "error", err)
		return nil, err
	}
	ts = append(ts, timeTools...)

	a, err := llmagent.New(llmagent.Config{
		Name:                "calendar_agent",
		Model:               model,
		Description:         "A calendar agent that can help you manage your calendar and schedule your events.",
		InstructionProvider: InstructionProvider, // 使用 InstructionProvider 替代静态 Instruction
		Tools:               ts,
	})
	if err != nil {
		slog.Error("Failed to create calendar agent", "error", err)
		return nil, err
	}

	return a, nil
}

// isDevelopment 检查当前是否为开发环境
func isDevelopment() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	return env == "development" || env == "dev"
}

func Launch(ctx context.Context, cfg *config.DeepSeekConfig, calendarService calendar.Service) error {
	otter, err := setupAgent(cfg, calendarService)
	if err != nil {
		return err
	}

	config := &launcher.Config{
		AgentLoader: adkagent.NewSingleLoader(otter),
	}

	// 构建启动选项：基础选项包含 web 和 api
	// 设置 write-timeout 为 5 分钟，以支持长时间的 SSE 流式响应
	options := []string{"web", "-port", "8081", "-read-timeout", "60s", "-write-timeout", "5m", "-idle-timeout", "60s", "api"}
	// 开发环境下额外添加 webui 选项
	if isDevelopment() {
		options = append(options, "-webui_address", "localhost:8081", "webui", "-api_server_address", "http://localhost:8081/api")
	}

	l := full.NewLauncher()
	return l.Execute(ctx, config, options)
}
