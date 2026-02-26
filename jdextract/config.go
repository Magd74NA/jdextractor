package jdextract

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	DeepSeekApiKey string `json:"deepseek_api_key"`
	DeepSeekModel  string `json:"deepseek_model"`
	Port           int    `json:"port"`
}

func CreateEmptyConfig(path string) error {
	emptyConfig := `
	{
		"deepseek_api_key": "example_key",
		"deepseek_model": "deepseek-chat",
		"port": 8080
	}
	`
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte(emptyConfig))
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func LoadConfig(f *os.File) (*Config, error) {
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)

	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
