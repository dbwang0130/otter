package agent

import (
	"bytes"
	_ "embed"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	"github.com/galilio/otter/internal/user"
	adkagent "google.golang.org/adk/agent"
	"gopkg.in/yaml.v3"
)

// userService 用户服务，用于获取用户配置
var userService user.Service

// SetUserService 设置用户服务
func SetUserService(service user.Service) {
	userService = service
}

//go:embed prompts/calendar_agent.md
var calendarInstruction string

//go:embed prompts/characters.yaml
var charactersYAML string

// Character 角色信息
type Character struct {
	Name          string    `yaml:"name"`
	Code          string    `yaml:"code"` // 角色代号
	Personality   string    `yaml:"personality"`
	SpeakingStyle string    `yaml:"speaking_style"`
	Examples      []Example `yaml:"examples"`
}

// Example 回复例子
type Example struct {
	Context string `yaml:"context"`
	Reply   string `yaml:"reply"`
}

// CharactersConfig 角色配置
type CharactersConfig struct {
	Characters []Character `yaml:"characters"`
}

// instructionData 模板数据
type instructionData struct {
	Character     string
	Personality   string
	SpeakingStyle string
	Examples      []Example
	Now           string
}

// loadCharactersConfig 加载所有角色配置
func loadCharactersConfig() (*CharactersConfig, error) {
	var config CharactersConfig
	if err := yaml.Unmarshal([]byte(charactersYAML), &config); err != nil {
		// 如果嵌入的文件为空，尝试从文件系统读取
		if charactersYAML == "" {
			charPath := filepath.Join("internal", "agent", "prompts", "characters.yaml")
			data, err := os.ReadFile(charPath)
			if err != nil {
				slog.Warn("Failed to read characters.yaml, using default", "error", err)
				return nil, err
			}
			if err := yaml.Unmarshal(data, &config); err != nil {
				slog.Warn("Failed to parse characters.yaml, using default", "error", err)
				return nil, err
			}
		} else {
			slog.Warn("Failed to parse embedded characters.yaml, using default", "error", err)
			return nil, err
		}
	}

	if len(config.Characters) == 0 {
		slog.Warn("No characters found in characters.yaml, using default")
		return nil, nil
	}

	return &config, nil
}

// loadCharacter 根据角色代号加载角色配置，如果 code 为空或未找到，返回第一个角色
func loadCharacter(code string) (*Character, error) {
	config, err := loadCharactersConfig()
	if err != nil || config == nil {
		return nil, err
	}

	// 如果提供了 code，尝试查找对应的角色
	if code != "" {
		for i := range config.Characters {
			if config.Characters[i].Code == code {
				return &config.Characters[i], nil
			}
		}
		slog.Warn("Character not found by code, using default", "code", code)
	}

	// 如果未找到或 code 为空，返回第一个角色
	if len(config.Characters) > 0 {
		return &config.Characters[0], nil
	}

	return nil, nil
}

// InstructionProvider 动态生成 agent 指令，使用 Go template 替换占位符
func InstructionProvider(ctx adkagent.ReadonlyContext) (string, error) {
	// 基础指令模板
	instruction := calendarInstruction
	if instruction == "" {
		instruction = "You are a calendar agent that can help you manage your calendar and schedule your events."
	}

	// 解析模板
	tmpl, err := template.New("instruction").Parse(instruction)
	if err != nil {
		slog.Error("Failed to parse instruction template", "error", err)
		return instruction, err
	}

	// 获取用户偏好角色代号
	var characterCode string
	if userService != nil {
		userID, err := strconv.ParseUint(ctx.UserID(), 10, 64)
		if err != nil {
			userID = 2
		}
		profile, err := userService.GetUserProfile(uint(userID))
		if err != nil {
			characterCode = "default" // 默认角色
		}
		characterCode = profile.PreferredCharacterCode
	}

	// 根据角色代号加载角色数据
	char, err := loadCharacter(characterCode)
	if err != nil {
		slog.Warn("Failed to load character, using default", "error", err, "code", characterCode)
	}

	// 准备模板数据
	data := instructionData{
		Now: time.Now().UTC().Format(time.RFC3339),
	}

	// 如果成功加载角色，填充角色数据
	if char != nil {
		data.Character = char.Name
		data.Personality = char.Personality
		data.SpeakingStyle = char.SpeakingStyle
		data.Examples = char.Examples
	} else {
		// 使用默认值
		data.Character = "calendar assistant"
		data.Personality = "Helpful, professional, and efficient"
		data.SpeakingStyle = "Clear, concise, and friendly"
		data.Examples = []Example{}
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		slog.Error("Failed to execute instruction template", "error", err)
		return instruction, err
	}

	result := buf.String()

	slog.Debug("Generated dynamic instruction",
		"user_id", ctx.UserID(),
		"session_id", ctx.SessionID(),
		"current_time", data.Now,
		"character", data.Character)

	return result, nil
}
