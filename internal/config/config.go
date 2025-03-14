package config

import (
	"github.com/lccxxo/bailuoli/internal/model"
	"time"
)

type Config struct {
	Server ServerConfig   `yaml:"server"`
	Log    LogConfig      `yaml:"log"`
	Routes []*model.Route `yaml:"routes"`
}

type ServerConfig struct {
	Addr            string        `yaml:"addr"`
	Mode            string        `yaml:"mode"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type LogConfig struct {
	Level    string   `yaml:"level"`
	Outputs  []string `yaml:"outputs"`
	Rotation struct {
		MaxSize    int  `yaml:"max_size"`
		MaxAge     int  `yaml:"max_age"`
		MaxBackups int  `yaml:"max_backups"`
		Compress   bool `yaml:"compress"`
	} `yaml:"rotation"`
}
