package jdextract

// ProgressStage identifies a step in the processing pipeline.
type ProgressStage string

const (
	StageFetching   ProgressStage = "fetching"
	StageParsing    ProgressStage = "parsing"
	StageGenerating ProgressStage = "generating"
	StageSaving     ProgressStage = "saving"
	StageComplete   ProgressStage = "complete"
	StageError      ProgressStage = "error"
)

// ProgressEvent is emitted at each stage boundary during processing.
type ProgressEvent struct {
	Stage   ProgressStage `json:"stage"`
	Message string        `json:"message,omitempty"`
	Dir     string        `json:"dir,omitempty"`
}

// ProgressFunc receives progress updates during processing.
type ProgressFunc func(ProgressEvent)
