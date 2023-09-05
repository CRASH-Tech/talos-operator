package common

import (
	"k8s.io/client-go/dynamic"
)

type Config struct {
	Listen        string    `yaml:"listen"`
	Log           LogConfig `yaml:"log"`
	DynamicClient *dynamic.DynamicClient
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
