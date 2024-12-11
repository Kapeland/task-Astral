package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// configPathEnv is set in Dockerfile
const configPathEnv = "CONFIG_PATH"

var cfg *Config

func GetConfig() Config {
	if cfg != nil {
		return *cfg
	}

	return Config{}
}

// Rest - contains parameter rest JSON connection.
type Rest struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Database - contains all parameters database connection.
type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
}

// Admin - contains admin's token
type Admin struct {
	Token string `yaml:"token"`
}

// Project - contains all project information.
type Project struct {
	Debug bool `yaml:"debug"`
}

// Logger - contains all parameters for logger configuration.
type Logger struct {
	Lvl     string  `yaml:"level"`
	LogRate float64 `yaml:"rate"`
}

type Config struct {
	Project  Project  `yaml:"project"`
	Rest     Rest     `yaml:"rest"`
	Database Database `yaml:"database"`
	Admin    Admin    `yaml:"admin"`
	Logger   Logger   `yaml:"logger"`
}

func ReadConfigYAML() error {
	if cfg != nil {
		return nil
	}
	filePath, exist := os.LookupEnv(configPathEnv)
	if !exist {
		return fmt.Errorf("env var %s does not exist", configPathEnv)
	}
	filePath = filepath.Clean(filePath)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)

	if err := decoder.Decode(&cfg); err != nil {
		return err
	}
	return nil
}
