package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	LLM      LLMConfig      `mapstructure:"llm"`
	Log      LogConfig      `mapstructure:"log"`
}

type LogConfig struct {
	LogLevel   string `mapstructure:"log_level"`
	LogFile    string `mapstructure:"log_file,omitempty"`
	MaxSize    int    `mapstructure:"max_size,omitempty"`    // 单个日志文件最大大小（MB），0表示不限制
	MaxBackups int    `mapstructure:"max_backups,omitempty"` // 保留的旧日志文件数量，0表示保留所有
	MaxAge     int    `mapstructure:"max_age,omitempty"`     // 保留旧日志文件的天数，0表示不删除
	Compress   bool   `mapstructure:"compress,omitempty"`    // 是否压缩轮转后的日志文件
	LocalTime  bool   `mapstructure:"local_time,omitempty"`  // 是否使用本地时间命名轮转文件
}

type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	Expiration        time.Duration `mapstructure:"expiration"`         // Access token过期时间
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"` // Refresh token过期时间
}

type ServerConfig struct {
	Port         int            `mapstructure:"port"`
	AgentPort    int            `mapstructure:"agent_port"`
	ReadTimeout  *time.Duration `mapstructure:"read_timeout,omitempty"`
	WriteTimeout *time.Duration `mapstructure:"write_timeout,omitempty"`
	IdleTimeout  *time.Duration `mapstructure:"idle_timeout,omitempty"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"` // 数据库连接 URL，格式: postgres://user:password@host:port/dbname?sslmode=disable
	MaxOpenConns    int           `mapstructure:"max_open_conns,omitempty"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns,omitempty"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime,omitempty"`
}

type LLMConfig struct {
	DeepSeek DeepSeekConfig `mapstructure:"deepseek"`
}

type DeepSeekConfig struct {
	APIKey  string `mapstructure:"api_key"`
	Model   string `mapstructure:"model"`
	BaseURL string `mapstructure:"base_url"`
}

// Load 加载配置文件
// configPath: 配置文件路径，如果为空则使用默认路径查找
func Load(configPath ...string) (*Config, error) {
	viper.SetConfigType("yaml")

	// 如果指定了配置文件路径，直接使用
	if len(configPath) > 0 && configPath[0] != "" {
		viper.SetConfigFile(configPath[0])
	} else {
		// 否则使用默认路径查找
		viper.SetConfigName("config")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 验证必需配置
	if config.Database.URL == "" {
		return nil, fmt.Errorf("数据库 URL 是必需的，请在配置文件中设置 database.url")
	}
	if config.LLM.DeepSeek.APIKey == "" {
		return nil, fmt.Errorf("DeepSeek API key 是必需的，请在配置文件中设置 llm.deepseek.api_key")
	}
	if config.LLM.DeepSeek.Model == "" {
		return nil, fmt.Errorf("DeepSeek model 是必需的，请在配置文件中设置 llm.deepseek.model")
	}
	if config.LLM.DeepSeek.BaseURL == "" {
		return nil, fmt.Errorf("DeepSeek base_url 是必需的，请在配置文件中设置 llm.deepseek.base_url")
	}

	// 应用默认值（如果配置文件中未设置）
	applyServerDefaults(&config.Server)
	applyDatabaseDefaults(&config.Database)
	applyLogDefaults(&config.Log)

	return &config, nil
}

// applyServerDefaults 应用服务器配置的默认值
func applyServerDefaults(server *ServerConfig) {
	if server.ReadTimeout == nil {
		defaultReadTimeout := 15 * time.Second
		server.ReadTimeout = &defaultReadTimeout
	}
	if server.WriteTimeout == nil {
		defaultWriteTimeout := 15 * time.Second
		server.WriteTimeout = &defaultWriteTimeout
	}
	if server.IdleTimeout == nil {
		defaultIdleTimeout := 60 * time.Second
		server.IdleTimeout = &defaultIdleTimeout
	}
}

// applyDatabaseDefaults 应用数据库配置的默认值
func applyDatabaseDefaults(database *DatabaseConfig) {
	if database.MaxOpenConns == 0 {
		database.MaxOpenConns = 25
	}
	if database.MaxIdleConns == 0 {
		database.MaxIdleConns = 5
	}
	if database.ConnMaxLifetime == 0 {
		database.ConnMaxLifetime = 5 * time.Minute
	}
}

func setDefaults() {
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.agent_port", 8081)
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	// database.url 没有默认值，必须设置
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	viper.SetDefault("jwt.secret", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration", "15m")          // Access token 15分钟
	viper.SetDefault("jwt.refresh_expiration", "168h") // Refresh token 7天 (168小时)

	// llm.deepseek 的所有字段都没有默认值，必须设置

	// log 配置默认值
	viper.SetDefault("log.log_level", "info")
	viper.SetDefault("log.max_size", 100)     // 100MB
	viper.SetDefault("log.max_backups", 5)    // 保留5个文件
	viper.SetDefault("log.max_age", 30)       // 保留30天
	viper.SetDefault("log.compress", false)   // 不压缩
	viper.SetDefault("log.local_time", false) // 使用UTC时间
}

// applyLogDefaults 应用日志配置的默认值
func applyLogDefaults(log *LogConfig) {
	if log.LogLevel == "" {
		log.LogLevel = "info"
	}
	if log.MaxSize == 0 {
		log.MaxSize = 100
	}
	if log.MaxBackups == 0 {
		log.MaxBackups = 5
	}
	if log.MaxAge == 0 {
		log.MaxAge = 30
	}
}
