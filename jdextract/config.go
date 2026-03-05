package jdextract

import (
	"encoding/json"
	"io"
	"os"
)

// Config holds runtime configuration loaded from config/config.json.
// The file is created with 0600 permissions because it contains the API key.
type Config struct {
	DeepSeekApiKey string `json:"deepseek_api_key"`
	DeepSeekModel  string `json:"deepseek_model"`
	Port           int    `json:"port"`
}

// CreateEmptyConfig writes a skeleton config.json with placeholder values to path.
// The file is created with 0600 permissions (owner read/write only) because it
// will contain the DeepSeek API key. If the file already exists it is truncated.
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

// SaveConfig writes cfg to path as indented JSON with 0600 permissions.
func SaveConfig(path string, cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadConfig reads and JSON-decodes a Config from an open file.
// The caller is responsible for opening the file; LoadConfig closes it.
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
