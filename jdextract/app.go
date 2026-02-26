package jdextract

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type PortablePaths struct {
	Root      string
	Jobs      string
	Data      string
	Config    string
	Templates string
}

type App struct {
	Paths PortablePaths
}

func getPortablePaths() (PortablePaths, error) {
	execPath, err := os.Executable()

	if err != nil {
		return PortablePaths{}, fmt.Errorf("error resolving exec: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return PortablePaths{}, fmt.Errorf("error resolving symlink: %w", err)
	}

	root := filepath.Dir(execPath)

	if runtime.GOOS == "darwin" {
		// Walk up until we find the .app bundle root
		for {
			if strings.HasSuffix(root, ".app") {
				// Found it, now go one more level up to get the container
				root = filepath.Dir(root)
				break
			}
			parent := filepath.Dir(root)
			if parent == root {
				// Hit filesystem root without finding .app
				break
			}
			root = parent
		}
	}

	paths := PortablePaths{
		Root:      root,
		Data:      filepath.Join(root, "data"),
		Jobs:      filepath.Join(root, "data", "jobs"),
		Config:    filepath.Join(root, "config"),
		Templates: filepath.Join(root, "config", "templates"),
	}

	return paths, nil
}

func (A *App) createExampleTemplates() error {
	// Example resume template
	exampleResume := `YOUR NAME
your.email@example.com | (555) 123-4567 | LinkedIn: linkedin.com/in/yourprofile

PROFESSIONAL SUMMARY
Results-driven professional with X years of experience in [your field].
Proven track record of [key achievement]. Skilled in [relevant skills].

EXPERIENCE

Job Title | Company Name | Month Year - Present
• Accomplishment-driven bullet point with quantifiable results (e.g., increased X by Y%)
• Another key responsibility or achievement demonstrating relevant skills
• Led cross-functional initiative resulting in [specific outcome]

Previous Role | Previous Company | Month Year - Month Year
• Key responsibility aligned with target job requirements
• Achievement showing relevant technical or soft skills
• Collaboration or leadership example

EDUCATION
Degree Name | University Name | Year
• Relevant coursework, honors, or activities (optional)

SKILLS
• Technical: [skill1, skill2, skill3]
• Tools: [tool1, tool2, tool3]
• Languages: [language1, language2]
`

	// Example cover letter template
	exampleCover := `YOUR NAME
your.email@example.com | (555) 123-4567
Date

Hiring Manager Name (or "Hiring Manager" if unknown)
Company Name
Company Address (optional)

Dear Hiring Manager,

OPENING PARAGRAPH: State the position you're applying for and express enthusiasm.
Mention how you learned about the opportunity and include a hook that demonstrates
your fit for the role.

BODY PARAGRAPH 1: Connect your experience to the job requirements. Highlight 2-3
of your most relevant achievements that directly relate to what they're looking for.
Use specific examples and quantifiable results where possible.

BODY PARAGRAPH 2: Demonstrate knowledge of the company and explain why you're
interested in this specific role. Show how your values align with theirs and what
unique perspective you bring.

CLOSING PARAGRAPH: Reiterate your interest and summarize why you're a strong fit.
Include a call to action (e.g., requesting an interview) and thank them for their
consideration.

Sincerely,
Your Name
`

	templates := []struct {
		name    string
		content string
	}{
		{"resume.txt", exampleResume},
		{"cover.txt", exampleCover},
	}

	for _, t := range templates {
		path := filepath.Join(A.Paths.Templates, t.name)
		// Only create if file doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(t.content), 0644); err != nil {
				return fmt.Errorf("cannot write %s: %w", t.name, err)
			}
		}
	}

	return nil
}

func (A *App) Setup() error {
	// Create directories if they don't exist
	for _, dir := range []string{A.Paths.Data, A.Paths.Config, A.Paths.Jobs, A.Paths.Templates} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}

	// Create example templates if they don't exist
	if err := A.createExampleTemplates(); err != nil {
		return fmt.Errorf("cannot create example templates: %w", err)
	}

	return nil
}

func NewApp() (*App, error) {
	paths, err := getPortablePaths()
	if err != nil {
		return nil, err
	}
	app := &App{
		Paths: paths,
	}

	err = app.Setup()

	if err != nil {
		return nil, err
	}
	return app, nil
}
