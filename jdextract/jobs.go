package jdextract

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	st "strings"
	"text/tabwriter"
	"time"
)

// ApplicationMeta is the persisted metadata for a processed job application.
type ApplicationMeta struct {
	Company string `json:"company"`
	Role    string `json:"role"`
	Score   int    `json:"score"`
	Tokens  int    `json:"tokens"`
	Date    string `json:"date"`
	Status  string `json:"status,omitempty"`
	Dir     string `json:"-"`
}

func (m *ApplicationMeta) SetDir(d string) { m.Dir = d }
func (m ApplicationMeta) GetDir() string   { return m.Dir }

var validStatuses = []string{"draft", "applied", "interviewing", "offer", "rejected"}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func currentDate() string {
	return time.Now().Format("2006-01-02")
}

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
	prefix := (currentDate())
	midfix := rand.Text()[:8]
	title = st.TrimSpace(st.ToValidUTF8(st.ToLower(title), ""))
	slug := slugRe.ReplaceAllString(title, "-")
	slug = st.Trim(slug, "-")

	if slug == "" {
		return prefix + "-" + midfix
	}
	return prefix + "-" + midfix + "-" + slug
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

// ListJobs reads all job directories, parses meta.json, and returns sorted results.
// Corrupt or missing meta.json entries are skipped with a warning to stderr.
func ListJobs(a *App) ([]ApplicationMeta, error) {
	jobs, err := a.Jobs.List()
	if err != nil {
		return nil, err
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Date > jobs[j].Date
	})
	return jobs, nil
}

// FormatJobs returns a tabular string listing of jobs.
func FormatJobs(jobs []ApplicationMeta) string {
	var buf st.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tCOMPANY\tROLE\tSCORE\tSTATUS\tDIR")
	for _, j := range jobs {
		status := j.Status
		if status == "" {
			status = "draft"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", j.Date, j.Company, j.Role, j.Score, status, j.Dir)
	}
	w.Flush()
	return buf.String()
}

// DeleteJob removes a job directory and all its contents by exact directory name.
func DeleteJob(a *App, dir string) error {
	return a.Jobs.Delete(dir)
}

// UpdateJobMeta updates editable metadata fields (company, role, date) for a job.
// id must be the exact directory name.
func UpdateJobMeta(a *App, id string, company, role, date *string) error {
	if !ValidID(id) {
		return fmt.Errorf("invalid job id %q", id)
	}
	m, err := a.Jobs.ReadMeta(id)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	if company != nil {
		m.Company = *company
	}
	if role != nil {
		m.Role = *role
	}
	if date != nil {
		m.Date = *date
	}
	return a.Jobs.WriteMeta(id, m)
}

// FindJobByPrefix finds a job directory by prefix, returning an error if there are
// zero or multiple matches.
func FindJobByPrefix(a *App, prefix string) (string, error) {
	return a.Jobs.FindByPrefix(prefix)
}

// UpdateJobStatus finds a job by directory prefix and updates its status in meta.json.
func UpdateJobStatus(a *App, prefix, status string) error {
	if !slices.Contains(validStatuses, status) {
		return fmt.Errorf("invalid status %q: must be one of %s", status, st.Join(validStatuses, ", "))
	}
	dir, err := a.Jobs.FindByPrefix(prefix)
	if err != nil {
		return err
	}
	m, err := a.Jobs.ReadMeta(dir)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	m.Status = status
	if err := a.Jobs.WriteMeta(dir, m); err != nil {
		return fmt.Errorf("write meta.json: %w", err)
	}
	fmt.Printf("Updated %s → %s\n", dir, status)
	return nil
}
