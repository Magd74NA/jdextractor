package jdextract

import (
	"fmt"
	"os"
	"path/filepath"
)

func (a *App) createExampleTemplates() error {
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
		path := filepath.Join(a.Paths.Templates, t.name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(t.content), 0644); err != nil {
				return fmt.Errorf("cannot write %s: %w", t.name, err)
			}
		}
	}

	return nil
}

func (a *App) Setup() error {
	for _, dir := range []string{a.Paths.Data, a.Paths.Config, a.Paths.Jobs, a.Paths.Templates} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create %s: %w", dir, err)
		}
	}

	if err := a.createExampleTemplates(); err != nil {
		return fmt.Errorf("cannot create example templates: %w", err)
	}

	return nil
}
