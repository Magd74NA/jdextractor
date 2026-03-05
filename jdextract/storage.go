package jdextract

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

func createApplicationDirectory(slug string, a *App) error {
	dirName := filepath.Join(a.Paths.Jobs, slug)
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			dirName = dirName + "col"
			err = os.Mkdir(dirName, 0755)
			if err != nil {
				return err
			}
		}
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

// ListJobs reads all job directories, parses meta.json, and returns sorted results.
// Corrupt or missing meta.json entries are skipped with a warning to stderr.
func ListJobs(a *App) ([]ApplicationMeta, error) {
	entries, err := os.ReadDir(a.Paths.Jobs)
	if err != nil {
		return nil, fmt.Errorf("read jobs directory: %w", err)
	}

	var jobs []ApplicationMeta
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metaPath := filepath.Join(a.Paths.Jobs, e.Name(), "meta.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", e.Name(), err)
			continue
		}
		var m ApplicationMeta
		if err := json.Unmarshal(data, &m); err != nil {
			fmt.Fprintf(os.Stderr, "warning: corrupt meta.json in %s: %v\n", e.Name(), err)
			continue
		}
		m.Dir = e.Name()
		jobs = append(jobs, m)
	}

	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].Date > jobs[j].Date
	})
	return jobs, nil
}

// PrintJobs writes a tabular listing of jobs to stdout.
func PrintJobs(jobs []ApplicationMeta) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tCOMPANY\tROLE\tSCORE\tSTATUS\tDIR")
	for _, j := range jobs {
		status := j.Status
		if status == "" {
			status = "draft"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", j.Date, j.Company, j.Role, j.Score, status, j.Dir)
	}
	w.Flush()
}

// UpdateJobStatus finds a job by directory prefix and updates its status in meta.json.
func UpdateJobStatus(a *App, prefix, status string) error {
	valid := false
	for _, s := range validStatuses {
		if s == status {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status %q: must be one of %s", status, st.Join(validStatuses, ", "))
	}

	entries, err := os.ReadDir(a.Paths.Jobs)
	if err != nil {
		return fmt.Errorf("read jobs directory: %w", err)
	}

	var matches []string
	for _, e := range entries {
		if e.IsDir() && st.HasPrefix(e.Name(), prefix) {
			matches = append(matches, e.Name())
		}
	}

	switch len(matches) {
	case 0:
		return fmt.Errorf("no job directory matches prefix %q", prefix)
	case 1:
		// ok
	default:
		return fmt.Errorf("prefix %q is ambiguous: matches %s", prefix, st.Join(matches, ", "))
	}

	metaPath := filepath.Join(a.Paths.Jobs, matches[0], "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	var m ApplicationMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("parse meta.json: %w", err)
	}
	m.Status = status
	out, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}
	if err := os.WriteFile(metaPath, out, 0644); err != nil {
		return fmt.Errorf("write meta.json: %w", err)
	}
	fmt.Printf("Updated %s → %s\n", matches[0], status)
	return nil
}
