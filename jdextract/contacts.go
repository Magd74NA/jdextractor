package jdextract

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	st "strings"
	"text/tabwriter"
	"time"
)

// ContactMeta is the persisted metadata for a networking contact.
type ContactMeta struct {
	Name          string              `json:"name"`
	Company       string              `json:"company,omitempty"`
	Role          string              `json:"role,omitempty"`
	Email         string              `json:"email,omitempty"`
	Phone         string              `json:"phone,omitempty"`
	LinkedIn      string              `json:"linkedin,omitempty"`
	Source        string              `json:"source,omitempty"`         // how met: event, referral, cold outreach, etc.
	Status        string              `json:"status"`                   // relationship stage
	FollowUpDate  string              `json:"follow_up_date,omitempty"` // YYYY-MM-DD
	LinkedJobs    []string            `json:"linked_jobs,omitempty"`    // job directory names
	Tags          []string            `json:"tags,omitempty"`           // freeform: recruiter, engineer, etc.
	Notes         string              `json:"notes,omitempty"`
	Conversations []Conversation `json:"conversations"`
	Created       string              `json:"created"` // YYYY-MM-DD
	Dir           string              `json:"-"`       // populated at read time, excluded from JSON
}

// Message is a single message within a conversation thread.
type Message struct {
	Sender  string `json:"sender"`  // freeform name: "me", "Jane Doe", etc.
	Content string `json:"content"`
	Date    string `json:"date"` // YYYY-MM-DD
}

// Conversation is a thread of messages with a contact.
type Conversation struct {
	Channel  string    `json:"channel,omitempty"` // email, linkedin, phone, in-person, event, other
	Summary  string    `json:"summary"`           // user-written, optionally LLM-regenerated
	Messages []Message `json:"messages"`
	Created  string    `json:"created"` // YYYY-MM-DD
}

// ContactUpdate holds optional fields for partial contact updates.
type ContactUpdate struct {
	Name         *string   `json:"name"`
	Company      *string   `json:"company"`
	Role         *string   `json:"role"`
	Email        *string   `json:"email"`
	Phone        *string   `json:"phone"`
	LinkedIn     *string   `json:"linkedin"`
	Source       *string   `json:"source"`
	Status       *string   `json:"status"`
	FollowUpDate *string   `json:"follow_up_date"`
	LinkedJobs   *[]string `json:"linked_jobs"`
	Tags         *[]string `json:"tags"`
	Notes        *string   `json:"notes"`
}

var validContactStatuses = []string{
	"new", "reached-out", "replied", "meeting-scheduled", "connected", "dormant",
}

func contactSlugify(name string) string {
	prefix := currentDate()
	midfix := rand.Text()[:8]
	name = st.TrimSpace(st.ToValidUTF8(st.ToLower(name), ""))
	slug := slugRe.ReplaceAllString(name, "-")
	slug = st.Trim(slug, "-")
	if slug == "" {
		return prefix + "-" + midfix
	}
	return prefix + "-" + midfix + "-" + slug
}

func createContactDirectory(slug string, a *App) (string, error) {
	dirName := filepath.Join(a.Paths.Contacts, slug)
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			dirName = dirName + "col"
			err = os.Mkdir(dirName, 0755)
			if err != nil {
				return "", err
			}
			return dirName, nil
		}
		return "", err
	}
	return dirName, nil
}

func readContactMeta(path string) (*ContactMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m ContactMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Conversations == nil {
		m.Conversations = []Conversation{}
	}
	return &m, nil
}

func writeContactMeta(path string, m *ContactMeta) error {
	data, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// CreateContact creates a new contact directory and writes meta.json. Returns the directory name.
func CreateContact(a *App, meta ContactMeta) (string, error) {
	if meta.Name == "" {
		return "", fmt.Errorf("contact name is required")
	}
	if meta.Status == "" {
		meta.Status = "new"
	}
	meta.Created = currentDate()
	if meta.Conversations == nil {
		meta.Conversations = []Conversation{}
	}

	slug := contactSlugify(meta.Name)
	dir, err := createContactDirectory(slug, a)
	if err != nil {
		return "", fmt.Errorf("create contact directory: %w", err)
	}

	metaPath := filepath.Join(dir, "meta.json")
	if err := writeContactMeta(metaPath, &meta); err != nil {
		return "", fmt.Errorf("write meta.json: %w", err)
	}

	return filepath.Base(dir), nil
}

// ListContacts reads all contact directories and returns contacts sorted by created date descending.
func ListContacts(a *App) ([]ContactMeta, error) {
	entries, err := os.ReadDir(a.Paths.Contacts)
	if err != nil {
		return nil, fmt.Errorf("read contacts directory: %w", err)
	}

	var contacts []ContactMeta
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		metaPath := filepath.Join(a.Paths.Contacts, e.Name(), "meta.json")
		m, err := readContactMeta(metaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", e.Name(), err)
			continue
		}
		m.Dir = e.Name()
		contacts = append(contacts, *m)
	}

	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Created > contacts[j].Created
	})
	return contacts, nil
}

