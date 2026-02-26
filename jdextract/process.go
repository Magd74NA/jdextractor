package jdextract

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type applicationMeta struct {
	Company string `json:"company"`
	Role    string `json:"role"`
	Score   int    `json:"score"`
	Tokens  int    `json:"tokens"`
	Date    string `json:"date"`
}

func (a *App) Process(ctx context.Context, rawText string) (string, error) {
	nodes := Parse(rawText)

	baseResume, err := fetchResume(a)
	if err != nil {
		return "", err
	}

	var baseCover *string
	if c, err := fetchCover(a); err == nil {
		baseCover = &c
	}

	company, role, resume, cover, score, tokens, err := GenerateAll(
		ctx,
		a.Config.DeepSeekApiKey,
		a.Config.DeepSeekModel,
		&a.Client,
		nodes,
		baseResume,
		baseCover,
	)
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	slug := slugify(nodes) //NEED TO UPDATE TO NOT USE NODES MAYBE?
	if err := createApplicationDirectory(slug, a); err != nil {
		return "", fmt.Errorf("create directory: %w", err)
	}
	dir := filepath.Join(a.Paths.Jobs, slug)

	if err := os.WriteFile(filepath.Join(dir, "resume.txt"), []byte(resume), 0644); err != nil {
		return "", fmt.Errorf("write resume: %w", err)
	}

	if cover != nil {
		if err := os.WriteFile(filepath.Join(dir, "cover.txt"), []byte(*cover), 0644); err != nil {
			return "", fmt.Errorf("write cover: %w", err)
		}
	}

	date := currentDate()

	meta := applicationMeta{Company: company, Role: role, Score: score, Tokens: tokens, Date: date}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644); err != nil {
		return "", fmt.Errorf("write meta: %w", err)
	}

	return dir, nil
}
