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
	KimiApiKey     string `json:"kimi_api_key"`
	KimiModel      string `json:"kimi_model"`
	Backend        string `json:"backend"` // "deepseek" (default) or "kimi"
	Port           int    `json:"port"`
}

type PromptConfig struct {
	TaskList     string `json:"task_list"`
	SystemPrompt string `json:"system_prompt"`
}

// NetworkingPromptConfig holds prompts for AI follow-up message generation.
type NetworkingPromptConfig struct {
	SystemPrompt string `json:"system_prompt"`
	TaskList     string `json:"task_list"`
}

func CreateEmptyNetworkingPromptConfig(path string) error {
	cfg := NetworkingPromptConfig{
		SystemPrompt: "You are a professional networking coach helping with job search outreach.\nYou will receive context about a contact, their conversation history, and relationship status.\n",
		TaskList:     "1. Analyze the conversation history and relationship stage.\n2. Draft a natural, personalized follow-up message appropriate for the channel and context.\n3. Suggest the best timing and channel for the follow-up.\n4. Keep the tone professional but warm — avoid sounding templated or generic.",
	}
	return SaveNetworkingPromptConfig(path, cfg)
}

// SaveNetworkingPromptConfig writes cfg to path as indented JSON with 0600 permissions.
func SaveNetworkingPromptConfig(path string, cfg NetworkingPromptConfig) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadNetworkingPromptConfig reads and JSON-decodes a NetworkingPromptConfig from an open file.
func LoadNetworkingPromptConfig(f *os.File) (*NetworkingPromptConfig, error) {
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cfg NetworkingPromptConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func CreateEmptyPromptConfig(path string) error {
	cfg := PromptConfig{
		SystemPrompt: "You are a professional resume writer and career coach.\nYou will receive a job description as a JSON array of classified lines, a base resume, and optionally a base cover letter.\n",
		TaskList:     "1. Extract the company name and role title from the job description.\n2. Rewrite the resume to align with the job — keep experience truthful, sharpen bullets to mirror the job's language and priorities.\n3. If a base cover letter is provided, draft a tailored cover letter for this role.\n4. Rate how well the base resume matches the job requirements on a scale of 1–10 (1 = poor fit, 10 = perfect fit).",
	}
	return SavePromptConfig(path, cfg)
}

// CreateEmptyConfig writes a skeleton config.json with placeholder values to path.
// The file is created with 0600 permissions (owner read/write only) because it
// will contain the DeepSeek API key. If the file already exists it is truncated.
func CreateEmptyConfig(path string) error {
	emptyConfig := `
	{
		"deepseek_api_key": "example_key",
		"deepseek_model": "deepseek-chat",
		"kimi_api_key": "",
		"kimi_model": "moonshotai/Kimi-K2.5",
		"backend": "deepseek",
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

// SavePromptConfig writes prompt cfg to path as indented JSON with 0600 permissions.
func SavePromptConfig(path string, cfg PromptConfig) error {
	data, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadPromptConfig reads and JSON-decodes a PromptConfig from an open file.
// The caller is responsible for opening the file; LoadPromptConfig closes it.
func LoadPromptConfig(f *os.File) (*PromptConfig, error) {
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var cfg PromptConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
