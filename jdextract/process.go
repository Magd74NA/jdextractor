package jdextract

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const batchConcurrency = 10

// BatchResult holds the outcome of a single URL in a batch run.
type BatchResult struct {
	URL string
	Dir string
	Err error
}

// ProcessBatch fetches and processes each URL concurrently (capped at batchConcurrency).
// Results are streamed to the returned channel as they complete; the channel is closed
// when all URLs are done. A failed URL does not affect the others.
func (a *App) ProcessBatch(ctx context.Context, urls []string) <-chan BatchResult {
	ch := make(chan BatchResult, len(urls))
	sem := make(chan struct{}, batchConcurrency)
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			raw, err := FetchJobDescription(ctx, url, &a.Client, 0)
			if err != nil {
				ch <- BatchResult{URL: url, Err: fmt.Errorf("fetch: %w", err)}
				return
			}
			dir, err := a.Process(ctx, raw)
			ch <- BatchResult{URL: url, Dir: dir, Err: err}
		}(url)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}

// Process runs the full generation pipeline for a single job description and
// returns the path to the output directory. rawText may come from any source
// (URL fetch, local file, or stdin) — routing is the caller's responsibility.
//
// Pipeline: Parse → load templates → GenerateAll (LLM) → create directory → write files.
// The LLM call is the only expensive step; no filesystem writes happen before it
// succeeds, so a failed generation leaves no partial state on disk.
func (a *App) Process(ctx context.Context, rawText string) (string, error) {
	return a.ProcessWithProgress(ctx, rawText, func(_ ProgressEvent) {})
}

// ProcessWithProgress is like Process but calls onProgress at each pipeline stage.
// During LLM generation, it also emits StageContent events with incremental text.
func (a *App) ProcessWithProgress(ctx context.Context, rawText string, onProgress func(ProgressEvent)) (string, error) {
	onProgress(ProgressEvent{Stage: StageParsing, Message: "Parsing job description\u2026"})
	nodes := Parse(rawText)

	baseResume, err := fetchResume(a)
	if err != nil {
		return "", err
	}

	var baseCover *string
	if c, err := fetchCover(a); err == nil {
		baseCover = &c
	}

	invoker := LLMInvoker(InvokeDeepseekApi)
	streamInvoker := StreamingLLMInvoker(InvokeDeepseekApiStream)
	apiKey := a.Config.DeepSeekApiKey
	model := a.Config.DeepSeekModel
	if a.Config.Backend == "kimi" {
		invoker = InvokeKimiApi
		streamInvoker = InvokeKimiApiStream
		apiKey = a.Config.KimiApiKey
		model = a.Config.KimiModel
	}

	onProgress(ProgressEvent{Stage: StageGenerating, Message: "Generating tailored resume\u2026"})
	onDelta := func(delta string) {
		onProgress(ProgressEvent{Stage: StageContent, Delta: delta})
	}
	company, role, resume, cover, score, tokens, err := GenerateAll(
		ctx,
		invoker,
		streamInvoker,
		apiKey,
		model,
		&a.Client,
		nodes,
		baseResume,
		baseCover,
		a.PromptConfig,
		onDelta,
	)
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	onProgress(ProgressEvent{Stage: StageSaving, Message: "Saving files\u2026"})
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

	meta := ApplicationMeta{Company: company, Role: role, Score: score, Tokens: tokens, Date: date}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaBytes, 0644); err != nil {
		return "", fmt.Errorf("write meta: %w", err)
	}

	return dir, nil
}
