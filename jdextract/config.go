package jdextract

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	DeepSeekApiKey string `json:"deepseek_api_key"`
	Port           int    `json:"port"`
}

func CreateEmptyConfig(path string) error {
	emptyConfig := `
	{
		"deepseek_api_key": "example_key",
		"port": 8080,
		"deepseek_model": "reasonerdeepseek-chat"
	}
	`
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	file.Write([]byte(emptyConfig))
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
