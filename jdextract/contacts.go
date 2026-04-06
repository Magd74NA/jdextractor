package jdextract

import (
	"crypto/rand"
	"fmt"
	"sort"
	st "strings"
	"text/tabwriter"
	"time"
)

// ContactMeta is the persisted metadata for a networking contact.
type ContactMeta struct {
	Name          string         `json:"name"`
	Company       string         `json:"company,omitempty"`
	Role          string         `json:"role,omitempty"`
	Email         string         `json:"email,omitempty"`
	Phone         string         `json:"phone,omitempty"`
	LinkedIn      string         `json:"linkedin,omitempty"`
	Source        string         `json:"source,omitempty"`         // how met: event, referral, cold outreach, etc.
	Status        string         `json:"status"`                   // relationship stage
	FollowUpDate  string         `json:"follow_up_date,omitempty"` // YYYY-MM-DD
	LinkedJobs    []string       `json:"linked_jobs,omitempty"`    // job directory names
	Tags          []string       `json:"tags,omitempty"`           // freeform: recruiter, engineer, etc.
	Notes         string         `json:"notes,omitempty"`
	Conversations []Conversation `json:"conversations"`
	Created       string         `json:"created"` // YYYY-MM-DD
	Dir           string         `json:"-"`       // populated at read time, excluded from JSON
}

// Message is a single message within a conversation thread.
type Message struct {
	Sender    string `json:"sender"` // freeform name: "me", "Jane Doe", etc.
	Content   string `json:"content"`
	Date      string `json:"date"`                // YYYY-MM-DD
	Generated bool   `json:"generated,omitempty"` // true if AI-drafted
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

func (m *ContactMeta) SetDir(d string) { m.Dir = d }
func (m ContactMeta) GetDir() string   { return m.Dir }

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

// mutateContact reads a contact's meta, applies fn, and writes it back.
func mutateContact(s *Store[ContactMeta], id string, fn func(*ContactMeta) error) error {
	if !ValidID(id) {
		return fmt.Errorf("invalid contact id %q", id)
	}
	m, err := s.ReadMeta(id)
	if err != nil {
		return fmt.Errorf("read meta.json: %w", err)
	}
	if err := fn(m); err != nil {
		return err
	}
	return s.WriteMeta(id, m)
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
	dir, err := a.Contacts.MkDir(slug)
	if err != nil {
		return "", fmt.Errorf("create contact directory: %w", err)
	}

	if err := a.Contacts.WriteMeta(dir, &meta); err != nil {
		return "", fmt.Errorf("write meta.json: %w", err)
	}

	return dir, nil
}

// ListContacts reads all contact directories and returns contacts sorted by created date descending.
func ListContacts(a *App) ([]ContactMeta, error) {
	contacts, err := a.Contacts.List()
	if err != nil {
		return nil, err
	}
	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Created > contacts[j].Created
	})
	return contacts, nil
}

// GetContact reads a single contact by its exact directory name.
func GetContact(a *App, dir string) (*ContactMeta, error) {
	return a.Contacts.Get(dir)
}

// UpdateContact applies partial updates to a contact's meta.json.
func UpdateContact(a *App, id string, updates ContactUpdate) error {
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

	return mutateContact(&a.Contacts, id, func(m *ContactMeta) error {
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
		return nil
	})
}

// DeleteContact removes a contact directory and all its contents.
func DeleteContact(a *App, dir string) error {
	return a.Contacts.Delete(dir)
}

// AddConversation appends a conversation thread to a contact's meta.json.
func AddConversation(a *App, dir string, conv Conversation) error {
	if conv.Summary == "" {
		return fmt.Errorf("conversation summary is required")
	}
	if conv.Created == "" {
		conv.Created = currentDate()
	}
	if conv.Messages == nil {
		conv.Messages = []Message{}
	}
	return mutateContact(&a.Contacts, dir, func(m *ContactMeta) error {
		m.Conversations = append(m.Conversations, conv)
		return nil
	})
}

// AddMessage appends a message to an existing conversation thread.
func AddMessage(a *App, dir string, convIndex int, msg Message) error {
	if msg.Content == "" {
		return fmt.Errorf("message content is required")
	}
	if msg.Sender == "" {
		return fmt.Errorf("message sender is required")
	}
	if msg.Date == "" {
		msg.Date = currentDate()
	}
	return mutateContact(&a.Contacts, dir, func(m *ContactMeta) error {
		if convIndex < 0 || convIndex >= len(m.Conversations) {
			return fmt.Errorf("conversation index %d out of range (0-%d)", convIndex, len(m.Conversations)-1)
		}
		m.Conversations[convIndex].Messages = append(m.Conversations[convIndex].Messages, msg)
		return nil
	})
}

