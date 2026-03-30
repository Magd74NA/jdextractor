package jdextract

// ProgressStage identifies a step in the processing pipeline.
type ProgressStage string

const (
	StageFetching   ProgressStage = "fetching"
	StageParsing    ProgressStage = "parsing"
	StageGenerating ProgressStage = "generating"
	StageContent    ProgressStage = "content"
	StageSaving     ProgressStage = "saving"
	StageComplete   ProgressStage = "complete"
	StageError      ProgressStage = "error"
)

// ProgressEvent is emitted at each stage boundary during processing.
// For StageContent events, Delta holds the incremental LLM output text.
type ProgressEvent struct {
	Stage   ProgressStage `json:"stage"`
	Message string        `json:"message,omitempty"`
	Dir     string        `json:"dir,omitempty"`
	Delta   string        `json:"delta,omitempty"`
}