// GetContact reads a single contact by its exact directory name.
func GetContact(a *App, dir string) (*ContactMeta, error) {
	if !validContactID(dir) {
		return nil, fmt.Errorf("invalid contact id %q", dir)
	}
	metaPath := filepath.Join(a.Paths.Contacts, dir, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("contact not found: %q", dir)
		}
		return nil, fmt.Errorf("read meta.json: %w", err)
	}
	m.Dir = dir
	return m, nil
}

// UpdateContact applies partial updates to a contact's meta.json.
func UpdateContact(a *App, id string, updates ContactUpdate) error {
	if !validContactID(id) {
		return fmt.Errorf("invalid contact id %q", id)
	}
	if updates.Status != nil {
		found := false
		for _, s := range validContactStatuses {
			if *updates.Status == s {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid status %q: must be one of %s", *updates.Status, st.Join(validContactStatuses, ", "))
		}
	}

	metaPath := filepath.Join(a.Paths.Contacts, id, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}

	if updates.Name != nil {
		m.Name = *updates.Name
	}
	if updates.Company != nil {
		m.Company = *updates.Company
	}
	if updates.Role != nil {
		m.Role = *updates.Role
	}
	if updates.Email != nil {
		m.Email = *updates.Email
	}
	if updates.Phone != nil {
		m.Phone = *updates.Phone
	}
	if updates.LinkedIn != nil {
		m.LinkedIn = *updates.LinkedIn
	}
	if updates.Source != nil {
		m.Source = *updates.Source
	}
	if updates.Status != nil {
		m.Status = *updates.Status
	}
	if updates.FollowUpDate != nil {
		m.FollowUpDate = *updates.FollowUpDate
	}
	if updates.LinkedJobs != nil {
		m.LinkedJobs = *updates.LinkedJobs
	}
	if updates.Tags != nil {
		m.Tags = *updates.Tags
	}
	if updates.Notes != nil {
		m.Notes = *updates.Notes
	}

	return writeContactMeta(metaPath, m)
}

// DeleteContact removes a contact directory and all its contents.
func DeleteContact(a *App, dir string) error {
	if !validContactID(dir) {
		return fmt.Errorf("invalid contact id %q", dir)
	}
	path := filepath.Join(a.Paths.Contacts, dir)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("contact not found: %q", dir)
	}
	return os.RemoveAll(path)
}

