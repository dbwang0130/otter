package llm

import (
	"github.com/galilio/otter/internal/common/config"
	"google.golang.org/adk/model"
)

// NewDeepSeekModel 创建 DeepSeek 模型实例
// modelName: 模型名称，如果为空则使用配置中的 model，默认 "deepseek-chat"
func NewDeepSeekModel(cfg *config.DeepSeekConfig) (model.LLM, error) {
	modelName := cfg.Model
	if modelName == "" {
		modelName = "deepseek-chat"
	}
	opts := []Option{
		WithAPIKey(cfg.APIKey),
		WithBaseURL(cfg.BaseURL),
		WithModelName(modelName),
	}
	return NewOpenAICompatModel(opts...)
}