// DeleteMessage removes a message by index from a conversation thread.
func DeleteMessage(a *App, dir string, convIndex, msgIndex int) error {
	return mutateContact(&a.Contacts, dir, func(m *ContactMeta) error {
		if convIndex < 0 || convIndex >= len(m.Conversations) {
			return fmt.Errorf("conversation index %d out of range (0-%d)", convIndex, len(m.Conversations)-1)
		}
		msgs := m.Conversations[convIndex].Messages
		if msgIndex < 0 || msgIndex >= len(msgs) {
			return fmt.Errorf("message index %d out of range (0-%d)", msgIndex, len(msgs)-1)
		}
		m.Conversations[convIndex].Messages = append(msgs[:msgIndex], msgs[msgIndex+1:]...)
		return nil
	})
}

// UpdateConversationSummary updates the summary of a conversation thread.
func UpdateConversationSummary(a *App, dir string, convIndex int, summary string) error {
	return mutateContact(&a.Contacts, dir, func(m *ContactMeta) error {
		if convIndex < 0 || convIndex >= len(m.Conversations) {
			return fmt.Errorf("conversation index %d out of range (0-%d)", convIndex, len(m.Conversations)-1)
		}
		m.Conversations[convIndex].Summary = summary
		return nil
	})
}

// DeleteConversation removes a conversation entry by index from a contact's meta.json.
func DeleteConversation(a *App, dir string, index int) error {
	return mutateContact(&a.Contacts, dir, func(m *ContactMeta) error {
		if index < 0 || index >= len(m.Conversations) {
			return fmt.Errorf("conversation index %d out of range (0-%d)", index, len(m.Conversations)-1)
		}
		m.Conversations = append(m.Conversations[:index], m.Conversations[index+1:]...)
		return nil
	})
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

// validFollowupChannels is the set of channels accepted by SendFollowup.
var validFollowupChannels = []string{"email", "linkedin", "phone", "in-person", "other"}

// SendFollowupInput holds the parameters for SendFollowup.
type SendFollowupInput struct {
	Content          string // required
	Channel          string // required; one of validFollowupChannels
	NextFollowUpDate string // optional; YYYY-MM-DD or ""
}

// SendFollowup logs a sent follow-up message to the contact's conversation
// thread (finding the most recent thread matching the channel, or creating one),
// clears or advances FollowUpDate, and persists the contact. Returns the updated ContactMeta.
func SendFollowup(a *App, id string, input SendFollowupInput) (*ContactMeta, error) {
	if input.Content == "" {
		return nil, fmt.Errorf("content is required")
	}
	validChannel := false
	for _, ch := range validFollowupChannels {
		if input.Channel == ch {
			validChannel = true
			break
		}
	}
	if !validChannel {
		return nil, fmt.Errorf("invalid channel %q: must be one of %s", input.Channel, st.Join(validFollowupChannels, ", "))
	}

	var updated *ContactMeta
	err := mutateContact(&a.Contacts, id, func(m *ContactMeta) error {
		// Find the most recent conversation whose channel matches.
		convIdx := -1
		for i := len(m.Conversations) - 1; i >= 0; i-- {
			if m.Conversations[i].Channel == input.Channel {
				convIdx = i
				break
			}
		}
		if convIdx == -1 {
			// Create a new conversation thread for this channel.
			m.Conversations = append(m.Conversations, Conversation{
				Channel:  input.Channel,
				Summary:  "",
				Messages: []Message{},
				Created:  currentDate(),
			})
			convIdx = len(m.Conversations) - 1
		}

		msg := Message{
			Sender:    "me",
			Content:   input.Content,
			Date:      currentDate(),
			Generated: true,
		}
		m.Conversations[convIdx].Messages = append(m.Conversations[convIdx].Messages, msg)
		m.FollowUpDate = input.NextFollowUpDate
		updated = m
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// FindContactByPrefix finds a contact directory by prefix, returning an error if
// there are zero or multiple matches.
func FindContactByPrefix(a *App, prefix string) (string, error) {
	return a.Contacts.FindByPrefix(prefix)
}
