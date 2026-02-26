package jdextract

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	st "strings"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(nodes []JobDescriptionNode) string {
	var title string

	for _, node := range nodes {
		switch node.NodeType {
		case NodeJinaTitle:
			title = st.TrimPrefix(node.Content, "Title:")
		case NodeJobTitle:
			title = st.TrimLeft(node.Content, "#* \t")
		}
		if title != "" {
			break
		}
	}

	prefix := rand.Text()[:8]
	title = st.TrimSpace(st.ToValidUTF8(st.ToLower(title), ""))
	slug := slugRe.ReplaceAllString(title, "-")
	slug = st.Trim(slug, "-")

	if slug == "" {
		return prefix
	}
	return prefix + "-" + slug
}

func createApplicationDirectory(slug string, a *App) error {
	dirName := filepath.Join(a.Paths.Jobs, slug)
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		//	if errors.Is(err, os.ErrExist) {
		//		dirName = dirName + "col"
		//		err = os.Mkdir(dirName, 0755)
		//		if err != nil {
		//			return err
		//		}
		//	}
		return err
	}
	return nil
}

func fetchCover(a *App) (string, error) {
	path := filepath.Join(a.Paths.Templates, "cover.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read cover letter template: %w", err)
	}
	return string(content), nil
}

func fetchResume(a *App) (string, error) {
	path := filepath.Join(a.Paths.Templates, "resume.txt")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read resume template: %w", err)
	}
	return string(content), nil
}

type applicationMeta struct {
	Company string `json:"company"`
	Role    string `json:"role"`
	Score   int    `json:"score"`
	Tokens  int    `json:"tokens"`
}

func ProcessApplication(ctx context.Context, a *App, nodes []JobDescriptionNode) error {
	baseResume, err := fetchResume(a)
	if err != nil {
		return err
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
		return fmt.Errorf("generate: %w", err)
	}

	slug := slugify(nodes)
	if err := createApplicationDirectory(slug, a); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	dir := filepath.Join(a.Paths.Jobs, slug)

	if err := os.WriteFile(filepath.Join(dir, "resume.txt"), []byte(resume), 0644); err != nil {
		return fmt.Errorf("write resume: %w", err)
	}

	if cover != nil {
		if err := os.WriteFile(filepath.Join(dir, "cover.txt"), []byte(*cover), 0644); err != nil {
			return fmt.Errorf("write cover: %w", err)
		}
	}

	meta := applicationMeta{Company: company, Role: role, Score: score, Tokens: tokens}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644); err != nil {
		return fmt.Errorf("write meta: %w", err)
	}

	return nil
}
