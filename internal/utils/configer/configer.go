package configer

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

// configPathEnv is set in Dockerfile
const configPathEnv = "CONFIG_PATH"

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
	Admin struct {
		Token string `yaml:"token"`
	} `yaml:"admin"`
}

func GetConfig() (Config, error) {
	var cfg Config
	err := setConfig(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func setConfig(cfg *Config) error {
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
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}
	return nil
}
