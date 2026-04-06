package jdextract

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// LoadJSON reads path and unmarshals into a new T.
func LoadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// SaveJSON marshals v as indented JSON and writes it to path with the given
// permissions using write-to-temp + rename for atomic updates.
func SaveJSON[T any](path string, v T, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+"-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}

func CreateEmptyNetworkingPromptConfig(path string) error {
	return SaveJSON(path, NetworkingPromptConfig{
		SystemPrompt: "You are a professional networking coach helping with job search outreach.\nYou will receive context about a contact, their conversation history, and relationship status.\n",
		TaskList:     "1. Analyze the conversation history and relationship stage.\n2. Draft a natural, personalized follow-up message appropriate for the channel and context.\n3. Suggest the best timing and channel for the follow-up.\n4. Keep the tone professional but warm — avoid sounding templated or generic.",
	}, 0600)
}

func CreateEmptyPromptConfig(path string) error {
	return SaveJSON(path, PromptConfig{
		SystemPrompt: "You are a professional resume writer and career coach.\nYou will receive a job description as a JSON array of classified lines, a base resume, and optionally a base cover letter.\n",
		TaskList:     "1. Extract the company name and role title from the job description.\n2. Rewrite the resume to align with the job — keep experience truthful, sharpen bullets to mirror the job's language and priorities.\n3. If a base cover letter is provided, draft a tailored cover letter for this role.\n4. Rate how well the base resume matches the job requirements on a scale of 1–10 (1 = poor fit, 10 = perfect fit).",
	}, 0600)
}

// CreateEmptyConfig writes a skeleton config.json with placeholder values to path.
// The file is created with 0600 permissions (owner read/write only) because it
// will contain the DeepSeek API key.
func CreateEmptyConfig(path string) error {
	return SaveJSON(path, Config{
		DeepSeekApiKey: "example_key",
		DeepSeekModel:  "deepseek-chat",
		KimiModel:      "moonshotai/Kimi-K2.5",
		Backend:        "deepseek",
		Port:           8080,
	}, 0600)
}
