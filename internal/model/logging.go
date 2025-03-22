package model

type LoggingConfig struct {
	Level    string   `yaml:"level"`
	Outputs  []string `yaml:"outputs"`
	Rotation struct {
		MaxSize    int  `yaml:"max_size"`
		MaxAge     int  `yaml:"max_age"`
		MaxBackups int  `yaml:"max_backups"`
		Compress   bool `yaml:"compress"`
	} `yaml:"rotation"`
}
