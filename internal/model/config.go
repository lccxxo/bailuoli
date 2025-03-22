package model

type Config struct {
	Server ServerConfig  `yaml:"server"`
	Log    LoggingConfig `yaml:"log"`
	Routes []*Route      `yaml:"routes"`
}
