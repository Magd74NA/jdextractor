package jdextractor

import (
	_ "embed"
	"fmt"
	"os"
	"time"
)

// App is the central orchestrator coordinating all jobauto operations
type App struct {
	Config *Config
	// We'll add LLM client and other dependencies here as we build them
}

// NewApp creates a new App instance with initialized configuration
func NewApp() *App {
	cfg, err := NewConfig()
	if err != nil {
		// Fatal since we can't recover from bad config paths
		panic(fmt.Sprintf("Failed to initialize config: %v", err))
	}

	return &App{
		Config: cfg,
	}
}

// Setup initializes the ~/.jobauto directory structure and creates
// example templates if they don't exist
func (a *App) Setup() error {
	// Create directories
	if err := a.Config.EnsureDirs(); err != nil {
		return err
	}

	// Create example resume template if it doesn't exist
	if _, err := os.Stat(a.Config.ResumeTemplate); os.IsNotExist(err) {
		if err := os.WriteFile(a.Config.ResumeTemplate, []byte(defaultResumeTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create resume template: %w", err)
		}
	}

	// Create example cover letter template if it doesn't exist
	if _, err := os.Stat(a.Config.CoverTemplate); os.IsNotExist(err) {
		if err := os.WriteFile(a.Config.CoverTemplate, []byte(defaultCoverTemplate), 0644); err != nil {
			return fmt.Errorf("failed to create cover template: %w", err)
		}
	}

	return nil
}

// ProcessJob fetches a job URL and generates a tailored resume
// This is a placeholder for the full implementation
func (a *App) ProcessJob(url string) (folderPath string, err error) {
	// Validate we have an API key
	if err := a.Config.ValidateAPIKey(); err != nil {
		return "", err
	}

	// TODO: Implement in generate.go
	// 1. Fetch URL content
	// 2. Extract company/role name
	// 3. Create run folder
	// 4. Call LLM to customize resume
	// 5. Write files

	return "", fmt.Errorf("ProcessJob not yet implemented (Phase 1)")
}

// MarkdownToPDF converts a markdown file to PDF using pandoc
// This is a placeholder for the full implementation
func (a *App) MarkdownToPDF(mdPath string) (pdfPath string, err error) {
	// TODO: Implement in pandoc.go
	return "", fmt.Errorf("MarkdownToPDF not yet implemented (Phase 2)")
}

// Serve starts the web UI server
// This is a placeholder for the full implementation
func (a *App) Serve(port string) error {
	// TODO: Implement in web.go
	return fmt.Errorf("Serve not yet implemented (Phase 3)")
}

// Helper to generate a run folder name from company and role
func (a *App) generateRunFolderName(company, role string) string {
	date := time.Now().Format("2006-01-02")
	// Sanitize names for filesystem
	company = sanitizeFilename(company)
	role = sanitizeFilename(role)
	return fmt.Sprintf("%s_%s_%s", date, company, role)
}

func sanitizeFilename(s string) string {
	// Basic sanitization - replace spaces with dashes, remove problematic chars
	// Full implementation would be more robust
	return s
}

// Embedded default templates
var defaultResumeTemplate = `# Your Name

## Contact
- Email: your.email@example.com
- Phone: (555) 123-4567
- LinkedIn: linkedin.com/in/yourprofile

## Experience

### Job Title — Company Name
*Dates*

- Bullet point describing achievement with metrics
- Bullet point highlighting relevant skills
- Bullet point showing leadership or initiative

### Previous Role — Previous Company
*Dates*

- Bullet point
- Bullet point

## Skills
- Skill 1, Skill 2, Skill 3

## Education
**Degree** — University, Year
`

var defaultCoverTemplate = `# Your Name
Your Address
Your Email | Your Phone

**Date**

Hiring Manager
Company Name

Dear Hiring Manager,

I am writing to express my interest in the [Position] role at [Company]. With my background in [relevant field], I am excited about the opportunity to contribute to [specific company goal or value].

In my current role at [Current Company], I have [specific achievement relevant to job description]. My experience has prepared me to [specific skill mentioned in job posting].

I am particularly drawn to [Company] because of [specific reason]. I believe my skills in [relevant skills] align well with your needs for this position.

Thank you for considering my application. I look forward to discussing how I can contribute to your team.

Sincerely,
Your Name
`
