package config

import (
	"github.com/lccxxo/bailuoli/internal/model"
	"gopkg.in/yaml.v3"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Load 加载配置（配置文件 + 环境变量） 环境变量 > 配置文件
func Load(path string) (*model.Config, error) {
	cfg, err := loadFromFile(path)
	if err != nil {
		return nil, err
	}

	// 环境变量覆盖
	if err := envconfig.Process("gateway", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadFromFile(path string) (*model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg model.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	setDefaults(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// 设置默认值
func setDefaults(cfg *model.Config) {
	if cfg.Server.Addr == "" {
		cfg.Server.Addr = ":8080"
	}
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30 * time.Second
	}
}

func validate(cfg *model.Config) error {
	// todo 验证配置文件字段的合法性

	return nil
}