// AddConversation appends a conversation thread to a contact's meta.json.
func AddConversation(a *App, dir string, conv Conversation) error {
	if !validContactID(dir) {
		return fmt.Errorf("invalid contact id %q", dir)
	}
	if conv.Summary == "" {
		return fmt.Errorf("conversation summary is required")
	}
	if conv.Created == "" {
		conv.Created = currentDate()
	}
	if conv.Messages == nil {
		conv.Messages = []Message{}
	}

	metaPath := filepath.Join(a.Paths.Contacts, dir, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	m.Conversations = append(m.Conversations, conv)
	return writeContactMeta(metaPath, m)
}

// AddMessage appends a message to an existing conversation thread.
func AddMessage(a *App, dir string, convIndex int, msg Message) error {
	if !validContactID(dir) {
		return fmt.Errorf("invalid contact id %q", dir)
	}
	if msg.Content == "" {
		return fmt.Errorf("message content is required")
	}
	if msg.Sender == "" {
		return fmt.Errorf("message sender is required")
	}
	if msg.Date == "" {
		msg.Date = currentDate()
	}

	metaPath := filepath.Join(a.Paths.Contacts, dir, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	if convIndex < 0 || convIndex >= len(m.Conversations) {
		return fmt.Errorf("conversation index %d out of range (0-%d)", convIndex, len(m.Conversations)-1)
	}
	m.Conversations[convIndex].Messages = append(m.Conversations[convIndex].Messages, msg)
	return writeContactMeta(metaPath, m)
}

// DeleteMessage removes a message by index from a conversation thread.
func DeleteMessage(a *App, dir string, convIndex, msgIndex int) error {
	if !validContactID(dir) {
		return fmt.Errorf("invalid contact id %q", dir)
	}
	metaPath := filepath.Join(a.Paths.Contacts, dir, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	if convIndex < 0 || convIndex >= len(m.Conversations) {
		return fmt.Errorf("conversation index %d out of range (0-%d)", convIndex, len(m.Conversations)-1)
	}
	msgs := m.Conversations[convIndex].Messages
	if msgIndex < 0 || msgIndex >= len(msgs) {
		return fmt.Errorf("message index %d out of range (0-%d)", msgIndex, len(msgs)-1)
	}
	m.Conversations[convIndex].Messages = append(msgs[:msgIndex], msgs[msgIndex+1:]...)
	return writeContactMeta(metaPath, m)
}

// UpdateConversationSummary updates the summary of a conversation thread.
func UpdateConversationSummary(a *App, dir string, convIndex int, summary string) error {
	if !validContactID(dir) {
		return fmt.Errorf("invalid contact id %q", dir)
	}
	metaPath := filepath.Join(a.Paths.Contacts, dir, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	if convIndex < 0 || convIndex >= len(m.Conversations) {
		return fmt.Errorf("conversation index %d out of range (0-%d)", convIndex, len(m.Conversations)-1)
	}
	m.Conversations[convIndex].Summary = summary
	return writeContactMeta(metaPath, m)
}

// DeleteConversation removes a conversation entry by index from a contact's meta.json.
func DeleteConversation(a *App, dir string, index int) error {
	if !validContactID(dir) {
		return fmt.Errorf("invalid contact id %q", dir)
	}
	metaPath := filepath.Join(a.Paths.Contacts, dir, "meta.json")
	m, err := readContactMeta(metaPath)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	if index < 0 || index >= len(m.Conversations) {
		return fmt.Errorf("conversation index %d out of range (0-%d)", index, len(m.Conversations)-1)
	}
	m.Conversations = append(m.Conversations[:index], m.Conversations[index+1:]...)
	return writeContactMeta(metaPath, m)
}

// ListOverdueFollowups returns contacts whose follow_up_date is today or earlier.
func ListOverdueFollowups(a *App) ([]ContactMeta, error) {
	contacts, err := ListContacts(a)
	if err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	var overdue []ContactMeta
	for _, c := range contacts {
		if c.FollowUpDate != "" && c.FollowUpDate <= today {
			overdue = append(overdue, c)
		}
	}
	return overdue, nil
}

// ListUpcomingFollowups returns contacts with follow_up_date within the next N days (exclusive of today).
func ListUpcomingFollowups(a *App, days int) ([]ContactMeta, error) {
	contacts, err := ListContacts(a)
	if err != nil {
		return nil, err
	}
	today := time.Now().Format("2006-01-02")
	cutoff := time.Now().AddDate(0, 0, days).Format("2006-01-02")
	var upcoming []ContactMeta
	for _, c := range contacts {
		if c.FollowUpDate != "" && c.FollowUpDate > today && c.FollowUpDate <= cutoff {
			upcoming = append(upcoming, c)
		}
	}
	return upcoming, nil
}

// FormatContacts returns a tabular string listing of contacts for CLI output.
func FormatContacts(contacts []ContactMeta) string {
	var buf st.Builder
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "CREATED\tNAME\tCOMPANY\tROLE\tSTATUS\tFOLLOW-UP\tDIR")
	for _, c := range contacts {
		followUp := c.FollowUpDate
		if followUp == "" {
			followUp = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			c.Created, c.Name, c.Company, c.Role, c.Status, followUp, c.Dir)
	}
	w.Flush()
	return buf.String()
}

// validContactID rejects contact IDs that could escape the contacts directory.
func validContactID(id string) bool {
	return id != "" && id != "." && !st.Contains(id, "/") && !st.Contains(id, "\\")
}

// FindContactByPrefix finds a contact directory by prefix, returning an error if
// there are zero or multiple matches.
func FindContactByPrefix(a *App, prefix string) (string, error) {
	entries, err := os.ReadDir(a.Paths.Contacts)
	if err != nil {
		return "", fmt.Errorf("read contacts directory: %w", err)
	}
	var matches []string
	for _, e := range entries {
		if e.IsDir() && st.HasPrefix(e.Name(), prefix) {
			matches = append(matches, e.Name())
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no contact directory matches prefix %q", prefix)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("prefix %q is ambiguous: matches %s", prefix, st.Join(matches, ", "))
	}
}
